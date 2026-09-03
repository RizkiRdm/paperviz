package services

import (
	"database/sql"
	"errors"
	"fmt"

	"paperviz/internal/repository"
)

// ErrForbidden is returned when user does not own collection.
var ErrForbidden = errors.New("forbidden: collection not owned by user")

// CreateCollection creates a new collection for a user.
func CreateCollection(db *sql.DB, id, userID, name string, createdAt int64) error {
	repo := repository.NewCollectionRepo(db)
	return repo.Insert(repository.Collection{
		ID:        id,
		UserID:    userID,
		Name:      name,
		CreatedAt: createdAt,
	})
}

// ListCollections returns collections belonging to a user.
func ListCollections(db *sql.DB, userID string) ([]repository.CollectionListItem, error) {
	repo := repository.NewCollectionRepo(db)
	return repo.ListByUser(userID)
}

// GetCollection returns collection by ID after verifying ownership.
func GetCollection(db *sql.DB, id, userID string) (*repository.Collection, error) {
	repo := repository.NewCollectionRepo(db)
	c, err := repo.Get(id)
	if err != nil {
		return nil, err
	}
	if c.UserID != userID {
		return nil, fmt.Errorf("user %s does not own collection %s: %w", userID, id, ErrForbidden)
	}
	return c, nil
}

// RenameCollection sets new name on collection after verifying ownership.
func RenameCollection(db *sql.DB, id, userID, name string) error {
	repo := repository.NewCollectionRepo(db)
	c, err := repo.Get(id)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return fmt.Errorf("user %s does not own collection %s: %w", userID, id, ErrForbidden)
	}
	return repo.UpdateName(id, name)
}

// DeleteCollection removes collection after verifying ownership.
func DeleteCollection(db *sql.DB, id, userID string) error {
	repo := repository.NewCollectionRepo(db)
	c, err := repo.Get(id)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return fmt.Errorf("user %s does not own collection %s: %w", userID, id, ErrForbidden)
	}
	return repo.Delete(id)
}

// AddDocumentToCollection adds document to collection after verifying ownership.
func AddDocumentToCollection(db *sql.DB, collectionID, userID, documentID string) error {
	repo := repository.NewCollectionRepo(db)
	c, err := repo.Get(collectionID)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return fmt.Errorf("user %s does not own collection %s: %w", userID, collectionID, ErrForbidden)
	}
	return repo.AddDocument(collectionID, documentID)
}

// RemoveDocumentFromCollection removes document from collection after verifying ownership.
func RemoveDocumentFromCollection(db *sql.DB, collectionID, userID, documentID string) error {
	repo := repository.NewCollectionRepo(db)
	c, err := repo.Get(collectionID)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return fmt.Errorf("user %s does not own collection %s: %w", userID, collectionID, ErrForbidden)
	}
	return repo.RemoveDocument(collectionID, documentID)
}

// ListCollectionDocuments returns documents in collection after verifying ownership.
func ListCollectionDocuments(db *sql.DB, collectionID, userID string) ([]repository.DocumentListItem, error) {
	repo := repository.NewCollectionRepo(db)
	c, err := repo.Get(collectionID)
	if err != nil {
		return nil, err
	}
	if c.UserID != userID {
		return nil, fmt.Errorf("user %s does not own collection %s: %w", userID, collectionID, ErrForbidden)
	}
	return repo.ListDocuments(collectionID)
}
