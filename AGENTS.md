# AGENTS.md — PaperViz

## Quick Rules
- MUST follow ARCHITECTURE.md layer boundaries (`handlers → services → repository/external`) without exception.
- MUST NOT introduce an ORM, job queue, message broker, or microservice split — see ARCHITECTURE.md Non-goals.
- MUST NOT persist uploaded PDF bytes to disk.
- MUST NOT route LLM calls through a gateway we do not control. Calls go direct to the user's own provider with their own key, resolved per request.
- MUST run tests before marking any task complete.
- MUST only inspect files directly related to task given. Do not scan the entire repository, do not read files in `@.gitignore`, do not inspect unrelated documents.

## Design System Rules
For any task that modifies or generates UI, styling, layout, or components:
1. Read `@DESIGN.md` completely before implementation.
2. Treat it as the absolute source of truth.
3. Do not invent any colors, spacing, or fonts outside of these rules.

## Project Context
PaperViz converts academic papers (PDF/pasted text/DOI/URL) into simplified-language versions, verified claims, structured research objects, and evidence-grounded figures. Product direction is agent-first, with web auth, model-key management, and manual ingestion as supporting human surfaces. PaperViz holds no third-party account: no OAuth provider, no payment processor, no model vendor. Full context in `PRODUCT.md`, `docs/PRD.md`, and `docs/overview.md`. Solo-dev project with ~1hr/day human oversight — agent autonomy within a phase is expected, but cross-phase scope changes require explicit human sign-off.

## Tech Stack
- Backend: Go 1.25, `chi` router, `modernc.org/sqlite` (no CGO), raw `database/sql`.
- Frontend: React 19 + Vite 8 + Tailwind CSS 4 + shadcn/ui, Recharts for chart rendering.
- LLM: provider-agnostic HTTP client with one backend per vendor. Keys are supplied per request (BYOK) from the user's own stored credential; the server holds no model vendor account.
- PDF processing: Go-native text/image extraction libraries (pinned exact versions in `go.mod`).
- No Docker/orchestration requirement for MVP — single binary + SQLite file.
- Agent access: `internal/mcp` (stdio MCP server, `cmd/mcp`), stateless text intake and research-data retrieval, separate interface sharing services/data with REST. See `docs/mcp-parity.md`.

## graphify
Code graph at `graphify-out/`. Query before grep/read.

Rules:
- `graphify query "<question>"` for scoped subgraph
- `graphify path "<A>" "<B>"` for relationships
- `graphify explain "<concept>"` for focused concepts
- GRAPH_REPORT.md only for broad architecture
- After code changes, run `graphify update .`

## Known Issues
- **Silent catches banned in frontend (12.1 precedent).** Every `catch` MUST handle: user-facing inline error + Retry, dev `console.error`, preserve inputs; `research-map.jsx` block is canonical standard.
- **Ponytail full-mode convention (12.1 precedent).** Ceiling comments only, format `// ponytail: <simplification> — ceiling: <limit> ; upgrade: <path>`; never logic with the marking; `comparison.go` 5 hits is canonical example.
- **Chart image serving exists.** Stored `image_blob` values are served through `GET /api/documents/:id/charts/:chartId/image`; the frontend still distinguishes original figure bytes from PaperViz interpretation. Keep endpoint MIME validation and document ownership scoping intact.
- **WAL mode required after DB reset.** Enabled via PRAGMA journal_mode=WAL + synchronous=NORMAL in `repository/db.go`. DB file must be deleted when schema changes (single flat migration). Already gitignored.
- **B3 (single-call verification) skipped** — requires live API key + real-document regression testing against B2 before shipping. Plan still describes it as optional.
- **Annotations require authentication.** Unauthenticated users cannot create/edit/delete annotations. Export endpoint also requires auth. Design decision: annotations are per-user research context, not collaborative.
- **Verification is not decoration (12.1 precedent).** `mismatch_detail` is evidence, not UI polish — always surface to user when status is `verification_failed`; pipeline populates claims table from verification output in the same tx, no separate LLM extraction step.
- **Service API keys are stored as digests.** `users.api_key_hash` holds a SHA-256 hex digest, never the key. `GET /api/auth/apikey` returns only `{configured: bool}`; the plaintext is returned once by `POST /api/auth/apikey` or `/regenerate`. A lost key is replaced, never recovered. Do not "helpfully" make it retrievable again.
- **Signup page restored.** Chunk 11.5 deleted it but login still linked to /signup. Restored with email+password.
- **Master refactor in progress (P01-P08).** Canonical flow Input→Processing→Understanding→Evidence→Figures→Source→Mgmt locked (P02). Routing/IA cleaned: dead routes removed, /dashboard→/account redirects fixed (P03). Result page decomposed 743→297 LOC into 6 components in `frontend/src/components/result/` (P04). Ingestion unified with source_type badge + inline error Retry (P05). Processing states 5 labels (P06). App service layer `internal/app/documents/service.go` + `readmodel.go` started, handler split `documents_create.go` (P07-P08). Plan at `.omo/plans/paperviz_master_refactor_plan.md`. Next: finish handler→repo decoupling (P09) and read-model aggregation (P10) before pipeline split.

