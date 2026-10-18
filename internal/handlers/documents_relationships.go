package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"paperviz/internal/app/documents"
	"paperviz/internal/services"
)

type paperRelationshipResponse struct {
	ID               string  `json:"id"`
	SourcePaperID    string  `json:"source_paper_id"`
	TargetPaperID    string  `json:"target_paper_id"`
	RelationshipType string  `json:"relationship_type"`
	EvidenceText     *string `json:"evidence_text,omitempty"`
	CreatedAt        int64   `json:"created_at"`
}

type createRelationshipRequest struct {
	TargetPaperID    string  `json:"target_paper_id"`
	RelationshipType string  `json:"relationship_type"`
	EvidenceText     *string `json:"evidence_text,omitempty"`
}

func (h *DocumentHandler) GetPaperRelationships(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rels, err := h.svc.Relationships(id)
	if errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		slog.Error("list paper relationships failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	resp := make([]paperRelationshipResponse, 0, len(rels))
	for _, rel := range rels {
		resp = append(resp, paperRelationshipResponse{
			ID: rel.ID, SourcePaperID: rel.SourcePaperID,
			TargetPaperID: rel.TargetPaperID, RelationshipType: rel.RelationshipType,
			EvidenceText: rel.EvidenceText, CreatedAt: rel.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *DocumentHandler) CreatePaperRelationship(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Existence is checked before the body is parsed, matching the endpoint's
	// historical precedence: a missing source paper answers 404 even for a
	// malformed payload.
	if _, err := h.svc.GetMetadata(id); errors.Is(err, documents.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	} else if err != nil {
		slog.Error("get document for create relationship failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	var req createRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if req.TargetPaperID == "" {
		writeError(w, http.StatusBadRequest, "target_paper_id_required")
		return
	}
	if req.RelationshipType == "" {
		writeError(w, http.StatusBadRequest, "relationship_type_required")
		return
	}

	rel, err := h.svc.CreateRelationship(id, req.TargetPaperID, req.RelationshipType, req.EvidenceText)
	switch {
	case errors.Is(err, documents.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
		return
	case errors.Is(err, documents.ErrInvalidRelationshipType):
		writeError(w, http.StatusBadRequest, "invalid_relationship_type")
		return
	case errors.Is(err, documents.ErrSelfReference):
		writeError(w, http.StatusBadRequest, "cannot_self_reference")
		return
	case errors.Is(err, documents.ErrTargetNotFound):
		writeError(w, http.StatusNotFound, "target_paper_not_found")
		return
	case err != nil:
		slog.Error("create paper relationship failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusCreated, paperRelationshipResponse{
		ID: rel.ID, SourcePaperID: rel.SourcePaperID,
		TargetPaperID: rel.TargetPaperID, RelationshipType: rel.RelationshipType,
		EvidenceText: rel.EvidenceText, CreatedAt: rel.CreatedAt,
	})
}

type researchMapResponse struct {
	DocumentID    string                                   `json:"document_id"`
	Relationships map[string][]researchMapRelationshipItem `json:"relationships"`
	TotalCount    int                                      `json:"total_count"`
}

type researchMapRelationshipItem struct {
	ID               string  `json:"id"`
	SourcePaperID    string  `json:"source_paper_id"`
	TargetPaperID    string  `json:"target_paper_id"`
	TargetPaperTitle string  `json:"target_paper_title"`
	RelationshipType string  `json:"relationship_type"`
	EvidenceText     *string `json:"evidence_text,omitempty"`
	CreatedAt        int64   `json:"created_at"`
}

// GetResearchMap handles GET /api/documents/:id/research-map.
func (h *DocumentHandler) GetResearchMap(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := services.GetResearchMap(h.db, id)
	if err != nil {
		if err.Error() == "document not found" {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("get research map failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	relationships := make(map[string][]researchMapRelationshipItem, len(result.Relationships))
	for relType, items := range result.Relationships {
		respItems := make([]researchMapRelationshipItem, 0, len(items))
		for _, item := range items {
			respItems = append(respItems, researchMapRelationshipItem{
				ID:               item.ID,
				SourcePaperID:    item.SourcePaperID,
				TargetPaperID:    item.TargetPaperID,
				TargetPaperTitle: item.TargetPaperTitle,
				RelationshipType: item.RelationshipType,
				EvidenceText:     item.EvidenceText,
				CreatedAt:        item.CreatedAt,
			})
		}
		relationships[relType] = respItems
	}

	writeJSON(w, http.StatusOK, researchMapResponse{
		DocumentID:    result.DocumentID,
		Relationships: relationships,
		TotalCount:    result.TotalCount,
	})
}
