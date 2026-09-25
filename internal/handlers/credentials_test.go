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

	"github.com/go-chi/chi/v5"

	"paperviz/internal/external"
	"paperviz/internal/repository"
)

const testModelKey = "AIzaSyD-SUPER-SECRET-USER-KEY-9999"

func openCredentialHandlerDB(t *testing.T) *sql.DB {
	t.Helper()
	migrations := make(map[int]string)
	for v, file := range map[int]string{
		1:  "001_init.sql",
		2:  "002_users.sql",
		20: "020_user_credentials.sql",
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

	for _, id := range []string{"u1", "u2"} {
		if err := repository.NewUserRepo(db).Insert(repository.User{
			ID: id, Email: id + "@example.com", PasswordHash: "x", CreatedAt: 1,
		}); err != nil {
			t.Fatalf("seed user %s: %v", id, err)
		}
	}
	return db
}

func testCipher(t *testing.T) *external.Cipher {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	c, err := external.NewCipher(key)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	return c
}

// asUser builds a request carrying a session user id, mirroring what
// AuthMiddleware injects.
func asUser(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
}

// withChiID injects a chi route context carrying {id}, so chi.URLParam
// resolves in a handler test without standing up a router.
func withChiID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func doJSON(t *testing.T, h http.HandlerFunc, method, body, userID string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, "/api/credentials", nil)
	} else {
		r = httptest.NewRequest(method, "/api/credentials", strings.NewReader(body))
	}
	if userID != "" {
		r = asUser(r, userID)
	}
	w := httptest.NewRecorder()
	h(w, r)
	return w
}

// TestCredentialCreateNeverEchoesKey is the core BYOK guarantee: the response
// must carry the hint and nothing else that could reconstruct the key.
func TestCredentialCreateNeverEchoesKey(t *testing.T) {
	db := openCredentialHandlerDB(t)
	h := NewCredentialHandler(db, testCipher(t))

	body := `{"provider":"gemini","api_key":"` + testModelKey + `"}`
	w := doJSON(t, h.Create, http.MethodPost, body, "u1")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body %s", w.Code, w.Body.String())
	}

	raw := w.Body.String()
	if strings.Contains(raw, testModelKey) {
		t.Fatalf("response leaked the plaintext key: %s", raw)
	}
	if strings.Contains(raw, "ciphertext") || strings.Contains(raw, "nonce") {
		t.Fatalf("response leaked stored secret material: %s", raw)
	}

	var got credentialResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.KeyHint != "9999" {
		t.Fatalf("key_hint = %q, want %q", got.KeyHint, "9999")
	}
	if got.ID == "" || got.Provider != "gemini" || got.Model == "" {
		t.Fatalf("incomplete response: %+v", got)
	}
	if !got.IsDefault {
		t.Fatal("first credential must become the default")
	}
}

// TestStoredCredentialDecrypts proves the handler and the read path derive the
// same AAD. A mismatch here makes every stored key permanently unreadable.
func TestStoredCredentialDecrypts(t *testing.T) {
	db := openCredentialHandlerDB(t)
	cipher := testCipher(t)
	h := NewCredentialHandler(db, cipher)

	w := doJSON(t, h.Create, http.MethodPost, `{"provider":"gemini","model":"gemini-2.5-flash","api_key":"`+testModelKey+`"}`, "u1")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; body %s", w.Code, w.Body.String())
	}

	stored, err := repository.NewCredentialRepo(db).GetDefault("u1")
	if err != nil {
		t.Fatalf("get default: %v", err)
	}
	plaintext, err := cipher.Decrypt(stored.Ciphertext, stored.Nonce,
		external.CredentialAAD(stored.UserID, external.Provider(stored.Provider), stored.Model))
	if err != nil {
		t.Fatalf("decrypt stored credential: %v", err)
	}
	if string(plaintext) != testModelKey {
		t.Fatalf("decrypted %q, want %q", plaintext, testModelKey)
	}
}

