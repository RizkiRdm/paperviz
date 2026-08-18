package services

import (
	"database/sql"
	"path/filepath"
	"testing"

	"paperviz/internal/repository"
)

func openTestServicesDB(t *testing.T) *sql.DB {
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
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSavePipelineResultWritesEvidence(t *testing.T) {
	db := openTestServicesDB(t)

	docID, err := repository.NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	docRepo := repository.NewDocumentRepo(db)
	if err := docRepo.Insert(repository.Document{ID: docID, CreatedAt: 1, LastAccessedAt: 1, Status: repository.StatusProcessing, SourceType: repository.SourceTypePDF, ReadingLevel: repository.ReadingLevelSimplified, OriginalText: "original"}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	output := PipelineOutput{
		Status:         pipelineStatusComplete,
		SimplifiedText: "simplified",
		Verify:         VerifyResult{OriginalClaims: []string{}, SimplifiedClaims: []string{}},
		Chapters:       []Chapter{{Title: "Intro", Summary: "summary", Excerpt: "excerpt"}},
		Charts: []Chart{
			{SourceMethod: chartSourceImageFallback, PageNumber: 3, SourceText: "page three context", DisplayOrder: 0, ChapterIndex: -1},
			{SourceMethod: chartSourceDataExtracted, PageNumber: 0, ChartData: `{"labels":[],"values":[]}`, Annotation: "from chapter", DisplayOrder: 1, ChapterIndex: 0},
		},
	}

	if err := savePipelineResult(db, docID, output); err != nil {
		t.Fatalf("save pipeline result: %v", err)
	}

	evidenceRepo := repository.NewEvidenceRepo(db)
	list, err := evidenceRepo.ListByPaper(docID)
	if err != nil {
		t.Fatalf("list evidence: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 evidence row for image-origin chart, got %d", len(list))
	}
	e := list[0]
	if e.Page == nil || *e.Page != 3 {
		t.Errorf("got page %v, want 3", e.Page)
	}
	if e.SourceText != "page three context" {
		t.Errorf("got source_text %q, want %q", e.SourceText, "page three context")
	}
	if e.SourceReference != "Figure on page 3" {
		t.Errorf("got source_reference %q, want %q", e.SourceReference, "Figure on page 3")
	}
	if e.FigureID == nil || *e.FigureID == "" {
		t.Errorf("expected figure_id to link to persisted chart, got %v", e.FigureID)
	}
}

func TestSavePipelineResultUnlinkedChartHasNilChapterID(t *testing.T) {
	db := openTestServicesDB(t)

	docID, err := repository.NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	docRepo := repository.NewDocumentRepo(db)
	if err := docRepo.Insert(repository.Document{
		ID:             docID,
		CreatedAt:      1,
		LastAccessedAt: 1,
		Status:         repository.StatusProcessing,
		SourceType:     repository.SourceTypePDF,
		ReadingLevel:   repository.ReadingLevelSimplified,
		OriginalText:   "original",
	}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	output := PipelineOutput{
		Status:         pipelineStatusComplete,
		SimplifiedText: "simplified",
		Verify:         VerifyResult{OriginalClaims: []string{}, SimplifiedClaims: []string{}},
		Chapters: []Chapter{
			{Title: "Chapter 1", Summary: "sum 1", Excerpt: "exc 1"},
			{Title: "Chapter 2", Summary: "sum 2", Excerpt: "exc 2"},
		},
		Charts: []Chart{
			{SourceMethod: chartSourceImageFallback, PageNumber: 1, SourceText: "page 1 context", DisplayOrder: 0, ChapterIndex: -1},
			{SourceMethod: chartSourceDataExtracted, PageNumber: 2, ChartData: `{"labels":[],"values":[]}`, Annotation: "ch 2 chart", DisplayOrder: 1, ChapterIndex: 1},
		},
	}

	if err := savePipelineResult(db, docID, output); err != nil {
		t.Fatalf("save pipeline result: %v", err)
	}

	chartRepo := repository.NewChartRepo(db)
	charts, err := chartRepo.ListByDocument(docID)
	if err != nil {
		t.Fatalf("list charts: %v", err)
	}

	if len(charts) != 2 {
		t.Fatalf("expected 2 charts, got %d", len(charts))
	}

	// Chart 0: ChapterIndex -1 -> ChapterID must be nil
	if charts[0].ChapterID != nil {
		t.Errorf("chart 0 ChapterID = %v, want nil", *charts[0].ChapterID)
	}

	// Chart 1: ChapterIndex 1 -> ChapterID must be non-nil
	if charts[1].ChapterID == nil {
		t.Errorf("chart 1 ChapterID is nil, want non-nil")
	}
}
