package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// ClaimRepo provides CRUD access to the claims table. One row per
// extracted claim from a paper, scoped to its parent paper.
type ClaimRepo struct {
	db dbExecutor
}

// NewClaimRepo returns a ClaimRepo backed by the given executor.
func NewClaimRepo(db dbExecutor) *ClaimRepo {
	return &ClaimRepo{db: db}
}

// Insert writes one claim row, linked to its parent paper.
func (r *ClaimRepo) Insert(c Claim) error {
	_, err := r.db.Exec(
		`INSERT INTO claims (id, paper_id, claim_text, claim_type, confidence, source_page, source_text, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.PaperID, c.ClaimText, c.ClaimType, c.Confidence, c.SourcePage, c.SourceText, c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert claim: %w", err)
	}
	return nil
}

// ListByPaper returns all claims for a paper, ordered chronologically.
func (r *ClaimRepo) ListByPaper(paperID string) ([]Claim, error) {
	rows, err := r.db.Query(
		`SELECT id, paper_id, claim_text, claim_type, confidence, source_page, source_text, created_at
		FROM claims WHERE paper_id = ? ORDER BY created_at ASC`, paperID,
	)
	if err != nil {
		return nil, fmt.Errorf("list claims: %w", err)
	}
	defer rows.Close()

	var list []Claim
	for rows.Next() {
		var c Claim
		if err := rows.Scan(&c.ID, &c.PaperID, &c.ClaimText, &c.ClaimType, &c.Confidence, &c.SourcePage, &c.SourceText, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan claim: %w", err)
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claims: %w", err)
	}
	return list, nil
}

// GetByID returns a single claim, scoped to its parent paper so a claim
// ID alone cannot be used to read another paper's claim.
// Returns ErrNotFound if no row matches either the id or the paper scope.
func (r *ClaimRepo) GetByID(paperID, claimID string) (*Claim, error) {
	row := r.db.QueryRow(
		`SELECT id, paper_id, claim_text, claim_type, confidence, source_page, source_text, created_at
		FROM claims WHERE id = ? AND paper_id = ?`, claimID, paperID,
	)
	var c Claim
	err := row.Scan(&c.ID, &c.PaperID, &c.ClaimText, &c.ClaimType, &c.Confidence, &c.SourcePage, &c.SourceText, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get claim: %w", err)
	}
	return &c, nil
}
