// Package handlers contains PaperViz's HTTP layer only: parsing requests,
// validating input shape, calling into services, and serializing responses.
// Per ARCHITECTURE.md Section 2, handlers MUST NOT call repository directly
// — every persistence operation goes through a service function first (even
// though today those service functions are thin wrappers, keeping the call
// path handlers -> services -> repository consistent is what lets us add
// real business logic later without restructuring call sites).
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
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

type createDocumentResponse struct {
	DocumentID string `json:"document_id"`
	Status     string `json:"status"`
}

type listDocumentResponse struct {
	Documents []documentSummary `json:"documents"`
	Total     int               `json:"total"`
	Limit     int               `json:"limit"`
	Offset    int               `json:"offset"`
}

type documentSummary struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Status           string `json:"status"`
	CreatedAt        int64  `json:"created_at"`
	SummaryPreview   string `json:"summary_preview"`
	ChartCount       int    `json:"chart_count"`
	ExplanationCount int    `json:"explanation_count"`
}

type documentStatsResponse struct {
	Total       int `json:"total"`
	Saved       int `json:"saved"`
	Collections int `json:"collections"`
}

type toggleSavedRequest struct {
	Saved bool `json:"saved"`
}

type updateTitleRequest struct {
	Title string `json:"title"`
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

	var pdfBytes []byte

	if hasFile {
		defer file.Close()
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
	}

	var userID *string
	if uid := UserIDFromContext(r.Context()); uid != "" {
		userID = &uid
	}

	intakeResult, errCode, err := services.ValidateAndInsert(h.db, readingLevel, hasFile, pdfBytes, pastedText, userID)
	if err != nil {
		if errCode == "no_text_layer" {
			writeError(w, http.StatusUnprocessableEntity, "no_text_layer")
			return
		}
		slog.Error("document intake validation/insertion failed", "error", err)
		writeError(w, http.StatusBadRequest, errCode)
		return
	}

	// Fire the rest of the pipeline in the background. Detached from the
	// request context (which dies when this handler returns) and given its
	// own bounded lifetime instead, so a slow Gemini call doesn't get
	// cancelled just because the HTTP response already went out.
	go h.runPipelineAndSave(intakeResult.DocumentID, services.PipelineInput{
		OriginalText: intakeResult.OriginalText,
		SourceType:   intakeResult.SourceType,
		ReadingLevel: readingLevel,
		PDFBytes:     intakeResult.PDFBytes,
	})

	writeJSON(w, http.StatusCreated, createDocumentResponse{DocumentID: intakeResult.DocumentID, Status: repository.StatusProcessing})
}

// runPipelineAndSave runs the full pipeline and persists the result. Errors
// are logged, not returned — there is no HTTP request left to answer by the
// time this runs; the client learns the outcome via the next GET poll.
func (h *DocumentHandler) runPipelineAndSave(documentID string, input services.PipelineInput) {
	services.RunPipelineAndPersist(h.db, h.gemini, documentID, input)
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
	ChapterID    *string         `json:"chapter_id,omitempty"`
	ImageURL     *string         `json:"image_url,omitempty"`
}

// claimDiffResponse is the wire shape for claim-diff verification data.
// OriginalClaims/SimplifiedClaims are pre-serialized JSON arrays from the
// repository — we wrap them in json.RawMessage to avoid double-marshal.
type claimDiffResponse struct {
	OriginalClaims   json.RawMessage `json:"original_claims,omitempty"`
	SimplifiedClaims json.RawMessage `json:"simplified_claims,omitempty"`
	MismatchDetected bool            `json:"mismatch_detected"`
	MismatchDetail   *string         `json:"mismatch_detail,omitempty"`
}

type chapterResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Content      string `json:"content"`
	DisplayOrder int    `json:"display_order"`
}

type evidenceResponse struct {
	ID              string  `json:"id"`
	Page            *int    `json:"page,omitempty"`
	FigureID        *string `json:"figure_id,omitempty"`
	TableID         *string `json:"table_id,omitempty"`
	Section         *string `json:"section,omitempty"`
	SourceText      string  `json:"source_text"`
	SourceReference string  `json:"source_reference"`
}

