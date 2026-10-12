package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSecurityHeadersCoversPostHog guards the CSP directives PostHog depends
// on. The analytics library loads, but without script-src and connect-src
// entries the browser refuses both its remote script and its event POST, so
// signup_completed, api_key_added, and paper_analysed never arrive while the
// CSP still looks healthy in a header dump.
func TestSecurityHeadersCoversPostHog(t *testing.T) {
	rec := httptest.NewRecorder()
	SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(
		rec, httptest.NewRequest(http.MethodGet, "/", nil))

	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("no Content-Security-Policy set")
	}

	for _, want := range []string{
		"script-src 'self' https://us-assets.i.posthog.com",
		"connect-src 'self' https://us.i.posthog.com https://us-assets.i.posthog.com",
	} {
		// Directive matching is exact-ish: check the directive and its value
		// appear together, because "connect-src 'self'" alone is the bug.
		parts := strings.Split(csp, ";")
		found := false
		for _, p := range parts {
			if strings.TrimSpace(p) == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CSP missing %q; full policy: %s", want, csp)
		}
	}

	if !strings.Contains(csp, "us.i.posthog.com") || !strings.Contains(csp, "us-assets.i.posthog.com") {
		t.Errorf("CSP must name both PostHog hosts; got: %s", csp)
	}
}

// TestSecurityHeadersBaseline keeps the existing baseline headers honest.
func TestSecurityHeadersBaseline(t *testing.T) {
	rec := httptest.NewRecorder()
	SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(
		rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want DENY", got)
	}
}
