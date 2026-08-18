package repository

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	migrations := make(map[int]string)
	for v, file := range map[int]string{
		1: "001_init.sql",
		2: "002_users.sql",
		3: "003_chapters.sql",
		4: "004_chapter_charts.sql",
		5: "005_evidence.sql",
		6: "006_document_title.sql",
		7: "007_saved_papers.sql",
	} {
		sqlStr, err := ReadMigration(filepath.Join("..", "..", "migrations", file))
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}
		migrations[v] = sqlStr
	}
	db, err := Open(":memory:", migrations)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestChartRepoGetByDocumentAndID(t *testing.T) {
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
		if err := docRepo.Insert(Document{ID: id, CreatedAt: now, LastAccessedAt: now, Status: StatusComplete, SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text"}); err != nil {
			t.Fatalf("insert document %s: %v", id, err)
		}
	}

	pageNum := 3
	chartID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	want := Chart{
		ID:           chartID,
		DocumentID:   docID,
		SourceMethod: ChartSourceImageFallback,
		ImageBlob:    []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a},
		Annotation:   ptr("annotation"),
		PageNumber:   &pageNum,
		DisplayOrder: 0,
	}
	chartRepo := NewChartRepo(db)
	if err := chartRepo.Insert(want); err != nil {
		t.Fatalf("insert chart: %v", err)
	}

	tests := []struct {
		name    string
		docID   string
		chartID string
		wantErr error
	}{
		{"found", docID, chartID, nil},
		{"wrong_document", otherDocID, chartID, ErrNotFound},
		{"unknown_chart", docID, "no-such-chart", ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := chartRepo.GetByDocumentAndID(tt.docID, tt.chartID)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("GetByDocumentAndID: %v", err)
				}
				if got.ID != want.ID || got.DocumentID != want.DocumentID {
					t.Errorf("got chart %+v, want id=%s doc=%s", got, want.ID, want.DocumentID)
				}
				if len(got.ImageBlob) != len(want.ImageBlob) {
					t.Errorf("got image_blob %d bytes, want %d", len(got.ImageBlob), len(want.ImageBlob))
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func ptr(s string) *string { return &s }
