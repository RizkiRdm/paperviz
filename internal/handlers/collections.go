package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/repository"
	"paperviz/internal/services"
)

type CollectionHandler struct {
	db *sql.DB
}

func NewCollectionHandler(db *sql.DB) *CollectionHandler {
	return &CollectionHandler{db: db}
}

type createCollectionRequest struct {
	Name string `json:"name"`
}

type collectionResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CreatedAt     int64  `json:"created_at"`
	DocumentCount int    `json:"document_count"`
}

type collectionDetailResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	CreatedAt int64             `json:"created_at"`
	Documents []documentSummary `json:"documents"`
}

type addDocumentRequest struct {
	DocumentID string `json:"document_id"`
}

func (h *CollectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name_required")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id, err := repository.NewID()
	if err != nil {
		slog.Error("generate collection id failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if err := services.CreateCollection(h.db, id, userID, req.Name, time.Now().Unix()); err != nil {
		slog.Error("create collection failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusCreated, collectionResponse{
		ID:        id,
		Name:      req.Name,
		CreatedAt: time.Now().Unix(),
	})
}

func (h *CollectionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	collections, err := services.ListCollections(h.db, userID)
	if err != nil {
		slog.Error("list collections failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"collections": collections})
}

// Get returns collection detail with documents for authenticated owner.
func (h *CollectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id := chi.URLParam(r, "id")

	col, err := services.GetCollection(h.db, id, userID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("get collection failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	docs, err := services.ListCollectionDocuments(h.db, id, userID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("list collection documents failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	summaries := make([]documentSummary, 0, len(docs))
	for _, d := range docs {
		summaries = append(summaries, documentSummary{
			ID:               d.ID,
			Title:            d.Title,
			Status:           d.Status,
			CreatedAt:        d.CreatedAt,
			SummaryPreview:   d.SummaryPreview,
			ChartCount:       d.ChartCount,
			ExplanationCount: d.ExplanationCount,
		})
	}

	writeJSON(w, http.StatusOK, collectionDetailResponse{
		ID:        col.ID,
		Name:      col.Name,
		CreatedAt: col.CreatedAt,
		Documents: summaries,
	})
}

// Rename changes collection name after verifying ownership.
func (h *CollectionHandler) Rename(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id := chi.URLParam(r, "id")

	var req createCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name_required")
		return
	}

	if err := services.RenameCollection(h.db, id, userID, req.Name); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("rename collection failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"name": req.Name})
}

// Delete removes collection owned by authenticated user.
func (h *CollectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id := chi.URLParam(r, "id")

	if err := services.DeleteCollection(h.db, id, userID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("delete collection failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// AddDocument adds document to collection owned by authenticated user.
func (h *CollectionHandler) AddDocument(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id := chi.URLParam(r, "id")

	var req addDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if req.DocumentID == "" {
		writeError(w, http.StatusBadRequest, "document_id_required")
		return
	}

	if err := services.AddDocumentToCollection(h.db, id, userID, req.DocumentID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("add document to collection failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

// RemoveDocument removes document from collection owned by authenticated user.
func (h *CollectionHandler) RemoveDocument(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id := chi.URLParam(r, "id")
	docID := chi.URLParam(r, "docId")

	if err := services.RemoveDocumentFromCollection(h.db, id, userID, docID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("remove document from collection failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}
