package repository

import "testing"

// TestSeededSampleDocument proves migration 023 seeds the guest sample as a
// complete anonymous document with a claim and supporting evidence, so
// DocumentAccessAllowed (user_id NULL) lets anonymous visitors read it.
// The ID literal must match SAMPLE_DOCUMENT_ID in frontend/src/lib/sample.js
// and the INSERT in migrations/000023_seed_sample.up.sql.
func TestSeededSampleDocument(t *testing.T) {
	db := openTestDB(t)

	doc, err := NewDocumentRepo(db).Get("SamplePaperCells")
	if err != nil {
		t.Fatalf("seeded sample missing: %v", err)
	}
	if doc.Status != StatusComplete {
		t.Errorf("status = %q, want %q", doc.Status, StatusComplete)
	}
	if doc.UserID != nil {
		t.Errorf("user_id = %v, want nil (anonymous read)", *doc.UserID)
	}
	if doc.SimplifiedText == nil || *doc.SimplifiedText == "" {
		t.Error("simplified_text empty, want renderable sample content")
	}

	var claims, evidence int
	if err := db.QueryRow(`SELECT COUNT(*) FROM claims WHERE paper_id = ?`, doc.ID).Scan(&claims); err != nil {
		t.Fatalf("count claims: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM evidence WHERE paper_id = ?`, doc.ID).Scan(&evidence); err != nil {
		t.Fatalf("count evidence: %v", err)
	}
	if claims == 0 {
		t.Error("seeded sample has no claims")
	}
	if evidence == 0 {
		t.Error("seeded sample has no evidence")
	}
}
