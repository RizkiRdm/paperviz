package services

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"paperviz/internal/repository"
)

// openTestTierDB creates an isolated file-based SQLite DB in t.TempDir()
// with all migrations applied, suitable for tier usage tests.
func openTestTierDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "tier_test.db")
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
	} {
		sqlStr, err := repository.ReadMigration(filepath.Join("..", "..", "migrations", file))
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}
		migrations[v] = sqlStr
	}
	db, err := repository.Open(dbPath, migrations)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestGetTier_NewUser(t *testing.T) {
	db := openTestTierDB(t)

	tier, err := GetTier(db, "new-fingerprint")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != TierFree {
		t.Errorf("tier: got %q, want %q", tier, TierFree)
	}
}

func TestGetTier_ExistingUser(t *testing.T) {
	db := openTestTierDB(t)

	// Insert a user with pro tier.
	_, err := db.Exec(`
		INSERT INTO user_tiers (fingerprint, tier, papers_used, last_reset)
		VALUES (?, ?, 0, ?)
	`, "pro-user", TierPro, time.Now().Unix())
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	tier, err := GetTier(db, "pro-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != TierPro {
		t.Errorf("tier: got %q, want %q", tier, TierPro)
	}
}

func TestCheckUsage_UnderLimit(t *testing.T) {
	db := openTestTierDB(t)

	// New user with 0 papers used — under limit.
	canCreate, used, err := CheckUsage(db, "new-fingerprint")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !canCreate {
		t.Error("expected canCreate=true for new user")
	}
	if used != 0 {
		t.Errorf("used: got %d, want 0", used)
	}
}

func TestCheckUsage_AtLimit(t *testing.T) {
	tests := []struct {
		name  string
		tier  string
		limit int
	}{
		{"free at limit", TierFree, LimitFree},
		{"pro at limit", TierPro, LimitPro},
		{"research at limit", TierResearch, LimitResearch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestTierDB(t)

			// Insert user at the exact limit.
			_, err := db.Exec(`
				INSERT INTO user_tiers (fingerprint, tier, papers_used, last_reset)
				VALUES (?, ?, ?, ?)
			`, "at-limit-"+tt.tier, tt.tier, tt.limit, time.Now().Unix())
			if err != nil {
				t.Fatalf("insert: %v", err)
			}

			canCreate, used, err := CheckUsage(db, "at-limit-"+tt.tier)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if canCreate {
				t.Error("expected canCreate=false at limit")
			}
			if used != tt.limit {
				t.Errorf("used: got %d, want %d", used, tt.limit)
			}
		})
	}
}

func TestIncrementUsage_CreatesRecord(t *testing.T) {
	db := openTestTierDB(t)

	err := IncrementUsage(db, "new-fingerprint")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var papersUsed int
	err = db.QueryRow(`SELECT papers_used FROM user_tiers WHERE fingerprint = ?`, "new-fingerprint").Scan(&papersUsed)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if papersUsed != 1 {
		t.Errorf("papers_used: got %d, want 1", papersUsed)
	}
}

func TestIncrementUsage_Increments(t *testing.T) {
	db := openTestTierDB(t)

	// Insert initial record.
	_, err := db.Exec(`
		INSERT INTO user_tiers (fingerprint, tier, papers_used, last_reset)
		VALUES (?, ?, 3, ?)
	`, "existing-user", TierFree, time.Now().Unix())
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	err = IncrementUsage(db, "existing-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var papersUsed int
	err = db.QueryRow(`SELECT papers_used FROM user_tiers WHERE fingerprint = ?`, "existing-user").Scan(&papersUsed)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if papersUsed != 4 {
		t.Errorf("papers_used: got %d, want 4", papersUsed)
	}
}

func TestResetMonthlyUsage_Resets(t *testing.T) {
	db := openTestTierDB(t)

	// Insert user with last_reset in previous month.
	lastMonth := time.Now().AddDate(0, -1, 0).Unix()
	_, err := db.Exec(`
		INSERT INTO user_tiers (fingerprint, tier, papers_used, last_reset)
		VALUES (?, ?, 5, ?)
	`, "old-reset-user", TierFree, lastMonth)
	if err != nil {
		t.Fatalf("insert old user: %v", err)
	}

	// Insert user with last_reset this month (should NOT reset).
	thisMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location()).Unix()
	_, err = db.Exec(`
		INSERT INTO user_tiers (fingerprint, tier, papers_used, last_reset)
		VALUES (?, ?, 3, ?)
	`, "current-reset-user", TierFree, thisMonth)
	if err != nil {
		t.Fatalf("insert current user: %v", err)
	}

	err = ResetMonthlyUsage(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Old reset user should have papers_used = 0.
	var used int
	err = db.QueryRow(`SELECT papers_used FROM user_tiers WHERE fingerprint = ?`, "old-reset-user").Scan(&used)
	if err != nil {
		t.Fatalf("query old user: %v", err)
	}
	if used != 0 {
		t.Errorf("old user papers_used: got %d, want 0", used)
	}

	// Current reset user should be unchanged.
	err = db.QueryRow(`SELECT papers_used FROM user_tiers WHERE fingerprint = ?`, "current-reset-user").Scan(&used)
	if err != nil {
		t.Fatalf("query current user: %v", err)
	}
	if used != 3 {
		t.Errorf("current user papers_used: got %d, want 3", used)
	}
}
