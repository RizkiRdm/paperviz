package analytics

import (
	"testing"

	"paperviz/internal/testutil"
)

// TestSummary covers the aggregate read: zeros on an empty database, and an
// error from a database that cannot answer.
func TestSummary(t *testing.T) {
	t.Run("empty database reports zeros", func(t *testing.T) {
		db := testutil.OpenTestDB(t)
		summary, err := Summary(db)
		if err != nil {
			t.Fatalf("summary: %v", err)
		}
		if summary.TotalPapers != 0 || summary.TotalFigures != 0 {
			t.Fatalf("expected zeros on empty db, got %+v", summary)
		}
	})

	t.Run("closed database errors", func(t *testing.T) {
		db := testutil.OpenTestDB(t)
		db.Close()
		if _, err := Summary(db); err == nil {
			t.Fatal("expected an error from a closed database")
		}
	})
}
