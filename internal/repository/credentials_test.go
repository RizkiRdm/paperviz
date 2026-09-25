package repository

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

// openCredentialTestDB loads only the migrations this package needs: documents
// and users, then the credentials table.
func openCredentialTestDB(t *testing.T) *sql.DB {
	t.Helper()
	migrations := make(map[int]string)
	for v, file := range map[int]string{
		1:  "001_init.sql",
		2:  "002_users.sql",
		20: "020_user_credentials.sql",
	} {
		sqlStr, err := ReadMigration(filepath.Join("..", "..", "migrations", file))
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}
		migrations[v] = sqlStr
	}
	db, err := Open(":memory:", migrations)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func seedUser(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	if err := NewUserRepo(db).Insert(User{ID: id, Email: id + "@example.com", PasswordHash: "x", CreatedAt: 1}); err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

func cred(id, userID, provider string, isDefault bool) Credential {
	return Credential{
		ID: id, UserID: userID, Provider: provider, Model: "m1",
		Ciphertext: []byte("ct"), Nonce: []byte("nonce"), KeyHint: "1234",
		IsDefault: isDefault, CreatedAt: 10, UpdatedAt: 10,
	}
}

func TestCredentialRepoInsertAndGet(t *testing.T) {
	db := openCredentialTestDB(t)
	seedUser(t, db, "u1")
	repo := NewCredentialRepo(db)

	if err := repo.Insert(cred("c1", "u1", "gemini", true)); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, err := repo.GetByID("u1", "c1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.UserID != "u1" || got.Provider != "gemini" || got.KeyHint != "1234" {
		t.Fatalf("unexpected credential: %+v", got)
	}
	if !got.IsDefault {
		t.Fatal("expected is_default true")
	}
	if string(got.Ciphertext) != "ct" || string(got.Nonce) != "nonce" {
		t.Fatalf("ciphertext roundtrip failed: %q %q", got.Ciphertext, got.Nonce)
	}

	if _, err := repo.GetByID("u1", "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

// TestCredentialRepoGetByIDIsUserScoped is the IDOR guard: a credential must
// be unreadable through another user's id even if the id is known.
func TestCredentialRepoGetByIDIsUserScoped(t *testing.T) {
	db := openCredentialTestDB(t)
	seedUser(t, db, "u1")
	seedUser(t, db, "u2")
	repo := NewCredentialRepo(db)

	if err := repo.Insert(cred("c1", "u1", "gemini", true)); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if _, err := repo.GetByID("u2", "c1"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-user read should be ErrNoRows, got %v", err)
	}
	if err := repo.Delete("u2", "c1"); err != nil {
		t.Fatalf("cross-user delete: %v", err)
	}
	if _, err := repo.GetByID("u1", "c1"); err != nil {
		t.Fatalf("cross-user delete must not remove the row: %v", err)
	}
}

func TestCredentialRepoListByUser(t *testing.T) {
	db := openCredentialTestDB(t)
	seedUser(t, db, "u1")
	seedUser(t, db, "u2")
	repo := NewCredentialRepo(db)

	for _, c := range []Credential{cred("c1", "u1", "gemini", false), cred("c2", "u1", "anthropic", false)} {
		if err := repo.Insert(c); err != nil {
			t.Fatalf("insert %s: %v", c.ID, err)
		}
	}
	if err := repo.Insert(cred("c3", "u2", "openai", false)); err != nil {
		t.Fatalf("insert c3: %v", err)
	}

	got, err := repo.ListByUser("u1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 credentials for u1, got %d", len(got))
	}
	for _, c := range got {
		if c.UserID != "u1" {
			t.Fatalf("leaked another user's credential: %+v", c)
		}
	}

	empty, err := repo.ListByUser("nobody")
	if err != nil {
		t.Fatalf("list nobody: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty list, got %d", len(empty))
	}
}

func TestCredentialRepoGetDefault(t *testing.T) {
	db := openCredentialTestDB(t)
	seedUser(t, db, "u1")
	repo := NewCredentialRepo(db)

	if _, err := repo.GetDefault("u1"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected ErrNoRows with no credentials, got %v", err)
	}

	if err := repo.Insert(cred("c1", "u1", "gemini", false)); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if _, err := repo.GetDefault("u1"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected ErrNoRows when none flagged default, got %v", err)
	}

	if err := repo.SetDefault("u1", "c1"); err != nil {
		t.Fatalf("set default: %v", err)
	}
	got, err := repo.GetDefault("u1")
	if err != nil {
		t.Fatalf("get default: %v", err)
	}
	if got.ID != "c1" {
		t.Fatalf("expected c1, got %s", got.ID)
	}
}

// TestCredentialRepoSetDefaultIsExclusive proves the one-default invariant:
// promoting a second credential must demote the first, never leave two.
func TestCredentialRepoSetDefaultIsExclusive(t *testing.T) {
	db := openCredentialTestDB(t)
	seedUser(t, db, "u1")
	repo := NewCredentialRepo(db)

	for _, id := range []string{"c1", "c2"} {
		if err := repo.Insert(cred(id, "u1", "gemini", id == "c1")); err != nil {
			t.Fatalf("insert %s: %v", id, err)
		}
	}
	if err := repo.SetDefault("u1", "c2"); err != nil {
		t.Fatalf("set default c2: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM user_credentials WHERE user_id = ? AND is_default = 1`, "u1").Scan(&count); err != nil {
		t.Fatalf("count defaults: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 default, got %d", count)
	}
	got, err := repo.GetDefault("u1")
	if err != nil {
		t.Fatalf("get default: %v", err)
	}
	if got.ID != "c2" {
		t.Fatalf("expected c2 as default, got %s", got.ID)
	}
}

// TestCredentialRepoSetDefaultUnknownID also proves a rejected promotion does
// not strip the user's existing default: ownership is verified before any
// write, so a bad id cannot leave the account with nothing to use.
func TestCredentialRepoSetDefaultUnknownID(t *testing.T) {
	db := openCredentialTestDB(t)
	seedUser(t, db, "u1")
	seedUser(t, db, "u2")
	repo := NewCredentialRepo(db)
	if err := repo.Insert(cred("mine", "u1", "gemini", true)); err != nil {
		t.Fatalf("insert mine: %v", err)
	}
	if err := repo.Insert(cred("theirs", "u2", "gemini", false)); err != nil {
		t.Fatalf("insert theirs: %v", err)
	}

	tests := []struct {
		name    string
		userID  string
		credID  string
		wantErr bool
	}{
		{"unknown credential", "u1", "nope", true},
		{"credential owned by another user", "u1", "theirs", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SetDefault(tt.userID, tt.credID)
			if tt.wantErr && !errors.Is(err, ErrCredentialNotFound) {
				t.Fatalf("expected ErrCredentialNotFound, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := repo.GetDefault("u1")
			if err != nil {
				t.Fatalf("existing default was lost: %v", err)
			}
			if got.ID != "mine" {
				t.Fatalf("expected default to stay 'mine', got %s", got.ID)
			}
		})
	}
}

func TestCredentialRepoDelete(t *testing.T) {
	db := openCredentialTestDB(t)
	seedUser(t, db, "u1")
	repo := NewCredentialRepo(db)
	if err := repo.Insert(cred("c1", "u1", "gemini", true)); err != nil {
		t.Fatalf("insert: %v", err)
	}

	if err := repo.Delete("u1", "c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID("u1", "c1"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected row gone, got %v", err)
	}
	if err := repo.Delete("u1", "c1"); err != nil {
		t.Fatalf("second delete should be a no-op: %v", err)
	}
}

// TestCredentialRepoCascadeOnUserDelete proves a deleted account takes its
// credentials with it — orphaned model keys must not survive their owner.
func TestCredentialRepoCascadeOnUserDelete(t *testing.T) {
	db := openCredentialTestDB(t)
	seedUser(t, db, "u1")
	repo := NewCredentialRepo(db)
	if err := repo.Insert(cred("c1", "u1", "gemini", true)); err != nil {
		t.Fatalf("insert: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM users WHERE id = ?`, "u1"); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM user_credentials`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected cascade delete, %d rows remain", count)
	}
}

// TestCredentialRepoRejectsOrphanUser proves the foreign key is enforced, so a
// bug upstream cannot attach a credential to a nonexistent account.
func TestCredentialRepoRejectsOrphanUser(t *testing.T) {
	db := openCredentialTestDB(t)
	err := NewCredentialRepo(db).Insert(cred("c1", "ghost", "gemini", true))
	if err == nil {
		t.Fatal("expected foreign key violation for unknown user")
	}
}
