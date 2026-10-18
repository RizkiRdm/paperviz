package documents

import (
	"errors"
	"testing"

	"paperviz/internal/testutil"
)

// TestValidReadingLevel guards the intake vocabulary against drift.
func TestValidReadingLevel(t *testing.T) {
	tests := []struct {
		level string
		want  bool
	}{
		{ReadingLevelSimplified, true},
		{ReadingLevelELI5, true},
		{"expert", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := ValidReadingLevel(tt.level); got != tt.want {
			t.Errorf("ValidReadingLevel(%q) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

// TestListClaims covers the sub-resource pattern every list endpoint shares:
// an existing document returns its rows, a missing document is ErrNotFound.
func TestListClaims(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.SeedClerkUser(t, db, "u1", "u1@test.com", 1)
	testutil.SeedOwnedDocument(t, db, "doc-1", "u1")
	svc := New(db, nil)

	tests := []struct {
		name    string
		docID   string
		wantErr error
	}{
		{name: "existing document lists claims", docID: "doc-1", wantErr: nil},
		{name: "missing document is not found", docID: "nope", wantErr: ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ListClaims(tt.docID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestStats covers the dashboard counters: zeros on an empty database, and an
// error from a database that cannot answer.
func TestStats(t *testing.T) {
	t.Run("empty database reports zeros", func(t *testing.T) {
		db := testutil.OpenTestDB(t)
		stats, err := New(db, nil).Stats("u1")
		if err != nil {
			t.Fatalf("stats: %v", err)
		}
		if stats.Total != 0 || stats.Saved != 0 || stats.Collections != 0 {
			t.Fatalf("expected zeros on empty db, got %+v", stats)
		}
	})

	t.Run("closed database errors", func(t *testing.T) {
		db := testutil.OpenTestDB(t)
		db.Close()
		if _, err := New(db, nil).Stats("u1"); err == nil {
			t.Fatal("expected an error from a closed database")
		}
	})
}

// TestRequireDocIsNotFoundSentinel proves the re-exported sentinel is the same
// value the repository returns, so errors.Is keeps working across the boundary.
func TestRequireDocIsNotFoundSentinel(t *testing.T) {
	db := testutil.OpenTestDB(t)
	_, err := New(db, nil).requireDoc("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