func TestCredentialCreateValidation(t *testing.T) {
	db := openCredentialHandlerDB(t)
	h := NewCredentialHandler(db, testCipher(t))

	tests := []struct {
		name     string
		body     string
		userID   string
		wantCode int
	}{
		{"unsupported provider", `{"provider":"llama","api_key":"k"}`, "u1", http.StatusBadRequest},
		{"empty provider", `{"provider":"","api_key":"k"}`, "u1", http.StatusBadRequest},
		{"missing api key", `{"provider":"gemini","api_key":""}`, "u1", http.StatusBadRequest},
		{"whitespace api key", `{"provider":"gemini","api_key":"   "}`, "u1", http.StatusBadRequest},
		{"malformed json", `{nope`, "u1", http.StatusBadRequest},
		{"unauthenticated", `{"provider":"gemini","api_key":"k"}`, "", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doJSON(t, h.Create, http.MethodPost, tt.body, tt.userID)
			if w.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d; body %s", w.Code, tt.wantCode, w.Body.String())
			}
		})
	}
}

// TestCredentialCreateDefaultsModel proves a key added without a model still
// produces a usable client.
func TestCredentialCreateDefaultsModel(t *testing.T) {
	db := openCredentialHandlerDB(t)
	h := NewCredentialHandler(db, testCipher(t))

	w := doJSON(t, h.Create, http.MethodPost, `{"provider":"anthropic","api_key":"`+testModelKey+`"}`, "u1")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; body %s", w.Code, w.Body.String())
	}
	var got credentialResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Model != external.ProviderAnthropic.DefaultModel() {
		t.Fatalf("model = %q, want %q", got.Model, external.ProviderAnthropic.DefaultModel())
	}
}

func TestCredentialListIsUserScoped(t *testing.T) {
	db := openCredentialHandlerDB(t)
	h := NewCredentialHandler(db, testCipher(t))

	doJSON(t, h.Create, http.MethodPost, `{"provider":"gemini","api_key":"`+testModelKey+`"}`, "u1")

	t.Run("owner sees own credential", func(t *testing.T) {
		w := doJSON(t, h.List, http.MethodGet, "", "u1")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if strings.Contains(w.Body.String(), testModelKey) {
			t.Fatalf("list leaked the key: %s", w.Body.String())
		}
		var got []credentialResponse
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 credential, got %d", len(got))
		}
	})

	t.Run("other user sees nothing", func(t *testing.T) {
		w := doJSON(t, h.List, http.MethodGet, "", "u2")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if got := strings.TrimSpace(w.Body.String()); got != "[]" {
			t.Fatalf("expected empty list for u2, got %s", got)
		}
	})

	t.Run("unauthenticated rejected", func(t *testing.T) {
		w := doJSON(t, h.List, http.MethodGet, "", "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
	})
}

// TestCredentialCrossUserAccessDenied covers the IDOR path: knowing another
// user's credential id must not let you promote or delete their key.
func TestCredentialCrossUserAccessDenied(t *testing.T) {
	db := openCredentialHandlerDB(t)
	h := NewCredentialHandler(db, testCipher(t))

	w := doJSON(t, h.Create, http.MethodPost, `{"provider":"gemini","api_key":"`+testModelKey+`"}`, "u1")
	var created credentialResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// u2 owns a credential of their own so the default-swap path is exercised.
	doJSON(t, h.Create, http.MethodPost, `{"provider":"openai","api_key":"`+testModelKey+`"}`, "u2")

	// u2 cannot delete u1's credential.
	r := asUser(httptest.NewRequest(http.MethodDelete, "/api/credentials/"+created.ID, nil), "u2")
	rec := httptest.NewRecorder()
	h.Delete(rec, withChiID(r, created.ID))
	if rec.Code != http.StatusOK {
		t.Fatalf("cross-user delete status = %d", rec.Code)
	}
	if _, err := repository.NewCredentialRepo(db).GetByID("u1", created.ID); err != nil {
		t.Fatalf("cross-user delete removed the row: %v", err)
	}

	// u2 cannot promote u1's credential to their own default.
	r2 := asUser(httptest.NewRequest(http.MethodPost, "/api/credentials/"+created.ID+"/default", nil), "u2")
	rec2 := httptest.NewRecorder()
	h.SetDefault(rec2, withChiID(r2, created.ID))
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("cross-user set-default status = %d, want 404", rec2.Code)
	}
}
