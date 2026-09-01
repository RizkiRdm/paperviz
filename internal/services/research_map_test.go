package services

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"paperviz/internal/repository"
)

// openTestResearchMapDB creates an in-memory SQLite DB with all migrations for research map tests.
func openTestResearchMapDB(t *testing.T) *sql.DB {
	t.Helper()
	migrations := make(map[int]string)
	for v, file := range map[int]string{
		1:  "001_init.sql",
		2:  "002_users.sql",
		3:  "003_chapters.sql",
		4:  "004_chapter_charts.sql",
		5:  "005_evidence.sql",
		6:  "006_document_title.sql",
		7:  "007_saved_papers.sql",
		8:  "008_research_collections.sql",
		9:  "009_share_tokens.sql",
		10: "010_document_share.sql",
		11: "011_share_referrals.sql",
		12: "012_usage_analytics.sql",
		13: "013_usage_tiers.sql",
		14: "014_structured_research_objects.sql",
		15: "015_evidence_graph.sql",
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

// insertTestDoc inserts a minimal document row for testing.
func insertTestDoc(t *testing.T, db *sql.DB, id, title string) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO documents (id, created_at, last_accessed_at, status, source_type, reading_level, title, original_text)
		VALUES (?, ?, ?, 'complete', 'pasted_text', 'simplified', ?, 'test text')`,
		id, now, now, title,
	)
	if err != nil {
		t.Fatalf("insert test doc %s: %v", id, err)
	}
}

func TestGetResearchMap_Empty(t *testing.T) {
	db := openTestResearchMapDB(t)
	insertTestDoc(t, db, "doc1", "Paper A")

	result, err := GetResearchMap(db, "doc1")
	if err != nil {
		t.Fatalf("GetResearchMap returned error: %v", err)
	}
	if result.DocumentID != "doc1" {
		t.Errorf("expected document_id doc1, got %s", result.DocumentID)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected 0 relationships, got %d", result.TotalCount)
	}
	for _, relType := range allRelationshipTypes {
		if len(result.Relationships[relType]) != 0 {
			t.Errorf("expected empty %s group, got %d items", relType, len(result.Relationships[relType]))
		}
	}
}

func TestGetResearchMap_NotFound(t *testing.T) {
	db := openTestResearchMapDB(t)

	_, err := GetResearchMap(db, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent document")
	}
	if err.Error() != "document not found" {
		t.Errorf("expected 'document not found', got %v", err)
	}
}

func TestGetResearchMap_Supporting(t *testing.T) {
	db := openTestResearchMapDB(t)
	insertTestDoc(t, db, "doc1", "Paper A")
	insertTestDoc(t, db, "doc2", "Paper B")

	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO paper_relationships (id, source_paper_id, target_paper_id, relationship_type, evidence_text, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"rel1", "doc1", "doc2", repository.PaperRelationshipSupporting, "same findings confirmed", now,
	)
	if err != nil {
		t.Fatalf("insert relationship: %v", err)
	}

	result, err := GetResearchMap(db, "doc1")
	if err != nil {
		t.Fatalf("GetResearchMap returned error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Fatalf("expected 1 relationship, got %d", result.TotalCount)
	}
	supporting := result.Relationships[repository.PaperRelationshipSupporting]
	if len(supporting) != 1 {
		t.Fatalf("expected 1 supporting relationship, got %d", len(supporting))
	}
	if supporting[0].TargetPaperTitle != "Paper B" {
		t.Errorf("expected target title 'Paper B', got %s", supporting[0].TargetPaperTitle)
	}
	if supporting[0].EvidenceText == nil || *supporting[0].EvidenceText != "same findings confirmed" {
		t.Errorf("expected evidence_text 'same findings confirmed', got %v", supporting[0].EvidenceText)
	}
}

func TestGetResearchMap_Bidirectional(t *testing.T) {
	db := openTestResearchMapDB(t)
	insertTestDoc(t, db, "doc1", "Paper A")
	insertTestDoc(t, db, "doc2", "Paper B")

	now := time.Now().Unix()
	// Relationship where doc1 is the TARGET, not the source.
	_, err := db.Exec(
		`INSERT INTO paper_relationships (id, source_paper_id, target_paper_id, relationship_type, evidence_text, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"rel1", "doc2", "doc1", repository.PaperRelationshipContradicting, "conflicting results", now,
	)
	if err != nil {
		t.Fatalf("insert relationship: %v", err)
	}

	result, err := GetResearchMap(db, "doc1")
	if err != nil {
		t.Fatalf("GetResearchMap returned error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Fatalf("expected 1 relationship, got %d", result.TotalCount)
	}
	contradicting := result.Relationships[repository.PaperRelationshipContradicting]
	if len(contradicting) != 1 {
		t.Fatalf("expected 1 contradicting relationship, got %d", len(contradicting))
	}
	if contradicting[0].SourcePaperID != "doc2" {
		t.Errorf("expected source doc2, got %s", contradicting[0].SourcePaperID)
	}
	if contradicting[0].TargetPaperTitle != "Paper B" {
		t.Errorf("expected target title 'Paper B', got %s", contradicting[0].TargetPaperTitle)
	}
}

func TestGetResearchMap_MultipleTypes(t *testing.T) {
	db := openTestResearchMapDB(t)
	insertTestDoc(t, db, "doc1", "Paper A")
	insertTestDoc(t, db, "doc2", "Paper B")
	insertTestDoc(t, db, "doc3", "Paper C")

	now := time.Now().Unix()
	rels := []struct {
		id, src, tgt, typ, evidence string
	}{
		{"r1", "doc1", "doc2", repository.PaperRelationshipSupporting, "supports"},
		{"r2", "doc1", "doc3", repository.PaperRelationshipContradicting, "contradicts"},
		{"r3", "doc2", "doc1", repository.PaperRelationshipCiting, "cites"},
	}
	for _, r := range rels {
		_, err := db.Exec(
			`INSERT INTO paper_relationships (id, source_paper_id, target_paper_id, relationship_type, evidence_text, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			r.id, r.src, r.tgt, r.typ, r.evidence, now,
		)
		if err != nil {
			t.Fatalf("insert relationship %s: %v", r.id, err)
		}
	}

	result, err := GetResearchMap(db, "doc1")
	if err != nil {
		t.Fatalf("GetResearchMap returned error: %v", err)
	}
	if result.TotalCount != 3 {
		t.Fatalf("expected 3 relationships, got %d", result.TotalCount)
	}
	if len(result.Relationships[repository.PaperRelationshipSupporting]) != 1 {
		t.Errorf("expected 1 supporting, got %d", len(result.Relationships[repository.PaperRelationshipSupporting]))
	}
	if len(result.Relationships[repository.PaperRelationshipContradicting]) != 1 {
		t.Errorf("expected 1 contradicting, got %d", len(result.Relationships[repository.PaperRelationshipContradicting]))
	}
	if len(result.Relationships[repository.PaperRelationshipCiting]) != 1 {
		t.Errorf("expected 1 citing, got %d", len(result.Relationships[repository.PaperRelationshipCiting]))
	}
}

func TestGetResearchMap_AllGroupsPresent(t *testing.T) {
	db := openTestResearchMapDB(t)
	insertTestDoc(t, db, "doc1", "Paper A")

	result, err := GetResearchMap(db, "doc1")
	if err != nil {
		t.Fatalf("GetResearchMap returned error: %v", err)
	}
	if len(result.Relationships) != len(allRelationshipTypes) {
		t.Errorf("expected %d relationship type groups, got %d", len(allRelationshipTypes), len(result.Relationships))
	}
	for _, relType := range allRelationshipTypes {
		if _, ok := result.Relationships[relType]; !ok {
			t.Errorf("missing relationship type group: %s", relType)
		}
	}
}
