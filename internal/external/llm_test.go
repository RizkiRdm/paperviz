package external

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fastTransport shrinks the retry schedule so backoff sleeps don't slow tests.
func fastTransport() *Transport {
	tr := NewTransport()
	tr.retries = 3
	tr.retryBudget = 2 * time.Second
	tr.backoffBase = 5 * time.Millisecond
	tr.backoffCeil = 20 * time.Millisecond
	return tr
}

// stubLLM returns a client whose backend always yields text, for exercising
// response handling without any provider.
func stubLLM(text string) *LLM {
	return fastTransport().newLLM(ProviderGemini, "test-model",
		func(context.Context, string, bool, int) (string, error) { return text, nil })
}

func TestGenerateRetriesOnRateLimitThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	tr := fastTransport()
	srv := tr.newLLM(ProviderGemini, "test-model",
		func(context.Context, string, bool, int) (string, error) {
			if calls.Add(1) == 1 {
				return "", &retryableError{msg: "rate limited (429): status=429"}
			}
			return "ok", nil
		})

	text, err := srv.Generate(context.Background(), "prompt", false, 0)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if text != "ok" {
		t.Fatalf("got text %q, want %q", text, "ok")
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls (1 retry), got %d", calls.Load())
	}
}

func TestGenerateGivesUpAfterMaxRetries(t *testing.T) {
	var calls atomic.Int32
	tr := fastTransport()
	srv := tr.newLLM(ProviderGemini, "test-model",
		func(context.Context, string, bool, int) (string, error) {
			calls.Add(1)
			return "", &retryableError{msg: "provider returned status 503"}
		})

	if _, err := srv.Generate(context.Background(), "prompt", false, 0); err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls.Load() != int32(tr.retries) {
		t.Fatalf("expected %d calls, got %d", tr.retries, calls.Load())
	}
}

func TestGenerateNonRetryable4xx(t *testing.T) {
	var calls atomic.Int32
	tr := fastTransport()
	srv := tr.newLLM(ProviderGemini, "test-model",
		func(context.Context, string, bool, int) (string, error) {
			calls.Add(1)
			return "", errors.New("provider returned status 400")
		})

	if _, err := srv.Generate(context.Background(), "prompt", false, 0); err == nil {
		t.Fatal("expected error for 400")
	}
	if calls.Load() != 1 {
		t.Fatalf("400 must not retry; got %d calls", calls.Load())
	}
}

// TestGenerateHonorsRetryAfter proves a server-suggested wait longer than our
// own backoff wins.
func TestGenerateHonorsRetryAfter(t *testing.T) {
	var calls atomic.Int32
	tr := fastTransport()
	tr.backoffBase = 10 * time.Millisecond
	srv := tr.newLLM(ProviderGemini, "test-model",
		func(context.Context, string, bool, int) (string, error) {
			if calls.Add(1) == 1 {
				return "", &retryableError{msg: "rate limited", retryAfter: 150 * time.Millisecond}
			}
			return "ok", nil
		})

	start := time.Now()
	if _, err := srv.Generate(context.Background(), "prompt", false, 0); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	elapsed := time.Since(start)
	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", calls.Load())
	}
	if elapsed < 150*time.Millisecond {
		t.Fatalf("Retry-After 150ms ignored; returned after %v", elapsed)
	}
}

// TestGenerateSerializesConcurrentCalls is the anti-429-storm guard: the single
// process-wide slot must keep in-flight calls at one regardless of user count.
func TestGenerateSerializesConcurrentCalls(t *testing.T) {
	var inflight, maxInflight atomic.Int32
	tr := fastTransport()
	srv := tr.newLLM(ProviderGemini, "test-model",
		func(context.Context, string, bool, int) (string, error) {
			cur := inflight.Add(1)
			for {
				prev := maxInflight.Load()
				if cur <= prev || maxInflight.CompareAndSwap(prev, cur) {
					break
				}
			}
			defer inflight.Add(-1)
			time.Sleep(30 * time.Millisecond)
			return "ok", nil
		})

	var wg sync.WaitGroup
	errs := make([]error, 4)
	for i := range errs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = srv.Generate(context.Background(), "prompt", false, 0)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
	}
	if m := maxInflight.Load(); m > 1 {
		t.Fatalf("expected serialized calls (max inflight 1), got %d", m)
	}
}

// TestGenerateContextCanceledWhileQueued proves a cancelled caller fails fast
// instead of blocking on the semaphore until the holder finishes.
func TestGenerateContextCanceledWhileQueued(t *testing.T) {
	release := make(chan struct{})
	tr := fastTransport()
	srv := tr.newLLM(ProviderGemini, "test-model",
		func(context.Context, string, bool, int) (string, error) {
			<-release
			return "ok", nil
		})

	done := make(chan struct{})
	go func() {
		srv.Generate(context.Background(), "first", false, 0)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond) // let the first call take the slot

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := srv.Generate(ctx, "second", false, 0)
	if err == nil {
		t.Fatal("expected error when context canceled while queued")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("expected deadline error, got %v", err)
	}
	close(release)
	<-done
}

