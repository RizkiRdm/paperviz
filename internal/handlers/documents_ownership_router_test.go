package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"

	"paperviz/internal/app/credentials"
	"paperviz/internal/external"
	"paperviz/internal/testutil"
)

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
	db := testutil.OpenTestDB(t)

	owner, other := "user_router_owner", "user_router_other"
	now := time.Now().Unix()
	testutil.SeedClerkUser(t, db, owner, owner+"@test.com", now)
	testutil.SeedClerkUser(t, db, other, other+"@test.com", now)
	testutil.SeedOwnedDocument(t, db, "doc-1", owner)

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

// TestSectionRoutesRespectOwnershipOverRouter guards the Wave G fix: section
// reads and relationship writes once bypassed auth, so any foreign or anonymous
// caller could read claims or post relationships on someone else's document.
func TestSectionRoutesRespectOwnershipOverRouter(t *testing.T) {
	db := testutil.OpenTestDB(t)

	owner, other := "user_sect_owner", "user_sect_other"
	now := time.Now().Unix()
	testutil.SeedClerkUser(t, db, owner, owner+"@test.com", now)
	testutil.SeedClerkUser(t, db, other, other+"@test.com", now)
	testutil.SeedOwnedDocument(t, db, "doc-sect", owner)

	router := withTestIdentity(NewRouter(db, credentials.NewResolver(db, external.NewTransport(), testCipher(t)), t.TempDir()))

	tests := []struct {
		name       string
		method     string
		path       string
		subject    string
		wantStatus int
	}{
		{name: "owner may read claims", method: http.MethodGet, path: "/api/documents/doc-sect/claims", subject: owner, wantStatus: http.StatusOK},
		{name: "other user forbidden from claims", method: http.MethodGet, path: "/api/documents/doc-sect/claims", subject: other, wantStatus: http.StatusForbidden},
		{name: "anonymous forbidden from claims", method: http.MethodGet, path: "/api/documents/doc-sect/claims", subject: "", wantStatus: http.StatusForbidden},
		{name: "other user forbidden from research map", method: http.MethodGet, path: "/api/documents/doc-sect/research-map", subject: other, wantStatus: http.StatusForbidden},
		{name: "anonymous relationship write rejected", method: http.MethodPost, path: "/api/papers/doc-sect/relationships", subject: "", wantStatus: http.StatusUnauthorized},
		{name: "other user forbidden from relationship write", method: http.MethodPost, path: "/api/papers/doc-sect/relationships", subject: other, wantStatus: http.StatusForbidden},
		{name: "anonymous compare rejected", method: http.MethodPost, path: "/api/documents/compare", subject: "", wantStatus: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
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
	db := testutil.OpenTestDB(t)
	router := withTestIdentity(NewRouter(db, credentials.NewResolver(db, external.NewTransport(), testCipher(t)), t.TempDir()))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/documents/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 (body %s)", rec.Code, rec.Body.String())
	}
}
