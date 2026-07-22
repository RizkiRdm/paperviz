package repository

import "fmt"

// ClaimDiffRepo provides CRUD access to the claim_diffs table. One row per
// document (ARCHITECTURE.md Section C: "Document (1) — (1) ClaimDiff").
type ClaimDiffRepo struct {
	db dbExecutor
}

func NewClaimDiffRepo(db dbExecutor) *ClaimDiffRepo {
	return &ClaimDiffRepo{db: db}
}

// Insert writes the claim-diff verification result for a document.
// originalClaims/simplifiedClaims are pre-serialized JSON arrays — the
// repository does not know or care about their internal structure, per
// ARCHITECTURE.md's "no business logic in repository" rule.
func (r *ClaimDiffRepo) Insert(cd ClaimDiff) error {
	mismatch := 0
	if cd.MismatchDetected {
		mismatch = 1
	}
	_, err := r.db.Exec(
		`INSERT INTO claim_diffs (id, document_id, original_claims, simplified_claims, mismatch_detected, mismatch_detail)
		VALUES (?, ?, ?, ?, ?, ?)`,
		cd.ID, cd.DocumentID, cd.OriginalClaims, cd.SimplifiedClaims, mismatch, cd.MismatchDetail,
	)
	if err != nil {
		return fmt.Errorf("insert claim_diff: %w", err)
	}
	return nil
}

// GetByDocument retrieves the claim-diff row for a document, if any exists
// (a "processing" or "failed" document may not have one yet).
func (r *ClaimDiffRepo) GetByDocument(documentID string) (*ClaimDiff, error) {
	row := r.db.QueryRow(
		`SELECT id, document_id, original_claims, simplified_claims, mismatch_detected, mismatch_detail
		FROM claim_diffs WHERE document_id = ?`, documentID,
	)
	var cd ClaimDiff
	var mismatch int
	err := row.Scan(&cd.ID, &cd.DocumentID, &cd.OriginalClaims, &cd.SimplifiedClaims, &mismatch, &cd.MismatchDetail)
	if err != nil {
		return nil, err // sql.ErrNoRows is a valid "not yet verified" state; callers check err directly
	}
	cd.MismatchDetected = mismatch == 1
	return &cd, nil
}
