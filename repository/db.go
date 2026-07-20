package repository

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// Open opens (creating if needed) the SQLite database at path and applies
// the schema migration if the documents table does not yet exist.
//
// ponytail: a single flat migration file is fine at MVP scale (one schema,
// no history to replay). If the schema needs to evolve post-launch, switch
// to a numbered migration runner that tracks applied versions in a
// schema_migrations table before adding a second migration file.
func Open(dbPath, migrationsSQL string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite allows only one writer at a time; a single connection avoids
	// SQLITE_BUSY errors under the low-concurrency MVP load this is designed for.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	var exists int
	err = db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='documents'`).Scan(&exists)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("check schema: %w", err)
	}

	if exists == 0 {
		if _, err := db.Exec(migrationsSQL); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply migration: %w", err)
		}
	}

	return db, nil
}

// ReadMigration loads the migration SQL file from disk. Kept separate from
// Open so callers control the path (and tests can pass an in-memory schema
// string directly without touching the filesystem).
func ReadMigration(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read migration %s: %w", path, err)
	}
	return string(b), nil
}
