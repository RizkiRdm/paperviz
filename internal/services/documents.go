package services

import (
	"database/sql"
	"errors"

	"paperviz/internal/apperr"
	"paperviz/internal/repository"
)

// AuthorizeDocumentAccess returns apperr.Permission when a document is owned by
// a different user. A document with no owner was uploaded while signed out, so
// anyone holding its direct URL may read and manage it; that keeps the
// no-account journey intact. Callers pass the authenticated user ID, which is
// empty only on anonymous routes.
func AuthorizeDocumentAccess(doc *repository.Document, userID string) error {
	if doc.UserID == nil || *doc.UserID == "" || *doc.UserID == userID {
		return nil
	}
	return apperr.Wrap(apperr.Permission, "document %s is owned by another user", doc.ID)
}

// getOwnedDocument loads a document and enforces the ownership rule in one step.
func getOwnedDocument(db *sql.DB, id, userID string) (*repository.Document, error) {
	doc, err := repository.NewDocumentRepo(db).Get(id)
	if err != nil {
		return nil, err
	}
	if err := AuthorizeDocumentAccess(doc, userID); err != nil {
		return nil, err
	}
	return doc, nil
}

// ToggleDocumentSaved sets or unsets the saved flag on a document the caller owns.
func ToggleDocumentSaved(db *sql.DB, id, userID string, saved bool) error {
	repo := repository.NewDocumentRepo(db)
	if _, err := getOwnedDocument(db, id, userID); err != nil {
		return err
	}
	return repo.ToggleSaved(id, saved)
}

// RenameDocument sets a custom title on a document the caller owns.
func RenameDocument(db *sql.DB, id, userID, title string) error {
	repo := repository.NewDocumentRepo(db)
	if _, err := getOwnedDocument(db, id, userID); err != nil {
		return err
	}
	return repo.UpdateTitle(id, title)
}

// DeleteDocument hard-deletes a document the caller owns, plus all related
// data (cascade).
func DeleteDocument(db *sql.DB, id, userID string) error {
	repo := repository.NewDocumentRepo(db)
	if _, err := getOwnedDocument(db, id, userID); err != nil {
		return err
	}
	return repo.DeleteDocument(id)
}

// DocumentAccessAllowed reports whether a user may read a document. It backs
// GET /api/documents/:id and the chart image endpoint, which serve read models
// and blobs rather than going through the write services above.
func DocumentAccessAllowed(db *sql.DB, id, userID string) error {
	_, err := getOwnedDocument(db, id, userID)
	return err
}

// IsDocumentNotFound reports whether an error is the repository's missing-row
// sentinel, so handlers can answer 404 instead of 500.
func IsDocumentNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}
