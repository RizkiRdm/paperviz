package external

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const okGeminiBody = `{"candidates":[{"content":{"parts":[{"text":"ok"}]}}]}`

// geminiStub wires the Gemini backend to a local server and returns a client
// with a shrunken retry schedule. The protocol layer is what these tests
// cover; retry policy lives in llm_test.go and needs no HTTP.
func geminiStub(t *testing.T, handler http.HandlerFunc) *LLM {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	tr := fastTransport()
	tr.httpClient = ts.Client()
	endpoint := ts.URL + "/v1beta/models/test-model:generateContent"
	return tr.newLLM(ProviderGemini, "test-model", newGeminiCall("test-key", endpoint, tr.httpClient))
}

// TestGeminiCallSendsRequest pins the wire format: model in the path, key in
// the x-goog-api-key header, and generationConfig only when asked for.
func TestGeminiCallSendsRequest(t *testing.T) {
	tests := []struct {
		name         string
		asJSON       bool
		maxTokens    int
		wantMIME     string
		wantMaxToken int
	}{
		{"plain text", false, 0, "", 0},
		{"json mode", true, 0, "application/json", 0},
		{"token cap only", false, 2048, "", 2048},
		{"json mode with token cap", true, 512, "application/json", 512},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotKey, gotBody string
			srv := geminiStub(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotKey = r.Header.Get("x-goog-api-key")
				b, _ := io.ReadAll(r.Body)
				gotBody = string(b)
				w.Write([]byte(okGeminiBody))
			})

			if _, err := srv.Generate(context.Background(), "hello", tt.asJSON, tt.maxTokens); err != nil {
				t.Fatalf("Generate: %v", err)
			}

			if !strings.Contains(gotPath, "test-model") {
				t.Errorf("model missing from path: %s", gotPath)
			}
			if gotKey != "test-key" {
				t.Errorf("api key header = %q, want %q", gotKey, "test-key")
			}
			if !strings.Contains(gotBody, `"text":"hello"`) {
				t.Errorf("prompt missing from body: %s", gotBody)
			}

			var req struct {
				GenerationConfig *struct {
					ResponseMIMEType string `json:"responseMimeType"`
					MaxOutputTokens  *int   `json:"maxOutputTokens"`
				} `json:"generationConfig"`
			}
			if err := json.Unmarshal([]byte(gotBody), &req); err != nil {
				t.Fatalf("request body is not valid json: %v", err)
			}

			if tt.wantMIME == "" && tt.wantMaxToken == 0 {
				if req.GenerationConfig != nil {
					t.Errorf("generationConfig sent when neither flag set: %s", gotBody)
				}
				return
			}
			if req.GenerationConfig == nil {
				t.Fatalf("generationConfig missing from body: %s", gotBody)
			}
			if req.GenerationConfig.ResponseMIMEType != tt.wantMIME {
				t.Errorf("responseMimeType = %q, want %q", req.GenerationConfig.ResponseMIMEType, tt.wantMIME)
			}
			if tt.wantMaxToken == 0 {
				if req.GenerationConfig.MaxOutputTokens != nil {
					t.Errorf("maxOutputTokens set but not requested: %d", *req.GenerationConfig.MaxOutputTokens)
				}
			} else if req.GenerationConfig.MaxOutputTokens == nil || *req.GenerationConfig.MaxOutputTokens != tt.wantMaxToken {
				t.Errorf("maxOutputTokens = %v, want %d", req.GenerationConfig.MaxOutputTokens, tt.wantMaxToken)
			}
		})
	}
}

