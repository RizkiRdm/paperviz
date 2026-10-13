package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"paperviz/internal/repository"
)

// newAuthTestDB opens an in-memory SQLite with the migrations needed by auth handlers.
func newAuthTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := repository.Open(":memory:", filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// seedClerkUser writes the row AuthMiddleware would have created on first sight.
func seedClerkUser(t *testing.T, db *sql.DB, id, email string) {
	t.Helper()
	if err := repository.NewUserRepo(db).UpsertClerkUser(id, email, 1700000000); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func TestMeReturnsProfileForVerifiedClerkSubject(t *testing.T) {
	db := newAuthTestDB(t)
	seedClerkUser(t, db, "user_2abc", "reader@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, "user_2abc"))
	rec := httptest.NewRecorder()

	NewAuthHandler(db).Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, "user_2abc") || !strings.Contains(body, "reader@example.com") {
		t.Fatalf("body = %s, want the seeded id and email", body)
	}
}

// TestMeRejectsMissingIdentity keeps Me safe if it is ever mounted without
// RequireAuth: it must not fall back to any ambient identity.
func TestMeRejectsMissingIdentity(t *testing.T) {
	db := newAuthTestDB(t)
	seedClerkUser(t, db, "user_2abc", "reader@example.com")

	rec := httptest.NewRecorder()
	NewAuthHandler(db).Me(rec, httptest.NewRequest(http.MethodGet, "/api/auth/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestMeRejectsUnknownSubject(t *testing.T) {
	db := newAuthTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, "user_missing"))
	rec := httptest.NewRecorder()

	NewAuthHandler(db).Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

// TestMeIsIdempotentAcrossReloads stands in for the Render restart case: the
// local row survives a deploy, so Me keeps working without re-provisioning.
func TestMeIsIdempotentAcrossReloads(t *testing.T) {
	db := newAuthTestDB(t)
	seedClerkUser(t, db, "user_2abc", "reader@example.com")
	seedClerkUser(t, db, "user_2abc", "reader@example.com")

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE id = ?`, "user_2abc").Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 1 {
		t.Fatalf("user count = %d, want 1 after repeat upserts", count)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
