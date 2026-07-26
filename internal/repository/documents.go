package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrNotFound is returned when a document lookup finds no matching row.
var ErrNotFound = errors.New("document not found")

// DocumentRepo provides CRUD access to the documents table. It contains no
// business logic — every method is a parameterized SQL operation only, per
// ARCHITECTURE.md Section 2 cross-layer rules.
//
// db is typed as dbExecutor (see charts.go) rather than *sql.DB so the same
// repo type works both standalone (simple reads like Get/TouchLastAccessed)
// and inside a *sql.Tx (the pipeline's multi-table write — see
// handlers/documents.go's use of NewDocumentRepo(tx) during save).
type DocumentRepo struct {
	db dbExecutor
}

func NewDocumentRepo(db dbExecutor) *DocumentRepo {
	return &DocumentRepo{db: db}
}

// Insert writes a new document row. Callers pass all fields explicitly;
// the repository does not infer or default anything.
func (r *DocumentRepo) Insert(d Document) error {
	_, err := r.db.Exec(
		`INSERT INTO documents
			(id, created_at, last_accessed_at, status, source_type, reading_level, original_text, simplified_text, error_message, chart_extraction_degraded)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.CreatedAt, d.LastAccessedAt, d.Status, d.SourceType, d.ReadingLevel, d.OriginalText, d.SimplifiedText, d.ErrorMessage, d.ChartExtractionDegraded,
	)
	if err != nil {
		return fmt.Errorf("insert document: %w", err)
	}
	return nil
}

// Get retrieves a document by ID. Returns ErrNotFound if no row matches.
func (r *DocumentRepo) Get(id string) (*Document, error) {
	row := r.db.QueryRow(
		`SELECT id, created_at, last_accessed_at, status, source_type, reading_level, original_text, simplified_text, error_message, chart_extraction_degraded
		FROM documents WHERE id = ?`, id,
	)
	var d Document
	err := row.Scan(&d.ID, &d.CreatedAt, &d.LastAccessedAt, &d.Status, &d.SourceType, &d.ReadingLevel, &d.OriginalText, &d.SimplifiedText, &d.ErrorMessage, &d.ChartExtractionDegraded)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	return &d, nil
}

// UpdateStatus sets status/simplified_text/error_message/chart_extraction_degraded for a document.
// Used by the pipeline to record terminal states (complete/failed/verification_failed).
func (r *DocumentRepo) UpdateStatus(id, status string, simplifiedText, errorMessage *string, chartExtractionDegraded bool) error {
	_, err := r.db.Exec(
		`UPDATE documents SET status = ?, simplified_text = ?, error_message = ?, chart_extraction_degraded = ? WHERE id = ?`,
		status, simplifiedText, errorMessage, chartExtractionDegraded, id,
	)
	if err != nil {
		return fmt.Errorf("update document status: %w", err)
	}
	return nil
}

// TouchLastAccessed updates last_accessed_at, extending the document's
// 7-day expiry window. Called on every successful GET (ARCHITECTURE.md
// Acceptance Scenario 5).
func (r *DocumentRepo) TouchLastAccessed(id string, now int64) error {
	_, err := r.db.Exec(`UPDATE documents SET last_accessed_at = ? WHERE id = ?`, now, id)
	if err != nil {
		return fmt.Errorf("touch last_accessed_at: %w", err)
	}
	return nil
}

// DeleteExpiredBefore deletes all documents whose last_accessed_at is older
// than cutoff (unix seconds). Charts and claim_diffs cascade via FK.
// Returns the number of documents deleted.
func (r *DocumentRepo) DeleteExpiredBefore(cutoff int64) (int64, error) {
	res, err := r.db.Exec(`DELETE FROM documents WHERE last_accessed_at < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete expired documents: %w", err)
	}
	return res.RowsAffected()
}
