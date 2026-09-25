package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"paperviz/internal/repository"
)

func openApiKeyTestDB(t *testing.T) *sql.DB {
	t.Helper()
	migrations := make(map[int]string)
	// 021 drops the oauth and billing columns, so 017 and 019 must be loaded
	// first: a DROP cannot succeed on a column that was never created.
	for v, file := range map[int]string{
		1:  "001_init.sql",
		2:  "002_users.sql",
		17: "017_oauth_columns.sql",
		18: "018_api_key_column.sql",
		19: "019_billing_columns.sql",
		21: "021_drop_oauth_billing.sql",
	} {
		sqlStr, err := repository.ReadMigration(filepath.Join("..", "..", "migrations", file))
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

	if err := repository.NewUserRepo(db).Insert(repository.User{
		ID: "u1", Email: "u1@example.com", PasswordHash: "x", CreatedAt: 1,
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return db
}

func callApiKey(t *testing.T, h http.HandlerFunc, method, userID string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, "/api/auth/apikey", nil)
	if userID != "" {
		r = asUser(r, userID)
	}
	w := httptest.NewRecorder()
	h(w, r)
	return w
}

// TestApiKeyIsNeverStoredInPlaintext is the point of this change: a database
// dump must not yield a usable service key.
func TestApiKeyIsNeverStoredInPlaintext(t *testing.T) {
	db := openApiKeyTestDB(t)
	h := NewApiKeyHandler(db)

	w := callApiKey(t, h.CreateApiKey, http.MethodPost, "u1")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; body %s", w.Code, w.Body.String())
	}

	var created apiKeyCreated
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.HasPrefix(created.Key, "pv_") {
		t.Fatalf("key %q missing pv_ prefix", created.Key)
	}

	var stored string
	if err := db.QueryRow(`SELECT COALESCE(api_key_hash, '') FROM users WHERE id = ?`, "u1").Scan(&stored); err != nil {
		t.Fatalf("read stored key: %v", err)
	}
	if stored == "" {
		t.Fatal("no key stored")
	}
	if stored == created.Key {
		t.Fatal("plaintext key was stored in the users table")
	}
	if stored != hashApiKey(created.Key) {
		t.Fatalf("stored value is not the digest of the issued key: %q", stored)
	}
	if strings.Contains(stored, "pv_") {
		t.Fatalf("stored value looks like a plaintext key: %q", stored)
	}
}

// TestGetApiKeyNeverReturnsTheKey covers the old behaviour directly: GET used
// to echo the key on every call.
func TestGetApiKeyNeverReturnsTheKey(t *testing.T) {
	db := openApiKeyTestDB(t)
	h := NewApiKeyHandler(db)

	if w := callApiKey(t, h.GetApiKey, http.MethodGet, "u1"); w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	} else if strings.Contains(w.Body.String(), "pv_") {
		t.Fatalf("GET leaked a key before one was issued: %s", w.Body.String())
	} else {
		var state apiKeyState
		if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if state.Configured {
			t.Fatal("reported configured before any key was issued")
		}
	}

	issued := callApiKey(t, h.CreateApiKey, http.MethodPost, "u1")
	var created apiKeyCreated
	if err := json.Unmarshal(issued.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	w := callApiKey(t, h.GetApiKey, http.MethodGet, "u1")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if strings.Contains(w.Body.String(), created.Key) {
		t.Fatalf("GET returned the key: %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "pv_") {
		t.Fatalf("GET returned key-shaped material: %s", w.Body.String())
	}
	var state apiKeyState
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !state.Configured {
		t.Fatal("expected configured=true after issuing a key")
	}
}

// TestRegenerateInvalidatesPreviousKey proves a lost or exposed key can be
// replaced and the old one stops resolving.
func TestRegenerateInvalidatesPreviousKey(t *testing.T) {
	db := openApiKeyTestDB(t)
	h := NewApiKeyHandler(db)

	first := callApiKey(t, h.CreateApiKey, http.MethodPost, "u1")
	var firstKey apiKeyCreated
	if err := json.Unmarshal(first.Body.Bytes(), &firstKey); err != nil {
		t.Fatalf("decode first: %v", err)
	}

	second := callApiKey(t, h.RegenerateApiKey, http.MethodPost, "u1")
	var secondKey apiKeyCreated
	if err := json.Unmarshal(second.Body.Bytes(), &secondKey); err != nil {
		t.Fatalf("decode second: %v", err)
	}
	if firstKey.Key == secondKey.Key {
		t.Fatal("regenerate returned the same key")
	}

	if _, err := LookupUserByApiKey(context.Background(), db, firstKey.Key); err == nil {
		t.Fatal("the superseded key still resolves")
	}
	userID, err := LookupUserByApiKey(context.Background(), db, secondKey.Key)
	if err != nil {
		t.Fatalf("new key does not resolve: %v", err)
	}
	if userID != "u1" {
		t.Fatalf("resolved user %q, want u1", userID)
	}
}

func TestApiKeyUnauthenticated(t *testing.T) {
	db := openApiKeyTestDB(t)
	h := NewApiKeyHandler(db)

	tests := []struct {
		name    string
		handler http.HandlerFunc
		method  string
	}{
		{"get", h.GetApiKey, http.MethodGet},
		{"create", h.CreateApiKey, http.MethodPost},
		{"regenerate", h.RegenerateApiKey, http.MethodPost},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := callApiKey(t, tt.handler, tt.method, "")
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", w.Code)
			}
		})
	}
}

func TestHashApiKey(t *testing.T) {
	key := "pv_abc123"
	if hashApiKey(key) != hashApiKey(key) {
		t.Fatal("digest is not deterministic")
	}
	if hashApiKey(key) == hashApiKey(key+"x") {
		t.Fatal("digest collides on a one-character change")
	}
	if len(hashApiKey(key)) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(hashApiKey(key)))
	}
	if strings.Contains(hashApiKey(key), "pv_") {
		t.Fatal("digest embeds the key prefix")
	}
}

func TestGenerateApiKeyIsUnpredictable(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 64; i++ {
		k, err := generateApiKey()
		if err != nil {
			t.Fatalf("generateApiKey: %v", err)
		}
		if seen[k] {
			t.Fatal("generateApiKey repeated a value")
		}
		seen[k] = true
		if !strings.HasPrefix(k, "pv_") || len(k) != 3+64 {
			t.Fatalf("unexpected key shape: %q", k)
		}
	}
}
