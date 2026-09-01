package repository

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestPaperRelationshipRepo(t *testing.T) {
	db := openTestDB(t)

	// Insert migration 015 for paper_relationships table
	if err := execMigration(db, 15, "015_evidence_graph.sql"); err != nil {
		t.Fatalf("apply migration 015: %v", err)
	}

	// Create two source documents for testing
	docID1, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	docID2, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	docRepo := NewDocumentRepo(db)
	for _, id := range []string{docID1, docID2} {
		if err := docRepo.Insert(Document{ID: id, CreatedAt: 1, LastAccessedAt: 1, Status: StatusComplete, SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text"}); err != nil {
			t.Fatalf("insert document %s: %v", id, err)
		}
	}

	repo := NewPaperRelationshipRepo(db)

	t.Run("insert success", func(t *testing.T) {
		rel := PaperRelationship{
			ID:               "pr-1",
			SourcePaperID:    docID1,
			TargetPaperID:    docID2,
			RelationshipType: "supporting",
			EvidenceText:     strPtr("evidence text"),
			CreatedAt:        1,
		}
		if err := repo.Insert(rel); err != nil {
			t.Fatalf("insert paper_relationship: %v", err)
		}
	})

	t.Run("insert duplicate id returns error", func(t *testing.T) {
		rel := PaperRelationship{
			ID:               "pr-1",
			SourcePaperID:    docID1,
			TargetPaperID:    docID2,
			RelationshipType: "contradicting",
			EvidenceText:     strPtr("other evidence"),
			CreatedAt:        2,
		}
		if err := repo.Insert(rel); err == nil {
			t.Errorf("expected unique constraint error, got nil")
		}
	})

	t.Run("get by source paper success", func(t *testing.T) {
		// Insert another relationship from docID1
		rel2 := PaperRelationship{
			ID:               "pr-2",
			SourcePaperID:    docID1,
			TargetPaperID:    docID2,
			RelationshipType: "citing",
			EvidenceText:     strPtr("citing evidence"),
			CreatedAt:        3,
		}
		if err := repo.Insert(rel2); err != nil {
			t.Fatalf("insert paper_relationship: %v", err)
		}

		list, err := repo.GetBySourcePaper(docID1)
		if err != nil {
			t.Fatalf("get by source paper: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("got %d rows, want 2", len(list))
		}
		if list[0].RelationshipType != "supporting" {
			t.Errorf("got relationship_type %q, want supporting", list[0].RelationshipType)
		}
		if list[0].EvidenceText == nil || *list[0].EvidenceText != "evidence text" {
			t.Errorf("got evidence_text %v, want %q", list[0].EvidenceText, "evidence text")
		}
	})

	t.Run("get by source paper empty case", func(t *testing.T) {
		list, err := repo.GetBySourcePaper("no-such-paper")
		if err != nil {
			t.Fatalf("get by source paper: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d rows, want 0", len(list))
		}
	})

	t.Run("get by target paper success", func(t *testing.T) {
		// Insert a relationship targeting docID1
		rel3 := PaperRelationship{
			ID:               "pr-3",
			SourcePaperID:    docID2,
			TargetPaperID:    docID1,
			RelationshipType: "similar_methodology",
			EvidenceText:     strPtr("methodology evidence"),
			CreatedAt:        4,
		}
		if err := repo.Insert(rel3); err != nil {
			t.Fatalf("insert paper_relationship: %v", err)
		}

		list, err := repo.GetByTargetPaper(docID1)
		if err != nil {
			t.Fatalf("get by target paper: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("got %d rows, want 1", len(list))
		}
		if list[0].RelationshipType != "similar_methodology" {
			t.Errorf("got relationship_type %q, want similar_methodology", list[0].RelationshipType)
		}
		if list[0].EvidenceText == nil || *list[0].EvidenceText != "methodology evidence" {
			t.Errorf("got evidence_text %v, want %q", list[0].EvidenceText, "methodology evidence")
		}
	})

	t.Run("get by target paper empty case", func(t *testing.T) {
		list, err := repo.GetByTargetPaper("no-such-paper")
		if err != nil {
			t.Fatalf("get by target paper: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d rows, want 0", len(list))
		}
	})

	t.Run("get relationships success bidirectional", func(t *testing.T) {
		// Get relationships between docID1 and docID2 - should find both directions
		list, err := repo.GetRelationships(docID1, docID2)
		if err != nil {
			t.Fatalf("get relationships: %v", err)
		}
		if len(list) != 3 {
			t.Fatalf("got %d rows, want 3", len(list))
		}
		// Verify we have relationships from both directions
		supportCount := 0
		citeCount := 0
		methodCount := 0
		for _, r := range list {
			switch r.RelationshipType {
			case "supporting":
				supportCount++
			case "citing":
				citeCount++
			case "similar_methodology":
				methodCount++
			}
		}
		if supportCount != 1 || citeCount != 1 || methodCount != 1 {
			t.Errorf("expected 1 supporting, 1 citing, 1 similar_methodology; got %d, %d, %d", supportCount, citeCount, methodCount)
		}
	})

	t.Run("get relationships not found", func(t *testing.T) {
		_, err := repo.GetRelationships("no-such-paper-1", "no-such-paper-2")
		if err != ErrNotFound {
			t.Errorf("got err=%v, want ErrNotFound", err)
		}
	})
}

// execMigration applies a single migration by version number
func execMigration(db *sql.DB, version int, filename string) error {
	sqlStr, err := ReadMigration(filepath.Join("..", "..", "migrations", filename))
	if err != nil {
		return err
	}
	// Check if already applied
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already applied
	}
	// Apply migration
	if _, err := db.Exec(sqlStr); err != nil {
		return err
	}
	// Record migration
	if _, err := db.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", version, unixNow()); err != nil {
		return err
	}
	return nil
}
