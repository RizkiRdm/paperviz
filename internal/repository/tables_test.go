package repository

import (
	"errors"
	"testing"
)

func TestPaperTableRepo_Insert(t *testing.T) {
	db := openTestDB(t)

	docID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	now := int64(1700000000)
	docRepo := NewDocumentRepo(db)
	if err := docRepo.Insert(Document{
		ID: docID, CreatedAt: now, LastAccessedAt: now, Status: StatusComplete,
		SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text",
	}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	pageNum := 2
	repo := NewPaperTableRepo(db)

	t.Run("success", func(t *testing.T) {
		id, err := NewID()
		if err != nil {
			t.Fatalf("new id: %v", err)
		}
		want := PaperTable{
			ID:           id,
			DocumentID:   docID,
			PageNumber:   &pageNum,
			Caption:      ptr("Table 1"),
			Headers:      `["col1","col2"]`,
			Rows:         `[["a","b"],["c","d"]]`,
			SourceText:   ptr("source"),
			DisplayOrder: 0,
		}
		if err := repo.Insert(want); err != nil {
			t.Fatalf("Insert: %v", err)
		}

		got, err := repo.GetByID(docID, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.ID != want.ID || got.DocumentID != want.DocumentID {
			t.Errorf("got id=%s doc=%s, want id=%s doc=%s", got.ID, got.DocumentID, want.ID, want.DocumentID)
		}
		if got.Headers != want.Headers || got.Rows != want.Rows {
			t.Errorf("got headers=%s rows=%s, want headers=%s rows=%s", got.Headers, got.Rows, want.Headers, want.Rows)
		}
	})

	t.Run("duplicate_id", func(t *testing.T) {
		id, err := NewID()
		if err != nil {
			t.Fatalf("new id: %v", err)
		}
		pt := PaperTable{
			ID:           id,
			DocumentID:   docID,
			Headers:      `[]`,
			Rows:         `[]`,
			DisplayOrder: 0,
		}
		if err := repo.Insert(pt); err != nil {
			t.Fatalf("first insert: %v", err)
		}
		if err := repo.Insert(pt); err == nil {
			t.Fatal("expected error on duplicate insert, got nil")
		}
	})
}

func TestPaperTableRepo_ListByDocument(t *testing.T) {
	db := openTestDB(t)

	docID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	emptyDocID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	now := int64(1700000000)
	docRepo := NewDocumentRepo(db)
	for _, id := range []string{docID, emptyDocID} {
		if err := docRepo.Insert(Document{
			ID: id, CreatedAt: now, LastAccessedAt: now, Status: StatusComplete,
			SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text",
		}); err != nil {
			t.Fatalf("insert document %s: %v", id, err)
		}
	}

	repo := NewPaperTableRepo(db)

	// Insert two tables for docID, ordered by display_order
	for i, caption := range []string{"Table 1", "Table 2"} {
		id, err := NewID()
		if err != nil {
			t.Fatalf("new id: %v", err)
		}
		if err := repo.Insert(PaperTable{
			ID:           id,
			DocumentID:   docID,
			Caption:      ptr(caption),
			Headers:      `["col"]`,
			Rows:         `[["val"]]`,
			DisplayOrder: i,
		}); err != nil {
			t.Fatalf("insert table %s: %v", caption, err)
		}
	}

	t.Run("success", func(t *testing.T) {
		tables, err := repo.ListByPaper(docID)
		if err != nil {
			t.Fatalf("ListByPaper: %v", err)
		}
		if len(tables) != 2 {
			t.Fatalf("got %d tables, want 2", len(tables))
		}
		if tables[0].DisplayOrder != 0 || tables[1].DisplayOrder != 1 {
			t.Errorf("wrong order: got [%d, %d], want [0, 1]", tables[0].DisplayOrder, tables[1].DisplayOrder)
		}
	})

	t.Run("empty", func(t *testing.T) {
		tables, err := repo.ListByPaper(emptyDocID)
		if err != nil {
			t.Fatalf("ListByPaper: %v", err)
		}
		if len(tables) != 0 {
			t.Errorf("got %d tables, want 0", len(tables))
		}
	})
}

func TestPaperTableRepo_GetByID(t *testing.T) {
	db := openTestDB(t)

	docID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	otherDocID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	now := int64(1700000000)
	docRepo := NewDocumentRepo(db)
	for _, id := range []string{docID, otherDocID} {
		if err := docRepo.Insert(Document{
			ID: id, CreatedAt: now, LastAccessedAt: now, Status: StatusComplete,
			SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text",
		}); err != nil {
			t.Fatalf("insert document %s: %v", id, err)
		}
	}

	pageNum := 5
	tableID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	repo := NewPaperTableRepo(db)
	want := PaperTable{
		ID:           tableID,
		DocumentID:   docID,
		PageNumber:   &pageNum,
		Caption:      ptr("Results Table"),
		Headers:      `["metric","value"]`,
		Rows:         `[["acc","0.95"],["f1","0.92"]]`,
		SourceText:   ptr("source text"),
		DisplayOrder: 1,
	}
	if err := repo.Insert(want); err != nil {
		t.Fatalf("insert table: %v", err)
	}

	tests := []struct {
		name    string
		docID   string
		tableID string
		wantErr error
	}{
		{"found", docID, tableID, nil},
		{"wrong_document", otherDocID, tableID, ErrNotFound},
		{"unknown_table", docID, "no-such-table", ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(tt.docID, tt.tableID)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("GetByID: %v", err)
				}
				if got.ID != want.ID || got.DocumentID != want.DocumentID {
					t.Errorf("got id=%s doc=%s, want id=%s doc=%s", got.ID, got.DocumentID, want.ID, want.DocumentID)
				}
				if got.Headers != want.Headers {
					t.Errorf("got headers=%s, want %s", got.Headers, want.Headers)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}
