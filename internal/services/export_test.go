package services

import (
	"database/sql"
	"path/filepath"
	"testing"

	"paperviz/internal/repository"
)

// openServiceTestDB opens an in-memory SQLite database with all migrations applied.
func openServiceTestDB(t *testing.T) *sql.DB {
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

// insertExportTestDoc creates a complete document for export tests.
func insertExportTestDoc(t *testing.T, db *sql.DB, docID, userID string) {
	t.Helper()
	userRepo := repository.NewUserRepo(db)
	if err := userRepo.Insert(repository.User{ID: userID, Email: userID + "@test.com", PasswordHash: "hash", CreatedAt: 1}); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	docRepo := repository.NewDocumentRepo(db)
	if err := docRepo.Insert(repository.Document{
		ID:             docID,
		CreatedAt:      1000,
		LastAccessedAt: 1000,
		Status:         repository.StatusComplete,
		SourceType:     repository.SourceTypePDF,
		ReadingLevel:   repository.ReadingLevelSimplified,
		Title:          "Export Test Paper",
		OriginalText:   "original text",
		UserID:         &userID,
	}); err != nil {
		t.Fatalf("insert document: %v", err)
	}
}

func TestExportResearchContext_NotFound(t *testing.T) {
	// Export nonexistent document and verify error is returned.
	db := openServiceTestDB(t)

	_, err := ExportResearchContext(db, "no-such-doc")
	if err == nil {
		t.Fatal("expected error for nonexistent document, got nil")
	}
}

func TestExportResearchContext_Empty(t *testing.T) {
	// Create document with no related data, export and verify empty arrays.
	db := openServiceTestDB(t)
	docID := "doc-export-empty"
	userID := "user-export-empty"
	insertExportTestDoc(t, db, docID, userID)

	export, err := ExportResearchContext(db, docID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if export.Document.ID != docID {
		t.Errorf("Document.ID = %q, want %q", export.Document.ID, docID)
	}
	if export.Document.Title != "Export Test Paper" {
		t.Errorf("Document.Title = %q, want %q", export.Document.Title, "Export Test Paper")
	}
	if len(export.Claims) != 0 {
		t.Errorf("Claims len = %d, want 0", len(export.Claims))
	}
	if len(export.Evidence) != 0 {
		t.Errorf("Evidence len = %d, want 0", len(export.Evidence))
	}
	if len(export.Tables) != 0 {
		t.Errorf("Tables len = %d, want 0", len(export.Tables))
	}
	if len(export.Methods) != 0 {
		t.Errorf("Methods len = %d, want 0", len(export.Methods))
	}
	if len(export.Results) != 0 {
		t.Errorf("Results len = %d, want 0", len(export.Results))
	}
	if len(export.Citations) != 0 {
		t.Errorf("Citations len = %d, want 0", len(export.Citations))
	}
	if len(export.Relationships) != 0 {
		t.Errorf("Relationships len = %d, want 0", len(export.Relationships))
	}
	if len(export.Annotations) != 0 {
		t.Errorf("Annotations len = %d, want 0", len(export.Annotations))
	}
	if len(export.Collections) != 0 {
		t.Errorf("Collections len = %d, want 0", len(export.Collections))
	}
}

func TestExportResearchContext_WithData(t *testing.T) {
	// Create doc with claims, annotations, and other data, verify export includes all.
	db := openServiceTestDB(t)
	docID := "doc-export-full"
	userID := "user-export-full"
	insertExportTestDoc(t, db, docID, userID)

	// Insert a claim.
	claimRepo := repository.NewClaimRepo(db)
	page := 1
	if err := claimRepo.Insert(repository.Claim{
		ID:         "cl-export-1",
		PaperID:    docID,
		ClaimText:  "Key finding",
		ClaimType:  "finding",
		Confidence: "high",
		SourcePage: &page,
		CreatedAt:  1001,
	}); err != nil {
		t.Fatalf("insert claim: %v", err)
	}

	// Insert an annotation.
	annID, err := CreateAnnotation(db, userID, docID, "paper", docID, "Important paper")
	if err != nil {
		t.Fatalf("create annotation: %v", err)
	}

	// Insert a method.
	methodRepo := repository.NewMethodRepo(db)
	if err := methodRepo.Insert(repository.Method{
		ID:         "meth-export-1",
		PaperID:    docID,
		MethodName: "RCT",
		MethodType: "experimental",
	}); err != nil {
		t.Fatalf("insert method: %v", err)
	}

	// Insert a result.
	resultRepo := repository.NewResultRepo(db)
	if err := resultRepo.Insert(repository.Result{
		ID:         "res-export-1",
		PaperID:    docID,
		ResultText: "Significant improvement",
		ResultType: "primary",
	}); err != nil {
		t.Fatalf("insert result: %v", err)
	}

	// Insert a citation.
	citationRepo := repository.NewCitationRepo(db)
	year := 2023
	if err := citationRepo.Insert(repository.Citation{
		ID:      "cit-export-1",
		PaperID: docID,
		Authors: strPtr("Smith et al."),
		Title:   strPtr("Related Work"),
		Year:    &year,
	}); err != nil {
		t.Fatalf("insert citation: %v", err)
	}

	export, err := ExportResearchContext(db, docID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	if len(export.Claims) != 1 {
		t.Fatalf("Claims len = %d, want 1", len(export.Claims))
	}
	if export.Claims[0].ClaimText != "Key finding" {
		t.Errorf("Claims[0].ClaimText = %q, want %q", export.Claims[0].ClaimText, "Key finding")
	}

	if len(export.Annotations) != 1 {
		t.Fatalf("Annotations len = %d, want 1", len(export.Annotations))
	}
	if export.Annotations[0].ID != annID {
		t.Errorf("Annotations[0].ID = %q, want %q", export.Annotations[0].ID, annID)
	}
	if export.Annotations[0].Content != "Important paper" {
		t.Errorf("Annotations[0].Content = %q, want %q", export.Annotations[0].Content, "Important paper")
	}

	if len(export.Methods) != 1 {
		t.Fatalf("Methods len = %d, want 1", len(export.Methods))
	}
	if export.Methods[0].MethodName != "RCT" {
		t.Errorf("Methods[0].MethodName = %q, want %q", export.Methods[0].MethodName, "RCT")
	}

	if len(export.Results) != 1 {
		t.Fatalf("Results len = %d, want 1", len(export.Results))
	}
	if export.Results[0].ResultText != "Significant improvement" {
		t.Errorf("Results[0].ResultText = %q, want %q", export.Results[0].ResultText, "Significant improvement")
	}

	if len(export.Citations) != 1 {
		t.Fatalf("Citations len = %d, want 1", len(export.Citations))
	}
	if export.Citations[0].Title == nil || *export.Citations[0].Title != "Related Work" {
		t.Errorf("Citations[0].Title = %v, want %q", export.Citations[0].Title, "Related Work")
	}
}
