package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"paperviz/internal/repository"
)

// TestLoadMigrationsRegistersAll guards the regression where migrations were
// never registered in LoadMigrations. Unregistered, repository queries fail
// with "no such column" on every request. Covers all 19 migrations and
// fails if a new migration file appears on disk without registration.
func TestLoadMigrationsRegistersChapterCharts(t *testing.T) {
	// go test runs with the package dir as cwd; migrations live at repo root.
	migrationsDir := filepath.Join("..", "..", "migrations")
	migrations, err := repository.LoadMigrations(migrationsDir)
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}

	want := map[int][]string{
		1:  {"CREATE TABLE documents", "CREATE TABLE charts"},
		2:  {"CREATE TABLE IF NOT EXISTS users"},
		3:  {"CREATE TABLE chapters"},
		4:  {"ALTER TABLE charts ADD COLUMN chapter_id"},
		5:  {"CREATE TABLE evidence"},
		6:  {"ADD COLUMN title"},
		7:  {"ADD COLUMN saved"},
		8:  {"CREATE TABLE collections", "CREATE TABLE document_collections"},
		9:  {"share_token", "visibility"},
		10: {"share_token"},
		11: {"share_visits", "share_conversions"},
		12: {"processing_time_ms", "analytics_events"},
		13: {"user_tiers"},
		14: {"claims", "paper_tables", "methods", "results", "citations"},
		15: {"claim_evidence", "paper_relationships"},
		16: {"annotations"},
		17: {"ALTER TABLE users ADD COLUMN oauth_provider", "ALTER TABLE users ADD COLUMN oauth_id"},
		18: {"api_key"},
		19: {"stripe_customer_id"},
	}

	if len(migrations) != len(want) {
		t.Fatalf("expected %d migrations, got %d", len(want), len(migrations))
	}

	// Guard against future drift: if a new .sql appears on disk without registration, fail.
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	sqlCount := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlCount++
		}
	}
	if sqlCount != len(want) {
		t.Fatalf("migrations on disk (%d .sql files) != registered (%d) — new migration file not registered in LoadMigrations", sqlCount, len(want))
	}
	if len(migrations) != sqlCount {
		t.Fatalf("loaded migrations (%d) != .sql files on disk (%d)", len(migrations), sqlCount)
	}

	for version, needles := range want {
		sql, ok := migrations[version]
		if !ok {
			t.Errorf("migration %d not registered", version)
			continue
		}
		for _, needle := range needles {
			if !strings.Contains(sql, needle) {
				t.Errorf("migration %d missing %q", version, needle)
			}
		}
	}
}
