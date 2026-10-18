package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/app/documents"
)

// GetClaims handles GET /api/documents/:id/claims. Returns all claims for a document.
func (h *DocumentHandler) GetClaims(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	claims, err := h.svc.ListClaims(id)
	if errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
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

	tables, err := h.svc.ListTables(id)
	if errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
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

	methods, err := h.svc.ListMethods(id)
	if errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
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

	results, err := h.svc.ListResults(id)
	if errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
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

	citations, err := h.svc.ListCitations(id)
	if errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
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
