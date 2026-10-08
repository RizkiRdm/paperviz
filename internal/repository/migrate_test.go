package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func migrationsDir(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "migrations")
}

// TestMigrateAppliesEveryFile proves the schema comes from the filesystem and
// not from a hardcoded list. It asserts the newest table exists, so a new
// migration that is never registered still lands.
func TestMigrateAppliesEveryFile(t *testing.T) {
	db, err := Open(":memory:", migrationsDir(t))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	version, dirty, err := Version(db, migrationsDir(t))
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if dirty {
		t.Fatal("schema is dirty after a clean migrate")
	}
	if version < 21 {
		t.Errorf("applied version = %d, want at least 21", version)
	}

	// user_credentials arrives in 020 and is what BYOK depends on.
	var name string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='user_credentials'`).Scan(&name)
	if err != nil {
		t.Fatalf("user_credentials missing after migrate: %v", err)
	}
}

// TestMigrateDownRoundTrip rolls the whole schema back to version 0 and
// forward again.
//
// This is the check that proves the .down.sql files are real. A down migration
// that silently does the wrong thing passes every other test in the repo,
// because nothing else ever goes backwards.
func TestMigrateDownRoundTrip(t *testing.T) {
	dir := migrationsDir(t)
	dbPath := filepath.Join(t.TempDir(), "roundtrip.db")

	db, err := Open(dbPath, dir)
	if err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	db.Close()

	down, err := Open(dbPath, dir)
	if err != nil {
		t.Fatalf("reopen for rollback: %v", err)
	}
	defer down.Close()

	if err := DownAll(down, dir); err != nil {
		t.Fatalf("roll back to version 0: %v", err)
	}

	version, _, err := Version(down, dir)
	if err != nil {
		t.Fatalf("read version after rollback: %v", err)
	}
	var dirty bool
	if version != 0 {
		t.Errorf("version after full rollback = %d, want 0", version)
	}

	// Every application table must be gone. schema_migrations and SQLite's own
	// bookkeeping tables are expected to remain.
	var remaining int
	err = down.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		  AND name != 'schema_migrations'`).Scan(&remaining)
	if err != nil {
		t.Fatalf("count remaining tables: %v", err)
	}
	if remaining != 0 {
		names := []string{}
		rows, qErr := down.Query(`
			SELECT name FROM sqlite_master
			WHERE type = 'table' AND name NOT LIKE 'sqlite_%' AND name != 'schema_migrations'`)
		if qErr == nil {
			defer rows.Close()
			for rows.Next() {
				var n string
				if rows.Scan(&n) == nil {
					names = append(names, n)
				}
			}
		}
		t.Errorf("%d table(s) survived a full rollback: %v", remaining, names)
	}

	// And forward again, to prove the down files left a schema the up files
	// can rebuild.
	if err := Migrate(down, dir); err != nil {
		t.Fatalf("re-migrate after rollback: %v", err)
	}
	version, dirty, err = Version(down, dir)
	if err != nil {
		t.Fatalf("read version after re-migrate: %v", err)
	}
	if dirty {
		t.Error("schema is dirty after re-migrating")
	}
	if version < 21 {
		t.Errorf("version after re-migrate = %d, want at least 21", version)
	}
}

// TestEveryMigrationHasADownFile guards the defect golang-migrate introduces:
// an .up.sql with no .down.sql still applies cleanly, so the asymmetry stays
// invisible until someone needs the rollback. Every migration must be
// reversible, which is what makes the rollback path real.
func TestEveryMigrationHasADownFile(t *testing.T) {
	dir := migrationsDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var ups, downs int
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			continue
		}
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			ups++
			base := strings.TrimSuffix(name, ".up.sql")
			if _, err := os.Stat(filepath.Join(dir, base+".down.sql")); err != nil {
				t.Errorf("migration %s has no matching .down.sql, so it cannot be rolled back", base)
			}
		case strings.HasSuffix(name, ".down.sql"):
			downs++
			base := strings.TrimSuffix(name, ".down.sql")
			if _, err := os.Stat(filepath.Join(dir, base+".up.sql")); err != nil {
				t.Errorf("migration %s.down.sql has no matching .up.sql", base)
			}
		default:
			t.Errorf("stray file %q in migrations: names must be {version}_{title}.up.sql / .down.sql", name)
		}
	}

	if ups == 0 {
		t.Fatal("no migrations found")
	}
	if ups != downs {
		t.Errorf("%d .up.sql but %d .down.sql", ups, downs)
	}
}
