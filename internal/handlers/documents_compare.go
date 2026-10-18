package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"paperviz/internal/app/documents"
	"paperviz/internal/services"
)

type compareRequest struct {
	DocumentIDs []string `json:"document_ids"`
}

type compareResponse struct {
	Papers         []services.PaperSummary        `json:"papers"`
	Dimensions     []services.ComparisonDimension `json:"dimensions"`
	Agreement      []string                       `json:"agreement"`
	Disagreement   []string                       `json:"disagreement"`
	EvidenceClaims []services.EvidenceClaim       `json:"evidence_claims"`
}

// Compare handles POST /api/documents/compare.
func (h *DocumentHandler) Compare(w http.ResponseWriter, r *http.Request) {
	var req compareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	if len(req.DocumentIDs) < 2 || len(req.DocumentIDs) > 10 {
		writeError(w, http.StatusBadRequest, "document_ids_must_be_2_to_10")
		return
	}

	client, ok := h.resolveClient(w, r)
	if !ok {
		return
	}

	comparison, err := h.svc.Compare(r.Context(), client, req.DocumentIDs)
	switch {
	case errors.Is(err, documents.ErrDocumentMissing):
		writeError(w, http.StatusNotFound, "document_not_found")
		return
	case errors.Is(err, documents.ErrExtractionFailed):
		writeError(w, http.StatusInternalServerError, "extraction_failed")
		return
	case errors.Is(err, documents.ErrComparisonFailed):
		writeError(w, http.StatusInternalServerError, "comparison_failed")
		return
	case err != nil:
		slog.Error("compare documents failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, compareResponse{
		Papers:         comparison.Papers,
		Dimensions:     comparison.Dimensions,
		Agreement:      comparison.Agreement,
		Disagreement:   comparison.Disagreement,
		EvidenceClaims: comparison.EvidenceClaims,
	})
}
