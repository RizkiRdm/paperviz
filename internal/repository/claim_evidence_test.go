package repository

import (
	"testing"
)

func TestClaimEvidenceRepo(t *testing.T) {
	db := openTestDB(t)

	// Set up prerequisite rows (FK dependencies: document → claim, document → evidence).
	docID, err := NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	docRepo := NewDocumentRepo(db)
	if err := docRepo.Insert(Document{ID: docID, CreatedAt: 1, LastAccessedAt: 1, Status: StatusComplete, SourceType: SourceTypePDF, ReadingLevel: ReadingLevelSimplified, OriginalText: "text"}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	page := 1
	claimRepo := NewClaimRepo(db)
	if err := claimRepo.Insert(Claim{ID: "ce-test-claim", PaperID: docID, ClaimText: "test claim", ClaimType: "finding", Confidence: "high", SourcePage: &page, CreatedAt: 1}); err != nil {
		t.Fatalf("insert claim: %v", err)
	}

	evRepo := NewEvidenceRepo(db)
	if err := evRepo.Insert(Evidence{ID: "ce-test-ev1", PaperID: docID, SourceText: "evidence one", SourceReference: "ref1"}); err != nil {
		t.Fatalf("insert evidence: %v", err)
	}
	if err := evRepo.Insert(Evidence{ID: "ce-test-ev2", PaperID: docID, SourceText: "evidence two", SourceReference: "ref2"}); err != nil {
		t.Fatalf("insert evidence: %v", err)
	}

	repo := NewClaimEvidenceRepo(db)

	t.Run("insert success", func(t *testing.T) {
		ce := ClaimEvidence{ID: "ce-1", ClaimID: "ce-test-claim", EvidenceID: "ce-test-ev1", RelationshipType: ClaimEvidenceRelationshipSupports, CreatedAt: 10}
		if err := repo.Insert(ce); err != nil {
			t.Fatalf("insert claim_evidence: %v", err)
		}
	})

	t.Run("insert duplicate id returns error", func(t *testing.T) {
		dup := ClaimEvidence{ID: "ce-1", ClaimID: "ce-test-claim", EvidenceID: "ce-test-ev2", RelationshipType: ClaimEvidenceRelationshipContradicts, CreatedAt: 11}
		if err := repo.Insert(dup); err == nil {
			t.Errorf("expected unique constraint error, got nil")
		}
	})

	t.Run("get by claim success", func(t *testing.T) {
		// Insert second row for same claim so we get multiple results.
		if err := repo.Insert(ClaimEvidence{ID: "ce-2", ClaimID: "ce-test-claim", EvidenceID: "ce-test-ev2", RelationshipType: ClaimEvidenceRelationshipClarifies, CreatedAt: 12}); err != nil {
			t.Fatalf("insert second claim_evidence: %v", err)
		}

		list, err := repo.GetByClaim("ce-test-claim")
		if err != nil {
			t.Fatalf("get by claim: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("got %d rows, want 2", len(list))
		}
		// Ordered by created_at ASC.
		if list[0].ID != "ce-1" || list[0].RelationshipType != ClaimEvidenceRelationshipSupports {
			t.Errorf("row 0: got id=%q rel=%q, want ce-1 supports", list[0].ID, list[0].RelationshipType)
		}
		if list[1].ID != "ce-2" || list[1].RelationshipType != ClaimEvidenceRelationshipClarifies {
			t.Errorf("row 1: got id=%q rel=%q, want ce-2 clarifies", list[1].ID, list[1].RelationshipType)
		}
	})

	t.Run("get by claim empty", func(t *testing.T) {
		list, err := repo.GetByClaim("no-such-claim")
		if err != nil {
			t.Fatalf("get by claim: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d rows, want 0", len(list))
		}
	})

	t.Run("get by evidence success", func(t *testing.T) {
		list, err := repo.GetByEvidence("ce-test-ev1")
		if err != nil {
			t.Fatalf("get by evidence: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("got %d rows, want 1", len(list))
		}
		if list[0].ClaimID != "ce-test-claim" || list[0].RelationshipType != ClaimEvidenceRelationshipSupports {
			t.Errorf("got claim_id=%q rel=%q, want ce-test-claim supports", list[0].ClaimID, list[0].RelationshipType)
		}
	})

	t.Run("get by evidence empty", func(t *testing.T) {
		list, err := repo.GetByEvidence("no-such-evidence")
		if err != nil {
			t.Fatalf("get by evidence: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d rows, want 0", len(list))
		}
	})
}
