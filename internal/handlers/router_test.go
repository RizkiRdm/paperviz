package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newStaticFixture builds a temp staticDir with a dummy index.html and assets/app.js.
func newStaticFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>spa-shell</html>"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log('app');"), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}
	return dir
}

func TestSpaNotFound(t *testing.T) {
	staticDir := newStaticFixture(t)
	fileServer := http.FileServer(http.Dir(staticDir))
	handler := spaNotFound(staticDir, fileServer)

	tests := []struct {
		name           string
		method         string
		target         string
		wantStatus     int
		wantCTPrefix   string
		wantBodySub    string
		wantBodyAbsent string
	}{
		{
			name:         "unknown_api_path_returns_json_404",
			method:       http.MethodGet,
			target:       "/api/nope",
			wantStatus:   http.StatusNotFound,
			wantCTPrefix: "application/json",
			wantBodySub:  "not_found",
		},
		{
			name:         "unknown_page_path_serves_index_html",
			method:       http.MethodGet,
			target:       "/research-paper-summarizer",
			wantStatus:   http.StatusOK,
			wantCTPrefix: "text/html",
			wantBodySub:  "spa-shell",
		},
		{
			name:         "deep_link_share_doc_serves_index_html",
			method:       http.MethodGet,
			target:       "/share/doc/tok123",
			wantStatus:   http.StatusOK,
			wantCTPrefix: "text/html",
			wantBodySub:  "spa-shell",
		},
		{
			name:         "existing_asset_served_from_disk",
			method:       http.MethodGet,
			target:       "/assets/app.js",
			wantStatus:   http.StatusOK,
			wantCTPrefix: "text/javascript",
			wantBodySub:  "console.log('app');",
		},
		{
			name:         "post_unknown_path_returns_json_404",
			method:       http.MethodPost,
			target:       "/nope",
			wantStatus:   http.StatusNotFound,
			wantCTPrefix: "application/json",
			wantBodySub:  "not_found",
		},
		{
			name:           "traversal_attempt_never_leaks_outside_static_dir",
			method:         http.MethodGet,
			target:         "/../../etc/passwd",
			wantStatus:     http.StatusNotFound,
			wantCTPrefix:   "application/json",
			wantBodySub:    "not_found",
			wantBodyAbsent: "root:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.target, nil))

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %q)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, tt.wantCTPrefix) {
				t.Errorf("Content-Type = %q, want prefix %q", ct, tt.wantCTPrefix)
			}
			if body := rec.Body.String(); !strings.Contains(body, tt.wantBodySub) {
				t.Errorf("body %q does not contain %q", body, tt.wantBodySub)
			}
			if tt.wantBodyAbsent != "" && strings.Contains(rec.Body.String(), tt.wantBodyAbsent) {
				t.Errorf("body leaked forbidden content %q: %q", tt.wantBodyAbsent, rec.Body.String())
			}
		})
	}
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

func TestSpaNotFoundHeadFallback(t *testing.T) {
	staticDir := newStaticFixture(t)
	handler := spaNotFound(staticDir, http.FileServer(http.Dir(staticDir)))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, "/dashboard", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("HEAD deep link status = %d, want %d", rec.Code, http.StatusOK)
	}
}
