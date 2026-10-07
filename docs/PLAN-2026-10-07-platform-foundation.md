# PaperViz Platform Foundation — Implementation Plan

> **Status:** proposed, awaiting owner approval. No code has been written.
> **Date:** 2026-10-07 · **Author:** QA audit + design grilling session

## The original brief

> Fix the QA findings, then add product analytics (PostHog), move the frontend to
> Cloudflare Pages, adopt Better Auth or Clerk for authentication, Convex for the
> backend, and R2 for image and PDF storage. Use free-tier infrastructure so a
> free-for-now product can reach other people. Use golang-migrate for migrations.

## What the grilling changed

Six technologies were evaluated. Four were dropped on evidence:

| Requested | Outcome | Reason |
|---|---|---|
| Convex | **Dropped** | No Go SDK. Go works only via generated OpenAPI, and Convex's own docs say queries are "not reactive/real-time". Also violates the written no-ORM rule. |
| Better Auth | **Dropped** | TypeScript server only. Would require adding a Node runtime to a Go-only service. |
| TypeScript integration layer | **Dropped** | Its only justification was Convex. With Convex gone, nothing needs it. |
| R2 for PDFs | **Dropped** | Violates `AGENTS.md`: "MUST NOT persist uploaded PDF bytes to disk". |
| Better Auth **or** Clerk | **Clerk** | Official Go SDK (`clerk-sdk-go/v2`). Free to 50,000 monthly retained users. |
| R2 for images | **Kept, narrowed** | Chart images only. No rule against it; BLOBs bloat the SQLite file. |
| golang-migrate | **Kept** | Its SQLite driver already uses `modernc.org/sqlite` — the same driver as the app. No CGO conflict. |
| Cloudflare Pages | **Kept, reshaped** | Single origin with the Go API proxied behind it, so no CORS and `SameSite=Lax` survives. |
| PostHog | **Kept** | Additive, three events. |

---

## 01 · What the audit found

A full QA pass ran before any of this was planned. The existing gates were **green** —
41/41 Playwright tests, 641 Go tests, clean `gofmt`, clean `go vet`, frontend lint passing —
yet a startup panic had shipped that takes down the entire agent surface.

| Severity | Finding | Evidence |
|---|---|---|
| **HIGH** | MCP server panics at startup; cannot boot | `internal/mcp/schemas.go` uses `jsonschema:"description=…"` in **74 fields**. SDK is `go-sdk v1.7.0`, which panics on that tag. Reproduced in a scratch module. |
| **HIGH** | The gate cannot catch it | `NewMCPServer` is called from exactly one place (`cmd/mcp/main.go`) and by **no test**. The 5 tests in `mcp_contract_test.go:38` hand-build the struct, so `registerTools` — the panicking function — never executes. |
| **MED** | MCP tests cover 12 of 21 migrations | Tests stop at `012_usage_analytics.sql`. The BYOK schema (`020_user_credentials`) and the OAuth/Stripe removal (`021`) are untested on the MCP path. |
| **MED** | Session cookie is `Secure` unconditionally | `auth.go:197,228` hardcode `Secure: true`. On any plain-HTTP origin the browser silently drops it: login appears to succeed, every call 401s. Invisible today because nothing is deployed. |
| **MED** | Weak-password rejection reported as a server fault | Backend returns `password_too_weak`; the frontend map has no entry, so `use-auth-submit.js` falls back to `internal_error` → *"Something went wrong."* |
| **LOW** | Malformed body reported as a size error | `documents_create.go:28` maps every `ParseMultipartForm` error to `file_too_large`, so a JSON body returns "file too large". |

**Docs that overstate current behavior** — the same class of bug as the panic, because the repo lies about its own safety:

- `ARCHITECTURE.md:163` says 19 migrations; **21** exist.
- `docs/mcp-parity.md:25` verifies the tool count with a `grep` that **passes on a server that panics before serving a single tool**.
- `ARCHITECTURE.md:249,255` still lists Google OAuth and Stripe as active controls; both were removed 2026-09-26.

---

## 02 · Target architecture

One Go process on Oracle Cloud Always Free, Cloudflare in front on a single origin.
One database writer. No service split. No ORM.

```
USER PLANE                     EDGE                        ORIGIN
┌──────────────┐          ┌─────────────────┐      ┌──────────────────────┐
│ Browser      │─HTTPS───▶│ Cloudflare      │─────▶│ Go process           │
│ React SPA    │          │ Pages + proxy   │ HTTP  │ Oracle Always Free   │
└──────────────┘          │ R2 (charts)     │      │                      │
┌──────────────┐          │ CDN · TLS · DNS │      │ chi router + SPA     │
│ Agent        │╌╌stdio╌╌▶└─────────────────┘      │ Clerk JWT verify     │
│ Claude Code  │  local                          │ handlers→app→repo    │
└──────────────┘                                  │ pipeline goroutine   │
                                                  └──────────┬───────────┘
                                                             │
                              ┌──────────────────────────────┴──────┐
                              │ SQLite + WAL (one writer)           │
                              │ Clerk (identity, JWKS cached)       │
                              │ PostHog (fire-and-forget)           │
                              │ BYOK provider (user's own key)      │
                              └─────────────────────────────────────┘
```

Single origin is the load-bearing choice: it keeps `SameSite=Lax`, needs no CORS
code (there is currently **zero** CORS config in the Go code), and keeps the cookie
fix optional rather than mandatory.

---

## 03 · Phases

### Phase 1 — Unblock the product surface

