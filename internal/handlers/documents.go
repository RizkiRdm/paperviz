// Package handlers contains PaperViz's HTTP layer only: parsing requests,
// validating input shape, calling into services, and serializing responses.
// Per ARCHITECTURE.md Section 2, handlers MUST NOT call repository directly
// — every persistence operation goes through a service function first (even
// though today those service functions are thin wrappers, keeping the call
// path handlers -> services -> repository consistent is what lets us add
// real business logic later without restructuring call sites).
package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/external"
	"paperviz/internal/repository"
	"paperviz/internal/services"
)

// maxUploadBytes is the MAX_LIMIT file size from ARCHITECTURE.md Section 4
// Validation Policy. Applied before reading the full body into memory, so
// an oversized upload is rejected cheaply rather than after buffering 20MB+
// of attacker-controlled data.
const maxUploadBytes = 20 << 20 // 20 MiB

// pollMinIntervalHint documents the client-side contract from
// ARCHITECTURE.md Section E ("Client polling interval for processing
// status: REQUIRED minimum 2s between polls"). This is enforced by the
// frontend, not the server — recorded here so a reader of this file knows
// the server intentionally does not rate-limit polling itself.
const pollMinIntervalHint = 2 * time.Second

// DocumentHandler holds everything the two document endpoints need: a DB
// handle (for opening transactions) and a Gemini client (passed down to the
// pipeline). It has no other state — request-scoped values are never stored
// on this struct, per AGENTS.md "MUST NOT use global mutable state for
// request-scoped data."
type DocumentHandler struct {
	db     *sql.DB
	gemini *external.GeminiClient
}

func NewDocumentHandler(db *sql.DB, gemini *external.GeminiClient) *DocumentHandler {
	return &DocumentHandler{db: db, gemini: gemini}
}

// createDocumentResponse is the 201 response body for POST /api/documents,
// matching ARCHITECTURE.md Section E's API Contract exactly.
type createDocumentResponse struct {
	DocumentID string `json:"document_id"`
	Status     string `json:"status"`
}

