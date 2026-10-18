// Package testutil provides shared database fixtures for tests that need a
// migrated schema. It exists so transport-layer tests seed rows through one
// helper instead of importing the repository package themselves.
package testutil

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"paperviz/internal/models"
	"paperviz/internal/repository"
)

// OpenTestDB opens an in-memory SQLite with every migration applied.
func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := repository.Open(":memory:", migrationsDir(t))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// migrationsDir walks up from the test's working directory so the helper
// works from any package depth (internal/handlers, internal/app/..., ...).
func migrationsDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 6; i++ {
		cand := filepath.Join(dir, "migrations")
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("migrations dir not found above test working directory")
	return ""
}

// SeedUser inserts a placeholder local user row (Clerk owns real credentials).
func SeedUser(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	if err := repository.NewUserRepo(db).Insert(repository.User{
		ID: id, Email: id + "@example.com", PasswordHash: "x", CreatedAt: 1,
	}); err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

// SeedClerkUser writes the row AuthMiddleware would have created on first
// sight of a verified Clerk identity.
func SeedClerkUser(t *testing.T, db *sql.DB, id, email string, createdAt int64) {
	t.Helper()
	if err := repository.NewUserRepo(db).UpsertClerkUser(id, email, createdAt); err != nil {
		t.Fatalf("seed clerk user %s: %v", id, err)
	}
}

// SeedOwnedDocument inserts a completed pasted-text document owned by userID.
func SeedOwnedDocument(t *testing.T, db *sql.DB, docID, userID string) {
	t.Helper()
	now := time.Now().Unix()
	owner := userID
	if err := repository.NewDocumentRepo(db).Insert(repository.Document{
		ID: docID, CreatedAt: now, LastAccessedAt: now,
		Status: repository.StatusComplete, SourceType: repository.SourceTypePastedText,
		ReadingLevel: models.ReadingLevelSimplified, OriginalText: "body", UserID: &owner,
	}); err != nil {
		t.Fatalf("seed document %s: %v", docID, err)
	}
}
