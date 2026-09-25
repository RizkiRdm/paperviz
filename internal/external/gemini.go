package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"

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

// newGeminiCall builds the Gemini backend: one request, no retry and no
// concurrency control. Both belong to Transport, so that every provider obeys
// the same schedule and the same single concurrency slot.
//
// endpoint is the fully-formatted URL. Tests point it at a local server;
// production formats geminiEndpoint with the model id.
func newGeminiCall(apiKey, endpoint string, httpClient *http.Client) llmCall {
	return func(ctx context.Context, prompt string, asJSON bool, maxTokens int) (string, error) {
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return "", fmt.Errorf("build gemini request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-goog-api-key", apiKey)

		resp, err := httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("gemini http call: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read gemini response: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			slog.Error("gemini rate limited by server", "stage", "gemini_generate")
			if isQuotaExhausted(body) {
				// Daily quota hit. Retrying cannot succeed, and under BYOK the
				// exhausted quota belongs to the user, not to us: fail fast so
				// the pipeline records the failure instead of holding a
				// concurrency slot for the whole retry budget.
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
}

// isQuotaExhausted reports whether a 429 body indicates daily-quota
// exhaustion rather than a per-minute rate limit. Gemini returns 429 for both
// and only the message distinguishes them: per-minute recovers on retry,
// daily quota only resets on a schedule.
func isQuotaExhausted(body []byte) bool {
	var eb geminiErrorBody
	if err := json.Unmarshal(body, &eb); err != nil || eb.Error.Message == "" {
		return false
	}
	msg := strings.ToLower(eb.Error.Message)
	return strings.Contains(msg, "quota") && (strings.Contains(msg, "daily") || strings.Contains(msg, "exhaust"))
}

// parseRetryAfter converts a Retry-After header in seconds into a duration.
// HTTP-date form is not parsed and yields 0.
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

// stripJSONFences removes a surrounding markdown code fence, if present.
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
