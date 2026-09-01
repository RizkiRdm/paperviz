package repository

import (
	"fmt"
)

// ClaimEvidenceRepo provides CRUD access to the claim_evidence join table.
// It holds no business logic — every method is a parameterized SQL operation.
type ClaimEvidenceRepo struct {
	db dbExecutor
}

// NewClaimEvidenceRepo returns a ClaimEvidenceRepo backed by the given executor.
func NewClaimEvidenceRepo(db dbExecutor) *ClaimEvidenceRepo {
	return &ClaimEvidenceRepo{db: db}
}

// Insert writes one claim_evidence row.
func (r *ClaimEvidenceRepo) Insert(ce ClaimEvidence) error {
	_, err := r.db.Exec(
		`INSERT INTO claim_evidence (id, claim_id, evidence_id, relationship_type, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		ce.ID, ce.ClaimID, ce.EvidenceID, ce.RelationshipType, ce.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert claim_evidence: %w", err)
	}
	return nil
}

// GetByClaim returns all claim_evidence rows for a given claim, ordered by creation time.
func (r *ClaimEvidenceRepo) GetByClaim(claimID string) ([]ClaimEvidence, error) {
	rows, err := r.db.Query(
		`SELECT id, claim_id, evidence_id, relationship_type, created_at
		FROM claim_evidence WHERE claim_id = ? ORDER BY created_at ASC`, claimID,
	)
	if err != nil {
		return nil, fmt.Errorf("get claim_evidence by claim: %w", err)
	}
	defer rows.Close()

	var list []ClaimEvidence
	for rows.Next() {
		var ce ClaimEvidence
		if err := rows.Scan(&ce.ID, &ce.ClaimID, &ce.EvidenceID, &ce.RelationshipType, &ce.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan claim_evidence: %w", err)
		}
		list = append(list, ce)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claim_evidence: %w", err)
	}
	return list, nil
}

// GetByEvidence returns all claim_evidence rows for a given evidence, ordered by creation time.
func (r *ClaimEvidenceRepo) GetByEvidence(evidenceID string) ([]ClaimEvidence, error) {
	rows, err := r.db.Query(
		`SELECT id, claim_id, evidence_id, relationship_type, created_at
		FROM claim_evidence WHERE evidence_id = ? ORDER BY created_at ASC`, evidenceID,
	)
	if err != nil {
		return nil, fmt.Errorf("get claim_evidence by evidence: %w", err)
	}
	defer rows.Close()

	var list []ClaimEvidence
	for rows.Next() {
		var ce ClaimEvidence
		if err := rows.Scan(&ce.ID, &ce.ClaimID, &ce.EvidenceID, &ce.RelationshipType, &ce.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan claim_evidence: %w", err)
		}
		list = append(list, ce)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claim_evidence: %w", err)
	}
	return list, nil
}
