package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/app/documents"
	"paperviz/internal/external"
	"paperviz/internal/services"
)

// Get handles GET /api/documents/:id.
func (h *DocumentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rm, err := h.svc.GetReadModel(id)
	if errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		slog.Error("get document failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	if err := services.AuthorizeDocumentAccess(&rm.Document, UserIDFromContext(r.Context())); err != nil {
		writeDocumentAccessError(w, err, "get document forbidden")
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

	if err := services.DocumentAccessAllowed(h.db, docID, UserIDFromContext(r.Context())); err != nil {
		writeDocumentAccessError(w, err, "get chart image forbidden")
		return
	}

	chart, err := h.svc.Chart(docID, chartID)
	if errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		slog.Error("get chart image failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	// Prefer external storage; fall back to the database column when no store
	// is configured, which keeps local development and unconfigured
	// deployments working without a second code path.
	image := chart.ImageBlob
	mime := ""
	if h.images != nil {
		blob, contentType, err := h.images.Get(r.Context(), chartImageKey(docID, chartID))
		switch {
		case err == nil:
			image, mime = blob, contentType
		case errors.Is(err, external.ErrImageNotFound):
			writeError(w, http.StatusNotFound, "not_found")
			return
		default:
			// Do not silently fall back to the column on a transport error: a
			// storage outage must not look like a working empty result.
			slog.Error("chart image fetch failed", "chart_id", chartID, "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
			return
		}
	}

	if len(image) == 0 {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}

	if mime == "" {
		mime = services.DetectImageMIME(image)
	}
	if mime == "" {
		slog.Error("chart image has unrecognized format", "chart_id", chartID)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(image); err != nil {
		slog.Error("write chart image failed", "chart_id", chartID, "error", err)
	}
}

// chartImageKey namespaces objects per document so a chart id from one document
// can never address another document's image.
func chartImageKey(documentID, chartID string) string {
	return documentID + "/" + chartID + ".img"
}
