# ARCHITECTURE.md — PaperViz (Agent-First)

> **Supersedes** the original `docs/ARCHITECTURE.md` blueprint (no-auth ephemeral MVP, no mention of auth/MCP/OAuth) and the stale addendum in `docs/architecture.md`. That addendum file should be deleted once this doc is adopted — keeping a changelog-style architecture-addendum file alongside a canonical one is exactly the kind of doc-drift source this rewrite exists to close off.
> **Last verified against code:** commit `68d213b`, 2026-09-12. Go backend could not be compiled in the audit sandbox (no access to the Go module proxy to fetch the pinned `go 1.25.0` toolchain) — findings below marked `[STATIC]` are confirmed via source cross-reference, not a compiler run. Re-verify with a real `go build ./...` before treating them as fully closed.

Blueprint Version: 2.0
Architecture Style: Monolith, single binary — **unchanged from v1.0, still the right call, do not split this into microservices for the pivot.**
System Scope: PDF/text ingestion, LLM-based simplification, chart re-visualization + grounding, structured research object extraction, **MCP server as primary agent interface**, human auth/billing/manual-upload as a thin secondary surface.

---

## 1. Context Lock

- Runtime: Go (pinned `go 1.25.0` in `go.mod`), Node 24+ (frontend build only).
- Database: SQLite via `modernc.org/sqlite` (no CGO).
- ORM: NOT_ALLOWED. Raw SQL with `database/sql`.
- Router: `chi`.
- Frontend: React 19 + Vite + Tailwind CSS v4 + shadcn/ui, SPA, static build served by the Go binary.
- Charting: Recharts (frontend only).
- LLM Provider: Google Gemini API, direct HTTP integration. Still NOT_ALLOWED to route through an internal gateway.
- Agent Interface: MCP over stdio (`internal/mcp`, `cmd/mcp`) — **this is now a first-class runtime target, not an optional add-on.** It must build and run correctly on every change to the shared service layer, same bar as the REST server.
- Auth: email/password (`[SHIPPED]`) + Google OAuth (`[BROKEN — see §7]`) + session cookies. `golang.org/x/oauth2` is an approved dependency for this.
- Dependency Direction: `handlers → services → repository → database`, `handlers/mcp → services` for the agent path. STRICT, unchanged. Repository MUST NOT import services or handlers.

---

## 2. Architectural Boundaries

