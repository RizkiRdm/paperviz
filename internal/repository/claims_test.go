package repository

import (
	"testing"
)

func TestClaimRepo(t *testing.T) {
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
	repo := NewClaimRepo(db)

	t.Run("insert and list", func(t *testing.T) {
		rows := []Claim{
			{ID: "cl-1", PaperID: docID, ClaimText: "claim one", ClaimType: "finding", Confidence: "high", SourcePage: &page, SourceText: stringPtr("source text one"), CreatedAt: 1},
			{ID: "cl-2", PaperID: docID, ClaimText: "claim two", ClaimType: "conclusion", Confidence: "medium", SourcePage: &page, SourceText: stringPtr("source text two"), CreatedAt: 2},
		}
		for _, r := range rows {
			if err := repo.Insert(r); err != nil {
				t.Fatalf("insert claim: %v", err)
			}
		}

		list, err := repo.ListByPaper(docID)
		if err != nil {
			t.Fatalf("list claims: %v", err)
		}
		if len(list) != len(rows) {
			t.Fatalf("got %d rows, want %d", len(list), len(rows))
		}
		if list[0].ClaimText != rows[0].ClaimText || list[0].ClaimType != rows[0].ClaimType {
			t.Errorf("got %+v, want claim_text=%q claim_type=%q", list[0], rows[0].ClaimText, rows[0].ClaimType)
		}
		if list[0].SourcePage == nil || *list[0].SourcePage != page {
			t.Errorf("got source_page %v, want %d", list[0].SourcePage, page)
		}
		if list[0].SourceText == nil || *list[0].SourceText != "source text one" {
			t.Errorf("got source_text %v, want %q", list[0].SourceText, "source text one")
		}
	})

	t.Run("list unknown paper returns empty", func(t *testing.T) {
		list, err := repo.ListByPaper("no-such-paper")
		if err != nil {
			t.Fatalf("list claims: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d rows, want 0", len(list))
		}
	})

	t.Run("insert duplicate id returns error", func(t *testing.T) {
		if err := repo.Insert(Claim{ID: "cl-1", PaperID: docID, ClaimText: "dup", ClaimType: "finding", Confidence: "high", CreatedAt: 3}); err == nil {
			t.Errorf("expected unique constraint error, got nil")
		}
	})

	t.Run("get by id success", func(t *testing.T) {
		got, err := repo.GetByID(docID, "cl-1")
		if err != nil {
			t.Fatalf("get claim: %v", err)
		}
		if got.ClaimText != "claim one" {
			t.Errorf("got claim_text=%q, want %q", got.ClaimText, "claim one")
		}
		if got.ClaimType != "finding" {
			t.Errorf("got claim_type=%q, want %q", got.ClaimType, "finding")
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		_, err := repo.GetByID(docID, "no-such-claim")
		if err != ErrNotFound {
			t.Errorf("got err=%v, want ErrNotFound", err)
		}
	})
}

func stringPtr(s string) *string {
	return &s
}
