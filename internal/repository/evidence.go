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
