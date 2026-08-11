package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadMigrationsRegistersChapterCharts guards the regression where
// 004_chapter_charts.sql (adds chapter_id to charts) was never registered in
// loadMigrations. Unregistered, the repository's chapter_id queries fail
// with "no such column" on every document GET.
func TestLoadMigrationsRegistersChapterCharts(t *testing.T) {
	// go test runs with the package dir as cwd; migrations live at repo root.
	migrationsDir := filepath.Join("..", "..", "migrations")
	migrations, err := loadMigrations(migrationsDir)
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}

	want := map[int][]string{
		1: {"CREATE TABLE documents", "CREATE TABLE charts"},
		2: {"CREATE TABLE IF NOT EXISTS users"},
		3: {"CREATE TABLE chapters"},
		4: {"ALTER TABLE charts ADD COLUMN chapter_id"},
	}

	if len(migrations) != len(want) {
		t.Fatalf("expected %d migrations, got %d", len(want), len(migrations))
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
