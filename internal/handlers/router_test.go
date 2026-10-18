package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"paperviz/internal/app/credentials"
	"paperviz/internal/external"
	"paperviz/internal/testutil"
)

// newTestRouter builds the real API router so route/CORS behaviour is asserted end to end.
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	db := testutil.OpenTestDB(t)
	return withTestIdentity(NewRouter(db, credentials.NewResolver(db, external.NewTransport(), testCipher(t))))
}

// TestAPIRootAndNoSPA asserts the binary serves the JSON API root and no longer falls back to an SPA shell.
func TestAPIRootAndNoSPA(t *testing.T) {
	handler := newTestRouter(t)

	t.Run("root_returns_json_healthcheck", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body: %q)", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		for _, want := range []string{`"status":"ok"`, `"service":"PaperViz API"`} {
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("body %q missing %q", rec.Body.String(), want)
			}
		}
	})

	// The SPA now lives on Cloudflare Pages; every non-API path must 404 as JSON.
	for _, target := range []string{"/index.html", "/upload", "/assets/app.js", "/api/nope"} {
		t.Run("no_spa_fallback_"+target, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want 404 (body: %q)", rec.Code, rec.Body.String())
			}
			body := rec.Body.String()
			if !strings.Contains(body, "not_found") {
				t.Errorf("body %q does not contain %q", body, "not_found")
			}
			if strings.Contains(body, "<html") {
				t.Errorf("body looks like an SPA shell: %q", body)
			}
		})
	}
}

// TestCORSMiddleware asserts the origin allowlist and preflight handling.
func TestCORSMiddleware(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://paperviz.pages.dev")

	tests := []struct {
		name        string
		origin      string
		wantAllowed bool
	}{
		{"allowlisted_pages_origin", "https://paperviz.pages.dev", true},
		{"always_allowed_local_dev", "http://localhost:5173", true},
		{"unknown_origin", "https://evil.example", false},
		{"no_origin_header", "", false},
	}

	handler := newTestRouter(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			got := rec.Header().Get("Access-Control-Allow-Origin")
			if tt.wantAllowed && got != tt.origin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.origin)
			}
			if !tt.wantAllowed && got != "" {
				t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
			}
		})
	}

	t.Run("preflight_returns_204_without_auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/documents", nil)
		req.Header.Set("Origin", "https://paperviz.pages.dev")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("preflight status = %d, want 204", rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "POST") {
			t.Errorf("Access-Control-Allow-Methods = %q, want it to contain POST", got)
		}
	})
}

func TestNoindexMiddleware(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name      string
		next      http.Handler
		wantValue string
	}{
		{"through_middleware_sets_header", noindexMiddleware(okHandler), "noindex, nofollow"},
		{"plain_handler_has_no_header", okHandler, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.next.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/share/doc/tok", nil))
			if got := rec.Header().Get("X-Robots-Tag"); got != tt.wantValue {
				t.Errorf("X-Robots-Tag = %q, want %q", got, tt.wantValue)
			}
		})
	}
}
