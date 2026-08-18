package repository

import (
	"testing"
)

func TestCollectionCRUD(t *testing.T) {
	db := openTestDB(t)
	repo := NewCollectionRepo(db)
	userRepo := NewUserRepo(db)

	if err := userRepo.Insert(User{ID: "user-1", Email: "test@example.com", PasswordHash: "hash", CreatedAt: 1000}); err != nil {
		t.Fatal(err)
	}

	// Create
	col := Collection{
		ID:        "col-1",
		UserID:    "user-1",
		Name:      "Test Collection",
		CreatedAt: 1000,
	}
	if err := repo.Insert(col); err != nil {
		t.Fatal(err)
	}

	// Get
	got, err := repo.Get("col-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Test Collection" {
		t.Errorf("expected name='Test Collection', got %q", got.Name)
	}

	// Update name
	if err := repo.UpdateName("col-1", "Renamed"); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.Get("col-1")
	if got.Name != "Renamed" {
		t.Errorf("expected name='Renamed', got %q", got.Name)
	}

	// List by user
	items, err := repo.ListByUser("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 collection, got %d", len(items))
	}

	// Delete
	if err := repo.Delete("col-1"); err != nil {
		t.Fatal(err)
	}
	_, err = repo.Get("col-1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestCollectionDocuments(t *testing.T) {
	db := openTestDB(t)
	colRepo := NewCollectionRepo(db)
	docRepo := NewDocumentRepo(db)
	userRepo := NewUserRepo(db)

	if err := userRepo.Insert(User{ID: "user-1", Email: "test@example.com", PasswordHash: "hash", CreatedAt: 1000}); err != nil {
		t.Fatal(err)
	}

	// Insert document
	doc := Document{
		ID: "doc-1", CreatedAt: 1000, LastAccessedAt: 1000,
		Status: StatusComplete, SourceType: SourceTypePDF,
		ReadingLevel: ReadingLevelSimplified, Title: "Test Doc",
		OriginalText: "original",
	}
	if err := docRepo.Insert(doc); err != nil {
		t.Fatal(err)
	}

	// Create collection
	col := Collection{
		ID: "col-1", UserID: "user-1", Name: "Test", CreatedAt: 1000,
	}
	if err := colRepo.Insert(col); err != nil {
		t.Fatal(err)
	}

	// Add document to collection
	if err := colRepo.AddDocument("col-1", "doc-1"); err != nil {
		t.Fatal(err)
	}

	// List documents in collection
	docs, err := colRepo.ListDocuments("col-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}

	// Remove document from collection
	if err := colRepo.RemoveDocument("col-1", "doc-1"); err != nil {
		t.Fatal(err)
	}
	docs, _ = colRepo.ListDocuments("col-1")
	if len(docs) != 0 {
		t.Errorf("expected 0 documents after remove, got %d", len(docs))
	}
}