## Agent-First Pivot (Chunk 11)
- **Agent-first direction locked.** PaperViz becomes agent-first, Context7-style. Human touches the product for 3 things only — auth, model-key management, and an optional manual-upload escape hatch. The MCP server is the actual product surface.
- **OAuth and Stripe are gone (2026-09-26).** Both existed only to satisfy a third-party account, which is a single point of catastrophic failure for a solo maintainer. Auth is email+password with `cmd/admin` recovery; model access is BYOK. See `.omo/plans/byok_auth_billing_removal.md`. Do not reintroduce either without an explicit owner decision.
- **Page cuts complete.** Removed `/dashboard`, `/pricing`, `/compare`, `/compare-research-papers`, `/research-paper-summarizer`, `/figure-explanation`, `/explain/:slug`. Kept `/`, `/agents`, `/login`, `/signup`, `/account`, `/upload`.
- **Landing page minimal.** Single above-the-fold section with 2 CTAs: "Add to Claude Code" and "Sign in".

## DB Reset Protocol
When schema changes (new column/table):
1. `kill` server process
2. `rm paperviz.db paperviz.db-wal paperviz.db-shm paperviz.db-journal` (all present)
3. `make dev` — fresh DB, fresh schema, fresh WAL
4. Data ephemeral (7-day expiry). No migration runner yet.

## Security & Hardening (Round 2 — applied July 2026)
1. A1: Removed `slog.Info("gemini debug", ...)` from `gemini.go:174` — was leaking model name + URL on every call.
2. A2: Added IP-based rate limiting on `POST /api/documents` (1 req/30s, burst 2) via `golang.org/x/time/rate`. New file `internal/handlers/ratelimit.go`. Only POST wrapped, GET unrestricted.
3. A3: Added `slog.Warn` logs on PDF extraction timeout branches in `pdf.go` — makes leaked goroutines observable in logs.
4. B1: Capped `maxImageChartsPerDocument = 5` in `pipeline.go` — bounds free-tier quota burn from image charts.
5. B2: Merged claim extraction from 2→1 Gemini call using `dualClaimExtractionPrompt` in `verification.go`. DiffClaims now 2 calls (down from 3).
6. C1-C3: Replaced full-text-scan chart pipeline with chapter-based approach:
   - New `internal/services/chapters.go`: `DetectChapters()` splits simplified text into ≤10 chapters
   - New `GenerateChapterCharts()` in `charts.go`: evidence→datasets→LLM plan, multiple charts per chapter
   - Old `ExtractChartsFromText`, `fullTextChartPrompt` removed. `textChartElem` purged 2026-09-04 (12.1 dead-code slice; C-series cleanup closed).
   - Image fallback path (`ReVisualizeCharts`) unchanged.
