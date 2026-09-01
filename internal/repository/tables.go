package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// PaperTableRepo provides CRUD access to the paper_tables table.
type PaperTableRepo struct {
	db dbExecutor
}

// NewPaperTableRepo returns a new PaperTableRepo backed by the given executor.
func NewPaperTableRepo(db dbExecutor) *PaperTableRepo {
	return &PaperTableRepo{db: db}
}

// Insert writes one paper_table row, linked to its parent document.
func (r *PaperTableRepo) Insert(t PaperTable) error {
	_, err := r.db.Exec(
		`INSERT INTO paper_tables (id, document_id, page_number, caption, headers, rows, source_text, display_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.DocumentID, t.PageNumber, t.Caption, t.Headers, t.Rows, t.SourceText, t.DisplayOrder,
	)
	if err != nil {
		return fmt.Errorf("insert paper_table: %w", err)
	}
	return nil
}

// ListByPaper returns all paper tables for a document, ordered for display.
func (r *PaperTableRepo) ListByPaper(documentID string) ([]PaperTable, error) {
	rows, err := r.db.Query(
		`SELECT id, document_id, page_number, caption, headers, rows, source_text, display_order
		FROM paper_tables WHERE document_id = ? ORDER BY display_order ASC`, documentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list paper_tables: %w", err)
	}
	defer rows.Close()

	var tables []PaperTable
	for rows.Next() {
		var t PaperTable
		if err := rows.Scan(&t.ID, &t.DocumentID, &t.PageNumber, &t.Caption, &t.Headers, &t.Rows, &t.SourceText, &t.DisplayOrder); err != nil {
			return nil, fmt.Errorf("scan paper_table: %w", err)
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate paper_tables: %w", err)
	}
	return tables, nil
}

// GetByID returns a single paper table, scoped to its parent document.
// Returns ErrNotFound if no row matches either the id or the document scope.
func (r *PaperTableRepo) GetByID(documentID, tableID string) (*PaperTable, error) {
	row := r.db.QueryRow(
		`SELECT id, document_id, page_number, caption, headers, rows, source_text, display_order
		FROM paper_tables WHERE id = ? AND document_id = ?`, tableID, documentID,
	)
	var t PaperTable
	err := row.Scan(&t.ID, &t.DocumentID, &t.PageNumber, &t.Caption, &t.Headers, &t.Rows, &t.SourceText, &t.DisplayOrder)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get paper_table: %w", err)
	}
	return &t, nil
}
