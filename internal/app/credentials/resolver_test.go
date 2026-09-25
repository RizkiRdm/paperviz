package credentials

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"paperviz/internal/external"
	"paperviz/internal/repository"
)

const testKey = "AIzaSyD-SUPER-SECRET-USER-KEY-9999"

func newTestResolver(t *testing.T) (*Resolver, *sql.DB, *external.Cipher) {
	t.Helper()
	migrations := make(map[int]string)
	for v, file := range map[int]string{
		1:  "001_init.sql",
		2:  "002_users.sql",
		20: "020_user_credentials.sql",
	} {
		sqlStr, err := repository.ReadMigration(filepath.Join("..", "..", "..", "migrations", file))
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}
		migrations[v] = sqlStr
	}
	db, err := repository.Open(":memory:", migrations)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	cipherKey := make([]byte, 32)
	for i := range cipherKey {
		cipherKey[i] = byte(i)
	}
	cipher, err := external.NewCipher(cipherKey)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	return NewResolver(db, external.NewTransport(), cipher), db, cipher
}

func seedUser(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	if err := repository.NewUserRepo(db).Insert(repository.User{
		ID: id, Email: id + "@example.com", PasswordHash: "x", CreatedAt: 1,
	}); err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

// addCredential stores an encrypted credential for a user.
func addCredential(t *testing.T, db *sql.DB, cipher *external.Cipher, userID, provider, model, key string, isDefault bool) string {
	t.Helper()
	p := external.Provider(provider)
	if model == "" {
		model = p.DefaultModel()
	}
	ct, nonce, err := cipher.Encrypt([]byte(key), external.CredentialAAD(userID, p, model))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	id := "cred-" + userID + "-" + provider
	now := time.Now().Unix()
	if err := repository.NewCredentialRepo(db).Insert(repository.Credential{
		ID: id, UserID: userID, Provider: provider, Model: model,
		Ciphertext: ct, Nonce: nonce, KeyHint: external.KeyHint(key),
		IsDefault: isDefault, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("insert credential: %v", err)
	}
	return id
}

func TestResolverReturnsClientForDefaultCredential(t *testing.T) {
	r, db, cipher := newTestResolver(t)
	seedUser(t, db, "u1")
	addCredential(t, db, cipher, "u1", "gemini", "", testKey, true)

	client, err := r.For(context.Background(), "u1")
	if err != nil {
		t.Fatalf("For: %v", err)
	}
	if client == nil {
		t.Fatal("expected a client")
	}
	if client.Provider() != external.ProviderGemini {
		t.Errorf("provider = %q, want gemini", client.Provider())
	}
	if client.Model() != external.ProviderGemini.DefaultModel() {
		t.Errorf("model = %q, want %q", client.Model(), external.ProviderGemini.DefaultModel())
	}
}

func TestResolverHonoursStoredModel(t *testing.T) {
	r, db, cipher := newTestResolver(t)
	seedUser(t, db, "u1")
	addCredential(t, db, cipher, "u1", "gemini", "gemini-3.5-flash", testKey, true)

	client, err := r.For(context.Background(), "u1")
	if err != nil {
		t.Fatalf("For: %v", err)
	}
	if client.Model() != "gemini-3.5-flash" {
		t.Fatalf("model = %q, want gemini-3.5-flash", client.Model())
	}
}

func TestResolverMissingCredential(t *testing.T) {
	r, db, cipher := newTestResolver(t)
	seedUser(t, db, "u1")
	seedUser(t, db, "u2")
	// u2 has a credential that is not flagged default, so GetDefault finds nothing.
	addCredential(t, db, cipher, "u2", "gemini", "", testKey, false)

	tests := []struct {
		name   string
		userID string
	}{
		{"empty user id", ""},
		{"user with no credentials", "u1"},
		{"user whose only credential is not default", "u2"},
		{"unknown user", "ghost"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := r.For(context.Background(), tt.userID)
			if !errors.Is(err, ErrNoCredential) {
				t.Fatalf("expected ErrNoCredential, got %v", err)
			}
			if client != nil {
				t.Fatal("expected no client alongside the error")
			}
		})
	}
}

// TestResolverRejectsUnreadableCredential covers the silent-rotation case: the
// row exists but the encryption key no longer opens it.
func TestResolverRejectsUnreadableCredential(t *testing.T) {
	_, db, cipher := newTestResolver(t)
	seedUser(t, db, "u1")
	addCredential(t, db, cipher, "u1", "gemini", "", testKey, true)

	// Swap in a different cipher, as a CREDENTIAL_ENCRYPTION_KEY rotation would.
	otherKey := make([]byte, 32)
	for i := range otherKey {
		otherKey[i] = byte(255 - i)
	}
	other, err := external.NewCipher(otherKey)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	rotated := NewResolver(db, external.NewTransport(), other)

	if _, err := rotated.For(context.Background(), "u1"); !errors.Is(err, ErrCredentialUnreadable) {
		t.Fatalf("expected ErrCredentialUnreadable, got %v", err)
	}
}

// TestResolverCredentialIsUserScoped proves one user's key is never handed to
// another user's request.
func TestResolverCredentialIsUserScoped(t *testing.T) {
	r, db, cipher := newTestResolver(t)
	seedUser(t, db, "u1")
	seedUser(t, db, "u2")
	addCredential(t, db, cipher, "u1", "gemini", "gemini-3.5-flash", testKey, true)

	if _, err := r.For(context.Background(), "u2"); !errors.Is(err, ErrNoCredential) {
		t.Fatalf("u2 resolved a credential: %v", err)
	}
	client, err := r.For(context.Background(), "u1")
	if err != nil {
		t.Fatalf("u1 could not resolve: %v", err)
	}
	if client.Model() != "gemini-3.5-flash" {
		t.Fatalf("u1 got the wrong credential: model %q", client.Model())
	}
}

func TestResolverRejectsProviderWithoutBackend(t *testing.T) {
	r, db, cipher := newTestResolver(t)
	seedUser(t, db, "u1")
	// Valid provider, but this build has no backend wired for it yet.
	addCredential(t, db, cipher, "u1", "anthropic", "", testKey, true)

	if _, err := r.For(context.Background(), "u1"); err == nil {
		t.Fatal("expected an error for a provider with no backend")
	} else if errors.Is(err, ErrNoCredential) {
		t.Fatalf("wrong error class: %v", err)
	}
}