7. D1: Annotations enforce per-user ownership — service layer checks `userID` matches before update/delete. 403 returned on ownership mismatch.
8. D2: Export endpoint excludes `OriginalText` and `SimplifiedText` — copyright compliance. Only structured metadata and user annotations are exported.
9. D3: Collections enforce per-user ownership — service Get/Rename/Delete/Add/Remove/ListDocuments take userID, ErrForbidden on mismatch; handler maps forbidden→403. Closes IDOR, mirrors D1.
10. E1: Added auth rate limiting on `POST /api/auth/signup` and `POST /api/auth/login` (5 req/60s, burst 3) via `rateLimitAuth` middleware reusing existing `ipRateLimiter` struct.
11. F1: Fixed silent zero-fill in chart rendering (`frontend/src/components/data-chart.jsx`) — missing values (undefined/null) now excluded from chart data instead of defaulted to 0; prevents misleading visualization where missing data points appear as real zeros.
12. G1: Chart engine rework (C2-C17) — evidence extraction pipeline replaces LLM value generation:
    - `internal/models/evidence.go`: NumericEvidence + EvidenceSource types
    - `internal/models/dataset.go`: CandidateDataset + DatasetPoint (group by metric+unit)
    - `internal/models/chart_spec.go`: ChartSpec + BarData/LineData/ScatterData/PieData + ChartProvenance
    - `internal/services/evidence_extract.go`: ExtractNumericEvidence (5 regex patterns, deterministic metric detection)
    - `internal/services/table_extract.go`: ExtractTableData (pipe/tab-separated tables)
    - `internal/services/dataset_build.go`: BuildCandidateDatasets (evidence → chart-ready datasets)
    - `internal/services/grounding.go`: ValidateGrounding (10 deterministic rules, unsupported → DO NOT RENDER)
    - `internal/services/charts.go`: GenerateChapterCharts (evidence→datasets→LLM plan, multi-chart per chapter)
    - Failure categories: EXTRACTION_ERROR, DATASET_ERROR, CHART_SELECTION_ERROR, GROUNDING_ERROR, SCHEMA_ERROR, RENDER_ERROR
    - Frontend: grounding status badge, provenance display, validation on render
    - Core principle: AI may transform evidence, but must never manufacture evidence
13. H1: OAuth state parameter was hardcoded string — CSRF vulnerability. Fixed: random nonce generated via crypto/rand, stored in httpOnly cookie (10min TTL), validated in callback. **Obsolete as of 2026-09-26: Google OAuth was removed, so this fix no longer has code to apply to. Kept for the audit trail; H2/H3 numbering depends on it.**
14. H2: hasMinComplexity function existed but was never called — weak passwords allowed. Fixed: added check after length validation, returns password_too_weak. (`internal/handlers/auth.go` Signup)
15. H3: Logout cookie missing Secure flag — cookie mismatch on HTTPS. Fixed: added Secure: true. (`internal/handlers/auth.go` Logout)

## Security & Hardening (Round 3 — BYOK, applied 2026-09-26)
Plan: `.omo/plans/byok_auth_billing_removal.md`. Threats addressed: (a) key theft, (c) mass signup, (d) database theft.

- **The server holds no model vendor credential.** Every model call runs on a key the user supplied, decrypted per request by `internal/app/credentials.Resolver`. There is no process-wide client and no fallback. Never reintroduce a server-side key, and never "temporarily" fall back to one when a user has no credential — `missing_credential` is the correct answer.
- **User model keys are encrypted at rest (AES-256-GCM).** The database is not the trust boundary: `CREDENTIAL_ENCRYPTION_KEY` lives in the process environment and is never persisted. A stolen dump is inert. AAD binds each blob to `user_id|provider|model`, so a row cannot be replayed under a different account. `CREDENTIAL_ENCRYPTION_KEY` fails loud at startup, unlike the peripheral integrations that used to be required — a wrong key silently makes every stored credential unreadable while the server looks healthy.
- **Never log, return, or store a plaintext model key.** The credential API returns `key_hint` (last 4 characters) and never the key. `CredentialAAD` must be used for both encrypt and decrypt; a mismatch makes credentials permanently unreadable with no way to tell "wrong key" from "wrong AAD".
- **Service API keys are stored as digests.** `users.api_key_hash` holds a SHA-256 hex digest, never the key. `GET /api/auth/apikey` returns only `{configured: bool}`; the plaintext is returned once by `POST /api/auth/apikey` or `/regenerate`. A lost key is replaced, never recovered. Do not "helpfully" make it retrievable again.
- **Ingestion can be stopped without a deploy.** `INGESTION_ENABLED=false` returns 503 on document creation while read, share, and library paths keep working. `PIPELINE_MAX_CONCURRENCY` (default 2) bounds concurrent pipelines. The slot is taken inside `RunPipelineAndPersist`, not at the request boundary: the pipeline runs in a detached goroutine, so a handler-side cap would bound submissions rather than CPU.
- **Only the Gemini backend is implemented.** `anthropic` and `openai` are accepted by the credential API and stored, but resolving one fails until their backends land. Do not present them as working in the UI.
- **No self-service password reset exists by design.** It would require an email provider account, which this project refuses to hold. Use `go run ./cmd/admin reset-password <email>`, which also revokes that account's live sessions.

