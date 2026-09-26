package external

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/anthropic"
	"github.com/zendev-sh/goai/provider/google"
	"github.com/zendev-sh/goai/provider/openai"
)

// jsonRequestSuffix asks for a bare JSON body. GoAI's structured output is
// schema-driven (ResponseFormat requires a Schema), so there is no equivalent
// of Gemini's responseMimeType flag. We ask by instruction instead and rely on
// ExtractJSON's fence-stripping and prefix recovery, which is what already
// handles models that wrap JSON in prose.
const jsonRequestSuffix = "Respond with a single valid JSON value and no other text."

// newGoAICall builds a provider call backed by github.com/zendev-sh/goai.
//
// The SDK's own retry is disabled. Transport.Generate owns the retry schedule,
// the total budget, and the single process-wide concurrency slot; a second
// retry layer underneath it would compound backoffs, which is the 429 storm the
// semaphore exists to prevent.
func newGoAICall(p Provider, model, apiKey string, hc *http.Client) llmCall {
	m := newGoAIModel(p, model, apiKey, hc)

	return func(ctx context.Context, prompt string, asJSON bool, maxTokens int) (string, error) {
		if asJSON {
			prompt += "\n\n" + jsonRequestSuffix
		}
		opts := []goai.Option{
			goai.WithPrompt(prompt),
			goai.WithMaxRetries(0),
		}
		if maxTokens > 0 {
			opts = append(opts, goai.WithMaxOutputTokens(maxTokens))
		}

		res, err := goai.GenerateText(ctx, m, opts...)
		if err != nil {
			return "", mapGoAIError(err)
		}
		return res.Text, nil
	}
}

// newGoAIModel constructs a model bound to one provider and one decrypted key.
//
// apiKey is always passed explicitly. Every GoAI provider falls back to a
// vendor environment variable when the key is empty, which in this server would
// silently route a request through an operator key — the one thing BYOK exists
// to prevent. Transport.For rejects an empty key before reaching here, and this
// constructor is only ever called from there.
func newGoAIModel(p Provider, model, apiKey string, hc *http.Client) provider.LanguageModel {
	switch p {
	case ProviderAnthropic:
		return anthropic.Chat(model,
			anthropic.WithAPIKey(apiKey),
			anthropic.WithHTTPClient(hc),
		)
	case ProviderOpenAI:
		return openai.Chat(model,
			openai.WithAPIKey(apiKey),
			openai.WithHTTPClient(hc),
		)
	default:
		return google.Chat(model,
			google.WithAPIKey(apiKey),
			google.WithHTTPClient(hc),
		)
	}
}

// mapGoAIError converts a GoAI API error into this package's error vocabulary so
// the existing retry schedule in Transport.Generate keeps working unchanged.
func mapGoAIError(err error) error {
	var apiErr *goai.APIError
	if !errors.As(err, &apiErr) {
		// Context cancellation, overflow, and transport failures pass through;
		// isRetryable already understands deadline errors.
		return err
	}

	// A 429 covers both a per-minute rate limit, which recovers on retry, and
	// exhausted daily quota, which does not. Under BYOK the exhausted quota
	// belongs to the user, so failing fast frees a concurrency slot instead of
	// spending the whole retry budget to reach the same outcome. The message is
	// the only signal that separates them. If a provider ever changes the
	// wording we fall back to retrying, which is the previous behaviour.
	if apiErr.StatusCode == http.StatusTooManyRequests && isDailyQuota(apiErr.Message) {
		return fmt.Errorf("daily quota exhausted (429): %s: %w", apiErr.Message, err)
	}

	if apiErr.IsRetryable {
		return &retryableError{
			msg:        fmt.Sprintf("provider returned status %d: %s", apiErr.StatusCode, apiErr.Message),
			retryAfter: retryAfterFromHeaders(apiErr.ResponseHeaders),
		}
	}
	return err
}

// isDailyQuota reports whether a 429 message describes exhausted daily quota
// rather than a per-minute rate limit.
func isDailyQuota(message string) bool {
	msg := strings.ToLower(message)
	return strings.Contains(msg, "quota") &&
		(strings.Contains(msg, "daily") || strings.Contains(msg, "exhaust"))
}

// retryAfterFromHeaders reads a server-suggested wait from the response
// headers GoAI preserved. Keys are matched case-insensitively because header
// casing is not guaranteed across the providers.
func retryAfterFromHeaders(headers map[string]string) time.Duration {
	for k, v := range headers {
		switch strings.ToLower(k) {
		case "retry-after":
			if secs, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && secs > 0 {
				return time.Duration(secs) * time.Second
			}
		case "retry-after-ms":
			if ms, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && ms > 0 {
				return time.Duration(ms) * time.Millisecond
			}
		}
	}
	return 0
}
