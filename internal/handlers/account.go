package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"
)

// AccountHandler manages account-related operations.
type AccountHandler struct {
	db *sql.DB
}

// NewAccountHandler creates a new AccountHandler.
func NewAccountHandler(db *sql.DB) *AccountHandler {
	return &AccountHandler{db: db}
}

// accountSummary is the wire shape for account summary responses.
type accountSummary struct {
	Email      string `json:"email"`
	UsageCount int    `json:"usage_count"`
}

// GetSummary handles GET /api/account/summary.
func (h *AccountHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var email string
	err := h.db.QueryRow(
		`SELECT email FROM users WHERE id = ?`, userID,
	).Scan(&email)
	if err != nil {
		slog.Error("get account summary failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	var usageCount int
	err = h.db.QueryRow(
		`SELECT COUNT(*) FROM documents WHERE user_id = ? AND created_at >= ?`,
		userID, beginningOfMonth(),
	).Scan(&usageCount)
	if err != nil {
		slog.Error("get usage count failed", "error", err)
		usageCount = 0
	}

	writeJSON(w, http.StatusOK, accountSummary{
		Email:      email,
		UsageCount: usageCount,
	})
}

func beginningOfMonth() int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
}
