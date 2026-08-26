package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"paperviz/internal/repository"
)

// AnalyticsHandler serves aggregate usage metrics for the dashboard.
type AnalyticsHandler struct {
	db *sql.DB
}

// NewAnalyticsHandler constructs an AnalyticsHandler backed by the given database.
func NewAnalyticsHandler(db *sql.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

// GetSummary returns the aggregate usage metrics as JSON.
func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewAnalyticsRepo(h.db)
	summary, err := repo.GetSummary()
	if err != nil {
		slog.Error("get analytics summary failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
