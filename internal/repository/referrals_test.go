package repository

import (
	"database/sql"
	"errors"
	"testing"
)

// insertReferralFixture inserts one document and one chart, setting a distinct
// share token on each so funnel counter updates can be asserted per table.
func insertReferralFixture(t *testing.T, db *sql.DB) (docToken, chartToken string) {
	t.Helper()
	docRepo := NewDocumentRepo(db)
	if err := docRepo.Insert(Document{
		ID: "ref-doc", CreatedAt: 1000, LastAccessedAt: 1000,
		Status: StatusComplete, SourceType: SourceTypePDF,
		ReadingLevel: ReadingLevelSimplified, Title: "Referral", OriginalText: "text",
	}); err != nil {
		t.Fatalf("insert document: %v", err)
	}
	if err := NewChartRepo(db).Insert(Chart{ID: "ref-chart", DocumentID: "ref-doc", SourceMethod: ChartSourceImageFallback}); err != nil {
		t.Fatalf("insert chart: %v", err)
	}
	docToken, chartToken = "doc-share-tok", "fig-share-tok"
	if err := docRepo.SetShareToken("ref-doc", docToken); err != nil {
		t.Fatalf("set doc share token: %v", err)
	}
	if err := NewChartRepo(db).SetShareToken("ref-chart", chartToken); err != nil {
		t.Fatalf("set chart share token: %v", err)
	}
	return docToken, chartToken
}

func queryCounter(t *testing.T, db *sql.DB, table, column, token string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT `+column+` FROM `+table+` WHERE share_token = ?`, token).Scan(&n); err != nil {
		t.Fatalf("query counter %s.%s: %v", table, column, err)
	}
	return n
}

func TestDocumentRepoIncrementShareCounters(t *testing.T) {
	db := openTestDB(t)
	repo := NewDocumentRepo(db)
	docToken, _ := insertReferralFixture(t, db)

	t.Run("visits increments from zero", func(t *testing.T) {
		if got := queryCounter(t, db, "documents", "share_visits", docToken); got != 0 {
			t.Fatalf("precondition: share_visits = %d, want 0", got)
		}
		if err := repo.IncrementShareVisits(docToken); err != nil {
			t.Fatalf("increment visits: %v", err)
		}
		if got := queryCounter(t, db, "documents", "share_visits", docToken); got != 1 {
			t.Errorf("share_visits: got %d, want 1", got)
		}
	})

	t.Run("conversions increments from zero", func(t *testing.T) {
		if err := repo.IncrementShareConversions(docToken); err != nil {
			t.Fatalf("increment conversions: %v", err)
		}
		if got := queryCounter(t, db, "documents", "share_conversions", docToken); got != 1 {
			t.Errorf("share_conversions: got %d, want 1", got)
		}
	})

	t.Run("unknown token is not found", func(t *testing.T) {
		if err := repo.IncrementShareVisits("no-such-token"); !errors.Is(err, ErrNotFound) {
			t.Errorf("visits unknown: got %v, want ErrNotFound", err)
		}
		if err := repo.IncrementShareConversions("no-such-token"); !errors.Is(err, ErrNotFound) {
			t.Errorf("conversions unknown: got %v, want ErrNotFound", err)
		}
	})
}

func TestChartRepoIncrementShareCounters(t *testing.T) {
	db := openTestDB(t)
	repo := NewChartRepo(db)
	_, chartToken := insertReferralFixture(t, db)

	t.Run("visits increments from zero", func(t *testing.T) {
		if got := queryCounter(t, db, "charts", "share_visits", chartToken); got != 0 {
			t.Fatalf("precondition: share_visits = %d, want 0", got)
		}
		if err := repo.IncrementShareVisits(chartToken); err != nil {
			t.Fatalf("increment visits: %v", err)
		}
		if got := queryCounter(t, db, "charts", "share_visits", chartToken); got != 1 {
			t.Errorf("share_visits: got %d, want 1", got)
		}
	})

	t.Run("conversions increments from zero", func(t *testing.T) {
		if err := repo.IncrementShareConversions(chartToken); err != nil {
			t.Fatalf("increment conversions: %v", err)
		}
		if got := queryCounter(t, db, "charts", "share_conversions", chartToken); got != 1 {
			t.Errorf("share_conversions: got %d, want 1", got)
		}
	})

	t.Run("unknown token is not found", func(t *testing.T) {
		if err := repo.IncrementShareVisits("no-such-token"); !errors.Is(err, ErrNotFound) {
			t.Errorf("visits unknown: got %v, want ErrNotFound", err)
		}
		if err := repo.IncrementShareConversions("no-such-token"); !errors.Is(err, ErrNotFound) {
			t.Errorf("conversions unknown: got %v, want ErrNotFound", err)
		}
	})
}
