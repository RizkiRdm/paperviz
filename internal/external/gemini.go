package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"

// geminiCallTimeout is the MAX_LIMIT per Gemini call (ARCHITECTURE.md Section 4).
const geminiCallTimeout = 120 * time.Second

// GeminiClient is a direct HTTP client for the Gemini API. No SDK, no
// gateway — per ARCHITECTURE.md Context Lock and AGENTS.md Quick Rules.
type GeminiClient struct {
	apiKey     string
	model      string
	endpoint   string // geminiEndpoint by default; overridable in tests
	httpClient *http.Client
	// sem serializes all Gemini calls through one client. Multiple document
	// pipelines run concurrently (one goroutine per upload); without a cap,
	// their overlapping calls blow past free-tier RPM limits and every call
	// 429s, which then compounds through the retry loop. A single-slot
	// semaphore keeps throughput at exactly one in-flight call.
	sem chan struct{}
	// Retry schedule. Production defaults below; tests shrink them so
	// backoff sleeps don't slow the suite.
	retries     int
	retryBudget time.Duration
	backoffBase time.Duration
	backoffCeil time.Duration
}

// Retry production defaults: up to 5 attempts, 90s total budget, exponential
// 2s/4s/8s/16s… capped at 32s per wait. These bound how long a persistently
// failing upstream can stall a document's pipeline (see Generate).
const (
	defaultRetries     = 5
	defaultRetryBudget = 90 * time.Second
	defaultBackoffBase = 2 * time.Second
	defaultBackoffCeil = 32 * time.Second
)

