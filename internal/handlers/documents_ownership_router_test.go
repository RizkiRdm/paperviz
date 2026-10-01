package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"paperviz/internal/app/credentials"
	"paperviz/internal/external"
	"paperviz/internal/repository"
)

// newOwnershipRouterDB opens an in-memory SQLite with the migrations a
// document read needs.
func newOwnershipRouterDB(t *testing.T) *sql.DB {
	t.Helper()
	migrations, err := repository.LoadMigrations(filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}
	db, err := repository.Open(":memory:", migrations)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// TestGetDocumentRespectsOwnershipOverRouter drives the real router. The
// service-level ownership test cannot catch a missing OptionalAuth on the
// route: without it the session never reaches the context, the owner is read
// as anonymous, and the owner is locked out of their own document.
func TestGetDocumentRespectsOwnershipOverRouter(t *testing.T) {
	db := newOwnershipRouterDB(t)
	users := repository.NewUserRepo(db)
	docs := repository.NewDocumentRepo(db)
	sessions := repository.NewSessionRepo(db)

	owner, other := "router-owner", "router-other"
	for _, u := range []string{owner, other} {
		if err := users.Insert(repository.User{ID: u, Email: u + "@test.com", PasswordHash: "h", CreatedAt: time.Now().Unix()}); err != nil {
			t.Fatalf("insert user %s: %v", u, err)
		}
		if err := sessions.Insert(repository.Session{Token: "tok-" + u, UserID: u, ExpiresAt: time.Now().Add(time.Hour).Unix()}); err != nil {
			t.Fatalf("insert session %s: %v", u, err)
		}
	}

	now := time.Now().Unix()
	ownerPtr := owner
	if err := docs.Insert(repository.Document{
		ID: "doc-1", CreatedAt: now, LastAccessedAt: now,
		Status: repository.StatusComplete, SourceType: repository.SourceTypePastedText,
		ReadingLevel: repository.ReadingLevelSimplified, OriginalText: "body", UserID: &ownerPtr,
	}); err != nil {
		t.Fatalf("insert document: %v", err)
	}

	router := NewRouter(db, credentials.NewResolver(db, external.NewTransport(), testCipher(t)), t.TempDir())

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{name: "owner may read", token: "tok-" + owner, wantStatus: http.StatusOK},
		{name: "other user is forbidden", token: "tok-" + other, wantStatus: http.StatusForbidden},
		{name: "anonymous is forbidden on owned document", token: "", wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/documents/doc-1", nil)
			if tt.token != "" {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: tt.token})
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