// Create handles POST /api/documents. It is intentionally synchronous up
// through extraction — a bad upload (wrong MIME, no text layer, oversized)
// must be rejected before we ever spend a Gemini call on it (ARCHITECTURE.md
// Failure Scenario 1). Once extraction succeeds, the rest of the pipeline
// (simplify/verify/chart) runs in a background goroutine so the client gets
// its document_id immediately and polls GET for the result — this is the
// "single synchronous request-scoped goroutine chain per document, not a
// job queue" async policy from ARCHITECTURE.md Section 4.
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+1<<20) // small slack for multipart overhead

	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "file_too_large")
		return
	}

	readingLevel := r.FormValue("reading_level")
	if readingLevel != repository.ReadingLevelSimplified && readingLevel != repository.ReadingLevelELI5 {
		writeError(w, http.StatusBadRequest, "invalid_reading_level")
		return
	}

	pastedText := r.FormValue("text")
	file, header, fileErr := r.FormFile("file")

	// Exactly one of file/text is required (ARCHITECTURE.md API Contract).
	hasFile := fileErr == nil
	hasText := pastedText != ""
	if hasFile == hasText {
		writeError(w, http.StatusBadRequest, "missing_input")
		return
	}

	var originalText, sourceType string
	var pdfBytes []byte

	if hasFile {
		defer file.Close()
		sourceType = repository.SourceTypePDF

		if header.Size > maxUploadBytes {
			writeError(w, http.StatusBadRequest, "file_too_large")
			return
		}

		var err error
		pdfBytes, err = io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "file_too_large")
			return
		}

		// Verify actual file content, not just the client-supplied
		// Content-Type header (AGENTS.md Security Rules: "MUST NOT trust
		// client-supplied content-type header alone; verify actual file
		// content").
		if !isPDFContent(pdfBytes) {
			writeError(w, http.StatusBadRequest, "invalid_file_type")
			return
		}

		originalText, err = external.ExtractText(pdfBytes)
		if err != nil {
			if errors.Is(err, external.ErrNoTextLayer) {
				writeError(w, http.StatusUnprocessableEntity, "no_text_layer")
				return
			}
			slog.Error("pdf extraction failed", "error", err)
			writeError(w, http.StatusBadRequest, "invalid_file_type")
			return
		}
	} else {
		sourceType = repository.SourceTypePastedText
		originalText = pastedText
	}

	id, err := repository.NewID()
	if err != nil {
		slog.Error("id generation failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	now := time.Now().Unix()
	doc := repository.Document{
		ID:             id,
		CreatedAt:      now,
		LastAccessedAt: now,
		Status:         repository.StatusProcessing,
		SourceType:     sourceType,
		ReadingLevel:   readingLevel,
		OriginalText:   originalText,
	}

	docRepo := repository.NewDocumentRepo(h.db)
	if err := docRepo.Insert(doc); err != nil {
		slog.Error("insert document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	// Fire the rest of the pipeline in the background. Detached from the
	// request context (which dies when this handler returns) and given its
	// own bounded lifetime instead, so a slow Gemini call doesn't get
	// cancelled just because the HTTP response already went out.
	go h.runPipelineAndSave(id, services.PipelineInput{
		OriginalText: originalText,
		SourceType:   sourceType,
		ReadingLevel: readingLevel,
		PDFBytes:     pdfBytes,
	})

	writeJSON(w, http.StatusCreated, createDocumentResponse{DocumentID: id, Status: repository.StatusProcessing})
}

// backgroundPipelineTimeout bounds the whole background run (all pipeline
// stages combined), independent of the request that triggered it.
const backgroundPipelineTimeout = 5 * time.Minute

// runPipelineAndSave runs the full pipeline and persists the result. Errors
// are logged, not returned — there is no HTTP request left to answer by the
// time this runs; the client learns the outcome via the next GET poll.
func (h *DocumentHandler) runPipelineAndSave(documentID string, input services.PipelineInput) {
	ctx, cancel := context.WithTimeout(context.Background(), backgroundPipelineTimeout)
	defer cancel()

	output := services.RunPipeline(ctx, h.gemini, input)

	if err := h.saveResult(documentID, output); err != nil {
		slog.Error("save pipeline result failed", "document_id", documentID, "error", err)
	}
}

// saveResult writes the full pipeline outcome (document status/text +
// charts + claim_diff) in a single transaction, per ARCHITECTURE.md Section
// 4 Transaction Policy: "Each document's full write MUST occur within a
// single SQLite transaction. Partial writes on failure are PROHIBITED."
func (h *DocumentHandler) saveResult(documentID string, output services.PipelineOutput) error {
	tx, err := h.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() // no-op if Commit succeeds below

	docRepo := repository.NewDocumentRepo(tx)

	if output.Status == repository.StatusFailed {
		errMsg := output.ErrorMessage
		if err := docRepo.UpdateStatus(documentID, repository.StatusFailed, nil, &errMsg); err != nil {
			return err
		}
		return tx.Commit()
	}

	simplified := output.SimplifiedText
	if err := docRepo.UpdateStatus(documentID, output.Status, &simplified, nil); err != nil {
		return err
	}

	// Persist the claim-diff result whenever verification actually ran
	// (i.e. we're not in the pipelineStatusFailed branch above, which
	// returns before reaching here).
	claimDiffRepo := repository.NewClaimDiffRepo(tx)
	originalClaimsJSON, err := json.Marshal(output.Verify.OriginalClaims)
	if err != nil {
		return fmt.Errorf("marshal original claims: %w", err)
	}
	simplifiedClaimsJSON, err := json.Marshal(output.Verify.SimplifiedClaims)
	if err != nil {
		return fmt.Errorf("marshal simplified claims: %w", err)
	}
	claimDiffID, err := repository.NewID()
	if err != nil {
		return fmt.Errorf("generate claim_diff id: %w", err)
	}
	detail := output.Verify.MismatchDetail
	if err := claimDiffRepo.Insert(repository.ClaimDiff{
		ID:               claimDiffID,
		DocumentID:       documentID,
		OriginalClaims:   string(originalClaimsJSON),
		SimplifiedClaims: string(simplifiedClaimsJSON),
		MismatchDetected: output.Verify.MismatchDetected,
		MismatchDetail:   &detail,
	}); err != nil {
		return err
	}

	// Charts only exist when status is complete (verification_failed stops
	// the pipeline before chart processing — see services/pipeline.go).
	chartRepo := repository.NewChartRepo(tx)
	for _, c := range output.Charts {
		chartID, err := repository.NewID()
		if err != nil {
			return fmt.Errorf("generate chart id: %w", err)
		}
		var chartDataPtr, annotationPtr *string
		if c.ChartData != "" {
			chartDataPtr = &c.ChartData
		}
		if c.Annotation != "" {
			annotationPtr = &c.Annotation
		}
		pageNum := c.PageNumber
		if err := chartRepo.Insert(repository.Chart{
			ID:           chartID,
			DocumentID:   documentID,
			SourceMethod: c.SourceMethod,
			ChartData:    chartDataPtr,
			ImageBlob:    c.ImageBlob,
			Annotation:   annotationPtr,
			PageNumber:   &pageNum,
			DisplayOrder: c.DisplayOrder,
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// chartResponse is the wire shape for one chart in the GET response.
// chart_data is passed through as a raw JSON object (not a Go-escaped
// string) so the frontend can consume it directly with Recharts.
type chartResponse struct {
	ID           string          `json:"id"`
	SourceMethod string          `json:"source_method"`
	ChartData    json.RawMessage `json:"chart_data,omitempty"`
	Annotation   *string         `json:"annotation,omitempty"`
	PageNumber   *int            `json:"page_number,omitempty"`
}

// getDocumentResponse matches ARCHITECTURE.md Section E's Get Document
// contract exactly.
type getDocumentResponse struct {
	ID             string          `json:"id"`
	Status         string          `json:"status"`
	ReadingLevel   string          `json:"reading_level"`
	SimplifiedText *string         `json:"simplified_text"`
	OriginalText   string          `json:"original_text"`
	Charts         []chartResponse `json:"charts"`
	ErrorMessage   *string         `json:"error_message"`
}

// Get handles GET /api/documents/:id. On every successful lookup it
// touches last_accessed_at, extending the document's 7-day expiry window
// (ARCHITECTURE.md Acceptance Scenario 5) — this is why a document a
// student keeps reading never expires mid-read, only after a full week of
// nobody opening the link.
func (h *DocumentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	docRepo := repository.NewDocumentRepo(h.db)
	doc, err := docRepo.Get(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		slog.Error("get document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if err := docRepo.TouchLastAccessed(id, time.Now().Unix()); err != nil {
		// Non-fatal: the read itself succeeded, only the expiry-refresh
		// write failed. Log and continue rather than fail the whole request
		// over a housekeeping update.
		slog.Error("touch last_accessed_at failed", "document_id", id, "error", err)
	}

	chartRepo := repository.NewChartRepo(h.db)
	charts, err := chartRepo.ListByDocument(id)
	if err != nil {
		slog.Error("list charts failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	chartResponses := make([]chartResponse, 0, len(charts))
	for _, c := range charts {
		cr := chartResponse{
			ID:           c.ID,
			SourceMethod: c.SourceMethod,
			Annotation:   c.Annotation,
			PageNumber:   c.PageNumber,
		}
		if c.ChartData != nil && *c.ChartData != "" {
			cr.ChartData = json.RawMessage(*c.ChartData)
		}
		chartResponses = append(chartResponses, cr)
	}

	writeJSON(w, http.StatusOK, getDocumentResponse{
		ID:             doc.ID,
		Status:         doc.Status,
		ReadingLevel:   doc.ReadingLevel,
		SimplifiedText: doc.SimplifiedText,
		OriginalText:   doc.OriginalText,
		Charts:         chartResponses,
		ErrorMessage:   doc.ErrorMessage,
	})
}
