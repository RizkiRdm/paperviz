# Engineering Log

This document records significant engineering work, not every coding session.

---

## 2026-09-12 — Auth Security Hardening (Signup Restore + 10 Security Fixes)

### Problem

Signup flow returned 404 (`/signup` route deleted in Chunk 11.5 but `login-page.jsx` still linked). Review of `internal/handlers/auth.go` found 10 audit findings, including HIGH OAuth CSRF (hardcoded `state="state"` with no validation).

### Context

* Agent-first pivot removed `/signup` (11.5 Page Cuts); `App.jsx` lost route, `LoginPage` kept `Link to="/signup"` and `navigate("/dashboard")` to deleted page.
* Auth offered email+password (`bcrypt DefaultCost`) + Google OAuth (`oauth2/google`) with session cookie `HttpOnly; Secure; SameSite=Lax; MaxAge 7d`. `hasMinComplexity` existed but never called. `oauth_state` not validated. Logout cookie missing `Secure`. Rate limiter map unbounded.
* `go test 422` passing, single binary + SQLite/WAL, `chi` router.

### Investigation

* Read `login-page.jsx:107` link, `App.jsx` routes (no `/signup`), `auth.go:285` (`AuthCodeURL("state")`), `hasMinComplexity:252` unused, `ratelimit.go:18` unbounded `map[string]*Limiter`, `sessions.go:59` `DeleteExpired` never called, `router.go:146` `/me` unwrapped, `users.go:80` empty `password_hash` for OAuth.

### Hypothesis

404 is route deletion without link cleanup. Security gaps are YAGNI simplifications that became vulnerabilities under agent-first auth reliance (OAuth foundation for Stripe/agents).

### Findings

* `/signup` file existed in `graphify-out` but deleted from `frontend/src/pages/`; route missing → `spaNotFound` fallback → NotFoundPage.
* OAuth `state` hardcoded → CSRF session fixation possible.
* `hasMinComplexity` dead → `aaaaaaaa` accepted.
* `ipRateLimiter.limiters` grows per spoofed IP, no eviction.
* `DeleteExpired` never invoked → table grows.
* `/me` brute-forceable (256-bit token, impractical but defense-in-depth).
* OAuth users `password_hash=""` → empty-password login risk.

### Decision

Restore signup + fix 10 findings in one hardening pass, prioritizing HIGH/MEDIUM, keeping small surface:

* Restore `signup-page.jsx` mirroring `login-page.jsx` tokens, add `App.jsx` `/signup` route, fix login redirect `/dashboard→/account`.
* OAuth: random `state` via `generateSessionToken()` (32B `crypto/rand` hex), store `oauth_state` cookie 10m TTL, validate in callback, clear after.
* Enable `hasMinComplexity` (now requires upper+lower+digit+special), return `password_too_weak`.
* Logout cookie add `Secure:true`.
* `ratelimit.go`: add `created map[string]time.Time` + `purgeOldEntries` + lazy TTL 5m in `getLimiter`.
* `sessions.go`: add `DeleteByUserID`, call `DeleteExpired` on startup (`cmd/server/main.go:121`), call `DeleteByUserID` after `createSessionAndSetCookie` in Login.
* `router.go`: wrap `/me` with `rateLimitAuth`.
* `users.go`: `UpsertByOAuth` sentinel `"oauth-only"`, add `PasswordVersion` field (reserved).

### Alternatives Considered

1. Merge signup into login toggle — rejected: user requested standalone page.
2. PKCE for OAuth — rejected: confidential client, `state` sufficient; upgrade later.
3. LRU cache for limiter — chosen TTL map for minimal dependency (no new dep, keeps `x/time/rate`).

### Trade-offs

**Pros**

* Closes CSRF with zero new dependency, leverages existing `generateSessionToken`.
* TTL map bounds memory without background goroutine.
* Special-char rule raises entropy with negligible UX cost (8-char min already).

**Cons**

* `DeleteByUserID` after creation invalidates other sessions (single-session policy) — may surprise users with multiple devices.
* `PasswordVersion` not enforced yet — future migration needed.

### Implementation

* `frontend/src/pages/signup-page.jsx` (new), `App.jsx:24` route, `login-page.jsx:28` redirect, `auth.go:70-73,197,292-341`, `ratelimit.go:12-58`, `sessions.go:67-74`, `users.go:8-94`, `router.go:146`, `main.go:121`.

### Verification

* `go test ./...` 422 passed in 8 packages (no regression).
* `frontend npm run build` ok (9.44s, chunks 561k).
* Manual: `GET /signup` now serves SignupPage; OAuth flow validates `state` (missing/invalid → 400); weak password `aaaaaaaa` → `password_too_weak`; logout `Set-Cookie` has `Secure`.

### Result

* After: `/signup` functional with email+password + Google OAuth; 10 findings fixed; `vite` build + `422 tests` green; docs updated (`PROJECT_STATE.md` 2026-09-12, `AGENTS.md` H1-H3).

### Lesson

Deletion test failed on page cuts — lazy route removal without link audit. Auth simplicity (`hasMinComplexity` dead, hardcoded state) is security debt; verification via `background_output` + `go test` caught no regression but manual state-mismatch test needed.

### Follow-up

* Enforce `PasswordVersion` on password change (no endpoint yet).
* PKCE if public client needed.
* Periodic `DeleteExpired` ticker (hourly) vs startup-only.
