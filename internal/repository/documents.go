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
			(id, created_at, last_accessed_at, status, source_type, reading_level, title, original_text, simplified_text, error_message, chart_extraction_degraded, processing_stage, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.CreatedAt, d.LastAccessedAt, d.Status, d.SourceType, d.ReadingLevel, d.Title, d.OriginalText, d.SimplifiedText, d.ErrorMessage, d.ChartExtractionDegraded, d.ProcessingStage, d.UserID,
	)
	if err != nil {
		return fmt.Errorf("insert document: %w", err)
	}
	return nil
}

// Get retrieves a document by ID. Returns ErrNotFound if no row matches.
func (r *DocumentRepo) Get(id string) (*Document, error) {
	row := r.db.QueryRow(
		`SELECT id, created_at, last_accessed_at, status, source_type, reading_level, title, original_text, simplified_text, error_message, chart_extraction_degraded, processing_stage, user_id, saved
		FROM documents WHERE id = ?`, id,
	)
	var d Document
	err := row.Scan(&d.ID, &d.CreatedAt, &d.LastAccessedAt, &d.Status, &d.SourceType, &d.ReadingLevel, &d.Title, &d.OriginalText, &d.SimplifiedText, &d.ErrorMessage, &d.ChartExtractionDegraded, &d.ProcessingStage, &d.UserID, &d.Saved)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	return &d, nil
}

// UpdateStatus sets status/simplified_text/error_message/chart_extraction_degraded for a document.
func (r *DocumentRepo) UpdateStatus(id, status string, simplifiedText, errorMessage *string, chartExtractionDegraded bool) error {
	_, err := r.db.Exec(
		`UPDATE documents SET status = ?, simplified_text = ?, error_message = ?, chart_extraction_degraded = ?, processing_stage = NULL WHERE id = ?`,
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

func (r *DocumentRepo) UpdateStage(id string, stage *string) error {
	_, err := r.db.Exec(`UPDATE documents SET processing_stage = ? WHERE id = ?`, stage, id)
	if err != nil {
		return fmt.Errorf("update processing stage: %w", err)
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

// ListByUser returns documents belonging to a user, ordered by most recent, paginated.
func (r *DocumentRepo) ListByUser(userID string, limit, offset int) ([]Document, error) {
	rows, err := r.db.Query(
		`SELECT id, created_at, last_accessed_at, status, source_type, reading_level, title, original_text, simplified_text, error_message, chart_extraction_degraded, processing_stage, user_id
		FROM documents WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list documents by user: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.CreatedAt, &d.LastAccessedAt, &d.Status, &d.SourceType, &d.ReadingLevel, &d.Title, &d.OriginalText, &d.SimplifiedText, &d.ErrorMessage, &d.ChartExtractionDegraded, &d.ProcessingStage, &d.UserID); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate documents: %w", err)
	}
	return docs, nil
}

// ListSummariesByUser returns lightweight history rows for a user's documents.
func (r *DocumentRepo) ListSummariesByUser(userID string, limit, offset int) ([]DocumentListItem, error) {
	rows, err := r.db.Query(
		`SELECT d.id, d.title, d.created_at, d.status, d.saved,
		        substr(coalesce(d.simplified_text, ''), 1, 240) AS summary_preview,
		        (SELECT COUNT(*) FROM charts c WHERE c.document_id = d.id) AS chart_count,
		        (SELECT COUNT(*) FROM charts c WHERE c.document_id = d.id AND c.annotation IS NOT NULL AND c.annotation != '') AS explanation_count
		FROM documents d WHERE d.user_id = ? ORDER BY d.created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list document summaries by user: %w", err)
	}
	defer rows.Close()

	var items []DocumentListItem
	for rows.Next() {
		var it DocumentListItem
		if err := rows.Scan(&it.ID, &it.Title, &it.CreatedAt, &it.Status, &it.Saved, &it.SummaryPreview, &it.ChartCount, &it.ExplanationCount); err != nil {
			return nil, fmt.Errorf("scan document summary: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document summaries: %w", err)
	}
	return items, nil
}

// ToggleSaved sets the saved flag on a document.
func (r *DocumentRepo) ToggleSaved(id string, saved bool) error {
	savedInt := 0
	if saved {
		savedInt = 1
	}
	_, err := r.db.Exec(`UPDATE documents SET saved = ? WHERE id = ?`, savedInt, id)
	if err != nil {
		return fmt.Errorf("toggle saved: %w", err)
	}
	return nil
}

// UpdateTitle sets a custom title on a document.
func (r *DocumentRepo) UpdateTitle(id string, title string) error {
	_, err := r.db.Exec(`UPDATE documents SET title = ? WHERE id = ?`, title, id)
	if err != nil {
		return fmt.Errorf("update title: %w", err)
	}
	return nil
}

// DeleteDocument hard-deletes a document. Related rows (charts, claim_diffs,
// chapters, evidence) cascade via FK constraints.
func (r *DocumentRepo) DeleteDocument(id string) error {
	_, err := r.db.Exec(`DELETE FROM documents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	return nil
}

// CountByUser returns the total number of documents for a user.
func (r *DocumentRepo) CountByUser(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM documents WHERE user_id = ?`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}
	return count, nil
}

// CountSavedByUser returns the number of saved documents for a user.
func (r *DocumentRepo) CountSavedByUser(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM documents WHERE user_id = ? AND saved = 1`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count saved documents: %w", err)
	}
	return count, nil
}
