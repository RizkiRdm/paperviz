package handlers

import (
	"fmt"
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

	if err := requireDirectives(csp,
		"script-src 'self' https://us-assets.i.posthog.com",
		"connect-src 'self' https://us.i.posthog.com https://us-assets.i.posthog.com",
	); err != nil {
		t.Error(err)
	}

	if !strings.Contains(csp, "us.i.posthog.com") || !strings.Contains(csp, "us-assets.i.posthog.com") {
		t.Errorf("CSP must name both PostHog hosts; got: %s", csp)
	}
}

// TestSecurityHeadersCoversClerk guards the CSP entries Clerk depends on.
//
// This is the one that shipped broken. clerk.browser.js was refused, so
// useAuth().getToken never resolved, so useCurrentUser's fetch never settled,
// so `checked` stayed false and every "Sign in" link silently vanished from
// the rendered page. Nothing threw and /healthz stayed 200: a CSP that looks
// correct in a header dump while blocking auth is indistinguishable from a
// working one by inspection alone.
func TestSecurityHeadersCoversClerk(t *testing.T) {
	rec := httptest.NewRecorder()
	SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(
		rec, httptest.NewRequest(http.MethodGet, "/", nil))

	csp := rec.Header().Get("Content-Security-Policy")

	if err := requireDirectives(csp,
		"script-src 'self' https://us-assets.i.posthog.com https://*.clerk.accounts.dev https://cdn.protect.clerk.com https://*.client.protect.clerk.com",
		"connect-src 'self' https://us.i.posthog.com https://us-assets.i.posthog.com https://*.clerk.accounts.dev https://cdn.protect.clerk.com https://*.client.protect.clerk.com",
	); err != nil {
		t.Error(err)
	}

	// Clerk's bot protection serves its bootstrap from cdn.protect.clerk.com and
	// then probes an instance-specific subdomain (<hash>-l.client.protect.clerk.com,
	// plus -p for the fallback probe), not the accounts.dev FAPI. Both blocked the
	// sign-in path; with cdn allowed but the subdomain not, no auth request reached
	// Clerk at all and the form showed a generic error.
	for _, origin := range []string{"cdn.protect.clerk.com", "*.client.protect.clerk.com"} {
		if !strings.Contains(csp, origin) {
			t.Errorf("CSP must allow %s, which Clerk's bot protection loads; got: %s", origin, csp)
		}
	}

	// Clerk's FAPI origin is derived from the publishable key, so it must be a
	// wildcard — a hardcoded instance hostname would break the build the moment
	// the key changes.
	if !strings.Contains(csp, "*.clerk.accounts.dev") {
		t.Errorf("CSP must allow the Clerk FAPI wildcard; got: %s", csp)
	}

	// Clerk loads parts of itself in a blob-backed worker.
	if !strings.Contains(csp, "worker-src 'self' blob:") {
		t.Errorf("CSP must allow blob: in worker-src for Clerk; got: %s", csp)
	}

	// frontend/index.html links font files on cdn.fontshare.com, distinct from
	// the api.fontshare.com already allowed in font-src.
	if !strings.Contains(csp, "cdn.fontshare.com") {
		t.Errorf("CSP must allow cdn.fontshare.com, which index.html links directly; got: %s", csp)
	}
}

// requireDirectives checks each want appears as a complete directive value.
// Matching is by prefix because a directive legitimately gains origins over
// time — adding Clerk to script-src must not require rewriting the PostHog
// assertion. The failure mode being guarded against is "connect-src 'self'"
// with the required hosts missing, so the listed sources must be present and
// ordered ahead of any extras.
func requireDirectives(csp string, wants ...string) error {
	parts := strings.Split(csp, ";")
	for _, want := range wants {
		fields := strings.Fields(want)
		directive, sources := fields[0], fields[1:]

		var got []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if f := strings.Fields(p); len(f) > 0 && f[0] == directive {
				got = f[1:]
			}
		}
		if len(got) < len(sources) {
			return fmt.Errorf("CSP %s missing required sources %v; found %v; full policy: %s",
				directive, sources, got, csp)
		}
		for i, s := range sources {
			if got[i] != s {
				return fmt.Errorf("CSP %s source %d = %q, want %q; found %v; full policy: %s",
					directive, i, got[i], s, got, csp)
			}
		}
	}
	return nil
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
