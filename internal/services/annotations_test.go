package services

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"paperviz/internal/repository"
)

// openTestDB opens an in-memory SQLite database with all migrations applied.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	migrations := make(map[int]string)
	for v, file := range map[int]string{
		1:  "001_init.sql",
		2:  "002_users.sql",
		3:  "003_chapters.sql",
		4:  "004_chapter_charts.sql",
		5:  "005_evidence.sql",
		6:  "006_document_title.sql",
		7:  "007_saved_papers.sql",
		8:  "008_research_collections.sql",
		9:  "009_share_tokens.sql",
		10: "010_document_share.sql",
		11: "011_share_referrals.sql",
		12: "012_usage_analytics.sql",
		13: "013_usage_tiers.sql",
		14: "014_structured_research_objects.sql",
		15: "015_evidence_graph.sql",
		16: "016_annotations.sql",
	} {
		sqlStr, err := repository.ReadMigration(filepath.Join("..", "..", "migrations", file))
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}
		migrations[v] = sqlStr
	}
	db, err := repository.Open(":memory:", migrations)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func insertAnnotationTestDoc(t *testing.T, db *sql.DB, docID, userID string) {
	t.Helper()
	userRepo := repository.NewUserRepo(db)
	if err := userRepo.Insert(repository.User{ID: userID, Email: userID + "@test.com", PasswordHash: "hash", CreatedAt: 1}); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	docRepo := repository.NewDocumentRepo(db)
	if err := docRepo.Insert(repository.Document{
		ID:             docID,
		CreatedAt:      time.Now().Unix(),
		LastAccessedAt: time.Now().Unix(),
		Status:         repository.StatusComplete,
		SourceType:     repository.SourceTypePDF,
		ReadingLevel:   repository.ReadingLevelSimplified,
		Title:          "Test Paper",
		OriginalText:   "original text",
		UserID:         &userID,
	}); err != nil {
		t.Fatalf("insert document: %v", err)
	}
}

func TestCreateAnnotation(t *testing.T) {
	// Create annotation and verify ID is returned with no error.
	db := openTestDB(t)
	userID := "user-ann-1"
	docID := "doc-ann-1"
	insertAnnotationTestDoc(t, db, docID, userID)

	id, err := CreateAnnotation(db, userID, docID, "paper", docID, "Great paper!")
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty annotation ID")
	}
}

func TestCreateAnnotation_InvalidType(t *testing.T) {
	// Create annotation with invalid targetType and verify error.
	db := openTestDB(t)
	userID := "user-ann-2"
	docID := "doc-ann-2"
	insertAnnotationTestDoc(t, db, docID, userID)

	_, err := CreateAnnotation(db, userID, docID, "invalid_type", docID, "content")
	if err == nil {
		t.Fatal("expected error for invalid targetType, got nil")
	}
}

func TestGetAnnotation(t *testing.T) {
	// Create annotation then retrieve it, verifying all fields match.
	db := openTestDB(t)
	userID := "user-ann-3"
	docID := "doc-ann-3"
	insertAnnotationTestDoc(t, db, docID, userID)

	id, err := CreateAnnotation(db, userID, docID, "claim", "claim-1", "Interesting claim")
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}

	got, err := GetAnnotation(db, id)
	if err != nil {
		t.Fatalf("get annotation: %v", err)
	}
	if got.ID != id {
		t.Errorf("ID = %q, want %q", got.ID, id)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %q, want %q", got.UserID, userID)
	}
	if got.DocumentID != docID {
		t.Errorf("DocumentID = %q, want %q", got.DocumentID, docID)
	}
	if got.TargetType != "claim" {
		t.Errorf("TargetType = %q, want %q", got.TargetType, "claim")
	}
	if got.TargetID != "claim-1" {
		t.Errorf("TargetID = %q, want %q", got.TargetID, "claim-1")
	}
	if got.Content != "Interesting claim" {
		t.Errorf("Content = %q, want %q", got.Content, "Interesting claim")
	}
}

func TestListAnnotations(t *testing.T) {
	// Create 3 annotations for same doc+user, verify list contains all.
	db := openTestDB(t)
	userID := "user-ann-4"
	docID := "doc-ann-4"
	insertAnnotationTestDoc(t, db, docID, userID)

	for i := 0; i < 3; i++ {
		_, err := CreateAnnotation(db, userID, docID, "paper", docID, "note")
		if err != nil {
			t.Fatalf("create annotation %d: %v", i, err)
		}
	}

	annotations, err := ListAnnotations(db, docID, userID)
	if err != nil {
		t.Fatalf("list annotations: %v", err)
	}
	if len(annotations) != 3 {
		t.Fatalf("got %d annotations, want 3", len(annotations))
	}
}

func TestListAnnotations_OtherUser(t *testing.T) {
	// Create annotation for different user, verify empty list for first user.
	db := openTestDB(t)
	userA := "user-ann-5a"
	userB := "user-ann-5b"
	docID := "doc-ann-5"
	insertAnnotationTestDoc(t, db, docID, userA)

	_, err := CreateAnnotation(db, userB, docID, "paper", docID, "user B note")
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}

	annotations, err := ListAnnotations(db, docID, userA)
	if err != nil {
		t.Fatalf("list annotations: %v", err)
	}
	if len(annotations) != 0 {
		t.Fatalf("got %d annotations for userA, want 0", len(annotations))
	}
}

func TestUpdateAnnotation(t *testing.T) {
	// Create annotation then update content, verify change is reflected.
	db := openTestDB(t)
	userID := "user-ann-6"
	docID := "doc-ann-6"
	insertAnnotationTestDoc(t, db, docID, userID)

	id, err := CreateAnnotation(db, userID, docID, "paper", docID, "original content")
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}

	err = UpdateAnnotation(db, id, userID, "updated content")
	if err != nil {
		t.Fatalf("update annotation: %v", err)
	}

	got, err := GetAnnotation(db, id)
	if err != nil {
		t.Fatalf("get annotation: %v", err)
	}
	if got.Content != "updated content" {
		t.Errorf("Content = %q, want %q", got.Content, "updated content")
	}
}

func TestUpdateAnnotation_WrongUser(t *testing.T) {
	// Try to update another user's annotation, verify error.
	db := openTestDB(t)
	userA := "user-ann-7a"
	userB := "user-ann-7b"
	docID := "doc-ann-7"
	insertAnnotationTestDoc(t, db, docID, userA)

	id, err := CreateAnnotation(db, userA, docID, "paper", docID, "user A note")
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}

	err = UpdateAnnotation(db, id, userB, "hijacked content")
	if err == nil {
		t.Fatal("expected unauthorized error, got nil")
	}
}

func TestDeleteAnnotation(t *testing.T) {
	// Create annotation then delete it, verify it's gone.
	db := openTestDB(t)
	userID := "user-ann-8"
	docID := "doc-ann-8"
	insertAnnotationTestDoc(t, db, docID, userID)

	id, err := CreateAnnotation(db, userID, docID, "paper", docID, "delete me")
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}

	err = DeleteAnnotation(db, id, userID)
	if err != nil {
		t.Fatalf("delete annotation: %v", err)
	}

	_, err = GetAnnotation(db, id)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}
