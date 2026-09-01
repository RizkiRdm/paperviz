package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// MethodRepo provides CRUD access to the methods table.
type MethodRepo struct {
	db dbExecutor
}

// NewMethodRepo returns a MethodRepo backed by the given db executor.
func NewMethodRepo(db dbExecutor) *MethodRepo {
	return &MethodRepo{db: db}
}

// Insert writes one method row, linked to its parent paper.
func (r *MethodRepo) Insert(m Method) error {
	_, err := r.db.Exec(
		`INSERT INTO methods (id, paper_id, method_name, description, method_type, source_page, source_text)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.PaperID, m.MethodName, m.Description, m.MethodType, m.SourcePage, m.SourceText,
	)
	if err != nil {
		return fmt.Errorf("insert method: %w", err)
	}
	return nil
}

// ListByPaper returns all methods for a paper, ordered by id ASC.
func (r *MethodRepo) ListByPaper(paperID string) ([]Method, error) {
	rows, err := r.db.Query(
		`SELECT id, paper_id, method_name, description, method_type, source_page, source_text
		FROM methods WHERE paper_id = ? ORDER BY id ASC`, paperID,
	)
	if err != nil {
		return nil, fmt.Errorf("list methods: %w", err)
	}
	defer rows.Close()

	var list []Method
	for rows.Next() {
		var m Method
		if err := rows.Scan(&m.ID, &m.PaperID, &m.MethodName, &m.Description, &m.MethodType, &m.SourcePage, &m.SourceText); err != nil {
			return nil, fmt.Errorf("scan method: %w", err)
		}
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate methods: %w", err)
	}
	return list, nil
}

// GetByID returns a single method, scoped to its parent paper.
// Returns ErrNotFound if no row matches either the id or the paper scope.
func (r *MethodRepo) GetByID(paperID, methodID string) (*Method, error) {
	row := r.db.QueryRow(
		`SELECT id, paper_id, method_name, description, method_type, source_page, source_text
		FROM methods WHERE id = ? AND paper_id = ?`, methodID, paperID,
	)
	var m Method
	err := row.Scan(&m.ID, &m.PaperID, &m.MethodName, &m.Description, &m.MethodType, &m.SourcePage, &m.SourceText)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get method: %w", err)
	}
	return &m, nil
}
