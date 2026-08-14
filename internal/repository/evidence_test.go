package repository

import (
	"testing"
)

func TestEvidenceRepo(t *testing.T) {
	db := openTestDB(t)

	docID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	docRepo := NewDocumentRepo(db)
	if err := docRepo.Insert(Document{ID: docID, CreatedAt: 1, LastAccessedAt: 1, Status: StatusComplete, SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text"}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	page := 3
	figID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	repo := NewEvidenceRepo(db)

	t.Run("insert and list", func(t *testing.T) {
		rows := []Evidence{
			{ID: "ev-1", PaperID: docID, Page: &page, FigureID: &figID, SourceText: "figure one context", SourceReference: "Figure on page 3"},
			{ID: "ev-2", PaperID: docID, Page: &page, FigureID: &figID, SourceText: "figure two context", SourceReference: "Figure on page 3"},
		}
		for _, r := range rows {
			if err := repo.Insert(r); err != nil {
				t.Fatalf("insert evidence: %v", err)
			}
		}

		list, err := repo.ListByPaper(docID)
		if err != nil {
			t.Fatalf("list evidence: %v", err)
		}
		if len(list) != len(rows) {
			t.Fatalf("got %d rows, want %d", len(list), len(rows))
		}
		if list[0].SourceText != rows[0].SourceText || list[0].SourceReference != rows[0].SourceReference {
			t.Errorf("got %+v, want source_text=%q reference=%q", list[0], rows[0].SourceText, rows[0].SourceReference)
		}
		if list[0].Page == nil || *list[0].Page != page {
			t.Errorf("got page %v, want %d", list[0].Page, page)
		}
		if list[0].FigureID == nil || *list[0].FigureID != figID {
			t.Errorf("got figure_id %v, want %s", list[0].FigureID, figID)
		}
	})

	t.Run("list unknown paper returns empty", func(t *testing.T) {
		list, err := repo.ListByPaper("no-such-paper")
		if err != nil {
			t.Fatalf("list evidence: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d rows, want 0", len(list))
		}
	})

	t.Run("insert with unknown paper returns error", func(t *testing.T) {
		if err := repo.Insert(Evidence{ID: "ev-bad", PaperID: "no-such-doc", SourceText: "text", SourceReference: "ref"}); err == nil {
			t.Errorf("expected foreign key error, got nil")
		}
	})
}
