package repository

import (
	"fmt"
)

// PaperRelationshipRepo provides CRUD access to the paper_relationships table.
type PaperRelationshipRepo struct {
	db dbExecutor
}

// NewPaperRelationshipRepo returns a new PaperRelationshipRepo.
func NewPaperRelationshipRepo(db dbExecutor) *PaperRelationshipRepo {
	return &PaperRelationshipRepo{db: db}
}

// Insert writes one paper_relationship row.
func (r *PaperRelationshipRepo) Insert(rel PaperRelationship) error {
	_, err := r.db.Exec(
		`INSERT INTO paper_relationships (id, source_paper_id, target_paper_id, relationship_type, evidence_text, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		rel.ID, rel.SourcePaperID, rel.TargetPaperID, rel.RelationshipType, rel.EvidenceText, rel.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert paper_relationship: %w", err)
	}
	return nil
}

// GetBySourcePaper returns all relationships where the given paper is the source.
func (r *PaperRelationshipRepo) GetBySourcePaper(paperID string) ([]PaperRelationship, error) {
	rows, err := r.db.Query(
		`SELECT id, source_paper_id, target_paper_id, relationship_type, evidence_text, created_at
		FROM paper_relationships WHERE source_paper_id = ? ORDER BY created_at ASC`, paperID,
	)
	if err != nil {
		return nil, fmt.Errorf("get paper_relationships: %w", err)
	}
	defer rows.Close()

	var list []PaperRelationship
	for rows.Next() {
		var rel PaperRelationship
		if err := rows.Scan(&rel.ID, &rel.SourcePaperID, &rel.TargetPaperID, &rel.RelationshipType, &rel.EvidenceText, &rel.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan paper_relationship: %w", err)
		}
		list = append(list, rel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate paper_relationships: %w", err)
	}
	return list, nil
}

// GetByTargetPaper returns all relationships where the given paper is the target.
func (r *PaperRelationshipRepo) GetByTargetPaper(paperID string) ([]PaperRelationship, error) {
	rows, err := r.db.Query(
		`SELECT id, source_paper_id, target_paper_id, relationship_type, evidence_text, created_at
		FROM paper_relationships WHERE target_paper_id = ? ORDER BY created_at ASC`, paperID,
	)
	if err != nil {
		return nil, fmt.Errorf("get paper_relationships: %w", err)
	}
	defer rows.Close()

	var list []PaperRelationship
	for rows.Next() {
		var rel PaperRelationship
		if err := rows.Scan(&rel.ID, &rel.SourcePaperID, &rel.TargetPaperID, &rel.RelationshipType, &rel.EvidenceText, &rel.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan paper_relationship: %w", err)
		}
		list = append(list, rel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate paper_relationships: %w", err)
	}
	return list, nil
}

// ListByPaper returns all relationships where the given paper is source or target.
func (r *PaperRelationshipRepo) ListByPaper(paperID string) ([]PaperRelationship, error) {
	rows, err := r.db.Query(
		`SELECT id, source_paper_id, target_paper_id, relationship_type, evidence_text, created_at
		FROM paper_relationships
		WHERE source_paper_id = ? OR target_paper_id = ?
		ORDER BY created_at ASC`, paperID, paperID,
	)
	if err != nil {
		return nil, fmt.Errorf("list paper_relationships: %w", err)
	}
	defer rows.Close()

	var list []PaperRelationship
	for rows.Next() {
		var rel PaperRelationship
		if err := rows.Scan(&rel.ID, &rel.SourcePaperID, &rel.TargetPaperID, &rel.RelationshipType, &rel.EvidenceText, &rel.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan paper_relationship: %w", err)
		}
		list = append(list, rel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate paper_relationships: %w", err)
	}
	return list, nil
}

// GetRelationships returns all relationships between two papers, bidirectional.
func (r *PaperRelationshipRepo) GetRelationships(sourcePaperID, targetPaperID string) ([]PaperRelationship, error) {
	rows, err := r.db.Query(
		`SELECT id, source_paper_id, target_paper_id, relationship_type, evidence_text, created_at
		FROM paper_relationships
		WHERE (source_paper_id = ? AND target_paper_id = ?)
		   OR (source_paper_id = ? AND target_paper_id = ?)`, sourcePaperID, targetPaperID, targetPaperID, sourcePaperID,
	)
	if err != nil {
		return nil, fmt.Errorf("get paper_relationships: %w", err)
	}
	defer rows.Close()

	var list []PaperRelationship
	for rows.Next() {
		var rel PaperRelationship
		if err := rows.Scan(&rel.ID, &rel.SourcePaperID, &rel.TargetPaperID, &rel.RelationshipType, &rel.EvidenceText, &rel.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan paper_relationship: %w", err)
		}
		list = append(list, rel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate paper_relationships: %w", err)
	}
	if len(list) == 0 {
		return nil, ErrNotFound
	}
	return list, nil
}
