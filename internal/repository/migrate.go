package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrationsTable is where golang-migrate records the applied version and
// whether the last attempt left the schema dirty.
const migrationsTable = "schema_migrations"

// migrator couples a migration source to db.
//
// Deliberately not closed through migrate.Close: that closes the database
// driver, and the driver wraps the caller's *sql.DB, so calling it would take
// the whole application down. Only the source file handle is released here.
type migrator struct {
	m   *migrate.Migrate
	src io.Closer
}

// newMigrator wires the migrations in migrationsDir to db.
//
// db is passed as an existing instance rather than a DSN. That matters for
// tests, which use SQLite's :memory: database — every connection to ":memory:"
// gets its own private database, so a driver opening its own connection would
// migrate a schema nothing else could see.
func newMigrator(db *sql.DB, migrationsDir string) (*migrator, error) {
	src, err := iofs.New(os.DirFS(migrationsDir), ".")
	if err != nil {
		return nil, fmt.Errorf("open migrations dir %s: %w", migrationsDir, err)
	}

	driver, err := migratesqlite.WithInstance(db, &migratesqlite.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		_ = src.Close()
		return nil, fmt.Errorf("open migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance(migrationsTable, src, migrationsTable, driver)
	if err != nil {
		_ = src.Close()
		return nil, fmt.Errorf("build migrator: %w", err)
	}
	return &migrator{m: m, src: src}, nil
}

// Migrate applies every migration in migrationsDir to db.
//
// Migration files are discovered from disk by golang-migrate's parser and named
// {version}_{title}.up.sql / .down.sql. Nothing here hardcodes a migration
// list: the hand-maintained map this replaced silently drifted once and shipped
// migrations 18-19 missing on deploy, so the filesystem is now the only
// inventory.
func Migrate(db *sql.DB, migrationsDir string) error {
	mr, err := newMigrator(db, migrationsDir)
	if err != nil {
		return err
	}
	defer mr.src.Close()

	if err := mr.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations from %s: %w", migrationsDir, err)
	}

	// A dirty database means a migration was interrupted. Reporting that as
	// success would let the server boot against a half-built schema.
	version, dirty, err := mr.m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("read migration version: %w", err)
	}
	if dirty {
		return fmt.Errorf("schema is dirty at version %d: a migration failed part-way and needs repair", version)
	}
	return nil
}

// Down rolls the schema back one version — the rollback path
// ARCHITECTURE.md records as missing.
func Down(db *sql.DB, migrationsDir string) error {
	mr, err := newMigrator(db, migrationsDir)
	if err != nil {
		return err
	}
	defer mr.src.Close()

	if err := mr.m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("roll back one migration: %w", err)
	}
	return nil
}

// DownTo rolls the schema back to target. target 0 is rejected: golang-migrate
// resolves it by looking for a version-0 down file, which cannot exist. Use
// DownAll to empty the schema.
func DownTo(db *sql.DB, migrationsDir string, target uint) error {
	if target == 0 {
		return errors.New("DownTo: use DownAll to roll back to version 0")
	}
	mr, err := newMigrator(db, migrationsDir)
	if err != nil {
		return err
	}
	defer mr.src.Close()

	if err := mr.m.Migrate(target); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("roll back to version %d: %w", target, err)
	}
	return nil
}

// DownAll rolls the schema back to version 0, dropping every table the
// migrations created.
func DownAll(db *sql.DB, migrationsDir string) error {
	mr, err := newMigrator(db, migrationsDir)
	if err != nil {
		return err
	}
	defer mr.src.Close()

	if err := mr.m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("roll back all migrations: %w", err)
	}
	return nil
}

// Version reports the applied migration version, or 0 when none is applied.
func Version(db *sql.DB, migrationsDir string) (uint, bool, error) {
	mr, err := newMigrator(db, migrationsDir)
	if err != nil {
		return 0, false, err
	}
	defer mr.src.Close()

	version, dirty, err := mr.m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, err
}
