package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/external"
	"paperviz/internal/repository"
)

// CredentialHandler manages the BYOK model keys a user supplies.
//
// PaperViz holds no AI vendor account, so these keys are the most sensitive
// data it stores. Two rules govern every method here: the plaintext key is
// never written to a log, and never returned in a response.
type CredentialHandler struct {
	db     *sql.DB
	cipher *external.Cipher
}

// NewCredentialHandler creates a CredentialHandler bound to a cipher.
func NewCredentialHandler(db *sql.DB, cipher *external.Cipher) *CredentialHandler {
	return &CredentialHandler{db: db, cipher: cipher}
}

// createCredentialRequest is the wire shape for POST /api/credentials.
type createCredentialRequest struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
}

// credentialResponse describes a stored credential. There is deliberately no
// field that can carry the key or the ciphertext.
type credentialResponse struct {
	ID        string `json:"id"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	KeyHint   string `json:"key_hint"`
	IsDefault bool   `json:"is_default"`
}

// Create handles POST /api/credentials.
func (h *CredentialHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var req createCredentialRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	provider := external.Provider(strings.ToLower(strings.TrimSpace(req.Provider)))
	if !provider.Valid() {
		writeError(w, http.StatusBadRequest, "unsupported_provider")
		return
	}

	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" {
		writeError(w, http.StatusBadRequest, "missing_api_key")
		return
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = provider.DefaultModel()
	}

	repo := repository.NewCredentialRepo(h.db)
	existing, err := repo.ListByUser(userID)
	if err != nil {
		slog.Error("list credentials before create failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	ciphertext, nonce, err := h.cipher.Encrypt([]byte(apiKey), external.CredentialAAD(userID, provider, model))
	if err != nil {
		slog.Error("encrypt credential failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	id, err := repository.NewID()
	if err != nil {
		slog.Error("generate credential id failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	now := time.Now().Unix()
	cred := repository.Credential{
		ID: id, UserID: userID, Provider: string(provider), Model: model,
		Ciphertext: ciphertext, Nonce: nonce, KeyHint: external.KeyHint(apiKey),
		// The first key a user adds becomes their default, otherwise they
		// would add a credential and still be told none is configured.
		IsDefault: len(existing) == 0,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Insert(cred); err != nil {
		slog.Error("insert credential failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusCreated, toCredentialResponse(cred))
}

// List handles GET /api/credentials.
func (h *CredentialHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	creds, err := repository.NewCredentialRepo(h.db).ListByUser(userID)
	if err != nil {
		slog.Error("list credentials failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	out := make([]credentialResponse, 0, len(creds))
	for _, c := range creds {
		out = append(out, toCredentialResponse(c))
	}
	writeJSON(w, http.StatusOK, out)
}

// SetDefault handles POST /api/credentials/{id}/default.
func (h *CredentialHandler) SetDefault(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id")
		return
	}

	err := repository.NewCredentialRepo(h.db).SetDefault(userID, id)
	if errors.Is(err, repository.ErrCredentialNotFound) {
		writeError(w, http.StatusNotFound, "credential_not_found")
		return
	}
	if err != nil {
		slog.Error("set default credential failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

// Delete handles DELETE /api/credentials/{id}.
func (h *CredentialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id")
		return
	}

	if err := repository.NewCredentialRepo(h.db).Delete(userID, id); err != nil {
		slog.Error("delete credential failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

// toCredentialResponse projects a stored credential onto the wire shape,
// dropping the ciphertext and nonce on the way out.
func toCredentialResponse(c repository.Credential) credentialResponse {
	return credentialResponse{
		ID:        c.ID,
		Provider:  c.Provider,
		Model:     c.Model,
		KeyHint:   c.KeyHint,
		IsDefault: c.IsDefault,
	}
}
