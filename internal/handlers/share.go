package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/services"
)

type ShareHandler struct {
	db *sql.DB
}

func NewShareHandler(db *sql.DB) *ShareHandler {
	return &ShareHandler{db: db}
}

type shareTokenResponse struct {
	ShareToken string `json:"share_token"`
	ShareURL   string `json:"share_url"`
}

type docShareURLResponse struct {
	ShareURL string `json:"share_url"`
}

func (h *ShareHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	chartID := chi.URLParam(r, "chartId")
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	token, err := services.GenerateShareToken(r.Context(), h.db, docID, chartID, userID)
	if err != nil {
		switch err.Error() {
		case "document not found", "chart not found":
			writeError(w, http.StatusNotFound, "not_found")
		case "unauthorized":
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			slog.Error("generate share token failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	writeJSON(w, http.StatusOK, shareTokenResponse{
		ShareToken: token,
		ShareURL:   "/share/fig/" + token,
	})
}

func (h *ShareHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	chartID := chi.URLParam(r, "chartId")
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err := services.RevokeShareToken(r.Context(), h.db, docID, chartID, userID)
	if err != nil {
		switch err.Error() {
		case "document not found":
			writeError(w, http.StatusNotFound, "not_found")
		case "unauthorized":
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			slog.Error("revoke share token failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ShareHandler) GetSharedFigure(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "shareToken")

	fig, err := services.GetSharedFigure(r.Context(), h.db, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "not found" {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("get shared figure failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, fig)
}

func (h *ShareHandler) GenerateDocToken(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	token, err := services.GenerateDocumentShareToken(r.Context(), h.db, docID, userID)
	if err != nil {
		switch err.Error() {
		case "document not found":
			writeError(w, http.StatusNotFound, "not_found")
		case "unauthorized":
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			slog.Error("generate document share token failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	writeJSON(w, http.StatusOK, docShareURLResponse{
		ShareURL: "/share/doc/" + token,
	})
}

func (h *ShareHandler) RevokeDocToken(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err := services.RevokeDocumentShareToken(r.Context(), h.db, docID, userID)
	if err != nil {
		switch err.Error() {
		case "document not found":
			writeError(w, http.StatusNotFound, "not_found")
		case "unauthorized":
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			slog.Error("revoke document share token failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ShareHandler) UpdateVisibility(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Visibility string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	if err := services.SetDocumentVisibility(r.Context(), h.db, docID, userID, req.Visibility); err != nil {
		switch err.Error() {
		case "invalid visibility":
			writeError(w, http.StatusBadRequest, "invalid_visibility")
		case "document not found":
			writeError(w, http.StatusNotFound, "not_found")
		case "unauthorized":
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			slog.Error("update visibility failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"visibility": req.Visibility})
}

func (h *ShareHandler) GetSharedPaper(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "shareToken")

	paper, err := services.GetSharedPaper(r.Context(), h.db, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "not found" {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("get shared paper failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, paper)
}
