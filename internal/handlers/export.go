package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/services"
)

// ExportHandler handles the document research-context export endpoint.
type ExportHandler struct {
	db *sql.DB
}

// NewExportHandler creates an ExportHandler with the given database connection.
func NewExportHandler(db *sql.DB) *ExportHandler {
	return &ExportHandler{db: db}
}

// Export handles GET /api/documents/{id}/export, returning a downloadable JSON export of the research context.
func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	documentID := chi.URLParam(r, "id")

	export, err := services.ExportResearchContext(h.db, documentID)
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
