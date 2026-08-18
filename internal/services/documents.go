package services

import (
	"database/sql"

	"paperviz/internal/repository"
)

// ToggleDocumentSaved sets or unsets the saved flag on a document.
func ToggleDocumentSaved(db *sql.DB, id string, saved bool) error {
	repo := repository.NewDocumentRepo(db)
	return repo.ToggleSaved(id, saved)
}

// RenameDocument sets a custom title on a document.
func RenameDocument(db *sql.DB, id, title string) error {
	repo := repository.NewDocumentRepo(db)
	return repo.UpdateTitle(id, title)
}

// DeleteDocument hard-deletes a document and all related data (cascade).
func DeleteDocument(db *sql.DB, id string) error {
	repo := repository.NewDocumentRepo(db)
	return repo.DeleteDocument(id)
}
