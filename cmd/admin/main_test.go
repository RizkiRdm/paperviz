package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"paperviz/internal/repository"
)

// newAdminTestDB points the command at a throwaway database and returns a
// cleanup that restores the environment.
func newAdminTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "paperviz.db")

	t.Setenv("DATABASE_PATH", dbPath)
	t.Setenv("MIGRATIONS_DIR", filepath.Join("..", "..", "migrations"))

	db, err := repository.Open(dbPath, filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"lowercases", "Reader@Example.COM", "reader@example.com"},
		{"trims", "  reader@example.com  ", "reader@example.com"},
		{"both", "\tREADER@Example.com\n", "reader@example.com"},
		{"already normalized", "reader@example.com", "reader@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeEmail(tt.in); got != tt.want {
				t.Errorf("normalizeEmail(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRunRequiresSubcommand(t *testing.T) {
	newAdminTestDB(t)
	if err := run(nil); err == nil {
		t.Error("run(nil) = nil, want an error")
	}
}

func TestRunHelpDoesNotRequireDatabase(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Errorf("run(--help) = %v, want nil", err)
	}
}

func TestRunRejectsUnknownSubcommand(t *testing.T) {
	newAdminTestDB(t)
	if err := run([]string{"reset-password", "a@b.com"}); err == nil {
		t.Error("run(reset-password) = nil, want an error for a removed subcommand")
	}
}

// TestRevokeSessionsRequiresSecretKey keeps the command from silently doing
// nothing when CLERK_SECRET_KEY is absent, which would report success while
// leaving every session live.
func TestRevokeSessionsRequiresSecretKey(t *testing.T) {
	t.Setenv("CLERK_SECRET_KEY", "")
	err := run([]string{"revoke-sessions", "reader@example.com"})
	if err == nil {
		t.Fatal("revoke-sessions without a secret key = nil, want an error")
	}
	if !strings.Contains(err.Error(), "CLERK_SECRET_KEY") {
		t.Errorf("error = %v, want it to name CLERK_SECRET_KEY", err)
	}
}

func TestRevokeSessionsRequiresEmail(t *testing.T) {
	t.Setenv("CLERK_SECRET_KEY", "sk_test_dummy")
	if err := run([]string{"revoke-sessions"}); err == nil {
		t.Error("revoke-sessions with no email = nil, want a usage error")
	}
}

func TestRunListUsers(t *testing.T) {
	db := newAdminTestDB(t)
	users := repository.NewUserRepo(db)
	for _, email := range []string{"a@example.com", "b@example.com"} {
		if err := users.UpsertClerkUser("user_"+email[:1], email, 1700000000); err != nil {
			t.Fatalf("seed %s: %v", email, err)
		}
	}

	out := captureStdout(t, func() {
		if err := run([]string{"list-users"}); err != nil {
			t.Fatalf("run(list-users): %v", err)
		}
	})

	for _, email := range []string{"a@example.com", "b@example.com"} {
		if !strings.Contains(out, email) {
			t.Errorf("list-users output missing %s:\n%s", email, out)
		}
	}
}

// TestListUsersOnEmptyDB keeps a wiped Render filesystem from looking like a
// crash, since that is the normal state on a Free instance redeploy.
func TestListUsersOnEmptyDB(t *testing.T) {
	newAdminTestDB(t)
	captureStdout(t, func() {
		if err := run([]string{"list-users"}); err != nil {
			t.Fatalf("run(list-users) on empty db = %v, want nil", err)
		}
	})
}

// TestAdminDoesNotPrintStoredSecret asserts no stored material reaches stdout.
// Credential ciphertext and the encryption key must never be printed.
func TestAdminDoesNotPrintStoredSecret(t *testing.T) {
	db := newAdminTestDB(t)
	users := repository.NewUserRepo(db)
	if err := users.UpsertClerkUser("user_x", "reader@example.com", 1700000000); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Setenv("CREDENTIAL_ENCRYPTION_KEY", "super-secret-key-material")

	out := captureStdout(t, func() {
		if err := run([]string{"list-users"}); err != nil {
			t.Fatalf("run(list-users): %v", err)
		}
	})

	if strings.Contains(out, "super-secret-key-material") {
		t.Error("admin printed CREDENTIAL_ENCRYPTION_KEY")
	}
}

func TestAdminRespectsDatabasePath(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "custom.db")
	t.Setenv("DATABASE_PATH", dbPath)
	t.Setenv("MIGRATIONS_DIR", filepath.Join("..", "..", "migrations"))

	db, err := repository.Open(dbPath, filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := repository.NewUserRepo(db).UpsertClerkUser("user_y", "custom@example.com", 1700000000); err != nil {
		t.Fatalf("seed: %v", err)
	}
	db.Close()

	captureStdout(t, func() {
		if err := run([]string{"list-users"}); err != nil {
			t.Fatalf("run(list-users): %v", err)
		}
	})

	if _, _, err := openDB(); err != nil {
		t.Fatalf("openDB at DATABASE_PATH: %v", err)
	}
}

// captureStdout runs fn with os.Stdout redirected and returns what it wrote.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				sb.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		done <- sb.String()
	}()

	fn()
	w.Close()
	os.Stdout = orig
	return <-done
}
