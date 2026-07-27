package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"net/http"
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
	httpClient *http.Client
}

// NewGeminiClient constructs a client with explicit transport timeouts so
// long-running generations don't hang on stale TCP connections. apiKey MUST
// come from an environment variable — never hardcoded (ARCHITECTURE.md §5).
func NewGeminiClient(apiKey, model string) *GeminiClient {
	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout:   geminiCallTimeout, // > per-attempt context (90s)
			Transport: &http.Transport{
				IdleConnTimeout:       90 * time.Second,
				ResponseHeaderTimeout: 100 * time.Second,
				DisableKeepAlives:     false,
			},
		},
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

// Generate sends a single prompt and returns the model's text response.
// If asJSON is true, the model is instructed to return only valid JSON
// (used by claim-extraction and chart-data-extraction prompts).
// If maxTokens > 0, the response is capped at that many output tokens
// (used by ELI5 mode to prevent connection resets on long generations).
//
// Retries up to maxRetries times with exponential backoff for transient
// failures (503, 429, timeout, connection reset). Client errors (400/401/403)
// return immediately.
const maxRetries = 10

func (c *GeminiClient) Generate(ctx context.Context, prompt string, asJSON bool, maxTokens int) (string, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
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

		// Non-retryable errors (4xx client errors, parse failures) → give up.
		if !isRetryable(lastErr) {
			return "", lastErr
		}

		if attempt < maxRetries-1 {
			backoff := time.Duration(2<<attempt) * time.Second // 2s, 4s, 8s…
			if backoff > 32*time.Second {
				backoff = 32 * time.Second
			}
			slog.Info("gemini retry backoff", "attempt", attempt+1, "wait_s", int(backoff.Seconds()))
			time.Sleep(backoff)
		}
	}
	return "", fmt.Errorf("gemini generate failed after %d attempts: %w", maxRetries, lastErr)
}

// isRetryable reports whether a Gemini error is worth retrying.
// 503 (server overload), 429 (rate limit), timeouts, and transient network
// errors (connection reset, refused, DNS failures) are retryable.
// 4xx client errors (400, 401, 403) are not — retrying won't help.
func isRetryable(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	errStr := err.Error()
	if strings.Contains(errStr, "status 503") || strings.Contains(errStr, "429") {
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

  url := fmt.Sprintf(geminiEndpoint, c.model)
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
		return "", fmt.Errorf("gemini rate limited (429): status=%d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
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
