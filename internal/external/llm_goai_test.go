package external

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/zendev-sh/goai"
)

// TestMapGoAIErrorDailyQuota is the case a typed status code cannot express.
// A 429 covers both a per-minute rate limit, which recovers, and exhausted
// daily quota, which does not. Under BYOK the exhausted quota is the user's,
// so retrying spends a whole budget to reach the same failure.
func TestMapGoAIErrorDailyQuota(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{"daily request quota", "You have exceeded your daily request quota"},
		{"quota exhausted", "Quota exhausted for this project"},
		{"mixed case", "DAILY QUOTA Reached"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &goai.APIError{
				Message:     tt.message,
				StatusCode:  http.StatusTooManyRequests,
				IsRetryable: true,
			}
			out := mapGoAIError(in)

			if isRetryable(out) {
				t.Fatal("daily quota must not be retryable")
			}
			if !strings.Contains(out.Error(), "quota") {
				t.Fatalf("error = %v, want it to mention quota", out)
			}
			// Must still wrap the original so the provider message is not lost.
			var apiErr *goai.APIError
			if !errors.As(out, &apiErr) {
				t.Fatal("mapped error no longer unwraps to *goai.APIError")
			}
		})
	}
}

// TestMapGoAIErrorPerMinuteRetries is the counterpart: a rate limit that will
// clear on its own must stay retryable and must carry Retry-After.
func TestMapGoAIErrorPerMinuteRetries(t *testing.T) {
	in := &goai.APIError{
		Message:         "Rate limit exceeded for requests",
		StatusCode:      http.StatusTooManyRequests,
		IsRetryable:     true,
		ResponseHeaders: map[string]string{"retry-after": "3"},
	}
	out := mapGoAIError(in)

	if !isRetryable(out) {
		t.Fatalf("per-minute 429 must stay retryable, got %v", out)
	}
	delay, ok := retryAfterFrom(out)
	if !ok || delay != 3*time.Second {
		t.Fatalf("Retry-After = %v (ok=%v), want 3s", delay, ok)
	}
}

func TestMapGoAIErrorServerErrors(t *testing.T) {
	for _, status := range []int{http.StatusServiceUnavailable, http.StatusInternalServerError, http.StatusBadGateway} {
		in := &goai.APIError{
			Message:     "upstream unavailable",
			StatusCode:  status,
			IsRetryable: true,
		}
		if out := mapGoAIError(in); !isRetryable(out) {
			t.Fatalf("status %d should be retryable, got %v", status, out)
		}
	}
}

func TestMapGoAIErrorNonRetryable(t *testing.T) {
	tests := []struct {
		name   string
		apiErr *goai.APIError
	}{
		{"bad request", &goai.APIError{Message: "bad", StatusCode: http.StatusBadRequest}},
		{"unauthorized", &goai.APIError{Message: "bad key", StatusCode: http.StatusUnauthorized}},
		{"forbidden", &goai.APIError{Message: "no access", StatusCode: http.StatusForbidden}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := mapGoAIError(tt.apiErr)
			if isRetryable(out) {
				t.Fatalf("%s must not be retryable, got %v", tt.name, out)
			}
			var apiErr *goai.APIError
			if !errors.As(out, &apiErr) {
				t.Fatal("non-retryable API error should pass through unchanged")
			}
		})
	}
}

// TestMapGoAIErrorPassesThroughNonAPIErrors covers context cancellation and
// context overflow, which carry no status code and must reach the caller's
// existing handling untouched.
func TestMapGoAIErrorPassesThroughNonAPIErrors(t *testing.T) {
	overflow := &goai.ContextOverflowError{Message: "prompt is too long"}
	if out := mapGoAIError(overflow); !errors.Is(out, overflow) {
		t.Fatal("context overflow should pass through unchanged")
	}
}

func TestIsDailyQuota(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{"daily quota", "You have exceeded your daily request quota", true},
		{"quota exhausted", "quota exhausted", true},
		{"per minute rate limit", "You have exceeded your per-minute request quota", false},
		{"no quota word", "daily limit reached", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDailyQuota(tt.message); got != tt.want {
				t.Fatalf("isDailyQuota(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestRetryAfterFromHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    time.Duration
	}{
		{"seconds", map[string]string{"retry-after": "5"}, 5 * time.Second},
		{"milliseconds", map[string]string{"retry-after-ms": "250"}, 250 * time.Millisecond},
		{"uppercase key", map[string]string{"Retry-After": "2"}, 2 * time.Second},
		{"mixed case key", map[string]string{"RETRY-AFTER": "4"}, 4 * time.Second},
		{"whitespace tolerated", map[string]string{"retry-after": " 6 "}, 6 * time.Second},
		{"nil headers", nil, 0},
		{"empty map", map[string]string{}, 0},
		{"absent", map[string]string{"content-type": "application/json"}, 0},
		{"garbage", map[string]string{"retry-after": "soon"}, 0},
		{"zero", map[string]string{"retry-after": "0"}, 0},
		{"negative", map[string]string{"retry-after": "-3"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := retryAfterFromHeaders(tt.headers); got != tt.want {
				t.Fatalf("retryAfterFromHeaders(%v) = %v, want %v", tt.headers, got, tt.want)
			}
		})
	}
}

// TestNewGoAIModelCoversEveryProvider guards against a provider that compiles
// but produces no model, which would surface as a nil dereference on first use
// rather than at wiring time.
func TestNewGoAIModelCoversEveryProvider(t *testing.T) {
	for _, p := range SupportedProviders {
		t.Run(string(p), func(t *testing.T) {
			m := newGoAIModel(p, p.DefaultModel(), "test-key", http.DefaultClient)
			if m == nil {
				t.Fatal("provider produced a nil model")
			}
			if m.ModelID() != p.DefaultModel() {
				t.Fatalf("ModelID = %q, want %q", m.ModelID(), p.DefaultModel())
			}
		})
	}
}

// TestTransportForBuildsEveryProvider proves the wiring reaches a real client for
// each supported provider, so a stored key is never dead on arrival.
func TestTransportForBuildsEveryProvider(t *testing.T) {
	tr := NewTransport()
	for _, p := range SupportedProviders {
		t.Run(string(p), func(t *testing.T) {
			llm, err := tr.For(p, "user-supplied-key", "")
			if err != nil {
				t.Fatalf("For(%s): %v", p, err)
			}
			if llm == nil || llm.Provider() != p {
				t.Fatalf("unexpected client: %+v", llm)
			}
			if llm.Model() != p.DefaultModel() {
				t.Fatalf("model = %q, want the provider default %q", llm.Model(), p.DefaultModel())
			}
		})
	}
}

// TestTransportForRejectsEmptyKey is a security control, not a validation
// nicety. Every GoAI provider falls back to a vendor environment variable when
// the key is empty, so an empty key reaching the backend would silently route a
// user's analysis through an operator credential. Transport.For is the only
// path to newGoAIModel, and it must refuse before that can happen.
func TestTransportForRejectsEmptyKey(t *testing.T) {
	tr := NewTransport()
	for _, p := range SupportedProviders {
		if _, err := tr.For(p, "", ""); err == nil {
			t.Fatalf("provider %q accepted an empty key", p)
		}
	}
}
