package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"

	"paperviz/internal/app/credentials"
	"paperviz/internal/external"
	"paperviz/internal/repository"
)

// newOwnershipRouterDB opens an in-memory SQLite with the migrations a
// document read needs.
func newOwnershipRouterDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := repository.Open(":memory:", filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// testIdentityHeader carries the Clerk subject a request should be treated as
// having. It stands in for a verified Clerk session token.
const testIdentityHeader = "X-Test-Clerk-Subject"

// withTestIdentity stamps verified-looking Clerk claims onto the request.
//
// The real token is signed by Clerk and cannot be forged in a unit test, but
// clerk.ContextWithSessionClaims is the exact function Clerk's own middleware
// calls after verification. Injecting here exercises AuthMiddleware and the
// route wiring without pulling a Clerk instance into the test.
func withTestIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subject := r.Header.Get(testIdentityHeader)
		if subject == "" {
			next.ServeHTTP(w, r)
			return
		}
		claims := &clerk.SessionClaims{RegisteredClaims: clerk.RegisteredClaims{Subject: subject}}
		ctx := clerk.ContextWithSessionClaims(context.WithValue(r.Context(), userIDKey, subject), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TestGetDocumentRespectsOwnershipOverRouter drives the real router. The
// service-level ownership test cannot catch a missing OptionalAuth on the
// route: without it the session never reaches the context, the owner is read
// as anonymous, and the owner is locked out of their own document.
func TestGetDocumentRespectsOwnershipOverRouter(t *testing.T) {
	db := newOwnershipRouterDB(t)
	users := repository.NewUserRepo(db)
	docs := repository.NewDocumentRepo(db)

	owner, other := "user_router_owner", "user_router_other"
	for _, u := range []string{owner, other} {
		if err := users.UpsertClerkUser(u, u+"@test.com", time.Now().Unix()); err != nil {
			t.Fatalf("provision user %s: %v", u, err)
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

	router := withTestIdentity(NewRouter(db, credentials.NewResolver(db, external.NewTransport(), testCipher(t)), t.TempDir()))

	tests := []struct {
		name       string
		subject    string
		wantStatus int
	}{
		{name: "owner may read", subject: owner, wantStatus: http.StatusOK},
		{name: "other user is forbidden", subject: other, wantStatus: http.StatusForbidden},
		{name: "anonymous is forbidden on owned document", subject: "", wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/documents/doc-1", nil)
			if tt.subject != "" {
				req.Header.Set(testIdentityHeader, tt.subject)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

// TestRequireAuthRejectsRequestWithoutClerkToken keeps RequireAuth from
// accepting a request that merely carries no session at all.
func TestRequireAuthRejectsRequestWithoutClerkToken(t *testing.T) {
	db := newOwnershipRouterDB(t)
	router := withTestIdentity(NewRouter(db, credentials.NewResolver(db, external.NewTransport(), testCipher(t)), t.TempDir()))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/documents/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 (body %s)", rec.Code, rec.Body.String())
	}
}
