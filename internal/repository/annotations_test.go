package repository

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func openAnnotationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db := openTestDB(t)
	annotationsSQL, err := ReadMigration(filepath.Join("..", "..", "migrations", "016_annotations.sql"))
	if err != nil {
		t.Fatalf("read annotations migration: %v", err)
	}
	if _, err := db.Exec(annotationsSQL); err != nil {
		t.Fatalf("run annotations migration: %v", err)
	}
	return db
}

// TestAnnotationRepo_Insert verifies that a single annotation can be inserted without error.
func TestAnnotationRepo_Insert(t *testing.T) {
	db := openAnnotationTestDB(t)
	repo := NewAnnotationRepo(db)

	a := Annotation{
		ID:         "ann-1",
		UserID:     "user-1",
		DocumentID: "doc-1",
		TargetType: "paper",
		TargetID:   "doc-1",
		Content:    "This paper is interesting",
		CreatedAt:  1000,
		UpdatedAt:  1000,
	}

	if err := repo.Insert(a); err != nil {
		t.Fatalf("Insert: unexpected error: %v", err)
	}
}

// TestAnnotationRepo_GetByID verifies that inserting then retrieving by ID returns matching fields.
func TestAnnotationRepo_GetByID(t *testing.T) {
	db := openAnnotationTestDB(t)
	repo := NewAnnotationRepo(db)

	want := Annotation{
		ID:         "ann-1",
		UserID:     "user-1",
		DocumentID: "doc-1",
		TargetType: "claim",
		TargetID:   "claim-1",
		Content:    "This claim is supported",
		CreatedAt:  1000,
		UpdatedAt:  1000,
	}

	if err := repo.Insert(want); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	got, err := repo.GetByID("ann-1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
	if got.UserID != want.UserID {
		t.Errorf("UserID = %q, want %q", got.UserID, want.UserID)
	}
	if got.DocumentID != want.DocumentID {
		t.Errorf("DocumentID = %q, want %q", got.DocumentID, want.DocumentID)
	}
	if got.TargetType != want.TargetType {
		t.Errorf("TargetType = %q, want %q", got.TargetType, want.TargetType)
	}
	if got.TargetID != want.TargetID {
		t.Errorf("TargetID = %q, want %q", got.TargetID, want.TargetID)
	}
	if got.Content != want.Content {
		t.Errorf("Content = %q, want %q", got.Content, want.Content)
	}
	if got.CreatedAt != want.CreatedAt {
		t.Errorf("CreatedAt = %d, want %d", got.CreatedAt, want.CreatedAt)
	}
	if got.UpdatedAt != want.UpdatedAt {
		t.Errorf("UpdatedAt = %d, want %d", got.UpdatedAt, want.UpdatedAt)
	}
}

// TestAnnotationRepo_GetByID_NotFound verifies that a nonexistent ID returns ErrNotFound.
func TestAnnotationRepo_GetByID_NotFound(t *testing.T) {
	db := openAnnotationTestDB(t)
	repo := NewAnnotationRepo(db)

	_, err := repo.GetByID("nonexistent-id")
	if err != ErrNotFound {
		t.Fatalf("GetByID nonexistent: got %v, want ErrNotFound", err)
	}
}

// TestAnnotationRepo_ListByDocument verifies that ListByDocument returns all annotations for a given document.
func TestAnnotationRepo_ListByDocument(t *testing.T) {
	db := openAnnotationTestDB(t)
	repo := NewAnnotationRepo(db)

	annotations := []Annotation{
		{ID: "ann-1", UserID: "user-1", DocumentID: "doc-1", TargetType: "paper", TargetID: "doc-1", Content: "Note 1", CreatedAt: 1000, UpdatedAt: 1000},
		{ID: "ann-2", UserID: "user-1", DocumentID: "doc-1", TargetType: "claim", TargetID: "claim-1", Content: "Note 2", CreatedAt: 1001, UpdatedAt: 1001},
		{ID: "ann-3", UserID: "user-2", DocumentID: "doc-1", TargetType: "paper", TargetID: "doc-1", Content: "Note 3", CreatedAt: 1002, UpdatedAt: 1002},
	}

	for _, a := range annotations {
		if err := repo.Insert(a); err != nil {
			t.Fatalf("Insert %s: %v", a.ID, err)
		}
	}

	got, err := repo.ListByDocument("doc-1")
	if err != nil {
		t.Fatalf("ListByDocument: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("ListByDocument: got %d annotations, want 3", len(got))
	}
}

// TestAnnotationRepo_ListByDocumentAndUser verifies that ListByDocumentAndUser filters by both document and user.
func TestAnnotationRepo_ListByDocumentAndUser(t *testing.T) {
	db := openAnnotationTestDB(t)
	repo := NewAnnotationRepo(db)

	annotations := []Annotation{
		{ID: "ann-1", UserID: "user-1", DocumentID: "doc-1", TargetType: "paper", TargetID: "doc-1", Content: "User 1 note", CreatedAt: 1000, UpdatedAt: 1000},
		{ID: "ann-2", UserID: "user-2", DocumentID: "doc-1", TargetType: "paper", TargetID: "doc-1", Content: "User 2 note", CreatedAt: 1001, UpdatedAt: 1001},
		{ID: "ann-3", UserID: "user-1", DocumentID: "doc-1", TargetType: "claim", TargetID: "claim-1", Content: "User 1 claim note", CreatedAt: 1002, UpdatedAt: 1002},
	}

	for _, a := range annotations {
		if err := repo.Insert(a); err != nil {
			t.Fatalf("Insert %s: %v", a.ID, err)
		}
	}

	got, err := repo.ListByDocumentAndUser("doc-1", "user-1")
	if err != nil {
		t.Fatalf("ListByDocumentAndUser: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListByDocumentAndUser: got %d annotations, want 2", len(got))
	}
	for _, a := range got {
		if a.UserID != "user-1" {
			t.Errorf("annotation %s has UserID %q, want user-1", a.ID, a.UserID)
		}
	}
}

// TestAnnotationRepo_Update verifies that Update changes the content and updated_at of an annotation.
func TestAnnotationRepo_Update(t *testing.T) {
	db := openAnnotationTestDB(t)
	repo := NewAnnotationRepo(db)

	a := Annotation{
		ID:         "ann-1",
		UserID:     "user-1",
		DocumentID: "doc-1",
		TargetType: "paper",
		TargetID:   "doc-1",
		Content:    "Original content",
		CreatedAt:  1000,
		UpdatedAt:  1000,
	}
	if err := repo.Insert(a); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	if err := repo.Update("ann-1", "Updated content"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID("ann-1")
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Content != "Updated content" {
		t.Errorf("Content = %q, want %q", got.Content, "Updated content")
	}
	if got.UpdatedAt <= 1000 {
		t.Errorf("UpdatedAt = %d, want > 1000", got.UpdatedAt)
	}
}

// TestAnnotationRepo_Delete verifies that Delete removes an annotation so GetByID returns ErrNotFound.
func TestAnnotationRepo_Delete(t *testing.T) {
	db := openAnnotationTestDB(t)
	repo := NewAnnotationRepo(db)

	a := Annotation{
		ID:         "ann-1",
		UserID:     "user-1",
		DocumentID: "doc-1",
		TargetType: "paper",
		TargetID:   "doc-1",
		Content:    "Delete me",
		CreatedAt:  1000,
		UpdatedAt:  1000,
	}
	if err := repo.Insert(a); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	if err := repo.Delete("ann-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID("ann-1")
	if err != ErrNotFound {
		t.Fatalf("GetByID after delete: got %v, want ErrNotFound", err)
	}
}
