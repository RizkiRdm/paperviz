package repository

import (
	"testing"
)

func TestMethodRepo_Insert(t *testing.T) {
	db := openTestDB(t)
	docRepo := NewDocumentRepo(db)
	methodRepo := NewMethodRepo(db)

	docID := "doc-methods-1"
	if err := docRepo.Insert(Document{
		ID: docID, CreatedAt: 1, LastAccessedAt: 1, Status: StatusComplete,
		SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text",
	}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	t.Run("success inserts method row", func(t *testing.T) {
		m := Method{
			ID: "m-1", PaperID: docID, MethodName: "Randomized controlled trial",
			Description: strPtr("Double-blind study with n=200"),
			MethodType: MethodTypeExperimental, SourcePage: intPtr(5),
			SourceText: strPtr("We conducted a randomized controlled trial"),
		}
		if err := methodRepo.Insert(m); err != nil {
			t.Fatalf("insert method: %v", err)
		}

		got, err := methodRepo.GetByID(docID, "m-1")
		if err != nil {
			t.Fatalf("get method: %v", err)
		}
		if got.MethodName != m.MethodName {
			t.Errorf("method_name: got %q, want %q", got.MethodName, m.MethodName)
		}
		if got.MethodType != MethodTypeExperimental {
			t.Errorf("method_type: got %q, want %q", got.MethodType, MethodTypeExperimental)
		}
		if got.Description == nil || *got.Description != "Double-blind study with n=200" {
			t.Errorf("description: got %v", got.Description)
		}
	})

	t.Run("duplicate ID returns error", func(t *testing.T) {
		dup := Method{
			ID: "m-1", PaperID: docID, MethodName: "Duplicate",
			MethodType: MethodTypeExperimental,
		}
		if err := methodRepo.Insert(dup); err == nil {
			t.Error("expected error for duplicate ID, got nil")
		}
	})
}

func TestMethodRepo_ListByPaper(t *testing.T) {
	db := openTestDB(t)
	docRepo := NewDocumentRepo(db)
	methodRepo := NewMethodRepo(db)

	docID := "doc-methods-2"
	if err := docRepo.Insert(Document{
		ID: docID, CreatedAt: 1, LastAccessedAt: 1, Status: StatusComplete,
		SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text",
	}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	t.Run("returns all methods ordered by id", func(t *testing.T) {
		methods := []Method{
			{ID: "m-a", PaperID: docID, MethodName: "Survey", MethodType: MethodTypeSurvey},
			{ID: "m-b", PaperID: docID, MethodName: "Interview", MethodType: MethodTypeQualitative},
			{ID: "m-c", PaperID: docID, MethodName: "Lab experiment", MethodType: MethodTypeExperimental},
		}
		for _, m := range methods {
			if err := methodRepo.Insert(m); err != nil {
				t.Fatalf("insert method %s: %v", m.ID, err)
			}
		}

		list, err := methodRepo.ListByPaper(docID)
		if err != nil {
			t.Fatalf("list methods: %v", err)
		}
		if len(list) != 3 {
			t.Fatalf("got %d methods, want 3", len(list))
		}
		if list[0].ID != "m-a" || list[1].ID != "m-b" || list[2].ID != "m-c" {
			t.Errorf("wrong order: got [%s, %s, %s]", list[0].ID, list[1].ID, list[2].ID)
		}
	})

	t.Run("unknown paper returns empty", func(t *testing.T) {
		list, err := methodRepo.ListByPaper("no-such-paper")
		if err != nil {
			t.Fatalf("list methods: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d methods, want 0", len(list))
		}
	})
}

func TestMethodRepo_GetByID(t *testing.T) {
	db := openTestDB(t)
	docRepo := NewDocumentRepo(db)
	methodRepo := NewMethodRepo(db)

	docID := "doc-methods-3"
	if err := docRepo.Insert(Document{
		ID: docID, CreatedAt: 1, LastAccessedAt: 1, Status: StatusComplete,
		SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text",
	}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	m := Method{
		ID: "m-g", PaperID: docID, MethodName: "Computational model",
		Description: strPtr("Neural network architecture"),
		MethodType: MethodTypeComputational, SourcePage: intPtr(12),
		SourceText: strPtr("We trained a neural network on the dataset"),
	}
	if err := methodRepo.Insert(m); err != nil {
		t.Fatalf("insert method: %v", err)
	}

	t.Run("success returns method", func(t *testing.T) {
		got, err := methodRepo.GetByID(docID, "m-g")
		if err != nil {
			t.Fatalf("get method: %v", err)
		}
		if got.MethodName != "Computational model" {
			t.Errorf("method_name: got %q, want %q", got.MethodName, "Computational model")
		}
		if got.MethodType != MethodTypeComputational {
			t.Errorf("method_type: got %q, want %q", got.MethodType, MethodTypeComputational)
		}
		if got.SourcePage == nil || *got.SourcePage != 12 {
			t.Errorf("source_page: got %v, want 12", got.SourcePage)
		}
	})

	t.Run("wrong paper ID returns not found", func(t *testing.T) {
		_, err := methodRepo.GetByID("wrong-paper", "m-g")
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("nonexistent ID returns not found", func(t *testing.T) {
		_, err := methodRepo.GetByID(docID, "no-such-method")
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func intPtr(i int) *int { return &i }
