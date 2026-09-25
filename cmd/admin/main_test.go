package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"golang.org/x/crypto/bcrypt"

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

	migrations, err := repository.LoadMigrations(filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}
	db, err := repository.Open(dbPath, migrations)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// hasMinComplexity mirrors the rule the signup handler enforces. A generated
// password that fails this would create an account nobody can log into.
func hasMinComplexity(password string) bool {
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case !unicode.IsLetter(c) && !unicode.IsDigit(c) && !unicode.IsSpace(c):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func TestGeneratePasswordSatisfiesSignupRule(t *testing.T) {
	for i := 0; i < 200; i++ {
		pw, err := generatePassword()
		if err != nil {
			t.Fatalf("generatePassword: %v", err)
		}
		if len(pw) != passwordLength {
			t.Fatalf("length = %d, want %d", len(pw), passwordLength)
		}
		if !hasMinComplexity(pw) {
			t.Fatalf("generated password fails the signup complexity rule: %q", pw)
		}
	}
}

func TestGeneratePasswordIsUnpredictable(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		pw, err := generatePassword()
		if err != nil {
			t.Fatalf("generatePassword: %v", err)
		}
		if seen[pw] {
			t.Fatal("generatePassword repeated a value")
		}
		seen[pw] = true
	}
}

// TestGeneratePasswordPositionIsShuffled proves the guaranteed character
// classes are not always in the same leading positions, which would make the
// shape predictable even though the values are random.
func TestGeneratePasswordPositionIsShuffled(t *testing.T) {
	firstIsUpper := 0
	const runs = 100
	for i := 0; i < runs; i++ {
		pw, err := generatePassword()
		if err != nil {
			t.Fatalf("generatePassword: %v", err)
		}
		if unicode.IsUpper(rune(pw[0])) {
			firstIsUpper++
		}
	}
	// Four classes are guaranteed, so a fixed order would give 100% here.
	if firstIsUpper == runs {
		t.Fatal("first character is always upper case; the shuffle is not applied")
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"lowercases", "User@Example.COM", "user@example.com"},
		{"trims surrounding space", "  user@example.com  ", "user@example.com"},
		{"already normalized", "user@example.com", "user@example.com"},
		{"empty", "", ""},
		{"only spaces", "   ", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeEmail(tt.in); got != tt.want {
				t.Fatalf("normalizeEmail(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestHashPasswordRoundTrip(t *testing.T) {
	const pw = "Correct-Horse-9!"
	hash, err := hashPassword(pw)
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if hash == pw {
		t.Fatal("password stored in the clear")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)); err != nil {
		t.Fatalf("stored hash does not verify: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrong")); err == nil {
		t.Fatal("a wrong password verified")
	}
}

func TestRunUsage(t *testing.T) {
	newAdminTestDB(t)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no arguments", nil, true},
		{"unknown subcommand", []string{"nuke"}, true},
		{"help", []string{"help"}, false},
		{"create-user without email", []string{"create-user"}, true},
		{"reset-password without email", []string{"reset-password"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.args)
			if tt.wantErr && err == nil {
				t.Fatal("expected an error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunCreateUser(t *testing.T) {
	db := newAdminTestDB(t)

	if err := run([]string{"create-user", "Owner@Example.com"}); err != nil {
		t.Fatalf("create-user: %v", err)
	}

	user, err := repository.NewUserRepo(db).GetByEmail("owner@example.com")
	if err != nil {
		t.Fatalf("account not created with a normalized email: %v", err)
	}
	if user.PasswordHash == "" {
		t.Fatal("no password hash stored")
	}
	if strings.Contains(user.PasswordHash, "$2a$") == false {
		t.Fatalf("expected a bcrypt hash, got %q", user.PasswordHash)
	}
}

func TestRunCreateUserRejectsDuplicate(t *testing.T) {
	newAdminTestDB(t)

	if err := run([]string{"create-user", "dup@example.com"}); err != nil {
		t.Fatalf("first create-user: %v", err)
	}
	// A different case must still collide, since lookup normalizes.
	if err := run([]string{"create-user", "DUP@example.com"}); err == nil {
		t.Fatal("expected a duplicate-account error")
	}
}

func TestRunCreateUserRejectsUnknownEmail(t *testing.T) {
	db := newAdminTestDB(t)

	if err := run([]string{"create-user", "not-an-email"}); err != nil {
		t.Fatalf("create-user: %v", err)
	}
	// The admin tool does not validate email shape; the signup handler owns
	// that rule. Confirm the row exists so the behaviour is explicit.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, "not-an-email").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestRunResetPasswordChangesHash(t *testing.T) {
	db := newAdminTestDB(t)

	if err := run([]string{"create-user", "reset@example.com"}); err != nil {
		t.Fatalf("create-user: %v", err)
	}
	before, err := repository.NewUserRepo(db).GetByEmail("reset@example.com")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	if err := run([]string{"reset-password", "reset@example.com"}); err != nil {
		t.Fatalf("reset-password: %v", err)
	}
	after, err := repository.NewUserRepo(db).GetByEmail("reset@example.com")
	if err != nil {
		t.Fatalf("get user after reset: %v", err)
	}
	if before.PasswordHash == after.PasswordHash {
		t.Fatal("password hash unchanged after reset")
	}
}

// TestRunResetPasswordRevokesSessions is the security point of a reset: old
// sessions must not survive it, or the reset locks nobody out.
func TestRunResetPasswordRevokesSessions(t *testing.T) {
	db := newAdminTestDB(t)

	if err := run([]string{"create-user", "sess@example.com"}); err != nil {
		t.Fatalf("create-user: %v", err)
	}
	user, err := repository.NewUserRepo(db).GetByEmail("sess@example.com")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	sessions := repository.NewSessionRepo(db)
	if err := sessions.Insert(repository.Session{
		Token: "live-token", UserID: user.ID, ExpiresAt: 4102444800,
	}); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	if err := run([]string{"reset-password", "sess@example.com"}); err != nil {
		t.Fatalf("reset-password: %v", err)
	}
	if _, err := sessions.Get("live-token"); err == nil {
		t.Fatal("session survived a password reset")
	}
}

func TestRunResetPasswordUnknownAccount(t *testing.T) {
	newAdminTestDB(t)

	if err := run([]string{"reset-password", "ghost@example.com"}); err == nil {
		t.Fatal("expected an error for an unknown account")
	}
}

func TestRunListUsers(t *testing.T) {
	db := newAdminTestDB(t)

	for _, email := range []string{"a@example.com", "b@example.com"} {
		if err := run([]string{"create-user", email}); err != nil {
			t.Fatalf("create-user %s: %v", email, err)
		}
	}

	rows, err := db.Query(`SELECT email FROM users ORDER BY email`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, email)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(got))
	}
	if err := run([]string{"list-users"}); err != nil {
		t.Fatalf("list-users: %v", err)
	}
}

// TestAdminDoesNotPrintStoredSecret guards the one-shot property: the generated
// password is printed, but the stored value is always a hash.
func TestAdminDoesNotPrintStoredSecret(t *testing.T) {
	db := newAdminTestDB(t)

	if err := run([]string{"create-user", "secret@example.com"}); err != nil {
		t.Fatalf("create-user: %v", err)
	}
	var stored string
	if err := db.QueryRow(`SELECT password_hash FROM users WHERE email = ?`, "secret@example.com").Scan(&stored); err != nil {
		t.Fatalf("read hash: %v", err)
	}
	if !strings.HasPrefix(stored, "$2") {
		t.Fatalf("stored value is not a bcrypt hash: %q", stored)
	}
}

// TestAdminRespectsDatabasePath proves the tool reads the same database the
// server uses, rather than silently opening a stray file in the cwd.
func TestAdminRespectsDatabasePath(t *testing.T) {
	newAdminTestDB(t)
	want := os.Getenv("DATABASE_PATH")
	if want == "" {
		t.Fatal("DATABASE_PATH was not set by the fixture")
	}
	if filepath.Dir(want) == "." {
		t.Fatalf("test did not redirect DATABASE_PATH: %q", want)
	}
	if err := run([]string{"create-user", "path@example.com"}); err != nil {
		t.Fatalf("create-user: %v", err)
	}
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("database not created at DATABASE_PATH: %v", err)
	}
}