// TestGeminiCallRateLimitIsRetryable covers per-minute throttling: retryable,
// carrying the server's Retry-After.
func TestGeminiCallRateLimitIsRetryable(t *testing.T) {
	srv := geminiStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"code":429,"message":"You have exceeded your per-minute request quota","status":"RESOURCE_EXHAUSTED"}}`))
	})

	_, err := srv.call(context.Background(), "prompt", false, 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if !isRetryable(err) {
		t.Fatalf("per-minute 429 must be retryable, got %v", err)
	}
	delay, ok := retryAfterFrom(err)
	if !ok || delay != 7*time.Second {
		t.Fatalf("Retry-After = %v (ok=%v), want 7s", delay, ok)
	}
}

// TestGeminiCallDailyQuotaFailsFast is the case Genkit's single
// status.ErrResourceExhausted cannot express: the exhausted quota is the
// user's, so retrying only holds a concurrency slot to fail the same way.
func TestGeminiCallDailyQuotaFailsFast(t *testing.T) {
	var calls atomic.Int32
	srv := geminiStub(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"code":429,"message":"You have exceeded your daily request quota","status":"RESOURCE_EXHAUSTED"}}`))
	})

	_, err := srv.Generate(context.Background(), "prompt", false, 0)
	if err == nil {
		t.Fatal("expected error for quota exhausted")
	}
	if !strings.Contains(err.Error(), "quota") {
		t.Fatalf("expected quota error, got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("quota exhaustion must not retry; got %d calls", calls.Load())
	}
}

func TestGeminiCallStatusMapping(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantRetry   bool
		wantErrPart string
	}{
		{"503 overload", http.StatusServiceUnavailable, `{"error":{"message":"overloaded"}}`, true, "status 503"},
		{"400 bad request", http.StatusBadRequest, `{"error":{"message":"bad"}}`, false, "status 400"},
		{"401 unauthorized", http.StatusUnauthorized, `{"error":{"message":"bad key"}}`, false, "status 401"},
		{"500 server error", http.StatusInternalServerError, `{"error":{"message":"boom"}}`, false, "status 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := geminiStub(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			})
			_, err := srv.call(context.Background(), "prompt", false, 0)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Errorf("error = %v, want it to mention %q", err, tt.wantErrPart)
			}
			if isRetryable(err) != tt.wantRetry {
				t.Errorf("isRetryable(%v) = %v, want %v", err, isRetryable(err), tt.wantRetry)
			}
		})
	}
}

func TestGeminiCallResponseErrors(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantErrPart string
	}{
		{"invalid json", `not json`, "unmarshal gemini response"},
		{"no candidates", `{"candidates":[]}`, "no candidates"},
		{"candidate without parts", `{"candidates":[{"content":{"parts":[]}}]}`, "no candidates"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := geminiStub(t, func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.body))
			})
			_, err := srv.Generate(context.Background(), "prompt", false, 0)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Fatalf("error = %v, want it to mention %q", err, tt.wantErrPart)
			}
		})
	}
}

// TestGeminiCallReadsFirstCandidate guards against returning an empty string
// when a response carries several parts.
func TestGeminiCallReadsFirstCandidate(t *testing.T) {
	srv := geminiStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"first"},{"text":"second"}]}}]}`))
	})
	got, err := srv.Generate(context.Background(), "prompt", false, 0)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "first" {
		t.Fatalf("got %q, want %q", got, "first")
	}
}

func TestRetryAfterParsing(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"seconds", "5", 5 * time.Second},
		{"empty", "", 0},
		{"garbage", "abc", 0},
		{"zero", "0", 0},
		{"negative", "-3", 0},
		{"http date unsupported", "Wed, 21 Oct 2015 07:28:00 GMT", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRetryAfter(tt.value); got != tt.want {
				t.Fatalf("parseRetryAfter(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsQuotaExhausted(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{"daily quota", `{"error":{"code":429,"message":"You have exceeded your daily request quota","status":"RESOURCE_EXHAUSTED"}}`, true},
		{"quota exhausted", `{"error":{"code":429,"message":"quota exhausted","status":"RESOURCE_EXHAUSTED"}}`, true},
		{"per minute rate limit", `{"error":{"code":429,"message":"You have exceeded your per-minute request quota","status":"RESOURCE_EXHAUSTED"}}`, false},
		{"unparseable", `not json`, false},
		{"empty message", `{"error":{"code":429,"message":"","status":"RESOURCE_EXHAUSTED"}}`, false},
		{"empty body", ``, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isQuotaExhausted([]byte(tt.body)); got != tt.want {
				t.Fatalf("isQuotaExhausted(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}
