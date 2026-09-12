# ADR-001: Auth Session Hardening — Single Session on Login + Strict Password Complexity

**Status:** Accepted

**Date:** 2026-09-12

---

## 1. Context

Agent-first pivot made OAuth foundation for Stripe (`stripe_customer_id` needs real `user_id`) and `/agents` API key prefill. Review of `internal/handlers/auth.go` after `/signup` 404 showed 10 findings: hardcoded OAuth `state="state"` (CSRF), `hasMinComplexity` never called (weak `aaaaaaaa` accepted), logout cookie missing `Secure`, `ipRateLimiter` unbounded map, expired sessions never swept, `/me` unwrapped, OAuth users `password_hash=""`, plus missing session invalidation on login.

## 2. Problem

* Should login invalidate existing sessions or allow concurrent sessions?
* Should password complexity require special character (vs upper+lower+digit only)?
* Must decisions bound rate-limiter memory without new dependency?

## 3. Decision

> We will:
> * On successful `Login`, create new session then `DeleteByUserID(userID)` — single active session per user (best-effort, non-fatal if delete fails).
> * Require `hasMinComplexity` to demand `upper && lower && digit && special (!IsLetter && !IsDigit && !IsSpace)`; return `password_too_weak` after length check.
> * Bound `ipRateLimiter` with `created map[string]time.Time` + 5-min TTL (lazy eviction in `getLimiter` + `purgeOldEntries`), no LRU dep.
> * Use random `state` via `generateSessionToken()` (32B `crypto/rand` hex) stored in `oauth_state` httpOnly Secure 10m cookie and validated in callback.

## 4. Alternatives

### Option A: Keep concurrent sessions (no invalidation)

Description: Allow multiple `session_token` rows per user, as before.

Advantages: Multi-device friendly.

Disadvantages: Session fixation risk retained; compromised token remains valid after password reuse/login.

### Option B: Single session + password_version revocation (chosen partially)

Description: Invalidate on login + add `PasswordVersion int64` column for future password-change revocation.

Advantages: Immediate fixation close; future hook for password-change global logout without extra table.

Disadvantages: Breaks multi-device expectation; `PasswordVersion` not yet enforced — migration debt.

### Option C: Strict LRU cache for limiter

Description: Use `hashicorp/golang-lru` bounded cache.

Advantages: Strict memory bound.

Disadvantages: New dependency, heavier; TTL map is simpler for single-process SQLite MVP.

## 5. Decision Rationale

* **Correctness/security:** CSRF `state` must be unpredictable per OAuth spec; hardcoded string violates. Single-session invalidation closes fixation with minimal code (one `DELETE WHERE user_id=?`), acceptable for MVP without multi-device requirement.
* **Maintainability:** TTL map reuses existing `x/time/rate` + `time.Time`, no dep; sentinel `"oauth-only"` prevents empty-password login without schema change.
* **Cost:** No new infra; `DeleteExpired` on startup suffices for 7-day TTL ephemeral data.

## 6. Consequences

### Positive

* OAuth flow CSRF-resistant; login fixation closed; weak passwords rejected; logout cookie correct on HTTPS; rate-limiter memory bounded.

### Negative

* Single session may surprise users with phone+laptop; needs UX note.

### New Risks

* `DeleteByUserID` after creation leaves window where old sessions remain if delete fails (logged but non-fatal).

## 7. Implementation Impact

* `internal/handlers/auth.go:70-73,146-157,197,263-277,292-341`
* `internal/handlers/ratelimit.go:12-58`
* `internal/repository/sessions.go:67-74`
* `internal/repository/users.go:8-94` (sentinel + PasswordVersion)
* `internal/handlers/router.go:146`
* `cmd/server/main.go:121`

## 8. Validation

* `go test ./...` 422 passed (8 packages) — no regression.
* Manual: OAuth `state` mismatch → 400 `invalid_oauth_state`; `curl -X POST /api/auth/signup` with `password=aaaaaaaa` → 400 `password_too_weak`.

## 9. Reversal / Migration

Reconsider if product needs concurrent sessions (collaborative/ multi-device). Then replace `DeleteByUserID` with session list + `RevokedAt` or `PasswordVersion` check per request, and add `/sessions` management UI.

## 10. Related

* `AGENTS.md` Security & Hardening H1–H3, G1.
* `docs/SECURITY.md` §3, §10.
* `docs/PROJECT_STATE.md` 2026-09-12 auth security hardening.
* Commit `68d213b`.

---

## Status History

| Date | Status | Reason |
|------|--------|--------|
| 2026-09-12 | Accepted | Hardening merged, 422 tests green |
