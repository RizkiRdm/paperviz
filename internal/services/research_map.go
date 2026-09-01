package services

import (
	"database/sql"
	"errors"
	"fmt"

	"paperviz/internal/repository"
)

// ResearchMapRelationship is one relationship enriched with the target paper's title.
type ResearchMapRelationship struct {
	ID               string  `json:"id"`
	SourcePaperID    string  `json:"source_paper_id"`
	TargetPaperID    string  `json:"target_paper_id"`
	TargetPaperTitle string  `json:"target_paper_title"`
	RelationshipType string  `json:"relationship_type"`
	EvidenceText     *string `json:"evidence_text,omitempty"`
	CreatedAt        int64   `json:"created_at"`
}

// ResearchMapResult groups all relationships for a document by type.
type ResearchMapResult struct {
	DocumentID    string                              `json:"document_id"`
	Relationships map[string][]ResearchMapRelationship `json:"relationships"`
	TotalCount    int                                 `json:"total_count"`
}

// allRelationshipTypes is the ordered list of relationship type keys for the response.
var allRelationshipTypes = []string{
	repository.PaperRelationshipSupporting,
	repository.PaperRelationshipContradicting,
	repository.PaperRelationshipCiting,
	repository.PaperRelationshipSimilarMethodology,
	repository.PaperRelationshipDifferentFindings,
}

// GetResearchMap returns all relationships for a document grouped by type, with paper titles.
func GetResearchMap(db *sql.DB, documentID string) (*ResearchMapResult, error) {
	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.Get(documentID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}

	_ = doc // document exists check only

	relRepo := repository.NewPaperRelationshipRepo(db)
	rels, err := relRepo.ListByPaper(documentID)
	if err != nil {
		return nil, fmt.Errorf("list relationships: %w", err)
	}

	// Collect unique paper IDs to batch-fetch titles.
	seen := make(map[string]bool)
	for _, rel := range rels {
		otherID := rel.TargetPaperID
		if rel.TargetPaperID == documentID {
			otherID = rel.SourcePaperID
		}
		seen[otherID] = true
	}

	titles := make(map[string]string, len(seen))
	for id := range seen {
		d, err := docRepo.Get(id)
		if err != nil {
			titles[id] = "Unknown Paper"
			continue
		}
		titles[id] = d.Title
	}

	// Group relationships by type.
	grouped := make(map[string][]ResearchMapRelationship, len(allRelationshipTypes))
	for _, t := range allRelationshipTypes {
		grouped[t] = []ResearchMapRelationship{}
	}

	for _, rel := range rels {
		otherID := rel.TargetPaperID
		if rel.TargetPaperID == documentID {
			otherID = rel.SourcePaperID
		}
		grouped[rel.RelationshipType] = append(grouped[rel.RelationshipType], ResearchMapRelationship{
			ID:               rel.ID,
			SourcePaperID:    rel.SourcePaperID,
			TargetPaperID:    rel.TargetPaperID,
			TargetPaperTitle: titles[otherID],
			RelationshipType: rel.RelationshipType,
			EvidenceText:     rel.EvidenceText,
			CreatedAt:        rel.CreatedAt,
		})
	}

	return &ResearchMapResult{
		DocumentID:    documentID,
		Relationships: grouped,
		TotalCount:    len(rels),
	}, nil
}