// TestGenerateBudgetCapsTotalRetryTime ensures the budget bounds total sleep,
// so a persistently failing upstream cannot stall a document for minutes.
func TestGenerateBudgetCapsTotalRetryTime(t *testing.T) {
	tr := fastTransport()
	srv := tr.newLLM(ProviderGemini, "test-model",
		func(context.Context, string, bool, int) (string, error) {
			return "", &retryableError{msg: "rate limited", retryAfter: 60 * time.Second}
		})

	start := time.Now()
	if _, err := srv.Generate(context.Background(), "prompt", false, 0); err == nil {
		t.Fatal("expected error from persistent 429")
	}
	// A 60s Retry-After would dominate the schedule; the budget must cut it off.
	if elapsed := time.Since(start); elapsed > tr.retryBudget+100*time.Millisecond {
		t.Fatalf("retry took %v, budget cap %v not enforced", elapsed, tr.retryBudget)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"503", errors.New("provider returned status 503"), true},
		{"429 rate limit", &retryableError{msg: "rate limited (429): status=429"}, true},
		{"400", errors.New("provider returned status 400"), false},
		{"daily quota exhausted", errors.New("quota exhausted (429): status=429"), false},
		{"connection reset", errors.New("http call: Post: connection reset by peer"), true},
		{"context deadline", context.DeadlineExceeded, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryable(tt.err); got != tt.want {
				t.Fatalf("isRetryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestProviderValid(t *testing.T) {
	tests := []struct {
		provider Provider
		want     bool
	}{
		{ProviderGemini, true},
		{ProviderAnthropic, true},
		{ProviderOpenAI, true},
		{"", false},
		{"llama", false},
		{"Gemini", false},
	}
	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			if got := tt.provider.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderDefaultModel(t *testing.T) {
	for _, p := range SupportedProviders {
		if p.DefaultModel() == "" {
			t.Fatalf("provider %q has no default model", p)
		}
	}
	// Unknown providers fall back rather than returning an empty model.
	if got := Provider("nope").DefaultModel(); got == "" {
		t.Fatal("unknown provider must still yield a usable default model")
	}
}

func TestTransportFor(t *testing.T) {
	tr := NewTransport()

	t.Run("gemini with key and model", func(t *testing.T) {
		llm, err := tr.For(ProviderGemini, "user-key", "gemini-2.5-flash")
		if err != nil {
			t.Fatalf("For: %v", err)
		}
		if llm.Provider() != ProviderGemini || llm.Model() != "gemini-2.5-flash" {
			t.Fatalf("unexpected client: %s/%s", llm.Provider(), llm.Model())
		}
	})

	t.Run("empty model falls back to provider default", func(t *testing.T) {
		llm, err := tr.For(ProviderGemini, "user-key", "")
		if err != nil {
			t.Fatalf("For: %v", err)
		}
		if llm.Model() != ProviderGemini.DefaultModel() {
			t.Fatalf("got model %q, want default %q", llm.Model(), ProviderGemini.DefaultModel())
		}
	})

	t.Run("unknown provider rejected", func(t *testing.T) {
		if _, err := tr.For("llama", "user-key", ""); err == nil {
			t.Fatal("expected error for unknown provider")
		}
	})

	t.Run("empty key rejected", func(t *testing.T) {
		if _, err := tr.For(ProviderGemini, "", ""); err == nil {
			t.Fatal("expected error for empty key")
		}
	})
}

func TestExtractJSON(t *testing.T) {
	type payload struct {
		Name  string   `json:"name"`
		Count int      `json:"count"`
		Tags  []string `json:"tags"`
	}

	tests := []struct {
		name    string
		raw     string
		want    payload
		wantErr bool
	}{
		{
			name: "plain object",
			raw:  `{"name":"a","count":2,"tags":["x"]}`,
			want: payload{Name: "a", Count: 2, Tags: []string{"x"}},
		},
		{
			name: "fenced object",
			raw:  "```json\n{\"name\":\"b\",\"count\":1,\"tags\":[]}\n```",
			want: payload{Name: "b", Count: 1, Tags: []string{}},
		},
		{
			name: "trailing prose is trimmed",
			raw:  `{"name":"c","count":3,"tags":["y"]} Hope this helps!`,
			want: payload{Name: "c", Count: 3, Tags: []string{"y"}},
		},
		{
			name: "leading whitespace",
			raw:  "\n\n  {\"name\":\"d\",\"count\":4,\"tags\":[]}",
			want: payload{Name: "d", Count: 4, Tags: []string{}},
		},
		{name: "empty response", raw: "", wantErr: true},
		{name: "null response", raw: "null", wantErr: true},
		{name: "empty array response", raw: "[]", wantErr: true},
		{name: "unparseable", raw: "not json at all", wantErr: true},
		{name: "truncated with no closing brace", raw: `{"name":"e"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractJSON[payload](context.Background(), stubLLM(tt.raw), "prompt", 0)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tt.want.Name || got.Count != tt.want.Count || len(got.Tags) != len(tt.want.Tags) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestExtractJSONPropagatesCallError(t *testing.T) {
	tr := fastTransport()
	srv := tr.newLLM(ProviderGemini, "test-model",
		func(context.Context, string, bool, int) (string, error) {
			return "", errors.New("provider returned status 400")
		})
	if _, err := ExtractJSON[map[string]any](context.Background(), srv, "prompt", 0); err == nil {
		t.Fatal("expected error from failing backend")
	}
}
