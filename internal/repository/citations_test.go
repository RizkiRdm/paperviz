package repository

import (
	"testing"
)

func TestCitationRepo(t *testing.T) {
	db := openTestDB(t)

	docID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	docRepo := NewDocumentRepo(db)
	if err := docRepo.Insert(Document{ID: docID, CreatedAt: 1, LastAccessedAt: 1, Status: StatusComplete, SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text"}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	repo := NewCitationRepo(db)
	year := 2024
	citedID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	authors := "Smith, Jones"
	title := "A Prior Work"
	venue := "ICML 2024"
	doi := "10.1234/example"
	url := "https://arxiv.org/abs/1234"
	srcPage := 5

	citation := Citation{
		ID:           "cit-1",
		PaperID:      docID,
		CitedPaperID: &citedID,
		Authors:      &authors,
		Title:        &title,
		Year:         &year,
		Venue:        &venue,
		DOI:          &doi,
		URL:          &url,
		SourcePage:   &srcPage,
	}

	t.Run("insert", func(t *testing.T) {
		if err := repo.Insert(citation); err != nil {
			t.Fatalf("insert citation: %v", err)
		}
	})

	t.Run("insert duplicate returns error", func(t *testing.T) {
		if err := repo.Insert(citation); err == nil {
			t.Errorf("expected duplicate ID error, got nil")
		}
	})

	t.Run("list by paper", func(t *testing.T) {
		year2 := 2025
		authors2 := "Lee"
		citation2 := Citation{
			ID:      "cit-2",
			PaperID: docID,
			Authors: &authors2,
			Year:    &year2,
			Title:   &title,
		}
		if err := repo.Insert(citation2); err != nil {
			t.Fatalf("insert citation: %v", err)
		}

		list, err := repo.ListByPaper(docID)
		if err != nil {
			t.Fatalf("list citations: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("got %d rows, want 2", len(list))
		}
		if list[0].ID != "cit-1" || list[1].ID != "cit-2" {
			t.Errorf("unexpected order: got IDs %q, %q", list[0].ID, list[1].ID)
		}
		if list[0].Authors == nil || *list[0].Authors != authors {
			t.Errorf("got authors %v, want %q", list[0].Authors, authors)
		}
	})

	t.Run("list unknown paper returns empty", func(t *testing.T) {
		list, err := repo.ListByPaper("no-such-paper")
		if err != nil {
			t.Fatalf("list citations: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d rows, want 0", len(list))
		}
	})

	t.Run("get by id", func(t *testing.T) {
		got, err := repo.GetByID(docID, "cit-1")
		if err != nil {
			t.Fatalf("get citation: %v", err)
		}
		if got.ID != citation.ID {
			t.Errorf("got ID %q, want %q", got.ID, citation.ID)
		}
		if got.Title == nil || *got.Title != *citation.Title {
			t.Errorf("got title %v, want %v", got.Title, citation.Title)
		}
		if got.DOI == nil || *got.DOI != *citation.DOI {
			t.Errorf("got doi %v, want %v", got.DOI, citation.DOI)
		}
	})

	t.Run("get by id not found returns ErrNotFound", func(t *testing.T) {
		_, err := repo.GetByID(docID, "no-such-citation")
		if err != ErrNotFound {
			t.Errorf("got err %v, want ErrNotFound", err)
		}
	})

	t.Run("get by id wrong paper returns ErrNotFound", func(t *testing.T) {
		_, err := repo.GetByID("wrong-paper", "cit-1")
		if err != ErrNotFound {
			t.Errorf("got err %v, want ErrNotFound", err)
		}
	})

	t.Run("insert with unknown paper returns error", func(t *testing.T) {
		if err := repo.Insert(Citation{ID: "cit-bad", PaperID: "no-such-doc"}); err == nil {
			t.Errorf("expected foreign key error, got nil")
		}
	})
}
