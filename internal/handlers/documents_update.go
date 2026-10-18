package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/services"
)

type toggleSavedRequest struct {
	Saved bool `json:"saved"`
}

type updateTitleRequest struct {
	Title string `json:"title"`
}

// ToggleSaved handles PUT /api/documents/:id/save
func (h *DocumentHandler) ToggleSaved(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := UserIDFromContext(r.Context())

	var req toggleSavedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	if err := services.ToggleDocumentSaved(h.db, id, userID, req.Saved); err != nil {
		writeDocumentAccessError(w, err, "toggle saved failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"saved": req.Saved})
}

// UpdateTitle handles PATCH /api/documents/:id
func (h *DocumentHandler) UpdateTitle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := UserIDFromContext(r.Context())

	var req updateTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title_required")
		return
	}

	if err := services.RenameDocument(h.db, id, userID, req.Title); err != nil {
		writeDocumentAccessError(w, err, "update title failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"title": req.Title})
}

// Delete handles DELETE /api/documents/:id
func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := UserIDFromContext(r.Context())

	if err := services.DeleteDocument(h.db, id, userID); err != nil {
		writeDocumentAccessError(w, err, "delete document failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
