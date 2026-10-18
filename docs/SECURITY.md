# Security

## 1. Security Scope

Covers authentication, authorization, input validation, file handling, rate limiting, external requests, secrets management, data isolation, and retention for PaperViz (Go `chi` + SQLite, React frontend, provider-agnostic model access under BYOK). Agent/MCP surface supports deterministic text intake and research-data retrieval; it does not expose user-owned management operations.

## 2. Threat Model

### Assets

* User data: account email, service API keys (`api_key_hash` column, SHA-256 digest), Clerk subject id
* Uploaded documents: PDF bytes (in-memory only, not persisted to disk), extracted text, simplified text, charts
* Structured research data: claims, tables, methods, results, citations, annotations, collections
* User model credentials: encrypted blobs in `user_credentials` (AES-256-GCM), `key_hint`
* Process secrets: `CREDENTIAL_ENCRYPTION_KEY` (server), `CLERK_SECRET_KEY` (server), `PAPERVIZ_API_KEY` (MCP only)
* Generated results: share tokens (`share_token` on documents/charts), public share pages

### Threat Actors

* Unauthenticated user (public share, API enumeration)
* Authenticated user (IDOR on documents/annotations/collections, quota bypass)
* Malicious upload (large PDF, malformed PDF, SSRF via DOI/URL import)
* Compromised Gemini / Crossref / Unpaywall dependency

## 3. Authentication

### Method

* **Clerk** owns signup, sign-in, session lifecycle, and password recovery. PaperViz has no local credential path: `/api/auth/signup` and `/api/auth/login` do not exist and a test asserts they return 404. See `docs/DECISIONS/ADR-003-clerk-auth.md`.

### Session / Token Strategy

* Clients send `Authorization: Bearer <clerk session token>`.
* `clerkhttp.WithHeaderAuthorization()` verifies the signature against Clerk's JWKS, caches keys, and re-fetches on rotation. A missing or unverifiable token yields 401; there is no fallback identity.
* `users.id` holds the Clerk subject. The local profile row is provisioned on the first authenticated request, not at signup, because Clerk owns the signup flow.
* The `sessions` table, `session_token` cookie, and bcrypt password handling were removed in migration 022 — a local session is a bearer token Clerk cannot revoke.

### Credential Storage

* PaperViz stores **no password**: signup, sign-in, password hashing, and password recovery are Clerk's responsibility (ADR-003). Locally the only credential material is the encrypted model-key ciphertext in `user_credentials` and the `api_key_hash` digest.

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
| Account / API key / model credentials | `RequireAuth` | `router.go` |
| Section + relationship reads (claims, tables, methods, results, citations, evidence-graph, research-map, relationships) | `OptionalAuth` + `RequireDocAccess` ownership guard — 403/404 for non-owners of private docs (test `TestSectionRoutesRespectOwnershipOverRouter`) | `documents.go` (`RequireDocAccess`) |
| MCP tools | Separate interface, deterministic intake/retrieval, `PAPERVIZ_API_KEY` required, no user-library access | `internal/mcp/server.go` |

`OptionalAuth` for document creation: attaches `userID` only if a valid Clerk token is present; anonymous allowed but usage tracked by fingerprint.

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

* **Current**: Document text is never concatenated into agent system prompts that grant tool control; `external/llm.go` uses text-only prompts with JSON schema expectations. No explicit injection filter.
* **Status**: Partially implemented — grounding validator prevents model-manufactured evidence but no dedicated prompt-injection sanitizer. Risk documented as known limitation.

## 8. Secrets

| Secret | Purpose | Storage |
|--------|---------|---------|
| `CREDENTIAL_ENCRYPTION_KEY` | AES-256-GCM key for stored user model keys | env, `os.Getenv`; never persisted |
| `PAPERVIZ_API_KEY` | MCP per-key auth | env (MCP) |
| `CLERK_SECRET_KEY` | Verifies session tokens, resolves account email | env, `os.Getenv`; never persisted |
| `VITE_CLERK_PUBLISHABLE_KEY` | Clerk browser SDK | env, frontend build; safe to expose |

