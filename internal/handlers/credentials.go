package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/app/credentials"
	"paperviz/internal/external"
)

// CredentialHandler manages the BYOK model keys a user supplies.
//
// PaperViz holds no AI vendor account, so these keys are the most sensitive
// data it stores. Two rules govern every method here: the plaintext key is
// never written to a log, and never returned in a response.
type CredentialHandler struct {
	store *credentials.Store
}

// NewCredentialHandler creates a CredentialHandler bound to a cipher.
func NewCredentialHandler(db *sql.DB, cipher *external.Cipher) *CredentialHandler {
	return &CredentialHandler{store: credentials.NewStore(db, cipher)}
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

	cred, err := h.store.Create(userID, string(provider), model, apiKey)
	if err != nil {
		slog.Error("create credential failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusCreated, toCredentialResponse(*cred))
}

// List handles GET /api/credentials.
func (h *CredentialHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	creds, err := h.store.List(userID)
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

	err := h.store.SetDefault(userID, id)
	if errors.Is(err, credentials.ErrNotFound) {
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

	if err := h.store.Delete(userID, id); err != nil {
		slog.Error("delete credential failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

// toCredentialResponse projects a stored credential onto the wire shape,
// dropping the ciphertext and nonce on the way out.
func toCredentialResponse(c credentials.Credential) credentialResponse {
	return credentialResponse{
		ID:        c.ID,
		Provider:  c.Provider,
		Model:     c.Model,
		KeyHint:   c.KeyHint,
		IsDefault: c.IsDefault,
	}
}

// writeCredentialError maps a credential resolution failure onto a response.
// It is shared by every endpoint that needs a model client, so a user missing
// a key gets the same actionable answer wherever they hit it.
func writeCredentialError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, credentials.ErrNoCredential):
		writeError(w, http.StatusBadRequest, "missing_credential")
	case errors.Is(err, credentials.ErrCredentialUnreadable):
		// Causes: a rotated CREDENTIAL_ENCRYPTION_KEY, or a credential row whose
		// provider/model no longer matches the AAD it was sealed with. Neither
		// resolves on retry, so this is a conflict rather than a server fault.
		slog.Error("model credential unreadable", "error", err)
		writeError(w, http.StatusConflict, "credential_unreadable")
	default:
		slog.Error("resolve model credential failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
	}
}
