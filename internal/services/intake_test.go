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
		5: "005_evidence.sql",
		6: "006_document_title.sql",
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
	if doc.Title != "Sample pasted research text" {
		t.Errorf("expected title 'Sample pasted research text', got %q", doc.Title)
	}
}

func TestDeriveTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"first line picked", "Quantum Entanglement in\nBosonic Systems\nAbstract", "Quantum Entanglement in"},
		{"leading blank lines skipped", "\n\n\n  First real line  \nMore text", "First real line"},
		{"empty string fallback", "", "Untitled paper"},
		{"whitespace only fallback", "   \n  \n  ", "Untitled paper"},
		{"200-char truncation", "The quick brown fox jumps over the lazy dog while the sly red fox leaps across the tall brown fence under the bright blue sky and over the green meadow beyond the distant purple hills where the golden sun sets slowly in the evening light", "The quick brown fox jumps over the lazy dog while the sly red fox leaps across the tall brown fence under the bright blue sky and over the green meadow beyond the distant purple hills where the golden"},
		{"single line no newline", "Direct title only", "Direct title only"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveTitle(tt.input)
			if got != tt.want {
				t.Errorf("deriveTitle(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
