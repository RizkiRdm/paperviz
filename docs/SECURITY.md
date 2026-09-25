# Security

## 1. Security Scope

Covers authentication, authorization, input validation, file handling, rate limiting, external requests, secrets management, data isolation, and retention for PaperViz (Go `chi` + SQLite, React frontend, Gemini API). Agent/MCP surface supports deterministic text intake and research-data retrieval; it does not expose user-owned management operations.

## 2. Threat Model

### Assets

* User data: email, password_hash, sessions, API keys (`api_key` column)
* Uploaded documents: PDF bytes (in-memory only, not persisted to disk), extracted text, simplified text, charts
* Structured research data: claims, tables, methods, results, citations, annotations, collections
* API credentials: `GEMINI_API_KEY`, `GOOGLE_CLIENT_ID/SECRET`, `STRIPE_SECRET_KEY/WEBHOOK_SECRET`, `PAPERVIZ_API_KEY` (MCP)
* Generated results: share tokens (`share_token` on documents/charts), public share pages
* Stripe customer/subscription state

### Threat Actors

* Unauthenticated user (public share, API enumeration)
* Authenticated user (IDOR on documents/annotations/collections, quota bypass)
* Malicious upload (large PDF, malformed PDF, SSRF via DOI/URL import)
* Compromised Gemini / Crossref / Unpaywall dependency

## 3. Authentication

### Method

* **Email + password**: `POST /api/auth/signup`, `POST /api/auth/login`. `isValidEmail` (3–254 chars, contains `@` and `.`); password ≥8 chars and `hasMinComplexity` requires upper + lower + digit + special (`!unicode.IsLetter && !IsDigit && !IsSpace`). `password_too_weak` on failure.
* **Google OAuth**: `GET /api/auth/google/login` → redirect with random `state` (32-byte `crypto/rand` hex, stored in `oauth_state` cookie `HttpOnly; Secure; SameSite=Lax; MaxAge=600`), `GET /api/auth/google/callback` validates `state` cookie vs query, clears cookie, exchanges code via `oauth2/google`, fetches `https://www.googleapis.com/oauth2/v2/userinfo`, `UpsertByOAuth` (sentinel `password_hash="oauth-only"` when empty, `ON CONFLICT(email) DO UPDATE`).

### Session / Token Strategy

* `generateSessionToken`: 32 random bytes → 64-char hex (`crypto/rand`).
* `sessions` table: `token PK, user_id, expires_at (7 days)`. `SessionRepo.Get` checks expiry.
* Cookie: `session_token` `HttpOnly; Secure; SameSite=Lax; MaxAge=7d; Path=/`. Logout clears with `Secure: true` and deletes row.
* **Login invalidation**: `Login` creates new session *then* `DeleteByUserID(userID)` (best-effort, non-fatal if fails). No global logout-all on password change (no password-change endpoint).
* **Expiry cleanup**: `DeleteExpired()` on server startup (`cmd/server/main.go:121`). No periodic sweeper.
* `PasswordVersion int64` field added to `User` struct (reserved, not yet enforced).

### Credential Storage

* `bcrypt.GenerateFromPassword` with `DefaultCost` (10). `CompareHashAndPassword` on login. Generic `invalid_credentials` on mismatch — no user enumeration via timing distinction (same 401 for missing email vs bad password). OAuth users never store Google token.

## 4. Authorization

No RBAC. Single `user` role. Access controlled by resource ownership.

### Resource Ownership

| Resource | Check | Handler enforcement |
|----------|-------|---------------------|
| Documents | `user_id` on insert; `RequireAuth` for save/rename/delete/list/stats | `documents.go` |
| Annotations | `service` checks `userID` matches annotation owner; `403 Forbidden` on mismatch | `annotations.go` |
| Collections | `collections.go` `Get/Rename/Delete/Add/Remove/ListDocuments(userID)` → `403` on `ErrForbidden` | `services/collections.go` + `handlers/collections.go` |
| Share tokens | Figures inherit document `visibility`; share service gates visibility/expired checks | `share.go` |
| Export | `ExportResearchContext` excludes `OriginalText`/`SimplifiedText` (copyright) | `export.go` |
| Account / API key / Billing | `RequireAuth` | `router.go` |
| MCP tools | Separate interface, deterministic intake/retrieval, `PAPERVIZ_API_KEY` required, no user-library access | `internal/mcp/server.go` |

