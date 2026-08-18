package repository

import (
	"fmt"
	"testing"
)

func TestDocumentRepoListSummariesByUser(t *testing.T) {
	db := openTestDB(t)
	userID := "user-list-1"
	docRepo := NewDocumentRepo(db)

	userRepo := NewUserRepo(db)
	if err := userRepo.Insert(User{ID: userID, Email: "list-test@test.com", PasswordHash: "hash", CreatedAt: 1}); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	t.Run("returns title summary preview and counts in desc order", func(t *testing.T) {
		docs := []Document{
			{ID: "doc-new", CreatedAt: 200, LastAccessedAt: 200, Status: StatusComplete, SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, Title: "New paper", OriginalText: "original", SimplifiedText: strPtr("Simplified version of new paper with enough text"), UserID: &userID},
			{ID: "doc-old", CreatedAt: 100, LastAccessedAt: 100, Status: StatusComplete, SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, Title: "Old paper", OriginalText: "original", SimplifiedText: strPtr("Old paper simplified"), UserID: &userID},
		}
		for _, d := range docs {
			if err := docRepo.Insert(d); err != nil {
				t.Fatalf("insert doc %s: %v", d.ID, err)
			}
		}

		chartRepo := NewChartRepo(db)
		ann := "Explains the figure clearly"
		charts := []Chart{
			{ID: "ch-1", DocumentID: "doc-new", SourceMethod: ChartSourceDataExtracted, Annotation: &ann, DisplayOrder: 0},
			{ID: "ch-2", DocumentID: "doc-new", SourceMethod: ChartSourceImageFallback, Annotation: &ann, DisplayOrder: 1},
			{ID: "ch-3", DocumentID: "doc-old", SourceMethod: ChartSourceImageFallback, Annotation: nil, DisplayOrder: 0},
		}
		for _, c := range charts {
			if err := chartRepo.Insert(c); err != nil {
				t.Fatalf("insert chart %s: %v", c.ID, err)
			}
		}

		items, err := docRepo.ListSummariesByUser(userID, 100, 0)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("got %d items, want 2", len(items))
		}

		if items[0].ID != "doc-new" || items[1].ID != "doc-old" {
			t.Errorf("wrong order: got [%s, %s], want [doc-new, doc-old]", items[0].ID, items[1].ID)
		}

		if items[0].Title != "New paper" {
			t.Errorf("title: got %q, want %q", items[0].Title, "New paper")
		}
		if items[0].SummaryPreview != "Simplified version of new paper with enough text" {
			t.Errorf("summary_preview: got %q", items[0].SummaryPreview)
		}
		if items[0].ChartCount != 2 {
			t.Errorf("chart_count: got %d, want 2", items[0].ChartCount)
		}
		if items[0].ExplanationCount != 2 {
			t.Errorf("explanation_count: got %d, want 2", items[0].ExplanationCount)
		}
		if items[1].ChartCount != 1 {
			t.Errorf("chart_count old: got %d, want 1", items[1].ChartCount)
		}
		if items[1].ExplanationCount != 0 {
			t.Errorf("explanation_count old: got %d, want 0", items[1].ExplanationCount)
		}
	})

	t.Run("empty user returns empty", func(t *testing.T) {
		items, err := docRepo.ListSummariesByUser("no-such-user", 100, 0)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("got %d items, want 0", len(items))
		}
	})

	t.Run("pagination limit offset", func(t *testing.T) {
		items, err := docRepo.ListSummariesByUser(userID, 1, 0)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("limit 1: got %d, want 1", len(items))
		}
		if items[0].ID != "doc-new" {
			t.Errorf("limit 1 first: got %s, want doc-new", items[0].ID)
		}

		items2, err := docRepo.ListSummariesByUser(userID, 1, 1)
		if err != nil {
			t.Fatalf("list offset: %v", err)
		}
		if len(items2) != 1 || items2[0].ID != "doc-old" {
			t.Errorf("offset 1: got %v, want [doc-old]", items2)
		}

		itemsPast, err := docRepo.ListSummariesByUser(userID, 100, 5)
		if err != nil {
			t.Fatalf("list past end: %v", err)
		}
		if len(itemsPast) != 0 {
			t.Errorf("past end: got %d, want 0", len(itemsPast))
		}
	})

	t.Run("user B not leaked to user A", func(t *testing.T) {
		userB := "user-list-2"
		if err := userRepo.Insert(User{ID: userB, Email: "list-test-b@test.com", PasswordHash: "hash", CreatedAt: 1}); err != nil {
			t.Fatalf("insert user B: %v", err)
		}
		if err := docRepo.Insert(Document{ID: "doc-b", CreatedAt: 300, LastAccessedAt: 300, Status: StatusComplete, SourceType: SourceTypePastedText, ReadingLevel: ReadingLevelELI5, Title: "User B paper", OriginalText: "text", UserID: &userB}); err != nil {
			t.Fatalf("insert user B doc: %v", err)
		}

		items, err := docRepo.ListSummariesByUser(userID, 100, 0)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		for _, it := range items {
			if it.ID == "doc-b" {
				t.Error("user B doc leaked to user A list")
			}
		}
	})
}

