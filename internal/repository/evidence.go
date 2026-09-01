package repository

import (
	"fmt"
)

type EvidenceRepo struct {
	db dbExecutor
}

func NewEvidenceRepo(db dbExecutor) *EvidenceRepo {
	return &EvidenceRepo{db: db}
}

func (r *EvidenceRepo) Insert(e Evidence) error {
	_, err := r.db.Exec(
		`INSERT INTO evidence (id, paper_id, page, figure_id, table_id, section, source_text, source_reference)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.PaperID, e.Page, e.FigureID, e.TableID, e.Section, e.SourceText, e.SourceReference,
	)
	if err != nil {
		return fmt.Errorf("insert evidence: %w", err)
	}
	return nil
}

func (r *EvidenceRepo) ListByPaper(paperID string) ([]Evidence, error) {
	rows, err := r.db.Query(
		`SELECT id, paper_id, page, figure_id, table_id, section, source_text, source_reference
		FROM evidence WHERE paper_id = ?`, paperID,
	)
	if err != nil {
		return nil, fmt.Errorf("list evidence by paper: %w", err)
	}
	defer rows.Close()

	var list []Evidence
	for rows.Next() {
		var e Evidence
		if err := rows.Scan(&e.ID, &e.PaperID, &e.Page, &e.FigureID, &e.TableID, &e.Section, &e.SourceText, &e.SourceReference); err != nil {
			return nil, fmt.Errorf("scan evidence: %w", err)
		}
		list = append(list, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate evidence: %w", err)
	}
	return list, nil
}

func (r *EvidenceRepo) GetClaims(evidenceID string) ([]Claim, error) {
	rows, err := r.db.Query(
		`SELECT c.id, c.paper_id, c.claim_text, c.claim_type, c.confidence, c.source_page, c.source_text, c.created_at
		FROM claims c INNER JOIN claim_evidence ce ON c.id = ce.claim_id WHERE ce.evidence_id = ? ORDER BY c.created_at ASC`, evidenceID,
	)
	if err != nil {
		return nil, fmt.Errorf("get claims for evidence: %w", err)
	}
	defer rows.Close()

	var claims []Claim
	for rows.Next() {
		var c Claim
		if err := rows.Scan(&c.ID, &c.PaperID, &c.ClaimText, &c.ClaimType, &c.Confidence, &c.SourcePage, &c.SourceText, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan claim: %w", err)
		}
		claims = append(claims, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claims: %w", err)
	}
	return claims, nil
}