`OptionalAuth` for document creation: attaches `userID` only if valid session; anonymous allowed but usage tracked by fingerprint.

## 5. Input Validation

### API Input

* `readJSON` with `http.MaxBytesReader(nil, r.Body, 1<<20)` (1 MiB JSON limit) in `respond.go`.
* Email: `strings.ToLower + TrimSpace`, `isValidEmail`.
* Password: length + `hasMinComplexity`; `invalid_email`, `password_too_short`, `password_too_weak`.
* Errors use `writeError` with generic codes (`invalid_request`, `internal_error`) — no stack traces to client; `slog.Error` server-side only.

### File Uploads

* Multipart: `ParseMultipartForm(20 << 20)` (20 MiB), `http.MaxBytesReader(w, r.Body, 20MiB+1MiB slack)`, `if header.Size > 20MiB` → reject. PDF bytes never persisted to disk (in-memory extraction via `pdfcpu`/`ledongthuc/pdf`, pinned `go.mod`).
* Type: handled as PDF only; text-layer extraction, no OCR.

### User-Provided URLs (DOI/URL import)

* `internal/services/import.go: FetchByDOI/FetchByURL` → `https`-only, SSRF protection: block private IPs (`10/8, 172.16/12, 192.168/16, 127/8`, etc.) via `net.ParseIP` + `IsPrivate`. `golang.org/x/net` not used for fetch — standard `net/http` with timeout.

### Model Output

* Gemini output validated via `ValidateGrounding` (10 rules, `unsupported → DO NOT RENDER`), `chartValues` lenient unmarshaler, grounding status badge. Pipeline populates `claims` from verification output in same tx.

## 6. File / Document Security

* **Size**: 20 MiB hard limit (`maxUploadBytes`).
* **Type**: PDF only; request `Content-Type` not trusted, bytes inspected by PDF parser.
* **Storage**: SQLite `documents` + `charts` (`image_blob BLOB` capped at 5 images/doc `maxImageChartsPerDocument`). No file on disk, no S3. Ephemeral 7-day TTL policy (document/architecture).
* **Isolation**: `WAL` + `PRAGMA synchronous=NORMAL`; DB file gitignored; reset requires deleting `paperviz.db*`.

## 7. Agent Security

### Tool Permissions

* MCP server (`internal/mcp`): stdio, `5` tools — `ingest_document`, `search_documents`, `get_document`, `get_figures`, `get_evidence`. `ingest_document` is deterministic text intake; the other tools retrieve shared research data. No tool calls Gemini.
* MCP is stateless and text-only. It does not carry REST session ownership or expose user library operations.

### Allowed / Restricted

* Allowed: deterministic pasted-text intake and research data retrieval through shared services/repositories (`docs/mcp-parity.md`).
* Restricted: no user-owned operations (`list/save/rename/delete/share/referral`) via MCP; no hidden LLM calls; `PAPERVIZ_API_KEY` auth, per-key rate limiting (ingest 5/min, read 30/min), 500 KiB input cap, and concurrent job limiter.

### Prompt Injection

* **Current**: Document text is never concatenated into agent system prompts that grant tool control; `external/gemini.go` uses text-only prompts with JSON schema expectations. No explicit injection filter.
* **Status**: Partially implemented — grounding validator prevents model-manufactured evidence but no dedicated prompt-injection sanitizer. Risk documented as known limitation.

## 8. Secrets

| Secret | Purpose | Storage |
|--------|---------|---------|
| `GEMINI_API_KEY` | LLM calls (direct HTTP, no gateway) | env, `os.Getenv` |
| `GOOGLE_CLIENT_ID/SECRET/REDIRECT_URL` | OAuth | env |
| `STRIPE_SECRET_KEY/WEBHOOK_SECRET/PRICE_PRO/PRICE_RESEARCH` | Billing | env |
| `PAPERVIZ_API_KEY` | MCP per-key auth | env (MCP) |
| `session_token`, `oauth_state` cookies | Session | `HttpOnly; Secure; SameSite=Lax` |