Unchanged from v1.0, and still correctly enforced in the code I read: `handlers` (HTTP request/response only), `services` (business logic), `repository` (SQLite access), `external` (Gemini client, PDF extraction), plus `mcp` (stdio tool adapters — reads from the same `services` layer, does not duplicate business logic; verified against `docs/mcp-parity.md`'s stated architecture rule and the actual `internal/mcp/tools.go` wiring).

**Addition for the agent-first pivot:** the MCP layer must be treated as a peer of the HTTP layer, not a lesser one. Any new service-layer capability that's meant to be agent-reachable needs an MCP tool added in the *same* change, not "later" — `AGENTS.md`'s own rule against shipping an MCP tool without demand signal is about *new speculative* tools, not about keeping existing capability in parity as REST evolves.

---

## 3. Data Model Contract

Base tables (documents, charts, claim_diffs) unchanged from v1.0. Since then, 16 migrations have added: users/sessions (002), chapters (003), chapter-linked charts (004), evidence (005), document title (006), saved papers (007), research collections (008), share tokens (009), document share (010), share referrals (011), usage analytics (012), usage tiers (013), structured research objects (014), evidence graph (015), annotations (016).

**Known gap, confirmed via migration + code cross-reference `[STATIC]`:** `internal/handlers/auth.go`'s `GoogleCallback` constructs a `repository.User{OAuthProvider: ..., OAuthID: ...}` literal and calls `userRepo.UpsertByOAuth(...)`. Neither the `oauth_provider`/`oauth_id` columns (no migration adds them) nor the struct fields/method (`internal/repository/users.go` only defines `ID, Email, PasswordHash, CreatedAt` with `Insert`/`GetByEmail`/`GetByID`) exist. **This is a build-breaking gap, not a schema nice-to-have — nothing OAuth-related can be exercised until it's closed.**

**Required next migration (017), not yet written:**
```sql
-- 017_oauth.sql — add OAuth identity columns to users
ALTER TABLE users ADD COLUMN oauth_provider TEXT;
ALTER TABLE users ADD COLUMN oauth_id TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_id) WHERE oauth_provider IS NOT NULL;
```
Decide at the same time whether `password_hash NOT NULL` stays as-is (OAuth-only users get an empty-string placeholder — verify the login path can never authenticate against that placeholder) or whether it becomes nullable with an explicit `auth_method` column. The latter is architecturally cleaner and is the recommended path; document the decision in `docs/decisions.md` either way, don't leave it implicit.

**API key storage for `/agents` prefill — not designed yet.** `AGENTS.md` currently claims an `api_key` column was added; it was not (grep-confirmed against every migration file). This needs an actual data-model decision (new column on `users`, or a separate `api_keys` table if you want rotation/revocation later) before Chunk 11.3 can be built for real.

---

## 4. Execution Constraints

Unchanged in principle from v1.0 (single synchronous request-scoped processing chain, transactional writes, structured logging, explicit timeouts) — these were good decisions for the original low-traffic human-upload model and remain good decisions in isolation. What needs updating is the concurrency ceiling, because the *traffic shape* changed.

### 4a. Concurrency — the agent-first risk `[STATIC, confirmed via source]`

- `internal/external/gemini.go:33,69` — `sem: make(chan struct{}, 1)`. One in-flight Gemini call for the *entire process*, all callers (REST and MCP) share it.
- `internal/repository/db.go:24` — `db.SetMaxOpenConns(1)`.

Under the pre-pivot model (occasional human upload), this was a reasonable simplification — nobody was going to notice serialized calls when traffic was sparse. Under the agent-first model, this is now the top architectural risk: **an agent calling multiple MCP tools in sequence on one paper, or looping `compare_papers` across several, will queue behind a single global lock the entire time.** This will read as "PaperViz is slow" to exactly the audience the pivot is trying to win.

**Required before any real agent-traffic push (i.e., before Chunk 11.8 distribution):**
- Raise the Gemini semaphore to a small bounded pool (start at 4, tune against your actual Gemini quota tier — don't guess, benchmark).
- Split SQLite access: keep a single writer connection (correct SQLite practice, don't change this part), but allow a small reader pool since WAL mode is already enabled (`journal_mode=WAL`, confirmed in `repository/db.go`).
- Add TTL/eviction to the IP rate limiter map (`internal/handlers/ratelimit.go`) — unrelated to the above two, but same "traffic shape changed" root cause: an unbounded per-IP map was fine for sporadic traffic, less fine under sustained agent call volume.

### 4b. Everything else in this section — unchanged
Transaction policy, logging policy (no full document text in logs), retry policy (1 retry on Gemini failure, no retry on PDF parse failure), validation policy (20MB PDF cap, MIME check, text-layer check) all remain correct and were spot-verified as still present in code. No changes recommended here.

---

## 5. Integration Contracts

### 5a. REST — unchanged in shape, see `docs/openapi.yaml` for the full contract (not reproduced here to avoid a second copy going stale — this doc references it, doesn't duplicate it).

### 5b. MCP — the primary integration contract for the pivot

- 6 tools currently registered (`internal/mcp/tools.go`): `analyze_paper`, `get_summary`, `get_figures`, `get_claims`, `get_evidence`, `compare_papers`.
- Parity map lives in `docs/mcp-parity.md` — **treat that file as living documentation that must be re-verified against `internal/mcp/tools.go` on every MCP-affecting change**, per its own stated rule (`grep -c "Name:" internal/mcp/tools.go` should match the parity table's row count — this is already listed as a verification command in `AGENTS.md`, keep running it).
- Deliberately **not** exposed to MCP: list/save/rename/delete a document, share/unshare, visibility toggle, referral tracking. These are human-preference operations, not research operations — this boundary is correct and should hold as the product evolves. Don't add a "convenience" MCP tool for these without a real demand signal and a documented decision, per `AGENTS.md`'s existing rule.
- MCP input is **text-only**, no PDF upload path — intentional, agents paste text rather than handle binary uploads. Documented in `docs/mcp-parity.md`, confirmed correct.

### 5c. `/agents` page contract — not yet built

The install page needs to produce, per connected client, a config block with the person's API key pre-filled from their session. **This requires the API-key data model decision from §3 to exist first** — there is currently no way to fetch "this signed-in user's API key" because there is no API key at all. Sequence: §3 data model decision → key issuance endpoint → `/agents` page consumes it. Don't build the page UI against a mocked key and leave the real wiring for later; that's exactly the pattern that produced the current OAuth situation (handler code written against a data model that was never actually created).

---

## 6. Non-Goals (unchanged, still enforced)

- No ORM, no job queue/message broker, no microservice split.
- No OCR for scanned PDFs.
- No routing LLM calls through a gateway other than direct Gemini API.
- No persisting uploaded PDF bytes to disk.
- No new MCP tool or SEO/marketing route without a demand signal first (`AGENTS.md` rule, carried forward — this repo has a documented pattern of shipping ahead of validation; the discipline to *not* do that is itself part of the architecture).

---

## 7. Current Known-Broken State (read this before starting any new chunk)

This section exists because the previous architecture doc had no mechanism for flagging "code exists but doesn't work," and that gap let `AGENTS.md` assert the OAuth foundation was "restored" when it in fact does not compile. Keep this section current — update it the moment a listed item is actually fixed and verified, and delete the entry rather than letting it go stale.

| Component | State | Evidence |
|---|---|---|
| Frontend build (`npx vite build`) | **Fails** | `App.jsx` imports `@/pages/account-page` and `@/pages/agents-page`, neither file exists in the repo. |
| `internal/handlers` package (backend) | **Does not compile** `[STATIC]` | `GoogleCallback` references `repository.User.OAuthProvider`/`.OAuthID` and `UserRepo.UpsertByOAuth`, none of which exist in `internal/repository/users.go`. |
| Google OAuth routes | **Not registered** | `GoogleLogin`/`GoogleCallback` defined but absent from `internal/handlers/router.go`. |
| `/agents` page | **0% implemented** | No file. |
| `/account` page | **0% implemented** | No file. |
| CI | **Does not exist** | No `.github/workflows` directory. This is why the above went uncaught. |

**Do not start Chunk 11.2 (Stripe) or continue Chunk 11.3/11.4 polish until every row in this table is cleared.** The dependency ordering in the original pivot plan (11.1 before 11.2/11.3) was correct — it just hasn't actually been satisfied yet, despite `AGENTS.md` describing 11.1 as done.