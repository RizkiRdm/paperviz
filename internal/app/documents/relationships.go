package documents

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"paperviz/internal/repository"
)

// Domain errors for the relationship endpoints. Handlers map each to the same
// status and code the inline orchestration used to produce.
var (
	// ErrInvalidRelationshipType means the type is not one of the five the
	// schema defines.
	ErrInvalidRelationshipType = errors.New("invalid relationship type")

	// ErrSelfReference means a paper was related to itself.
	ErrSelfReference = errors.New("cannot relate a paper to itself")

	// ErrTargetNotFound means the target paper does not exist.
	ErrTargetNotFound = errors.New("target paper not found")
)

// validRelationshipTypes mirrors the paper_relationships CHECK values the API
// accepts; it moved here so the transport layer never sees domain enums.
var validRelationshipTypes = map[string]bool{
	repository.PaperRelationshipSupporting:         true,
	repository.PaperRelationshipContradicting:      true,
	repository.PaperRelationshipCiting:             true,
	repository.PaperRelationshipSimilarMethodology: true,
	repository.PaperRelationshipDifferentFindings:  true,
}

// Relationships returns every relationship where the paper is source or target.
func (s *Service) Relationships(id string) ([]repository.PaperRelationship, error) {
	if _, err := s.requireDoc(id); err != nil {
		return nil, err
	}
	return repository.NewPaperRelationshipRepo(s.db).ListByPaper(id)
}

// CreateRelationship validates and inserts one paper-to-paper relationship.
func (s *Service) CreateRelationship(sourceID, targetID, relType string, evidenceText *string) (*repository.PaperRelationship, error) {
	if _, err := s.requireDoc(sourceID); err != nil {
		return nil, err
	}
	if !validRelationshipTypes[relType] {
		return nil, ErrInvalidRelationshipType
	}
	if targetID == sourceID {
		return nil, ErrSelfReference
	}
	if _, err := repository.NewDocumentRepo(s.db).Get(targetID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTargetNotFound
		}
		slog.Error("get target document failed", "error", err)
		return nil, err
	}
	relID, err := repository.NewID()
	if err != nil {
		return nil, fmt.Errorf("generate relationship id: %w", err)
	}
	rel := repository.PaperRelationship{
		ID:               relID,
		SourcePaperID:    sourceID,
		TargetPaperID:    targetID,
		RelationshipType: relType,
		EvidenceText:     evidenceText,
		CreatedAt:        time.Now().Unix(),
	}
	if err := repository.NewPaperRelationshipRepo(s.db).Insert(rel); err != nil {
		return nil, err
	}
	return &rel, nil
}