Branch `fix/qa-round-1`, six fixes, one commit each. MCP first, because the agent
surface is the product.

| # | Fix | Files |
|---|---|---|
| 1 | Strip the `description=` prefix from 74 struct tags | `internal/mcp/schemas.go` |
| 2 | Rewrite `newTestServer` to call `NewMCPServer`; extend to 21 migrations; assert 5 registered tools | `internal/mcp/mcp_contract_test.go` |
| 3 | Make `Secure` follow an env flag | `internal/handlers/auth.go` |
| 4 | Add the missing `password_too_weak` mapping | `frontend/src/pages/signup-page.jsx` |
| 5 | Split size vs content-type errors | `internal/handlers/documents_create.go` |
| 6 | Header reflects `/api/auth/me` | `frontend/src/pages/upload-page.jsx` |
| 7 | Add MCP handshake + frontend lint to CI | `.github/workflows/*` |

**Fix 1** — the tag form:

```go
// panics: SDK rejects WORD= prefixed tags
Text string `json:"text" jsonschema:"description=Paper text to ingest"`

// correct: the description is the tag body itself
Text string `json:"text" jsonschema:"Paper text to ingest"`
```

**Fix 2 matters more than fix 1.** The tags were not the whole problem — the reason it
shipped is that no test ever ran the code that panics. Rewriting the helper closes
that class of bug permanently, and the new assertion also replaces the `grep` check
in `docs/mcp-parity.md`.

Merge to `main` only when every verification row is green.

### Phase 2 — Standardize schema tooling

Branch `chore/golang-migrate`, isolated because it touches every test that builds an
in-memory database.

- Replace the custom applier in `internal/repository/migrations.go`
- Reconcile 21 existing migrations into golang-migrate's versioned scheme
- Update the DB Reset Protocol in `AGENTS.md`
- **Buys real `Down` rollback**, which `ARCHITECTURE.md:173` explicitly says is missing

### Phase 3 — Deploy and instrument

The testable half. No third-party keys or domain required.

- Oracle Cloud Always Free instance (2 ARM cores, 24 GB RAM)
- Cloudflare single origin: Pages serves the SPA, proxies `/api/*` to the VM
- PostHog: `signup_completed`, `api_key_added`, `paper_analysed`
- R2 for chart images, replacing `image_blob` in SQLite
- ADR-002

### Phase 4 — Hand identity to Clerk

Last, because it needs a live domain and Clerk keys.

- `clerk-sdk-go/v2` middleware; verified via **cached JWKS**, so MCP parity survives a Clerk outage
- Store Clerk's `user_…` ID as-is so all FKs work unchanged
- Drop `users.password_hash`; keep `users` for `created_at` and `api_key_hash`
- MCP service keys stay local digest-based — **not** Clerk M2M tokens
- 12 e2e auth tests need rewriting

---

## 04 · Verification gate

Phase 1 is not done until every row passes.

| Check | Requirement |
|---|---|
| `gofmt -l .` | Must return empty |
| `go vet ./...` | Clean |
| `go build ./cmd/{server,mcp,admin}` | All three compile |
| `go test ./...` | 641 existing + new MCP registration test |
| **MCP stdio handshake** | Launch `cmd/mcp`, send `initialize` + `tools/list`, assert 5 tools. **This is the check that was missing.** |
| Frontend `lint` + `build` | Per the documented gate |
| Playwright e2e | 41/41 must not regress |
| Browser re-test | Weak password, no-credential, provider failure, rate limits |
| Container image build | **Currently blocked** — see below |

---

## 05 · Open items needing the owner

### Disk space blocks Phase 3

`/home` is **99% full — 2.3 GB free**, and podman's storage graph root lives there.
This is what killed the container build during the audit. It will also block Oracle
provisioning and `npm ci`.

`/` is a different filesystem with 19 GB free, so moving podman storage may be the real
fix. Pruning images would reclaim ~6 GB but deletes local state, so it needs your
explicit approval first.

### Branch hygiene

`main` carries **45 uncommitted files**, including `upload-page.jsx` which fix 6
touches. Proposed: name that WIP on a `wip/` branch first, then branch clean for the
fixes, so your work and the fixes never interleave.

### The policy reversal, stated plainly

`AGENTS.md` says PaperViz holds **no third-party account: no OAuth provider, no payment
processor, no model vendor.** Clerk makes that false.

This is deliberate, not drift. The reasoning on record: OAuth is the one dependency
worth an account because it removes all password-security maintenance, and Hobby is
free to 50,000 monthly retained users. If that trade stops being true, email/password
can return.

It must be recorded in `AGENTS.md` **and** ADR-002 together, or the repo will contain
two contradictory statements about its own security posture.

---

## 06 · ADR-002 scope

One ADR covering every reversible policy change and, just as importantly, **every
rejected option**. The dropped list is the valuable part — without it, the next agent
will propose Convex again.

- Clerk as OAuth provider (reverses the 2026-09-26 decision)
- Oracle Cloud as host; Go backend stays a stateful long-lived process
- Single-origin deploy, preserving `SameSite=Lax`
- golang-migrate adopted
- R2 for chart images only
- MCP agent keys remain local digest-based
- **Dropped:** Convex, Better Auth, TypeScript integration layer, Workers, R2-for-PDFs

---

## Summary

Four phases, ordered so each is verifiable before the next begins. Phase 1 fixes a
product surface that currently cannot start. Phases 2–4 are platform work, sequenced so
the testable half ships before the part that needs external keys.

**No code has been written. This plan needs owner approval before execution.**
