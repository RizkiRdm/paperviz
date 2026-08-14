package services

import (
	"path/filepath"
	"testing"

	"paperviz/internal/repository"
)

func TestValidateAndInsert(t *testing.T) {
	migrations := make(map[int]string)
	for v, file := range map[int]string{
		1: "001_init.sql",
		2: "002_users.sql",
		3: "003_chapters.sql",
		4: "004_chapter_charts.sql",
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
	defer db.Close()

	result, errCode, err := ValidateAndInsert(db, repository.ReadingLevelSimplified, false, nil, "Sample pasted research text", nil)
	if err != nil {
		t.Fatalf("ValidateAndInsert failed: %v (errCode: %s)", err, errCode)
	}

	if result.DocumentID == "" {
		t.Errorf("expected non-empty document ID")
	}
	if result.SourceType != repository.SourceTypePastedText {
		t.Errorf("expected source_type pasted_text, got %s", result.SourceType)
	}
	if result.OriginalText != "Sample pasted research text" {
		t.Errorf("expected original text preserved")
	}

	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.Get(result.DocumentID)
	if err != nil {
		t.Fatalf("get inserted document: %v", err)
	}
	if doc.Status != repository.StatusProcessing {
		t.Errorf("expected status 'processing', got %s", doc.Status)
	}
}
