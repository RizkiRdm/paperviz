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
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/app/credentials"
	"paperviz/internal/app/documents"
	"paperviz/internal/external"
	"paperviz/internal/repository"
	"paperviz/internal/services"
)

// DocumentHandler holds everything the two document endpoints need: a DB
// handle (for opening transactions) and a credential resolver, which produces
// the model client for whoever is making the request. It has no other state —
// request-scoped values are never stored on this struct, per AGENTS.md "MUST
// NOT use global mutable state for request-scoped data."
type DocumentHandler struct {
	db       *sql.DB
	provider *credentials.Resolver
}

func NewDocumentHandler(db *sql.DB, provider *credentials.Resolver) *DocumentHandler {
	return &DocumentHandler{db: db, provider: provider}
}

// resolveClient builds a model client for the requesting user and writes the
// error response itself, returning false when the caller should stop. Every
// endpoint that reaches a model goes through here, so an unconfigured key
// produces the same answer everywhere instead of a per-handler variant.
func (h *DocumentHandler) resolveClient(w http.ResponseWriter, r *http.Request) (*external.LLM, bool) {
	client, err := h.provider.For(r.Context(), UserIDFromContext(r.Context()))
	if err != nil {
		writeCredentialError(w, err)
		return nil, false
	}
	return client, true
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

// Get handles GET /api/documents/:id.
func (h *DocumentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	svc := documents.New(h.db, h.provider)
	rm, err := svc.GetReadModel(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		slog.Error("get document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	doc := rm.Document
	charts := rm.Charts
	chapters := rm.Chapters
	claimDiff := rm.ClaimDiff
	evidence := rm.Evidence
	type chartResp struct {
		ID           string  `json:"id"`
		DocumentID   string  `json:"document_id"`
		SourceMethod string  `json:"source_method"`
		ChartData    *string `json:"chart_data"`
		Annotation   *string `json:"annotation"`
		PageNumber   *int    `json:"page_number"`
		DisplayOrder int     `json:"display_order"`
		ChapterID    *string `json:"chapter_id"`
		ImageURL     *string `json:"image_url"`
	}
	chartResps := make([]chartResp, 0, len(charts))
	for _, c := range charts {
		var imageURL *string
		if len(c.ImageBlob) > 0 {
			u := "/api/documents/" + doc.ID + "/charts/" + c.ID + "/image"
			imageURL = &u
		}
		chartResps = append(chartResps, chartResp{
			ID: c.ID, DocumentID: c.DocumentID, SourceMethod: c.SourceMethod,
			ChartData: c.ChartData, Annotation: c.Annotation, PageNumber: c.PageNumber,
			DisplayOrder: c.DisplayOrder, ChapterID: c.ChapterID, ImageURL: imageURL,
		})
	}
	type chapterResp struct {
		ID           string `json:"id"`
		DocumentID   string `json:"document_id"`
		Title        string `json:"title"`
		Summary      string `json:"summary"`
		Excerpt      string `json:"excerpt"`
		DisplayOrder int    `json:"display_order"`
	}
	chapterResps := make([]chapterResp, 0, len(chapters))
	for _, ch := range chapters {
		chapterResps = append(chapterResps, chapterResp{
			ID: ch.ID, DocumentID: ch.DocumentID, Title: ch.Title,
			Summary: ch.Summary, Excerpt: ch.Excerpt, DisplayOrder: ch.DisplayOrder,
		})
	}
	type claimDiffResp struct {
		ID               string  `json:"id"`
		DocumentID       string  `json:"document_id"`
		OriginalClaims   string  `json:"original_claims"`
		SimplifiedClaims string  `json:"simplified_claims"`
		MismatchDetected bool    `json:"mismatch_detected"`
		MismatchDetail   *string `json:"mismatch_detail"`
	}
	var claimDiffVal *claimDiffResp
	if claimDiff != nil {
		claimDiffVal = &claimDiffResp{
			ID: claimDiff.ID, DocumentID: claimDiff.DocumentID,
			OriginalClaims: claimDiff.OriginalClaims, SimplifiedClaims: claimDiff.SimplifiedClaims,
			MismatchDetected: claimDiff.MismatchDetected, MismatchDetail: claimDiff.MismatchDetail,
		}
	}
	evResps := make([]evidenceResponse, 0, len(evidence))
	for _, e := range evidence {
		evResps = append(evResps, evidenceResponse{
			ID: e.ID, Page: e.Page, FigureID: e.FigureID, TableID: e.TableID,
			Section: e.Section, SourceText: e.SourceText, SourceReference: e.SourceReference,
		})
	}
	resp := map[string]any{
		"id": doc.ID, "title": doc.Title, "status": doc.Status,
		"source_type": doc.SourceType, "reading_level": doc.ReadingLevel,
		"created_at": doc.CreatedAt, "last_accessed_at": doc.LastAccessedAt,
		"original_text": doc.OriginalText, "simplified_text": doc.SimplifiedText,
		"error_message": doc.ErrorMessage, "chart_extraction_degraded": doc.ChartExtractionDegraded,
		"processing_stage": doc.ProcessingStage, "processing_time_ms": doc.ProcessingTimeMs,
		"user_id": doc.UserID, "saved": doc.Saved, "visibility": doc.Visibility,
		"share_token": doc.ShareToken, "charts": chartResps, "chapters": chapterResps,
		"claim_diff": claimDiffVal, "evidence": evResps,
	}
	writeJSON(w, http.StatusOK, resp)
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

type compareRequest struct {
	DocumentIDs []string `json:"document_ids"`
}

type compareResponse struct {
	Papers         []services.PaperSummary        `json:"papers"`
	Dimensions     []services.ComparisonDimension `json:"dimensions"`
	Agreement      []string                       `json:"agreement"`
	Disagreement   []string                       `json:"disagreement"`
	EvidenceClaims []services.EvidenceClaim       `json:"evidence_claims"`
}

// GetClaims handles GET /api/documents/:id/claims. Returns all claims for a document.
func (h *DocumentHandler) GetClaims(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	docRepo := repository.NewDocumentRepo(h.db)
	if _, err := docRepo.Get(id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	} else if err != nil {
		slog.Error("get document for claims failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	claimRepo := repository.NewClaimRepo(h.db)
	claims, err := claimRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list claims failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	type claimResp struct {
		ID         string  `json:"id"`
		ClaimText  string  `json:"claim_text"`
		ClaimType  string  `json:"claim_type"`
		Confidence string  `json:"confidence"`
		SourcePage *int    `json:"source_page,omitempty"`
		SourceText *string `json:"source_text,omitempty"`
		CreatedAt  int64   `json:"created_at"`
	}
	resp := make([]claimResp, 0, len(claims))
	for _, c := range claims {
		resp = append(resp, claimResp{
			ID: c.ID, ClaimText: c.ClaimText, ClaimType: c.ClaimType,
			Confidence: c.Confidence, SourcePage: c.SourcePage,
			SourceText: c.SourceText, CreatedAt: c.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetTables handles GET /api/documents/:id/tables. Returns all paper tables for a document.
func (h *DocumentHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	docRepo := repository.NewDocumentRepo(h.db)
	if _, err := docRepo.Get(id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	} else if err != nil {
		slog.Error("get document for tables failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	tableRepo := repository.NewPaperTableRepo(h.db)
	tables, err := tableRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list tables failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	type tableResp struct {
		ID           string  `json:"id"`
		PageNumber   *int    `json:"page_number,omitempty"`
		Caption      *string `json:"caption,omitempty"`
		Headers      string  `json:"headers"`
		Rows         string  `json:"rows"`
		SourceText   *string `json:"source_text,omitempty"`
		DisplayOrder int     `json:"display_order"`
	}
	resp := make([]tableResp, 0, len(tables))
	for _, t := range tables {
		resp = append(resp, tableResp{
			ID: t.ID, PageNumber: t.PageNumber, Caption: t.Caption,
			Headers: t.Headers, Rows: t.Rows, SourceText: t.SourceText,
			DisplayOrder: t.DisplayOrder,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetMethods handles GET /api/documents/:id/methods. Returns all methods for a document.
func (h *DocumentHandler) GetMethods(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	docRepo := repository.NewDocumentRepo(h.db)
	if _, err := docRepo.Get(id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	} else if err != nil {
		slog.Error("get document for methods failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	methodRepo := repository.NewMethodRepo(h.db)
	methods, err := methodRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list methods failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	type methodResp struct {
		ID          string  `json:"id"`
		MethodName  string  `json:"method_name"`
		Description *string `json:"description,omitempty"`
		MethodType  string  `json:"method_type"`
		SourcePage  *int    `json:"source_page,omitempty"`
		SourceText  *string `json:"source_text,omitempty"`
	}
	resp := make([]methodResp, 0, len(methods))
	for _, m := range methods {
		resp = append(resp, methodResp{
			ID: m.ID, MethodName: m.MethodName, Description: m.Description,
			MethodType: m.MethodType, SourcePage: m.SourcePage, SourceText: m.SourceText,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetResults handles GET /api/documents/:id/results. Returns all results for a document.
func (h *DocumentHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	docRepo := repository.NewDocumentRepo(h.db)
	if _, err := docRepo.Get(id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	} else if err != nil {
		slog.Error("get document for results failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	resultRepo := repository.NewResultRepo(h.db)
	results, err := resultRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list results failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	type resultResp struct {
		ID                   string  `json:"id"`
		ResultText           string  `json:"result_text"`
		ResultType           string  `json:"result_type"`
		SupportingEvidenceID *string `json:"supporting_evidence_id,omitempty"`
		SourcePage           *int    `json:"source_page,omitempty"`
		SourceText           *string `json:"source_text,omitempty"`
	}
	resp := make([]resultResp, 0, len(results))
	for _, res := range results {
		resp = append(resp, resultResp{
			ID: res.ID, ResultText: res.ResultText, ResultType: res.ResultType,
			SupportingEvidenceID: res.SupportingEvidenceID, SourcePage: res.SourcePage,
			SourceText: res.SourceText,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetCitations handles GET /api/documents/:id/citations. Returns all citations for a document.
func (h *DocumentHandler) GetCitations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	docRepo := repository.NewDocumentRepo(h.db)
	if _, err := docRepo.Get(id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	} else if err != nil {
		slog.Error("get document for citations failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	citationRepo := repository.NewCitationRepo(h.db)
	citations, err := citationRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list citations failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	type citationResp struct {
		ID           string  `json:"id"`
		CitedPaperID *string `json:"cited_paper_id,omitempty"`
		Authors      *string `json:"authors,omitempty"`
		Title        *string `json:"title,omitempty"`
		Year         *int    `json:"year,omitempty"`
		Venue        *string `json:"venue,omitempty"`
		DOI          *string `json:"doi,omitempty"`
		URL          *string `json:"url,omitempty"`
		SourcePage   *int    `json:"source_page,omitempty"`
	}
	resp := make([]citationResp, 0, len(citations))
	for _, c := range citations {
		resp = append(resp, citationResp{
			ID: c.ID, CitedPaperID: c.CitedPaperID, Authors: c.Authors,
			Title: c.Title, Year: c.Year, Venue: c.Venue, DOI: c.DOI,
			URL: c.URL, SourcePage: c.SourcePage,
		})
	}
	writeJSON(w, http.StatusOK, resp)
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

type claimWithEvidenceResponse struct {
	ID         string             `json:"id"`
	ClaimText  string             `json:"claim_text"`
	ClaimType  string             `json:"claim_type"`
	Confidence string             `json:"confidence"`
	SourcePage *int               `json:"source_page,omitempty"`
	SourceText *string            `json:"source_text,omitempty"`
	CreatedAt  int64              `json:"created_at"`
	Evidence   []evidenceResponse `json:"evidence"`
}

type egTableResp struct {
	ID           string  `json:"id"`
	PageNumber   *int    `json:"page_number,omitempty"`
	Caption      *string `json:"caption,omitempty"`
	Headers      string  `json:"headers"`
	Rows         string  `json:"rows"`
	SourceText   *string `json:"source_text,omitempty"`
	DisplayOrder int     `json:"display_order"`
}

type egMethodResp struct {
	ID          string  `json:"id"`
	MethodName  string  `json:"method_name"`
	Description *string `json:"description,omitempty"`
	MethodType  string  `json:"method_type"`
	SourcePage  *int    `json:"source_page,omitempty"`
	SourceText  *string `json:"source_text,omitempty"`
}

type egResultResp struct {
	ID                   string  `json:"id"`
	ResultText           string  `json:"result_text"`
	ResultType           string  `json:"result_type"`
	SupportingEvidenceID *string `json:"supporting_evidence_id,omitempty"`
	SourcePage           *int    `json:"source_page,omitempty"`
	SourceText           *string `json:"source_text,omitempty"`
}

type egCitationResp struct {
	ID           string  `json:"id"`
	CitedPaperID *string `json:"cited_paper_id,omitempty"`
	Authors      *string `json:"authors,omitempty"`
	Title        *string `json:"title,omitempty"`
	Year         *int    `json:"year,omitempty"`
	Venue        *string `json:"venue,omitempty"`
	DOI          *string `json:"doi,omitempty"`
	URL          *string `json:"url,omitempty"`
	SourcePage   *int    `json:"source_page,omitempty"`
}

type evidenceGraphResponse struct {
	Claims    []claimWithEvidenceResponse `json:"claims"`
	Tables    []egTableResp               `json:"tables"`
	Methods   []egMethodResp              `json:"methods"`
	Results   []egResultResp              `json:"results"`
	Citations []egCitationResp            `json:"citations"`
}

func (h *DocumentHandler) GetEvidenceGraph(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	docRepo := repository.NewDocumentRepo(h.db)
	if _, err := docRepo.Get(id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	} else if err != nil {
		slog.Error("get document for evidence graph failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	claimRepo := repository.NewClaimRepo(h.db)
	claims, err := claimRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list claims for evidence graph failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	claimResponses := make([]claimWithEvidenceResponse, 0, len(claims))
	for _, c := range claims {
		evidence, err := claimRepo.GetEvidence(c.ID)
		if err != nil {
			slog.Error("get evidence for claim failed", "claim_id", c.ID, "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		evResponses := make([]evidenceResponse, 0, len(evidence))
		for _, e := range evidence {
			evResponses = append(evResponses, evidenceResponse{
				ID:              e.ID,
				Page:            e.Page,
				FigureID:        e.FigureID,
				TableID:         e.TableID,
				Section:         e.Section,
				SourceText:      e.SourceText,
				SourceReference: e.SourceReference,
			})
		}
		claimResponses = append(claimResponses, claimWithEvidenceResponse{
			ID: c.ID, ClaimText: c.ClaimText, ClaimType: c.ClaimType,
			Confidence: c.Confidence, SourcePage: c.SourcePage,
			SourceText: c.SourceText, CreatedAt: c.CreatedAt,
			Evidence: evResponses,
		})
	}

	tableRepo := repository.NewPaperTableRepo(h.db)
	tables, err := tableRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list tables for evidence graph failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	tableResponses := make([]egTableResp, 0, len(tables))
	for _, t := range tables {
		tableResponses = append(tableResponses, egTableResp{
			ID: t.ID, PageNumber: t.PageNumber, Caption: t.Caption,
			Headers: t.Headers, Rows: t.Rows, SourceText: t.SourceText,
			DisplayOrder: t.DisplayOrder,
		})
	}

	methodRepo := repository.NewMethodRepo(h.db)
	methods, err := methodRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list methods for evidence graph failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	methodResponses := make([]egMethodResp, 0, len(methods))
	for _, m := range methods {
		methodResponses = append(methodResponses, egMethodResp{
			ID: m.ID, MethodName: m.MethodName, Description: m.Description,
			MethodType: m.MethodType, SourcePage: m.SourcePage, SourceText: m.SourceText,
		})
	}

	resultRepo := repository.NewResultRepo(h.db)
	results, err := resultRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list results for evidence graph failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	resultResponses := make([]egResultResp, 0, len(results))
	for _, res := range results {
		resultResponses = append(resultResponses, egResultResp{
			ID: res.ID, ResultText: res.ResultText, ResultType: res.ResultType,
			SupportingEvidenceID: res.SupportingEvidenceID, SourcePage: res.SourcePage,
			SourceText: res.SourceText,
		})
	}

	citationRepo := repository.NewCitationRepo(h.db)
	citations, err := citationRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list citations for evidence graph failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	citationResponses := make([]egCitationResp, 0, len(citations))
	for _, c := range citations {
		citationResponses = append(citationResponses, egCitationResp{
			ID: c.ID, CitedPaperID: c.CitedPaperID, Authors: c.Authors,
			Title: c.Title, Year: c.Year, Venue: c.Venue, DOI: c.DOI,
			URL: c.URL, SourcePage: c.SourcePage,
		})
	}

	writeJSON(w, http.StatusOK, evidenceGraphResponse{
		Claims:    claimResponses,
		Tables:    tableResponses,
		Methods:   methodResponses,
		Results:   resultResponses,
		Citations: citationResponses,
	})
}

type paperRelationshipResponse struct {
	ID               string  `json:"id"`
	SourcePaperID    string  `json:"source_paper_id"`
	TargetPaperID    string  `json:"target_paper_id"`
	RelationshipType string  `json:"relationship_type"`
	EvidenceText     *string `json:"evidence_text,omitempty"`
	CreatedAt        int64   `json:"created_at"`
}

type createRelationshipRequest struct {
	TargetPaperID    string  `json:"target_paper_id"`
	RelationshipType string  `json:"relationship_type"`
	EvidenceText     *string `json:"evidence_text,omitempty"`
}

var validRelationshipTypes = map[string]bool{
	repository.PaperRelationshipSupporting:         true,
	repository.PaperRelationshipContradicting:      true,
	repository.PaperRelationshipCiting:             true,
	repository.PaperRelationshipSimilarMethodology: true,
	repository.PaperRelationshipDifferentFindings:  true,
}

func (h *DocumentHandler) GetPaperRelationships(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	docRepo := repository.NewDocumentRepo(h.db)
	if _, err := docRepo.Get(id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	} else if err != nil {
		slog.Error("get document for relationships failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	relRepo := repository.NewPaperRelationshipRepo(h.db)
	rels, err := relRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list paper relationships failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	resp := make([]paperRelationshipResponse, 0, len(rels))
	for _, rel := range rels {
		resp = append(resp, paperRelationshipResponse{
			ID: rel.ID, SourcePaperID: rel.SourcePaperID,
			TargetPaperID: rel.TargetPaperID, RelationshipType: rel.RelationshipType,
			EvidenceText: rel.EvidenceText, CreatedAt: rel.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *DocumentHandler) CreatePaperRelationship(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	docRepo := repository.NewDocumentRepo(h.db)
	if _, err := docRepo.Get(id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	} else if err != nil {
		slog.Error("get document for create relationship failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	var req createRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if req.TargetPaperID == "" {
		writeError(w, http.StatusBadRequest, "target_paper_id_required")
		return
	}
	if req.RelationshipType == "" {
		writeError(w, http.StatusBadRequest, "relationship_type_required")
		return
	}
	if !validRelationshipTypes[req.RelationshipType] {
		writeError(w, http.StatusBadRequest, "invalid_relationship_type")
		return
	}

	if req.TargetPaperID == id {
		writeError(w, http.StatusBadRequest, "cannot_self_reference")
		return
	}

	if _, err := docRepo.Get(req.TargetPaperID); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "target_paper_not_found")
		return
	} else if err != nil {
		slog.Error("get target document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	relID, err := repository.NewID()
	if err != nil {
		slog.Error("generate relationship id failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	rel := repository.PaperRelationship{
		ID:               relID,
		SourcePaperID:    id,
		TargetPaperID:    req.TargetPaperID,
		RelationshipType: req.RelationshipType,
		EvidenceText:     req.EvidenceText,
		CreatedAt:        time.Now().Unix(),
	}

	relRepo := repository.NewPaperRelationshipRepo(h.db)
	if err := relRepo.Insert(rel); err != nil {
		slog.Error("insert paper relationship failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusCreated, paperRelationshipResponse{
		ID: rel.ID, SourcePaperID: rel.SourcePaperID,
		TargetPaperID: rel.TargetPaperID, RelationshipType: rel.RelationshipType,
		EvidenceText: rel.EvidenceText, CreatedAt: rel.CreatedAt,
	})
}

type researchMapResponse struct {
	DocumentID    string                                   `json:"document_id"`
	Relationships map[string][]researchMapRelationshipItem `json:"relationships"`
	TotalCount    int                                      `json:"total_count"`
}

type researchMapRelationshipItem struct {
	ID               string  `json:"id"`
	SourcePaperID    string  `json:"source_paper_id"`
	TargetPaperID    string  `json:"target_paper_id"`
	TargetPaperTitle string  `json:"target_paper_title"`
	RelationshipType string  `json:"relationship_type"`
	EvidenceText     *string `json:"evidence_text,omitempty"`
	CreatedAt        int64   `json:"created_at"`
}

// GetResearchMap handles GET /api/documents/:id/research-map.
func (h *DocumentHandler) GetResearchMap(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := services.GetResearchMap(h.db, id)
	if err != nil {
		if err.Error() == "document not found" {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("get research map failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	relationships := make(map[string][]researchMapRelationshipItem, len(result.Relationships))
	for relType, items := range result.Relationships {
		respItems := make([]researchMapRelationshipItem, 0, len(items))
		for _, item := range items {
			respItems = append(respItems, researchMapRelationshipItem{
				ID:               item.ID,
				SourcePaperID:    item.SourcePaperID,
				TargetPaperID:    item.TargetPaperID,
				TargetPaperTitle: item.TargetPaperTitle,
				RelationshipType: item.RelationshipType,
				EvidenceText:     item.EvidenceText,
				CreatedAt:        item.CreatedAt,
			})
		}
		relationships[relType] = respItems
	}

	writeJSON(w, http.StatusOK, researchMapResponse{
		DocumentID:    result.DocumentID,
		Relationships: relationships,
		TotalCount:    result.TotalCount,
	})
}

// Compare handles POST /api/documents/compare.
func (h *DocumentHandler) Compare(w http.ResponseWriter, r *http.Request) {
	var req compareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	if len(req.DocumentIDs) < 2 || len(req.DocumentIDs) > 10 {
		writeError(w, http.StatusBadRequest, "document_ids_must_be_2_to_10")
		return
	}

	client, ok := h.resolveClient(w, r)
	if !ok {
		return
	}

	docRepo := repository.NewDocumentRepo(h.db)
	var papers []services.PaperSummary

	for _, docID := range req.DocumentIDs {
		doc, err := docRepo.Get(docID)
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "document_not_found")
			return
		}
		if err != nil {
			slog.Error("get document for comparison failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
			return
		}

		summary, err := services.ExtractPaperSummary(r.Context(), client, doc.ID, doc.Title, doc.OriginalText)
		if err != nil {
			slog.Error("extract paper summary failed", "document_id", docID, "error", err)
			writeError(w, http.StatusInternalServerError, "extraction_failed")
			return
		}
		papers = append(papers, summary)
	}

	comparison, err := services.ComparePapers(r.Context(), client, papers)
	if err != nil {
		slog.Error("compare papers failed", "error", err)
		writeError(w, http.StatusInternalServerError, "comparison_failed")
		return
	}

	// Track comparison event for analytics (Chunk 6.1).
	eventID, _ := repository.NewID()
	if eventID != "" {
		_, _ = h.db.Exec(
			`INSERT INTO analytics_events (id, event_type, entity_id, metadata, created_at) VALUES (?, 'comparison', ?, ?, ?)`,
			eventID, strings.Join(req.DocumentIDs, ","), "", time.Now().Unix(),
		)
	}

	writeJSON(w, http.StatusOK, compareResponse{
		Papers:         comparison.Papers,
		Dimensions:     comparison.Dimensions,
		Agreement:      comparison.Agreement,
		Disagreement:   comparison.Disagreement,
		EvidenceClaims: comparison.EvidenceClaims,
	})
}
