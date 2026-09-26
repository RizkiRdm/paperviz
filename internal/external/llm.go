package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Provider identifies a model vendor. PaperViz holds no vendor account of its
// own: every request runs on a key the user supplied (BYOK), so a Provider is
// a routing label, not a configured integration.
type Provider string

const (
	ProviderGemini    Provider = "gemini"
	ProviderAnthropic Provider = "anthropic"
	ProviderOpenAI    Provider = "openai"
)

// SupportedProviders lists every accepted provider, in display order.
var SupportedProviders = []Provider{ProviderGemini, ProviderAnthropic, ProviderOpenAI}

// Valid reports whether p is a provider this build can route to.
func (p Provider) Valid() bool {
	for _, s := range SupportedProviders {
		if p == s {
			return true
		}
	}
	return false
}

// DefaultModel returns the model used when a user supplies a key without
// naming a model. Chosen for cost and latency on structured extraction.
func (p Provider) DefaultModel() string {
	switch p {
	case ProviderAnthropic:
		return "claude-sonnet-4-6"
	case ProviderOpenAI:
		return "gpt-5.5"
	default:
		return "gemini-2.5-flash-lite"
	}
}

// llmCall performs one provider request with no retry or concurrency control.
// Backends implement only this; Transport owns everything around it.
type llmCall func(ctx context.Context, prompt string, asJSON bool, maxTokens int) (string, error)

// Transport holds the provider-agnostic machinery every request shares: the
// HTTP client, the retry schedule, and the process-wide concurrency slot.
// Constructed once at startup.
//
// The single semaphore slot is deliberate. Concurrent document pipelines
// otherwise self-inflict rate limits, and once upstream starts 429ing the
// retry loop compounds it. One slot keeps throughput at one in-flight call
// regardless of how many users are active.
type Transport struct {
	httpClient  *http.Client
	sem         chan struct{}
	retries     int
	retryBudget time.Duration
	backoffBase time.Duration
	backoffCeil time.Duration
}

// Retry production defaults: up to 5 attempts, 90s total budget, exponential
// 2s/4s/8s/16s... capped at 32s per wait. These bound how long a persistently
// failing upstream can stall a document's pipeline.
const (
	defaultRetries     = 5
	defaultRetryBudget = 90 * time.Second
	defaultBackoffBase = 2 * time.Second
	defaultBackoffCeil = 32 * time.Second
)

// llmCallTimeout is the per-attempt ceiling for a single provider call
// (ARCHITECTURE.md Section 4). Must exceed defaultRetryBudget.
const llmCallTimeout = 120 * time.Second

// NewTransport builds the shared Transport with explicit transport timeouts so
// long generations cannot hang on stale TCP connections.
func NewTransport() *Transport {
	return &Transport{
		httpClient: &http.Client{
			Timeout: llmCallTimeout,
			Transport: &http.Transport{
				IdleConnTimeout:       90 * time.Second,
				ResponseHeaderTimeout: 100 * time.Second,
				DisableKeepAlives:     false,
			},
		},
		sem:         make(chan struct{}, 1),
		retries:     defaultRetries,
		retryBudget: defaultRetryBudget,
		backoffBase: defaultBackoffBase,
		backoffCeil: defaultBackoffCeil,
	}
}

// HTTPClient exposes the shared client so a backend inherits the tuned
// timeouts instead of constructing its own.
func (t *Transport) HTTPClient() *http.Client { return t.httpClient }

// LLM is a client bound to one provider, model, and decrypted user key.
//
// It is never a process-wide singleton. Under BYOK every user supplies their
// own key, so an LLM is constructed per request from a credential that was
// decrypted moments earlier and discarded with it.
type LLM struct {
	provider Provider
	model    string
	tr       *Transport
	call     llmCall
}

// Provider reports which vendor this client routes to.
func (l *LLM) Provider() Provider { return l.provider }

// Model reports the model id in use.
func (l *LLM) Model() string { return l.model }

// newLLM assembles a client around a backend call. Tests use this directly to
// exercise retry and concurrency behaviour without any HTTP.
func (t *Transport) newLLM(p Provider, model string, call llmCall) *LLM {
	return &LLM{provider: p, model: model, tr: t, call: call}
}

// For builds a client for one provider and one user-supplied key. An empty
// model falls back to the provider default.
//
// The key is the caller's decrypted BYOK credential. It is captured in the
// returned backend closure and must never be logged, stored, or included in
// any error this or the backend produces.
func (t *Transport) For(p Provider, apiKey, model string) (*LLM, error) {
	if !p.Valid() {
		return nil, fmt.Errorf("unsupported provider %q", p)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("empty api key for provider %q", p)
	}
	if model == "" {
		model = p.DefaultModel()
	}
	return t.newLLM(p, model, newGoAICall(p, model, apiKey, t.httpClient)), nil
}