// getDocumentResponse matches ARCHITECTURE.md Section E's Get Document
// contract exactly.
type getDocumentResponse struct {
	ID                      string             `json:"id"`
	Title                   string             `json:"title"`
	Status                  string             `json:"status"`
	ReadingLevel            string             `json:"reading_level"`
	SimplifiedText          *string            `json:"simplified_text"`
	OriginalText            string             `json:"original_text"`
	Charts                  []chartResponse    `json:"charts"`
	ErrorMessage            *string            `json:"error_message"`
	ChartExtractionDegraded bool               `json:"chart_extraction_degraded"`
	ProcessingStage         *string            `json:"processing_stage,omitempty"`
	ClaimDiff               *claimDiffResponse `json:"claim_diff,omitempty"`
	Chapters                []chapterResponse  `json:"chapters,omitempty"`
	Evidence                []evidenceResponse `json:"evidence,omitempty"`
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
			ChapterID:    c.ChapterID,
		}
		if c.ChartData != nil && *c.ChartData != "" {
			cr.ChartData = json.RawMessage(*c.ChartData)
		}
		if len(c.ImageBlob) > 0 {
			imageURL := fmt.Sprintf("/api/documents/%s/charts/%s/image", id, c.ID)
			cr.ImageURL = &imageURL
		}
		chartResponses = append(chartResponses, cr)
	}

	claimDiffRepo := repository.NewClaimDiffRepo(h.db)
	var claimDiffResp *claimDiffResponse
	if cd, err := claimDiffRepo.GetByDocument(id); err == nil {
		claimDiffResp = &claimDiffResponse{
			OriginalClaims:   json.RawMessage(cd.OriginalClaims),
			SimplifiedClaims: json.RawMessage(cd.SimplifiedClaims),
			MismatchDetected: cd.MismatchDetected,
			MismatchDetail:   cd.MismatchDetail,
		}
	}

	chapterRepo := repository.NewChapterRepo(h.db)
	chapters, err := chapterRepo.ListByDocument(id)
	if err != nil {
		slog.Error("list chapters failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	chapterResponses := make([]chapterResponse, 0, len(chapters))
	for _, ch := range chapters {
		chapterResponses = append(chapterResponses, chapterResponse{
			ID:           ch.ID,
			Title:        ch.Title,
			Summary:      ch.Summary,
			Content:      ch.Excerpt,
			DisplayOrder: ch.DisplayOrder,
		})
	}

	evidenceRepo := repository.NewEvidenceRepo(h.db)
	evidence, err := evidenceRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list evidence failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	evidenceResponses := make([]evidenceResponse, 0, len(evidence))
	for _, e := range evidence {
		evidenceResponses = append(evidenceResponses, evidenceResponse{
			ID:              e.ID,
			Page:            e.Page,
			FigureID:        e.FigureID,
			TableID:         e.TableID,
			Section:         e.Section,
			SourceText:      e.SourceText,
			SourceReference: e.SourceReference,
		})
	}

	writeJSON(w, http.StatusOK, getDocumentResponse{
		ID:                      doc.ID,
		Title:                   doc.Title,
		Status:                  doc.Status,
		ReadingLevel:            doc.ReadingLevel,
		SimplifiedText:          doc.SimplifiedText,
		OriginalText:            doc.OriginalText,
		Charts:                  chartResponses,
		ErrorMessage:            doc.ErrorMessage,
		ChartExtractionDegraded: doc.ChartExtractionDegraded,
		ProcessingStage:         doc.ProcessingStage,
		ClaimDiff:               claimDiffResp,
		Chapters:                chapterResponses,
		Evidence:                evidenceResponses,
	})
}

func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	docRepo := repository.NewDocumentRepo(h.db)
	docs, err := docRepo.ListSummariesByUser(userID, limit, offset)
	if err != nil {
		slog.Error("list documents failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	summaries := make([]documentSummary, 0, len(docs))
	for _, d := range docs {
		summaries = append(summaries, documentSummary{
			ID:               d.ID,
			Title:            d.Title,
			Status:           d.Status,
			CreatedAt:        d.CreatedAt,
			SummaryPreview:   d.SummaryPreview,
			ChartCount:       d.ChartCount,
			ExplanationCount: d.ExplanationCount,
		})
	}

	writeJSON(w, http.StatusOK, listDocumentResponse{
		Documents: summaries,
		Total:     len(summaries),
		Limit:     limit,
		Offset:    offset,
	})
}

// Stats handles GET /api/documents/stats
func (h *DocumentHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	docRepo := repository.NewDocumentRepo(h.db)

	total, err := docRepo.CountByUser(userID)
	if err != nil {
		slog.Error("count documents failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	saved, err := docRepo.CountSavedByUser(userID)
	if err != nil {
		slog.Error("count saved failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	colRepo := repository.NewCollectionRepo(h.db)
	collections, err := colRepo.ListByUser(userID)
	if err != nil {
		slog.Error("count collections failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, documentStatsResponse{
		Total:       total,
		Saved:       saved,
		Collections: len(collections),
	})
}

// ToggleSaved handles PUT /api/documents/:id/save
func (h *DocumentHandler) ToggleSaved(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req toggleSavedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	if err := services.ToggleDocumentSaved(h.db, id, req.Saved); err != nil {
		slog.Error("toggle saved failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"saved": req.Saved})
}

// UpdateTitle handles PATCH /api/documents/:id
func (h *DocumentHandler) UpdateTitle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req updateTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title_required")
		return
	}

	if err := services.RenameDocument(h.db, id, req.Title); err != nil {
		slog.Error("update title failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"title": req.Title})
}

// Delete handles DELETE /api/documents/:id
func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := services.DeleteDocument(h.db, id); err != nil {
		slog.Error("delete document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetChartImage serves the original chart image bytes for a document's
// chart. The lookup is scoped to the parent document so a bare chart ID
// cannot be used to read another document's figure; a missing image or
// unrecognized format both resolve to 404/500 rather than leaking partial
// state.
func (h *DocumentHandler) GetChartImage(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	chartID := chi.URLParam(r, "chartId")

	chartRepo := repository.NewChartRepo(h.db)
	chart, err := chartRepo.GetByDocumentAndID(docID, chartID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		slog.Error("get chart image failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	if len(chart.ImageBlob) == 0 {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}

	mime := detectImageMIME(chart.ImageBlob)
	if mime == "" {
		slog.Error("chart image has unrecognized format", "chart_id", chartID)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(chart.ImageBlob); err != nil {
		slog.Error("write chart image failed", "chart_id", chartID, "error", err)
	}
}