---

## Coding Conventions
- MUST use one line comment in every function made.
- MUST follow standard Go formatting (`gofmt`) — non-negotiable, run before every commit.
- MUST use explicit error returns (`if err != nil`) — no panics for expected error paths (validation failures, API timeouts). Panics reserved for truly unrecoverable states (e.g., failed DB connection at startup).
- MUST NOT use global mutable state for request-scoped data (document ID, request context) — pass explicitly.
- Service functions MUST be pure with respect to side effects where possible: `func Simplify(text string, level string) (string, error)` — no hidden reads from global config inside business logic functions; pass config explicitly.
- Naming: Go idiomatic (`CamelCase` exported, `camelCase` unexported). No Hungarian notation, no abbreviation-heavy naming.
- Logging: structured, via standard library `log/slog` — MUST NOT log full document text content (see ARCHITECTURE.md Logging Policy).

## Testing Rules
- Every service function MUST have a table-driven unit test: minimum 1 success case, 1 error case.
- Pipeline integration test REQUIRED covering all 4 Acceptance Scenarios and 4 Failure Scenarios listed in ARCHITECTURE.md Section 6.
- MUST run `go test ./...` before considering any task in PLAN.md complete.
- Claim-diff verification MUST be tested against the Phase 0 corrupted-passage case (PLAN.md) to confirm it actually catches injected errors — a verification system that never fails its own test is not proven.

## Git Rules
- Every task do MUST commit and push to remote repository.

## Agent-Integration Rules (discovered 2026-09-09, Chunk 10 audit)
- [DO] Treat MCP as an additive interface layer only. It shares the service layer with REST (`docs/mcp-parity.md` architecture rule) — it does not replace or gate the consumer web app.
- [DO] Keep MCP tools stateless and scoped to deterministic intake and research-data retrieval (`ingest_document`, `search_documents`, `get_document`, `get_figures`, `get_evidence`). No user-library access or user-owned management operations by design.
- [DON'T] Expose user-owned operations (list, save, rename, delete, share, referral) through MCP. These stay REST-only, human-only — this was a deliberate decision, not an oversight (`docs/mcp-parity.md`).
- [DON'T] Add a new MCP tool or a new SEO/marketing route without a demand signal first (see Chunk 10.5/10.6 checkpoints). This repo has a documented pattern of shipping ahead of validation — don't repeat it here.
- [DO] Keep `docs/mcp-parity.md` and `goals/chunk-7-4-mcp/plan.md` in sync with `internal/mcp/tools.go`. Docs listing unimplemented tools is worse than no docs — it misleads the next agent session.
- [DO] Prefer one canonical route per feature. `/compare` and `/compare-research-papers` rendering the same component was route debt, not intentional design — don't introduce a second copy again.

## Verification Commands
- `go build ./cmd/server ./cmd/mcp ./cmd/admin` — all three binaries compile
- `go test ./...` — backend test suite
- `go vet ./...` — vet must be clean
- `gofmt -l .` — formatting check, must return empty
- `grep -c "Name:" internal/mcp/tools.go` — actual MCP tool count, cross-check against `docs/mcp-parity.md` table row count
- `npm --prefix frontend run lint && npm --prefix frontend run build` — frontend gate (`oxlint`, then `vite build`)