// Generate sends one prompt and returns the model's text.
//
// Retries transient failures with exponential backoff, bounded by a total
// budget so a stuck upstream cannot stall the pipeline. Provider-agnostic:
// the backend call is retried here, never inside the backend.
func (l *LLM) Generate(ctx context.Context, prompt string, asJSON bool, maxTokens int) (string, error) {
	// Serialize calls so concurrent pipelines cannot self-inflict rate limits.
	// Context-aware: a caller that cancels while queued fails fast instead of
	// silently holding the slot.
	select {
	case l.tr.sem <- struct{}{}:
		defer func() { <-l.tr.sem }()
	case <-ctx.Done():
		return "", fmt.Errorf("%s generate: waiting for capacity: %w", l.provider, ctx.Err())
	}

	var lastErr error
	budgetStart := time.Now()
	for attempt := 0; attempt < l.tr.retries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, l.tr.retryBudget)
		start := time.Now()
		text, err := l.call(attemptCtx, prompt, asJSON, maxTokens)
		cancel()
		duration := time.Since(start).Milliseconds()

		slog.Info("llm call",
			"provider", string(l.provider),
			"model", l.model,
			"attempt", attempt+1,
			"duration_ms", duration,
			"success", err == nil,
			"prompt_bytes", len(prompt),
		)

		if err == nil {
			return text, nil
		}

		if errors.Is(err, context.DeadlineExceeded) {
			slog.Error("llm call timeout (client-side, not rate limit)",
				"provider", string(l.provider), "attempt", attempt+1)
			lastErr = fmt.Errorf("%s client timeout: %w", l.provider, err)
		} else {
			lastErr = err
		}

		// Non-retryable errors (4xx client errors, quota exhaustion, parse
		// failures) give up immediately.
		if !isRetryable(lastErr) {
			return "", lastErr
		}

		if attempt == l.tr.retries-1 {
			break
		}

		backoff := time.Duration(2<<attempt) * l.tr.backoffBase
		if backoff > l.tr.backoffCeil {
			backoff = l.tr.backoffCeil
		}
		// Honor the server's Retry-After when it is more conservative than
		// our exponential schedule.
		if retryAfter, ok := retryAfterFrom(lastErr); ok && retryAfter > backoff {
			backoff = retryAfter
		}
		// Never sleep past the overall budget: a dying upstream must not hold
		// the pipeline hostage.
		if remaining := l.tr.retryBudget - time.Since(budgetStart); backoff > remaining {
			backoff = remaining
		}
		if backoff <= 0 {
			break
		}

		slog.Info("llm retry backoff",
			"provider", string(l.provider), "attempt", attempt+1, "wait_s", int(backoff.Seconds()))
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return "", fmt.Errorf("%s generate: retry aborted: %w", l.provider, ctx.Err())
		}
	}
	return "", fmt.Errorf("%s generate failed after %d attempts: %w", l.provider, l.tr.retries, lastErr)
}

// retryableError marks an error worth retrying and carries the server-suggested
// wait (Retry-After) so the schedule can honor it.
type retryableError struct {
	msg        string
	retryAfter time.Duration
}

func (e *retryableError) Error() string { return e.msg }

// isRetryable reports whether a provider error is worth another attempt.
// 503 overload, per-minute rate limits, timeouts, and transient network errors
// (connection reset, refused, DNS failure) are retryable. 4xx client errors
// and exhausted daily quota are not: retrying cannot help.
func isRetryable(err error) bool {
	var re *retryableError
	if errors.As(err, &re) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	errStr := err.Error()
	if strings.Contains(errStr, "status 503") || strings.Contains(errStr, "status 429") {
		return true
	}
	// Transient network-level errors, typically surfaced by long generations
	// tripping idle timeouts on intermediate proxies.
	if strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") {
		return true
	}
	return false
}

// retryAfterFrom extracts a server-suggested wait from a retryable error.
func retryAfterFrom(err error) (time.Duration, bool) {
	var re *retryableError
	if errors.As(err, &re) && re.retryAfter > 0 {
		return re.retryAfter, true
	}
	return 0, false
}

// ExtractJSON sends a JSON-requested prompt, strips markdown code fences,
// recovers truncated candidate JSON, and unmarshals into T.
func ExtractJSON[T any](ctx context.Context, client *LLM, prompt string, maxTokens int) (T, error) {
	var zero T
	raw, err := client.Generate(ctx, prompt, true, maxTokens)
	if err != nil {
		return zero, fmt.Errorf("%s generate json: %w", client.provider, err)
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		return zero, fmt.Errorf("empty json response")
	}

	trimmed = stripJSONFences(trimmed)

	var parsed T
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		// Models often append prose or a second object after valid JSON.
		// Retry against the largest well-formed prefix before giving up.
		var closer string
		switch {
		case strings.HasPrefix(trimmed, "{"):
			closer = "}"
		case strings.HasPrefix(trimmed, "["):
			closer = "]"
		}
		if closer != "" {
			if idx := strings.LastIndex(trimmed, closer); idx > 0 {
				if err := json.Unmarshal([]byte(trimmed[:idx+1]), &parsed); err == nil {
					return parsed, nil
				}
			}
		}
		return zero, fmt.Errorf("unmarshal json response %q: %w", trimmed, err)
	}
	return parsed, nil
}

// stripJSONFences removes a surrounding markdown code fence, which models add
// even when told to return bare JSON.
func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	nl := strings.Index(s, "\n")
	if nl < 0 {
		return s
	}
	s = s[nl+1:]
	if idx := strings.LastIndex(s, "```"); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}
