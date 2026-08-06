package repository

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"time"

	_ "modernc.org/sqlite"
)

// Open opens (creating if needed) the SQLite database at path and applies
// any pending migrations. Migration tracking uses a schema_migrations table
// to ensure each migration runs exactly once, in order.
func Open(dbPath string, migrations map[int]string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite allows only one writer at a time; a single connection avoids
	// SQLITE_BUSY errors under the low-concurrency MVP load this is designed for.
	db.SetMaxOpenConns(1)

	// WAL mode + synchronous=NORMAL: write-ahead logging avoids the fsync
	// overhead of rollback journals on every GET poll (which calls TouchLastAccessed
	// — a write). synchronous=NORMAL is acceptable for ephemeral data (7-day expiry).
	// busy_timeout=5000 prevents SQLITE_BUSY if MaxOpenConns is raised later.
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set synchronous mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	// Ensure schema_migrations table exists
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL
	)`); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema_migrations: %w", err)
	}

	// Get applied migrations
	applied := make(map[int]bool)
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			db.Close()
			return nil, fmt.Errorf("scan schema_migrations: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		db.Close()
		return nil, fmt.Errorf("iterate schema_migrations: %w", err)
	}

	// Sort migration versions
	versions := make([]int, 0, len(migrations))
	for v := range migrations {
		versions = append(versions, v)
	}
	sort.Ints(versions)

	// Apply pending migrations
	for _, version := range versions {
		if applied[version] {
			continue
		}

		migrationSQL := migrations[version]
		if _, err := db.Exec(migrationSQL); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply migration %d: %w", version, err)
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, unixNow()); err != nil {
			db.Close()
			return nil, fmt.Errorf("record migration %d: %w", version, err)
		}
	}

	return db, nil
}

// ReadMigration loads a migration SQL file from disk.
func ReadMigration(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read migration %s: %w", path, err)
	}
	return string(b), nil
}

// unixNow returns the current Unix timestamp in seconds.
func unixNow() int64 {
	return time.Now().Unix()
}
