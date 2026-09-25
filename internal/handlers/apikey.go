package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
)

// ApiKeyHandler issues and rotates the service key agents use to reach
// PaperViz.
//
// Only the key's SHA-256 digest is stored, so a database dump or a SQL
// injection yields nothing usable. SHA-256 rather than bcrypt or argon2 is
// deliberate: these keys carry 256 bits of entropy from crypto/rand, so there
// is no dictionary to attack and a slow KDF would only add latency to every
// authenticated request. A lost key is replaced with regenerate, never
// recovered.
type ApiKeyHandler struct {
	db *sql.DB
}

// NewApiKeyHandler creates a new ApiKeyHandler.
func NewApiKeyHandler(db *sql.DB) *ApiKeyHandler {
	return &ApiKeyHandler{db: db}
}

// apiKeyState is the wire shape for GET /api/auth/apikey. It reports whether a
// key exists and nothing more: the plaintext is returned exactly once, when
// the key is issued.
type apiKeyState struct {
	Configured bool `json:"configured"`
}

// apiKeyCreated is the wire shape returned when a key is issued. This is the
// only response that ever contains the plaintext.
type apiKeyCreated struct {
	Key string `json:"key"`
}

// GetApiKey handles GET /api/auth/apikey. Reports whether a key is configured.
func (h *ApiKeyHandler) GetApiKey(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var stored string
	err := h.db.QueryRow(`SELECT COALESCE(api_key, '') FROM users WHERE id = ?`, userID).Scan(&stored)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "user_not_found")
		return
	}
	if err != nil {
		slog.Error("get api key state failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, apiKeyState{Configured: stored != ""})
}

// CreateApiKey handles POST /api/auth/apikey. Issues a key and returns it
// once. The response is the only place the plaintext exists outside the
// process that generated it.
func (h *ApiKeyHandler) CreateApiKey(w http.ResponseWriter, r *http.Request) {
	h.issueKey(w, r)
}

// RegenerateApiKey handles POST /api/auth/apikey/regenerate. Issues a new key
// and invalidates the previous one, for when a key is lost or exposed.
func (h *ApiKeyHandler) RegenerateApiKey(w http.ResponseWriter, r *http.Request) {
	h.issueKey(w, r)
}

// issueKey generates a key, stores only its digest, and returns the plaintext.
func (h *ApiKeyHandler) issueKey(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	key, err := generateApiKey()
	if err != nil {
		slog.Error("generate api key failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if _, err := h.db.Exec(`UPDATE users SET api_key = ? WHERE id = ?`, hashApiKey(key), userID); err != nil {
		slog.Error("store api key digest failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusCreated, apiKeyCreated{Key: key})
}

// generateApiKey creates a cryptographically secure random service key.
func generateApiKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pv_" + hex.EncodeToString(b), nil
}

// hashApiKey returns the hex SHA-256 digest stored in place of the key. The
// users.api_key column holds this digest, never the key itself.
func hashApiKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// LookupUserByApiKey resolves the owner of a presented service key by its
// digest, so an authenticated transport never needs to store or compare a
// plaintext key. Returns sql.ErrNoRows for an unknown or revoked key.
func LookupUserByApiKey(ctx context.Context, db *sql.DB, presented string) (string, error) {
	var userID string
	err := db.QueryRowContext(ctx,
		`SELECT id FROM users WHERE api_key = ? AND api_key != ''`, hashApiKey(presented),
	).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}
