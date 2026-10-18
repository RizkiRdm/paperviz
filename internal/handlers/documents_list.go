package handlers

import (
	"net/http"
	"strconv"
)

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

	docs, err := h.svc.List(userID, limit, offset)
	if err != nil {
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

	stats, err := h.svc.Stats(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, documentStatsResponse{
		Total:       stats.Total,
		Saved:       stats.Saved,
		Collections: stats.Collections,
	})
}
