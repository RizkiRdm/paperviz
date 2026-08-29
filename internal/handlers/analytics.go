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

// TrackPricingView records a pricing page view event.
func (h *AnalyticsHandler) TrackPricingView(w http.ResponseWriter, r *http.Request) {
	fp := GetFingerprint(r)
	slog.Info("pricing view", "fingerprint", fp)
	w.WriteHeader(http.StatusNoContent)
}

// TrackUpgradeIntent records when a user clicks an upgrade CTA.
func (h *AnalyticsHandler) TrackUpgradeIntent(w http.ResponseWriter, r *http.Request) {
	fp := GetFingerprint(r)
	slog.Info("upgrade intent", "fingerprint", fp)
	w.WriteHeader(http.StatusNoContent)
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
