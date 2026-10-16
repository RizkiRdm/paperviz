package documents

import (
	"errors"
	"testing"

	"paperviz/internal/repository"
	"paperviz/internal/testutil"
)

// TestCreateRelationship covers the domain rules that moved out of the
// handler: type vocabulary, self-reference, and target existence.
func TestCreateRelationship(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.SeedClerkUser(t, db, "u1", "u1@test.com", 1)
	testutil.SeedOwnedDocument(t, db, "src", "u1")
	testutil.SeedOwnedDocument(t, db, "tgt", "u1")
	svc := New(db, nil)

	tests := []struct {
		name       string
		sourceID   string
		targetID   string
		relType    string
		wantErr    error
		wantStatus int
	}{
		{
			name:     "valid relationship is inserted",
			sourceID: "src", targetID: "tgt",
			relType: repository.PaperRelationshipSupporting, wantErr: nil,
		},
		{
			name:     "unknown type is rejected",
			sourceID: "src", targetID: "tgt",
			relType: "friends", wantErr: ErrInvalidRelationshipType,
		},
		{
			name:     "self reference is rejected",
			sourceID: "src", targetID: "src",
			relType: repository.PaperRelationshipCiting, wantErr: ErrSelfReference,
		},
		{
			name:     "missing source is not found",
			sourceID: "ghost", targetID: "tgt",
			relType: repository.PaperRelationshipCiting, wantErr: ErrNotFound,
		},
		{
			name:     "missing target is target not found",
			sourceID: "src", targetID: "ghost",
			relType: repository.PaperRelationshipCiting, wantErr: ErrTargetNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, err := svc.CreateRelationship(tt.sourceID, tt.targetID, tt.relType, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if rel.ID == "" || rel.SourcePaperID != tt.sourceID || rel.TargetPaperID != tt.targetID {
					t.Fatalf("unexpected relationship: %+v", rel)
				}
				stored, err := repository.NewPaperRelationshipRepo(db).ListByPaper(tt.sourceID)
				if err != nil || len(stored) == 0 {
					t.Fatalf("relationship not persisted: %v, %v", stored, err)
				}
			}
		})
	}
}

// TestRelationshipsRequiresDocument keeps the list endpoint 404-safe.
func TestRelationshipsRequiresDocument(t *testing.T) {
	db := testutil.OpenTestDB(t)
	_, err := New(db, nil).Relationships("ghost")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
