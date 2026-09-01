package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// ResultRepo provides CRUD access to the results table.
type ResultRepo struct {
	db dbExecutor
}

// NewResultRepo returns a ResultRepo backed by the given executor.
func NewResultRepo(db dbExecutor) *ResultRepo {
	return &ResultRepo{db: db}
}

// Insert writes one result row, linked to its parent paper.
func (r *ResultRepo) Insert(res Result) error {
	_, err := r.db.Exec(
		`INSERT INTO results (id, paper_id, result_text, result_type, supporting_evidence_id, source_page, source_text)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		res.ID, res.PaperID, res.ResultText, res.ResultType, res.SupportingEvidenceID, res.SourcePage, res.SourceText,
	)
	if err != nil {
		return fmt.Errorf("insert result: %w", err)
	}
	return nil
}

// ListByPaper returns all results for a paper, ordered by id ascending.
func (r *ResultRepo) ListByPaper(paperID string) ([]Result, error) {
	rows, err := r.db.Query(
		`SELECT id, paper_id, result_text, result_type, supporting_evidence_id, source_page, source_text
		FROM results WHERE paper_id = ? ORDER BY id ASC`, paperID,
	)
	if err != nil {
		return nil, fmt.Errorf("list results: %w", err)
	}
	defer rows.Close()

	var list []Result
	for rows.Next() {
		var res Result
		if err := rows.Scan(&res.ID, &res.PaperID, &res.ResultText, &res.ResultType, &res.SupportingEvidenceID, &res.SourcePage, &res.SourceText); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		list = append(list, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate results: %w", err)
	}
	return list, nil
}

// GetByID returns a single result, scoped to its parent paper.
// Returns ErrNotFound if no row matches either the id or the paper scope.
func (r *ResultRepo) GetByID(paperID, resultID string) (*Result, error) {
	row := r.db.QueryRow(
		`SELECT id, paper_id, result_text, result_type, supporting_evidence_id, source_page, source_text
		FROM results WHERE id = ? AND paper_id = ?`, resultID, paperID,
	)
	var res Result
	err := row.Scan(&res.ID, &res.PaperID, &res.ResultText, &res.ResultType, &res.SupportingEvidenceID, &res.SourcePage, &res.SourceText)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get result: %w", err)
	}
	return &res, nil
}