func strPtr(s string) *string { return &s }

func TestToggleSaved(t *testing.T) {
	db := openTestDB(t)
	repo := NewDocumentRepo(db)

	doc := Document{
		ID: "test-doc-1", CreatedAt: 1000, LastAccessedAt: 1000,
		Status: StatusComplete, SourceType: SourceTypePDF,
		ReadingLevel: ReadingLevelSimplified, Title: "Test",
		OriginalText: "original",
	}
	if err := repo.Insert(doc); err != nil {
		t.Fatal(err)
	}

	// Toggle on
	if err := repo.ToggleSaved("test-doc-1", true); err != nil {
		t.Fatal(err)
	}
	d, _ := repo.Get("test-doc-1")
	if !d.Saved {
		t.Error("expected saved=true after toggle")
	}

	// Toggle off
	if err := repo.ToggleSaved("test-doc-1", false); err != nil {
		t.Fatal(err)
	}
	d, _ = repo.Get("test-doc-1")
	if d.Saved {
		t.Error("expected saved=false after toggle")
	}
}

func TestUpdateTitle(t *testing.T) {
	db := openTestDB(t)
	repo := NewDocumentRepo(db)

	doc := Document{
		ID: "test-doc-2", CreatedAt: 1000, LastAccessedAt: 1000,
		Status: StatusComplete, SourceType: SourceTypePDF,
		ReadingLevel: ReadingLevelSimplified, Title: "Original",
		OriginalText: "original",
	}
	if err := repo.Insert(doc); err != nil {
		t.Fatal(err)
	}

	if err := repo.UpdateTitle("test-doc-2", "New Title"); err != nil {
		t.Fatal(err)
	}
	d, _ := repo.Get("test-doc-2")
	if d.Title != "New Title" {
		t.Errorf("expected title='New Title', got %q", d.Title)
	}
}

func TestDeleteDocument(t *testing.T) {
	db := openTestDB(t)
	repo := NewDocumentRepo(db)

	doc := Document{
		ID: "test-doc-3", CreatedAt: 1000, LastAccessedAt: 1000,
		Status: StatusComplete, SourceType: SourceTypePDF,
		ReadingLevel: ReadingLevelSimplified, Title: "Delete Me",
		OriginalText: "original",
	}
	if err := repo.Insert(doc); err != nil {
		t.Fatal(err)
	}

	if err := repo.DeleteDocument("test-doc-3"); err != nil {
		t.Fatal(err)
	}
	_, err := repo.Get("test-doc-3")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestCountByUser(t *testing.T) {
	db := openTestDB(t)
	repo := NewDocumentRepo(db)

	userRepo := NewUserRepo(db)
	if err := userRepo.Insert(User{ID: "user-1", Email: "count@test.com", PasswordHash: "hash", CreatedAt: 1}); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		doc := Document{
			ID: fmt.Sprintf("doc-%d", i), CreatedAt: 1000, LastAccessedAt: 1000,
			Status: StatusComplete, SourceType: SourceTypePDF,
			ReadingLevel: ReadingLevelSimplified, Title: fmt.Sprintf("Doc %d", i),
			OriginalText: "original", UserID: strPtr("user-1"),
		}
		if err := repo.Insert(doc); err != nil {
			t.Fatal(err)
		}
	}

	count, err := repo.CountByUser("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("expected 3 documents, got %d", count)
	}
}

func TestCountSavedByUser(t *testing.T) {
	db := openTestDB(t)
	repo := NewDocumentRepo(db)

	userRepo := NewUserRepo(db)
	if err := userRepo.Insert(User{ID: "user-1", Email: "saved@test.com", PasswordHash: "hash", CreatedAt: 1}); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		doc := Document{
			ID: fmt.Sprintf("doc-%d", i), CreatedAt: 1000, LastAccessedAt: 1000,
			Status: StatusComplete, SourceType: SourceTypePDF,
			ReadingLevel: ReadingLevelSimplified, Title: fmt.Sprintf("Doc %d", i),
			OriginalText: "original", UserID: strPtr("user-1"),
		}
		if err := repo.Insert(doc); err != nil {
			t.Fatal(err)
		}
	}
	repo.ToggleSaved("doc-0", true)
	repo.ToggleSaved("doc-1", true)

	count, err := repo.CountSavedByUser("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2 saved documents, got %d", count)
	}
}
