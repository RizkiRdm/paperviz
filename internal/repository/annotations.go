package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// Annotation represents a user annotation on a paper or claim.
type Annotation struct {
	ID         string
	UserID     string
	DocumentID string
	TargetType string // "paper" or "claim"
	TargetID   string
	Content    string
	CreatedAt  int64
	UpdatedAt  int64
}

// AnnotationRepo provides CRUD access to the annotations table.
type AnnotationRepo struct {
	db *sql.DB
}

// NewAnnotationRepo creates a new AnnotationRepo.
func NewAnnotationRepo(db *sql.DB) *AnnotationRepo {
	return &AnnotationRepo{db: db}
}

// Insert creates a new annotation.
func (r *AnnotationRepo) Insert(a Annotation) error {
	_, err := r.db.Exec(
		`INSERT INTO annotations (id, user_id, document_id, target_type, target_id, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.UserID, a.DocumentID, a.TargetType, a.TargetID, a.Content, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert annotation: %w", err)
	}
	return nil
}

// GetByID retrieves an annotation by ID. Returns ErrNotFound if no row matches.
func (r *AnnotationRepo) GetByID(id string) (*Annotation, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, document_id, target_type, target_id, content, created_at, updated_at FROM annotations WHERE id = ?`, id,
	)
	var a Annotation
	err := row.Scan(&a.ID, &a.UserID, &a.DocumentID, &a.TargetType, &a.TargetID, &a.Content, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get annotation by id: %w", err)
	}
	return &a, nil
}

// ListByDocument returns all annotations for a document, ordered by most recent.
func (r *AnnotationRepo) ListByDocument(documentID string) ([]Annotation, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, document_id, target_type, target_id, content, created_at, updated_at FROM annotations WHERE document_id = ? ORDER BY created_at DESC`,
		documentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list annotations by document: %w", err)
	}
	defer rows.Close()

	var annotations []Annotation
	for rows.Next() {
		var a Annotation
		if err := rows.Scan(&a.ID, &a.UserID, &a.DocumentID, &a.TargetType, &a.TargetID, &a.Content, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan annotation: %w", err)
		}
		annotations = append(annotations, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate annotations: %w", err)
	}
	return annotations, nil
}

// ListByDocumentAndUser returns annotations for a document owned by a specific user.
func (r *AnnotationRepo) ListByDocumentAndUser(documentID, userID string) ([]Annotation, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, document_id, target_type, target_id, content, created_at, updated_at FROM annotations WHERE document_id = ? AND user_id = ? ORDER BY created_at DESC`,
		documentID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list annotations by document and user: %w", err)
	}
	defer rows.Close()

	var annotations []Annotation
	for rows.Next() {
		var a Annotation
		if err := rows.Scan(&a.ID, &a.UserID, &a.DocumentID, &a.TargetType, &a.TargetID, &a.Content, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan annotation: %w", err)
		}
		annotations = append(annotations, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate annotations: %w", err)
	}
	return annotations, nil
}

// Update sets new content and updated_at on an annotation.
func (r *AnnotationRepo) Update(id, content string) error {
	now := time.Now().Unix()
	res, err := r.db.Exec(
		`UPDATE annotations SET content = ?, updated_at = ? WHERE id = ?`,
		content, now, id,
	)
	if err != nil {
		return fmt.Errorf("update annotation: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update annotation rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes an annotation by ID.
func (r *AnnotationRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM annotations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete annotation: %w", err)
	}
	return nil
}
