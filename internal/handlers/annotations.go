package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/services"
)

// AnnotationHandler handles annotation CRUD for documents.
type AnnotationHandler struct {
	db *sql.DB
}

// NewAnnotationHandler returns a new AnnotationHandler.
func NewAnnotationHandler(db *sql.DB) *AnnotationHandler {
	return &AnnotationHandler{db: db}
}

// createAnnotationRequest is the wire format for creating an annotation.
type createAnnotationRequest struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Content    string `json:"content"`
}

// updateAnnotationRequest is the wire format for updating an annotation.
type updateAnnotationRequest struct {
	Content string `json:"content"`
}

// annotationResponse is the wire format for a single annotation.
type annotationResponse struct {
	ID         string `json:"id"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Content    string `json:"content"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// List returns all annotations for a document owned by the authenticated user.
func (h *AnnotationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	documentID := chi.URLParam(r, "id")

	annotations, err := services.ListAnnotations(h.db, documentID, userID)
	if err != nil {
		slog.Error("list annotations failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	resp := make([]annotationResponse, 0, len(annotations))
	for _, a := range annotations {
		resp = append(resp, annotationResponse{
			ID:         a.ID,
			TargetType: a.TargetType,
			TargetID:   a.TargetID,
			Content:    a.Content,
			CreatedAt:  a.CreatedAt,
			UpdatedAt:  a.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"annotations": resp})
}

// Create adds a new annotation on a paper or claim target.
func (h *AnnotationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	documentID := chi.URLParam(r, "id")

	var req createAnnotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if req.TargetType == "" || req.TargetID == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "missing_required_field")
		return
	}
	if req.TargetType != "paper" && req.TargetType != "claim" {
		writeError(w, http.StatusBadRequest, "invalid_target_type")
		return
	}

	id, err := services.CreateAnnotation(h.db, userID, documentID, req.TargetType, req.TargetID, req.Content)
	if err != nil {
		slog.Error("create annotation failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	annotation, err := services.GetAnnotation(h.db, id)
	if err != nil {
		slog.Error("get annotation after create failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusCreated, annotationResponse{
		ID:         annotation.ID,
		TargetType: annotation.TargetType,
		TargetID:   annotation.TargetID,
		Content:    annotation.Content,
		CreatedAt:  annotation.CreatedAt,
		UpdatedAt:  annotation.UpdatedAt,
	})
}

// Update changes the content of an existing annotation owned by the user.
func (h *AnnotationHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	annotationID := chi.URLParam(r, "annotationId")

	var req updateAnnotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content_required")
		return
	}

	if err := services.UpdateAnnotation(h.db, annotationID, userID, req.Content); err != nil {
		slog.Error("update annotation failed", "error", err)
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	annotation, err := services.GetAnnotation(h.db, annotationID)
	if err != nil {
		slog.Error("get annotation after update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, annotationResponse{
		ID:         annotation.ID,
		TargetType: annotation.TargetType,
		TargetID:   annotation.TargetID,
		Content:    annotation.Content,
		CreatedAt:  annotation.CreatedAt,
		UpdatedAt:  annotation.UpdatedAt,
	})
}

// Delete removes an annotation owned by the authenticated user.
func (h *AnnotationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	annotationID := chi.URLParam(r, "annotationId")

	if err := services.DeleteAnnotation(h.db, annotationID, userID); err != nil {
		slog.Error("delete annotation failed", "error", err)
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
