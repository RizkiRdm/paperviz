package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/services"
)

// exportService abstracts research-context export for transport layer.
type exportService interface {
	Export(documentID string) (*services.ResearchExport, error)
}

// ExportHandler handles research-context export.
type ExportHandler struct {
	svc exportService
}

// NewExportHandler creates ExportHandler delegating to services layer.
func NewExportHandler(db *sql.DB) *ExportHandler {
	return &ExportHandler{svc: services.NewExportService(db)}
}

// Export handles GET /api/documents/{id}/export.
func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	documentID := chi.URLParam(r, "id")
	export, err := h.svc.Export(documentID)
	if err != nil {
		if strings.Contains(err.Error(), "document not found") {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("export research context failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="research-context-`+documentID+`.json"`)
	writeJSON(w, http.StatusOK, export)
}
