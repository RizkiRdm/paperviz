# TASK-1 — Auth Rate Limiting

## Metadata
- id: TASK-1
- priority: P0
- depends_on: none
- estimated_scope: 1 file edit + 1 new function, no new dependency
- session_type: single-session, isolated

## Context (read this, do not skip)

This repo already has a working IP-based rate limiter used on document
creation. It lives in `internal/handlers/ratelimit.go` as the struct
`ipRateLimiter`. Do NOT invent a new rate-limiting mechanism. Reuse the
existing struct and pattern.

`/api/auth/signup` and `/api/auth/login` currently have ZERO rate limiting.
This is a confirmed gap, not a guess — verified by reading
`internal/handlers/router.go` directly.

## Mandatory Investigation Steps (do these BEFORE editing any file)

Run these commands and read the output before writing any code:

```
grep -n "rateLimitDocumentCreate" internal/handlers/router.go
```

```
cat internal/handlers/ratelimit.go
```

```
grep -n "r.Post(\"/signup\"\|r.Post(\"/login\"" internal/handlers/router.go
```

Confirm for yourself:
1. `rateLimitDocumentCreate` is a middleware function that wraps a handler
   using `newIPRateLimiter(rate.Every(...), burst)`.
2. `/api/auth/signup` and `/api/auth/login` are currently registered with
   NO middleware in front of them (bare `r.Post(...)`).

Do NOT modify any code during this step.

## Scope — What To Change

### File 1: `internal/handlers/ratelimit.go`

Add ONE new middleware function, following the exact same pattern as
`rateLimitDocumentCreate` in the same file. Name it `rateLimitAuth`.

Requirements for `rateLimitAuth`:
- Uses the existing `ipRateLimiter` struct — do not create a new struct.
- Uses `clientIP(r)` — the existing helper in the same file — do not
  reimplement IP extraction.
- Rate: 5 requests per 60 seconds per IP, burst 3.
  (`rateLimitDocumentCreate` uses `rate.Every(30*1000_000_000)` i.e.
  nanoseconds-based `rate.Every` — follow the same numeric style, do not
  switch to a different time API.)
- On limit exceeded: call `writeError(w, http.StatusTooManyRequests, "rate_limited")`
  — this is the exact same error call `rateLimitDocumentCreate` already uses.
  Reuse it verbatim, do not invent a new error code string.

### File 2: `internal/handlers/router.go`

Find this block:

```go
r.Route("/api/auth", func(r chi.Router) {
    r.Post("/signup", authHandler.Signup)
    r.Post("/login", authHandler.Login)
    r.Post("/logout", authHandler.Logout)
    r.Get("/me", authHandler.Me)
})
```

Change ONLY the `/signup` and `/login` lines to wrap with the new
middleware, using chi's `.With()` pattern exactly as already used
elsewhere in this same file (see the `docHandler.Create` line for the
exact syntax to copy):

```go
r.With(rateLimitAuth).Post("/signup", authHandler.Signup)
r.With(rateLimitAuth).Post("/login", authHandler.Login)
```

Do NOT touch `/logout` or `/me` — those stay bare, no rate limit.

## Out of Scope — Do NOT Do These

- Do NOT modify `rateLimitDocumentCreate` or its rate/burst values.
- Do NOT modify `internal/handlers/auth.go` (the Signup/Login handler
  logic itself). This task is routing + middleware only.
- Do NOT add rate limiting to any other route not listed above.
- Do NOT add a new Go dependency. `golang.org/x/time/rate` is already
  imported in `ratelimit.go` — reuse it.
- Do NOT change session cookie logic, password hashing, or validation
  logic in `auth.go`.

## Validation — Run These Commands And Paste Raw Output As Proof

```
gofmt -l internal/handlers/
```
Expected: no output (empty = clean).

```
go build ./...
```
Expected: exits 0, no output.

```
go vet ./...
```
Expected: exits 0, no output.

```
go test ./internal/handlers/...
```
Expected: `ok` line, no FAIL.

```
go test ./...
```
Expected: all packages `ok`, same or higher pass count than before this
change. If the pass count drops, STOP and report — do not force-fix by
deleting or skipping a test.

## Acceptance Criteria

- [ ] `rateLimitAuth` function exists in `internal/handlers/ratelimit.go`,
      reusing `ipRateLimiter` and `clientIP`.
- [ ] `/api/auth/signup` and `/api/auth/login` are wrapped with
      `rateLimitAuth` in `router.go`. `/logout` and `/me` are unchanged.
- [ ] `gofmt -l internal/handlers/` outputs nothing.
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` all pass — paste
      raw terminal output as completion proof.
- [ ] Manual proof (if a local server can be started): send 6 POST
      requests to `/api/auth/login` from the same client within 60
      seconds. The 6th response must be HTTP 429 with body containing
      `"rate_limited"`. Paste the raw curl/response output. If a local
      server cannot be started in this session, state that explicitly
      instead of skipping this line silently.

## Completion Report Format

When done, report in this exact structure:

```
TASK-1 STATUS: <done | blocked | partial>
Files changed: <list>
Test result: <paste raw `go test ./...` tail>
Manual 429 check: <done with output | not run — reason>
```
