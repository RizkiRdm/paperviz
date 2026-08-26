package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"paperviz/internal/services"
)

// UsageHandler handles usage-related API endpoints.
type UsageHandler struct {
	ts *services.TierService
}

// NewUsageHandler creates a UsageHandler with the given TierService.
func NewUsageHandler(ts *services.TierService) *UsageHandler {
	return &UsageHandler{ts: ts}
}

// usageResponse is the wire shape for GET /api/usage.
type usageResponse struct {
	Tier      string `json:"tier"`
	PapersUsed int   `json:"papers_used"`
	Limit     int    `json:"limit"`
	ResetDate string `json:"reset_date"`
}

// GetUsage handles GET /api/usage, returning current tier and paper usage stats.
func (h *UsageHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	fp := GetFingerprint(r)

	tier, err := h.ts.GetTier(fp)
	if err != nil {
		slog.Error("get tier failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	canCreate, papersUsed, err := h.ts.CheckUsage(fp)
	if err != nil {
		slog.Error("check usage failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	_ = canCreate

	limit := services.LimitFree
	if tier == services.TierPro {
		limit = services.LimitPro
	} else if tier == services.TierResearch {
		limit = services.LimitResearch
	}

	resetDate := nextResetDate()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(usageResponse{
		Tier:       tier,
		PapersUsed: papersUsed,
		Limit:      limit,
		ResetDate:  resetDate,
	}); err != nil {
		slog.Error("encode usage response failed", "error", err)
	}
}

// nextResetDate returns the first day of the next month as a date string.
func nextResetDate() string {
	now := time.Now().UTC()
	year, month, _ := now.Date()
	if month == 12 {
		year++
		month = 1
	} else {
		month++
	}
	return fmt.Sprintf("%04d-%02d-01", year, month)
}
