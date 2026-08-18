package services

import (
	"database/sql"

	"paperviz/internal/repository"
)

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

// GetCollection returns a collection by ID.
func GetCollection(db *sql.DB, id string) (*repository.Collection, error) {
	repo := repository.NewCollectionRepo(db)
	return repo.Get(id)
}

// RenameCollection sets a new name on a collection.
func RenameCollection(db *sql.DB, id, name string) error {
	repo := repository.NewCollectionRepo(db)
	return repo.UpdateName(id, name)
}

// DeleteCollection removes a collection.
func DeleteCollection(db *sql.DB, id string) error {
	repo := repository.NewCollectionRepo(db)
	return repo.Delete(id)
}

// AddDocumentToCollection adds a document to a collection.
func AddDocumentToCollection(db *sql.DB, collectionID, documentID string) error {
	repo := repository.NewCollectionRepo(db)
	return repo.AddDocument(collectionID, documentID)
}

// RemoveDocumentFromCollection removes a document from a collection.
func RemoveDocumentFromCollection(db *sql.DB, collectionID, documentID string) error {
	repo := repository.NewCollectionRepo(db)
	return repo.RemoveDocument(collectionID, documentID)
}

// ListCollectionDocuments returns documents in a collection.
func ListCollectionDocuments(db *sql.DB, collectionID string) ([]repository.DocumentListItem, error) {
	repo := repository.NewCollectionRepo(db)
	return repo.ListDocuments(collectionID)
}