*Exposure prevention*: No `slog.Info("gemini debug"... )` (removed A1); `slog.Error` never logs full document text (ARCHITECTURE Logging Policy); no hardcoded secrets; `.env.example` documents required vars; MCP requires `PAPERVIZ_API_KEY`.

## 9. Data Isolation

* SQLite single file; tenant isolation via `user_id` column + ownership checks in service layer (`ErrForbidden → 403`). No RLS (SQLite). `X-Frame-Options`, `CSP` headers via `SecurityHeaders` middleware.

## 10. Rate Limiting

* `ipRateLimiter` (`golang.org/x/time/rate`, `map[string]*rate.Limiter + created map[string]time.Time`, `sync.Mutex`, TTL 5 min eviction via `purgeOldEntries` + lazy refresh in `getLimiter`).

| Scope | Endpoints | Limit |
|-------|-----------|-------|
| Document create | `POST /api/documents`, `POST /api/import/doi|url` | `1 req/30s, burst 2` (`rateLimitDocumentCreate`) |
| Auth | `POST /api/auth/signup|login`, `GET /api/auth/me` | `5 req/60s, burst 3` (`rateLimitAuth`) |
| MCP (per-key) | `ingest_document 5/min; read/search 30/min` | token bucket (`internal/mcp/ratelimit.go`) |

*Only POST wrapped for documents; GET unrestricted. No user-based limiter beyond IP.

## 11. External Services

| Service | Data Sent | Risk | Mitigation |
|---------|-----------|------|------------|
| Gemini API (direct HTTP) | Extracted text / evidence (never full PDF bytes to disk) | Prompt data exfil, quota burn | 5 image charts/doc cap, `dualClaimExtractionPrompt` reduces calls, structured JSON validation |
| Google OAuth (`googleapis.com/oauth2/v2/userinfo`) | `code` exchange, bearer token | Token leakage | `oauth_state` CSRF nonce, `Secure` cookies, token not persisted |
| Crossref / Unpaywall (DOI import) | DOI / URL | SSRF, large fetch | https-only, private-IP block, 100 MiB read cap |
| Stripe | checkout/portal, webhook | Webhook spoof | HMAC via `stripe-go/v79` (verify signature) |

## 12. Logging & Sensitive Data

* `log/slog` structured JSON. `slogRequestLogger`: `id, method, path, status, duration` with `Warn` ≥400, `Error` ≥500. PDF timeout branches log `Warn` (observable leaked goroutines). Gemini calls log stage/duration, not text.
* **Must not log**: full document text, `password`, `password_hash`, `session_token`, `oauth_state`, Gemini URL with key, Stripe secrets.

## 13. Data Retention

* Documents ephemeral 7-day expiry (last activity), single flat migration, DB reset protocol: `rm paperviz.db*` + `make dev`.
* Share links inherit visibility + share_token lifecycle (`unlisted→private` clears token, `private` clears token).

## 14. Security Testing

| Area | Method | Status |
|------|--------|--------|
| Authentication | table-driven `go test ./...` (422 passing) + manual OAuth CSRF verification (`state` mismatch → 400) | Implemented |
| Authorization | ownership checks covered by `collections_test.go`, `annotations` tests (`ErrForbidden` → 403) | Implemented |
| Input validation | `isValidEmail`, `hasMinComplexity` unit paths; `MaxBytesReader` limit | Implemented |
| File handling | `pdf.go` timeout + size tests | Implemented |
| Prompt injection | Grounding validator regression tests (unsupported → not rendered) | Partial |

## 15. Known Security Risks

* OAuth: no PKCE (authorization-code flow without PKCE; `state` is CSRF-only, not code-verifier). Low for confidential client.
* Rate limiter in-memory only; resets on restart; no distributed limiter.
* `PasswordVersion` exists but not enforced — password change does not revoke sessions (no endpoint).
* Prompt injection: no dedicated sanitizer; relies on grounding validator.
* Expired sessions cleaned only on startup, not periodically.
* Stripe webhook must be configured manually in dashboard (`/api/billing/webhook`).

## 16. Security Decisions

See `DECISIONS/` (e.g., `ADR-001` auth hardening) and `AGENTS.md` Security & Hardening (A1–H3).
