package handlers

import (
	"net/http"
	"testing"
)

// TestCookieSecure covers the COOKIE_SECURE flag, which decides whether the
// session cookie carries the Secure attribute. A Secure cookie on a plain-HTTP
// origin is silently dropped by the browser: login appears to succeed and every
// later request 401s.
func TestCookieSecure(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		want  bool
	}{
		{name: "unset defaults to secure", set: false, want: true},
		{name: "explicit true", value: "true", set: true, want: true},
		{name: "explicit false for plain http", value: "false", set: true, want: false},
		{name: "1 is true", value: "1", set: true, want: true},
		{name: "0 is false", value: "0", set: true, want: false},
		{name: "invalid value falls back to secure", value: "yes-please", set: true, want: true},
		{name: "empty string is treated as unset", value: "", set: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv("COOKIE_SECURE", tt.value)
			}
			if got := cookieSecure(); got != tt.want {
				t.Errorf("cookieSecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSessionCookie_Attributes pins the attributes a browser depends on. The
// logout cookie must match these exactly or the browser keeps the original.
func TestSessionCookie_Attributes(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "false")

	c := sessionCookie("token-value", 3600)

	if c.Name != "session_token" {
		t.Errorf("Name = %q, want session_token", c.Name)
	}
	if c.Value != "token-value" {
		t.Errorf("Value = %q, want token-value", c.Value)
	}
	if c.Path != "/" {
		t.Errorf("Path = %q, want /", c.Path)
	}
	if c.MaxAge != 3600 {
		t.Errorf("MaxAge = %d, want 3600", c.MaxAge)
	}
	if !c.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", c.SameSite)
	}
	if c.Secure {
		t.Error("Secure must follow COOKIE_SECURE=false")
	}

	t.Setenv("COOKIE_SECURE", "true")
	if secure := sessionCookie("token-value", 3600); !secure.Secure {
		t.Error("Secure must be set when COOKIE_SECURE=true")
	}
}
