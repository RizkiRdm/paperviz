package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"net/http"
)

// ApiKeyHandler manages API key generation and retrieval.
type ApiKeyHandler struct {
	db *sql.DB
}

// NewApiKeyHandler creates a new ApiKeyHandler.
func NewApiKeyHandler(db *sql.DB) *ApiKeyHandler {
	return &ApiKeyHandler{db: db}
}

// apiKeyResponse is the wire shape for API key responses.
type apiKeyResponse struct {
	Key string `json:"key"`
}

// GetApiKey handles GET /api/auth/apikey. Returns the user's API key.
func (h *ApiKeyHandler) GetApiKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var key string
	err := h.db.QueryRow(
		`SELECT api_key FROM users WHERE id = ?`, userID,
	).Scan(&key)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "user_not_found")
		return
	}
	if err != nil {
		slog.Error("get api key failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if key == "" {
		key, err = h.generateApiKey()
		if err != nil {
			slog.Error("generate api key failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		_, err = h.db.Exec(`UPDATE users SET api_key = ? WHERE id = ?`, key, userID)
		if err != nil {
			slog.Error("save api key failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
			return
		}
	}

	writeJSON(w, http.StatusOK, apiKeyResponse{Key: key})
}

// RegenerateApiKey handles POST /api/auth/apikey/regenerate. Generates a new API key.
func (h *ApiKeyHandler) RegenerateApiKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	key, err := h.generateApiKey()
	if err != nil {
		slog.Error("generate api key failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	_, err = h.db.Exec(`UPDATE users SET api_key = ? WHERE id = ?`, key, userID)
	if err != nil {
		slog.Error("save api key failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, apiKeyResponse{Key: key})
}

// generateApiKey creates a cryptographically secure random API key.
func (h *ApiKeyHandler) generateApiKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pv_" + hex.EncodeToString(b), nil
}
