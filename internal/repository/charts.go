package repository

import (
	"database/sql"
	"fmt"
)

// ChartRepo provides CRUD access to the charts table. Like DocumentRepo,
// it holds no business logic — every method is a parameterized SQL
// operation, and callers (the pipeline, via handlers) decide what data to
// pass in.
type ChartRepo struct {
	db dbExecutor
}

// dbExecutor is satisfied by both *sql.DB and *sql.Tx, letting repository
// methods run either standalone or inside the pipeline's single write
// transaction (ARCHITECTURE.md Section 4 Transaction Policy) without
// duplicating each method for both cases.
type dbExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func NewChartRepo(db dbExecutor) *ChartRepo {
	return &ChartRepo{db: db}
}

// Insert writes one chart row, linked to its parent document.
func (r *ChartRepo) Insert(c Chart) error {
	_, err := r.db.Exec(
		`INSERT INTO charts (id, document_id, source_method, chart_data, image_blob, annotation, page_number, display_order, chapter_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.DocumentID, c.SourceMethod, c.ChartData, c.ImageBlob, c.Annotation, c.PageNumber, c.DisplayOrder, c.ChapterID,
	)
	if err != nil {
		return fmt.Errorf("insert chart: %w", err)
	}
	return nil
}

// ListByDocument returns all charts for a document, ordered for display.
func (r *ChartRepo) ListByDocument(documentID string) ([]Chart, error) {
	rows, err := r.db.Query(
		`SELECT id, document_id, source_method, chart_data, image_blob, annotation, page_number, display_order, chapter_id
		FROM charts WHERE document_id = ? ORDER BY display_order ASC`, documentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list charts: %w", err)
	}
	defer rows.Close()

	var charts []Chart
	for rows.Next() {
		var c Chart
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.SourceMethod, &c.ChartData, &c.ImageBlob, &c.Annotation, &c.PageNumber, &c.DisplayOrder, &c.ChapterID); err != nil {
			return nil, fmt.Errorf("scan chart: %w", err)
		}
		charts = append(charts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate charts: %w", err)
	}
	return charts, nil
}
