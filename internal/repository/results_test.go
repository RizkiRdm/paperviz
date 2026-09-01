package repository

import (
	"testing"
)

func TestResultRepo(t *testing.T) {
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
	evID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	evRepo := NewEvidenceRepo(db)
	if err := evRepo.Insert(Evidence{ID: evID, PaperID: docID, Page: &page, SourceText: "evidence text", SourceReference: "ref"}); err != nil {
		t.Fatalf("insert evidence: %v", err)
	}

	repo := NewResultRepo(db)

	t.Run("insert success", func(t *testing.T) {
		res := Result{
			ID:                   "res-1",
			PaperID:              docID,
			ResultText:           "The study found a 15% improvement",
			ResultType:           ResultTypePrimary,
			SupportingEvidenceID: &evID,
			SourcePage:           &page,
			SourceText:           strPtr("source passage"),
		}
		if err := repo.Insert(res); err != nil {
			t.Fatalf("insert result: %v", err)
		}
	})

	t.Run("insert duplicate returns error", func(t *testing.T) {
		res := Result{
			ID:         "res-1",
			PaperID:    docID,
			ResultText: "duplicate",
			ResultType: ResultTypeSecondary,
		}
		if err := repo.Insert(res); err == nil {
			t.Errorf("expected duplicate ID error, got nil")
		}
	})

	t.Run("list by paper", func(t *testing.T) {
		res2 := Result{
			ID:         "res-2",
			PaperID:    docID,
			ResultText: "Secondary finding",
			ResultType: ResultTypeSecondary,
			SourcePage: &page,
		}
		if err := repo.Insert(res2); err != nil {
			t.Fatalf("insert result: %v", err)
		}

		list, err := repo.ListByPaper(docID)
		if err != nil {
			t.Fatalf("list results: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("got %d rows, want 2", len(list))
		}
		if list[0].ID != "res-1" || list[1].ID != "res-2" {
			t.Errorf("unexpected order: got IDs %q, %q", list[0].ID, list[1].ID)
		}
		if list[0].ResultText != "The study found a 15% improvement" {
			t.Errorf("got result_text %q, want expected text", list[0].ResultText)
		}
		if list[0].SupportingEvidenceID == nil || *list[0].SupportingEvidenceID != evID {
			t.Errorf("got evidence_id %v, want %s", list[0].SupportingEvidenceID, evID)
		}
		if list[0].SourcePage == nil || *list[0].SourcePage != page {
			t.Errorf("got source_page %v, want %d", list[0].SourcePage, page)
		}
	})

	t.Run("list unknown paper returns empty", func(t *testing.T) {
		list, err := repo.ListByPaper("no-such-paper")
		if err != nil {
			t.Fatalf("list results: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d rows, want 0", len(list))
		}
	})

	t.Run("get by id", func(t *testing.T) {
		got, err := repo.GetByID(docID, "res-1")
		if err != nil {
			t.Fatalf("get result: %v", err)
		}
		if got.ID != "res-1" {
			t.Errorf("got ID %q, want res-1", got.ID)
		}
		if got.ResultText != "The study found a 15% improvement" {
			t.Errorf("got result_text %q, want expected text", got.ResultText)
		}
		if got.ResultType != ResultTypePrimary {
			t.Errorf("got result_type %q, want %q", got.ResultType, ResultTypePrimary)
		}
	})

	t.Run("get by id not found returns ErrNotFound", func(t *testing.T) {
		_, err := repo.GetByID(docID, "no-such-result")
		if err != ErrNotFound {
			t.Errorf("got err %v, want ErrNotFound", err)
		}
	})

	t.Run("get by id wrong paper returns ErrNotFound", func(t *testing.T) {
		_, err := repo.GetByID("wrong-paper", "res-1")
		if err != ErrNotFound {
			t.Errorf("got err %v, want ErrNotFound", err)
		}
	})

	t.Run("insert with unknown paper returns error", func(t *testing.T) {
		if err := repo.Insert(Result{ID: "res-bad", PaperID: "no-such-doc", ResultText: "text", ResultType: ResultTypePrimary}); err == nil {
			t.Errorf("expected foreign key error, got nil")
		}
	})
}
