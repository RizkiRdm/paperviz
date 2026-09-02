package services

import (
	"database/sql"
	"fmt"
	"time"

	"paperviz/internal/repository"
)

// CreateAnnotation creates a new annotation on a paper or claim.
func CreateAnnotation(db *sql.DB, userID, documentID, targetType, targetID, content string) (string, error) {
	if targetType != "paper" && targetType != "claim" {
		return "", fmt.Errorf("invalid targetType %q: must be \"paper\" or \"claim\"", targetType)
	}
	id, err := repository.NewID()
	if err != nil {
		return "", fmt.Errorf("generate annotation id: %w", err)
	}
	now := time.Now().Unix()
	repo := repository.NewAnnotationRepo(db)
	err = repo.Insert(repository.Annotation{
		ID:         id,
		UserID:     userID,
		DocumentID: documentID,
		TargetType: targetType,
		TargetID:   targetID,
		Content:    content,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetAnnotation returns an annotation by ID.
func GetAnnotation(db *sql.DB, id string) (*repository.Annotation, error) {
	repo := repository.NewAnnotationRepo(db)
	return repo.GetByID(id)
}

// ListAnnotations returns annotations for a document owned by a user.
func ListAnnotations(db *sql.DB, documentID, userID string) ([]repository.Annotation, error) {
	repo := repository.NewAnnotationRepo(db)
	return repo.ListByDocumentAndUser(documentID, userID)
}

// UpdateAnnotation updates annotation content after verifying ownership.
func UpdateAnnotation(db *sql.DB, id, userID, content string) error {
	repo := repository.NewAnnotationRepo(db)
	existing, err := repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return fmt.Errorf("unauthorized: user %s does not own annotation %s", userID, id)
	}
	return repo.Update(id, content)
}

// DeleteAnnotation removes an annotation after verifying ownership.
func DeleteAnnotation(db *sql.DB, id, userID string) error {
	repo := repository.NewAnnotationRepo(db)
	existing, err := repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return fmt.Errorf("unauthorized: user %s does not own annotation %s", userID, id)
	}
	return repo.Delete(id)
}