// NewGeminiClient constructs a client with explicit transport timeouts so
// long-running generations don't hang on stale TCP connections. apiKey MUST
// come from an environment variable — never hardcoded (ARCHITECTURE.md §5).
func NewGeminiClient(apiKey, model string) *GeminiClient {
	return &GeminiClient{
		apiKey:   apiKey,
		model:    model,
		endpoint: geminiEndpoint,
		httpClient: &http.Client{
			Timeout: geminiCallTimeout, // > per-attempt context (90s)
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

type geminiRequest struct {
	Contents         []geminiContent   `json:"contents"`
	GenerationConfig *generationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type generationConfig struct {
	ResponseMIMEType string `json:"responseMimeType,omitempty"`
	MaxOutputTokens  *int   `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// geminiErrorBody is the error envelope Gemini returns for non-2xx responses.
type geminiErrorBody struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// Generate sends a single prompt and returns the model's text response.
// If asJSON is true, the model is instructed to return only valid JSON
// (used by claim-extraction and chart-data-extraction prompts).
// If maxTokens > 0, the response is capped at that many output tokens
// (used by ELI5 mode to prevent connection resets on long generations).
//
// Retries transient failures (503, per-minute rate limits, timeout,
// connection reset) with exponential backoff, honoring Retry-After when the
// server sends one, and bounded by a total budget so a stuck upstream can't
// stall the pipeline for minutes. Non-retryable errors (400/401/403, daily
// quota exhaustion) return immediately — quota_exceeded never recovers by
// retrying, per Gemini API error semantics.
func (c *GeminiClient) Generate(ctx context.Context, prompt string, asJSON bool, maxTokens int) (string, error) {
	// Serialize calls so concurrent document pipelines can't self-inflict
	// rate limits. Context-aware: if the caller cancels while queued, fail
	// fast instead of silently holding the slot.
	select {
	case c.sem <- struct{}{}:
		defer func() { <-c.sem }()
	case <-ctx.Done():
		return "", fmt.Errorf("gemini generate: waiting for capacity: %w", ctx.Err())
	}

	var lastErr error
	budgetStart := time.Now()
	for attempt := 0; attempt < c.retries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
		start := time.Now()
		text, err := c.generateOnce(attemptCtx, prompt, asJSON, maxTokens)
		cancel()
		duration := time.Since(start).Milliseconds()

		slog.Info("gemini call",
			"stage", "gemini_generate",
			"attempt", attempt+1,
			"duration_ms", duration,
			"success", err == nil,
			"prompt_bytes", len(prompt),
		)

		if err == nil {
			return text, nil
		}

		if errors.Is(err, context.DeadlineExceeded) {
			slog.Error("Gemini call timeout (client-side, not rate limit)",
				"stage", "gemini_generate", "attempt", attempt+1)
			lastErr = fmt.Errorf("gemini client timeout: %w", err)
		} else {
			lastErr = err
		}

		// Non-retryable errors (4xx client errors, quota exhaustion, parse
		// failures) → give up immediately.
		if !isRetryable(lastErr) {
			return "", lastErr
		}

		// Last attempt exhausted → give up, no point sleeping.
		if attempt == c.retries-1 {
			break
		}

		backoff := time.Duration(2<<attempt) * c.backoffBase
		if backoff > c.backoffCeil {
			backoff = c.backoffCeil
		}
		// Honor the server's Retry-After when it's more conservative than
		// our exponential schedule.
		if retryAfter, ok := retryAfterFrom(lastErr); ok && retryAfter > backoff {
			backoff = retryAfter
		}
		// Never sleep past the overall retry budget — a dying upstream must
		// not hold the pipeline hostage.
		if remaining := c.retryBudget - time.Since(budgetStart); backoff > remaining {
			backoff = remaining
		}
		if backoff <= 0 {
			break
		}

		slog.Info("gemini retry backoff", "attempt", attempt+1, "wait_s", int(backoff.Seconds()))
		// Context-aware sleep: an expired pipeline timeout cancels the
		// wait instead of blocking until the backoff elapses.
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return "", fmt.Errorf("gemini generate: retry aborted: %w", ctx.Err())
		}
	}
	return "", fmt.Errorf("gemini generate failed after %d attempts: %w", c.retries, lastErr)
}

// retryableError marks an error worth retrying and optionally carries the
// server-suggested wait (Retry-After) so the caller can honor it.
type retryableError struct {
	msg        string
	retryAfter time.Duration
}

func (e *retryableError) Error() string { return e.msg }

// isRetryable reports whether a Gemini error is worth retrying.
// 503 (server overload), per-minute 429 rate limits, timeouts, and transient
// network errors (connection reset, refused, DNS failures) are retryable.
// 4xx client errors (400, 401, 403) and daily quota exhaustion are not —
// retrying won't help.
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
	// Transient network-level errors — typically "connection reset by peer",
	// "connection refused", or "no such host". These happen when long
	// generations trip idle timeouts on intermediate proxies.
	if strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") {
		return true
	}
	return false
}

func retryAfterFrom(err error) (time.Duration, bool) {
	var re *retryableError
	if errors.As(err, &re) && re.retryAfter > 0 {
		return re.retryAfter, true
	}
	return 0, false
}

// isQuotaExhausted reports whether a 429 body indicates daily-quota
// exhaustion. Gemini returns 429 for both per-minute rate limits (retry
// helps) and daily quota exhaustion (retrying is pointless — the quota only
// resets on a schedule). The two are only distinguishable via the message.
func isQuotaExhausted(body []byte) bool {
	var eb geminiErrorBody
	if err := json.Unmarshal(body, &eb); err != nil || eb.Error.Message == "" {
		return false
	}
	msg := strings.ToLower(eb.Error.Message)
	return strings.Contains(msg, "quota") && (strings.Contains(msg, "daily") || strings.Contains(msg, "exhaust"))
}

// parseRetryAfter converts a Retry-After header value (seconds, or HTTP-date
// which we don't parse) into a duration. Returns 0 if absent/unparseable.
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	secs, err := strconv.Atoi(v)
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

func (c *GeminiClient) generateOnce(ctx context.Context, prompt string, asJSON bool, maxTokens int) (string, error) {
	reqBody := geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
	}
	if asJSON || maxTokens > 0 {
		cfg := &generationConfig{}
		if asJSON {
			cfg.ResponseMIMEType = "application/json"
		}
		if maxTokens > 0 {
			cfg.MaxOutputTokens = &maxTokens
		}
		reqBody.GenerationConfig = cfg
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf(c.endpoint, c.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini http call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		slog.Error("Gemini rate limited by server", "stage", "gemini_generate")
		if isQuotaExhausted(body) {
			// Daily quota hit — retrying cannot succeed; fail fast so the
			// pipeline marks the document failed instead of sleeping ~minutes.
			return "", fmt.Errorf("gemini quota exhausted (429): status=%d", resp.StatusCode)
		}
		return "", &retryableError{
			msg:        fmt.Sprintf("gemini rate limited (429): status=%d", resp.StatusCode),
			retryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
		}
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusServiceUnavailable {
			return "", &retryableError{msg: fmt.Sprintf("gemini returned status %d", resp.StatusCode)}
		}
		return "", fmt.Errorf("gemini returned status %d", resp.StatusCode)
	}

	var parsed geminiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("unmarshal gemini response: %w", err)
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini response had no candidates")
	}

	return parsed.Candidates[0].Content.Parts[0].Text, nil
}

// ExtractJSON sends a JSON-requested prompt to Gemini, strips markdown code fences,
// handles candidate JSON recovery (substring trimming), and unmarshals into T.
func ExtractJSON[T any](ctx context.Context, client *GeminiClient, prompt string, maxTokens int) (T, error) {
	var zero T
	raw, err := client.Generate(ctx, prompt, true, maxTokens)
	if err != nil {
		return zero, fmt.Errorf("gemini generate json: %w", err)
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		return zero, fmt.Errorf("empty json response")
	}

	trimmed = stripJSONFences(trimmed)

	var parsed T
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		if strings.HasPrefix(trimmed, "{") {
			if idx := strings.LastIndex(trimmed, "}"); idx > 0 {
				candidate := trimmed[:idx+1]
				if err := json.Unmarshal([]byte(candidate), &parsed); err == nil {
					return parsed, nil
				}
			}
		} else if strings.HasPrefix(trimmed, "[") {
			if idx := strings.LastIndex(trimmed, "]"); idx > 0 {
				candidate := trimmed[:idx+1]
				if err := json.Unmarshal([]byte(candidate), &parsed); err == nil {
					return parsed, nil
				}
			}
		}
		return zero, fmt.Errorf("unmarshal json response %q: %w", trimmed, err)
	}

	return parsed, nil
}

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
