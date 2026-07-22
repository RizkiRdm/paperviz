# PAPERVIZ — GEMINI 503 RETRY FIX (Chunk 4)

Context: Chunks 2 and 3 were already correct (the per-attempt context is
working—as evidenced by attempt-2 receiving a response in 1308 ms instead of 0 ms).
The current error is “Gemini returned status 503” — this is an actual error from
Google (the Gemini server is overloaded), NOT a rate limit, NOT a code bug.
However, the current retry limit is only 2 attempts with too short a delay, so it still often
fails completely because the 503 error hasn’t resolved by the time the second retry is made.

Work on CHUNK 4 only. Do not change any other parts.

---

## CHUNK 4 — Increase retry count + exponential backoff for 503

TASK:
In the Gemini retry function (which was previously fixed in Chunk 2),
add the following:
1. Increase the maximum retry count to 4 (from 2).
2. Backoff intervals between attempts: 2s, 4s, 8s (increasing with each attempt), not
   retrying immediately without a pause.
3. Retry ONLY if the response status is 503 or 429 or a
   timeout error. For other errors (400, 401, etc.—client errors that won’t
   improve with a retry), do not retry; return the error immediately.

PATTERN EXAMPLE (adjust variable names to match the original code):

```go
const maxRetries = 4

func geminiGenerateWithRetry(prompt string) (*GeminiResponse, error) {
    var lastErr error
    for attempt := 1; attempt <= maxRetries; attempt++ {
        ctx, cancel := context.WithTimeout (context.Background(), 90*time.Second)
        resp, err := callGemini(ctx, prompt)
        cancel()

        if err == nil {
            return resp, nil
        }
        lastErr = err

        if !isRetryable(err) {
            return nil, err
        }

        if attempt < maxRetries {
            backoff := time.Duration(attempt*attempt) * time.Second // 1, 4, 9s -> can be adjusted
            log.Info(“gemini retry backoff”, ‘attempt’, attempt, “wait”, backoff)
            time.Sleep(backoff)
        }
    }
    return nil, fmt.Errorf(“gemini generate failed after %d attempts: %w”, maxRetries, lastErr)
}

func isRetryable(err error) bool {
    if errors.Is(err, context.DeadlineExceeded) {
        return true
    }
    var httpErr *GeminiHTTPError // adjust to match the error struct in the original code
    if errors.As(err, &httpErr) {
        return httpErr.StatusCode == 503 || httpErr.StatusCode == 429
    }
    return false
}
```

MANDATORY RULES:
- The backoff MUST increase with each attempt (not a fixed/constant delay every time).
- If the original code doesn’t have a struct or type to distinguish HTTP
  error statuses (503 vs. 400, etc.), DO NOT create a new, complex abstraction —
  just check `resp.StatusCode` directly where the error is generated,
  then wrap the error message with the status code inside it (example:
  `fmt.Errorf(“gemini returned status %d”, resp.StatusCode)` is already
  in the code; use that for detection—don’t create a new system).
- DO NOT retry for status codes 400/401/403 (client errors; retrying
  is pointless).
- DO NOT change the signature of public functions called from outside
  (the “verify” pipeline stage must still be able to call these functions in
  the same way as before).
- DO NOT touch any logic outside of this retry/backoff mechanism.

PROOF OF COMPLETION:
- Paste the diff (before/after).
- Run `go build ./...` — paste the raw output; it must succeed.
- Run `go vet ./...` — paste the raw output; it must succeed.
- If there are unit tests related to the retry logic, run them and paste the
  raw results. DO NOT modify existing test assertions to make them
  pass by force.

NOTES:
- This DOES NOT fix the 503 itself (that’s a Google-side issue,
  beyond our control). This fix simply makes the system more resilient
  to transient 503s by waiting longer and
  trying more times before giving up on the pipeline.