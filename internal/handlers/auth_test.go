package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"paperviz/internal/repository"
)

// newAuthTestDB opens an in-memory SQLite with the migrations needed by auth handlers.
func newAuthTestDB(t *testing.T) *sql.DB {
	t.Helper()
	migrations := make(map[int]string)
	for v, file := range map[int]string{
		1:  "001_init.sql",
		2:  "002_users.sql",
		16: "016_annotations.sql",
		17: "017_oauth_columns.sql",
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
	return db
}

// TestLoginSessionSurvives is the BUG-01 regression: a session created by
// Login must not be deleted by the login flow itself (session invalidation
// on re-login deletes OLD sessions only, before the new one is created).
func TestLoginSessionSurvives(t *testing.T) {
	db := newAuthTestDB(t)
	auth := NewAuthHandler(db)

	// Signup a user (creates session too).
	signupBody, _ := json.Marshal(map[string]string{
		"email":    "bug01@test.com",
		"password": "Sup3rSecure!",
	})
	signup := httptest.NewRecorder()
	auth.Signup(signup, httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewReader(signupBody)))
	if signup.Code != http.StatusCreated {
		t.Fatalf("signup status = %d, want 201 (body: %s)", signup.Code, signup.Body.String())
	}

	// Login must leave a valid session behind.
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "bug01@test.com",
		"password": "Sup3rSecure!",
	})
	login := httptest.NewRecorder()
	auth.Login(login, httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody)))
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200 (body: %s)", login.Code, login.Body.String())
	}

	// The session cookie from login must authenticate /api/auth/me.
	var cookie *http.Cookie
	for _, c := range login.Result().Cookies() {
		if c.Name == "session_token" && c.Value != "" {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatalf("login did not set session_token cookie (headers: %v)", login.Header())
	}

	me := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.AddCookie(cookie)
	auth.Me(me, meReq)
	if me.Code != http.StatusOK {
		t.Fatalf("me after login status = %d, want 200 (BUG-01: session self-destructed)", me.Code)
	}

	// DB-level guarantee: exactly one session remains for the user.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 1 {
		t.Fatalf("sessions count = %d, want 1", count)
	}
}
