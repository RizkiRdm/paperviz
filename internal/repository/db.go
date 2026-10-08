package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Open opens (creating if needed) the SQLite database at path and brings it up
// to date with the migrations found in migrationsDir.
func Open(dbPath string, migrationsDir string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite allows only one writer at a time; a single connection avoids
	// SQLITE_BUSY errors under the low-concurrency MVP load this is designed for.
	db.SetMaxOpenConns(1)

	// A ":memory:" database lives inside its connection, so the pool must never
	// retire that connection or the schema disappears. MaxIdleConns(1) holds it
	// and a zero lifetime means "never expire".
	if isMemoryPath(dbPath) {
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(0)
	}

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

	if err := Migrate(db, migrationsDir); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// isMemoryPath reports whether dbPath names a private in-memory database.
// Such a database belongs to a single connection, which drives the pool
// settings above.
func isMemoryPath(dbPath string) bool {
	base := dbPath
	if i := strings.IndexByte(base, '?'); i >= 0 {
		base = base[:i]
	}
	return base == ":memory:" || strings.HasPrefix(base, "file::memory:") || strings.Contains(base, "mode=memory")
}

// unixNow returns the current Unix timestamp in seconds.
func unixNow() int64 {
	return time.Now().Unix()
}