*User model keys* are the other secret class. They are supplied by the user through the web UI, stored as AES-256-GCM ciphertext with AAD bound to `user_id|provider|model`, and returned only as `key_hint` (last 4 characters). The server holds no model vendor credential of its own, so no server-side key may be introduced — a user with no credential gets `missing_credential`, which is the correct answer.

*Exposure prevention*: No `slog.Info("gemini debug"... )` (removed A1); `slog.Error` never logs full document text (ARCHITECTURE Logging Policy); no plaintext model key is ever logged, returned, or stored; no hardcoded secrets; `.env.example` documents required vars; MCP requires `PAPERVIZ_API_KEY`.

## 9. Data Isolation

* SQLite single file; tenant isolation via `user_id` column + ownership checks in service layer (`ErrForbidden → 403`). No RLS (SQLite). `X-Frame-Options`, `CSP` headers via `SecurityHeaders` middleware.

## 10. Rate Limiting

* `ipRateLimiter` (`golang.org/x/time/rate`, `map[string]*rate.Limiter + created map[string]time.Time`, `sync.Mutex`, TTL 5 min eviction via `purgeOldEntries` + lazy refresh in `getLimiter`).

| Scope | Endpoints | Limit |
|-------|-----------|-------|
| Document create | `POST /api/documents`, `POST /api/import/doi|url` | `1 req/30s, burst 2` (`rateLimitDocumentCreate`) |
| MCP (per-key) | `ingest_document 5/min; read/search 30/min` | token bucket (`internal/mcp/ratelimit.go`) |

*Only POST wrapped for documents; GET unrestricted. No user-based limiter beyond IP.

## 11. External Services

| Service | Data Sent | Risk | Mitigation |
|---------|-----------|------|------------|
| Model providers (Gemini / Anthropic / OpenAI, direct via `goai`, per-request user key) | Extracted text / evidence (never full PDF bytes to disk) | Prompt data exfil, quota burn (user's own spend) | 5 image charts/doc cap, `dualClaimExtractionPrompt` reduces calls, structured JSON validation; no server-held provider credential |
| Clerk (authentication) | Account email, session-token verification claims | Identity SaaS dependency, lock-in | JWKS signature verification, no local credential path, ADR-003 risk register |
| Crossref / Unpaywall (DOI import) | DOI / URL | SSRF, large fetch | https-only, private-IP block, 100 MiB read cap |

## 12. Logging & Sensitive Data

* `log/slog` structured JSON. `slogRequestLogger`: `id, method, path, status, duration` with `Warn` ≥400, `Error` ≥500. PDF timeout branches log `Warn` (observable leaked goroutines). Gemini calls log stage/duration, not text.
* **Must not log**: full document text, plaintext model keys, `CREDENTIAL_ENCRYPTION_KEY`, `CLERK_SECRET_KEY`, `PAPERVIZ_API_KEY`, Gemini URL with key.

## 13. Data Retention

* Documents ephemeral 7-day expiry (last activity), single flat migration, DB reset protocol: `rm paperviz.db*` + `make dev`.
* Share links inherit visibility + share_token lifecycle (`unlisted→private` clears token, `private` clears token).

## 14. Security Testing

| Area | Method | Status |
|------|--------|--------|
| Authentication | table-driven `go test ./...` + Clerk token verification tests; local signup/login 404 asserted by test | Implemented |
| Authorization | ownership checks covered by `collections_test.go`, `annotations` tests (`ErrForbidden` → 403); router-level section guard `TestSectionRoutesRespectOwnershipOverRouter` | Implemented |
| Input validation | `isValidEmail` unit paths; `MaxBytesReader` limit | Implemented |
| File handling | `pdf.go` timeout + size tests | Implemented |
| Prompt injection | Grounding validator regression tests (unsupported → not rendered) | Partial |

## 15. Known Security Risks

* Clerk lock-in: a Clerk outage or account suspension blocks all sign-in; no local fallback path may be added (ADR-003).
* Rate limiter in-memory only; resets on restart; no distributed limiter.
* Prompt injection: no dedicated sanitizer; relies on grounding validator.

## 16. Security Decisions

See `DECISIONS/` (e.g., `ADR-001` auth hardening) and `AGENTS.md` Security & Hardening (A1–H3).
