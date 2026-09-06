# AGENTS.md — PaperViz 
## Quick Rules 
- MUST follow ARCHITECTURE.md layer boundaries (`handlers → services → repository/external`) without exception.
- MUST NOT introduce an ORM, job queue, message broker, or microservice split — see ARCHITECTURE.md Non-goals.- MUST NOT persist uploaded PDF bytes to disk.
- MUST NOT route LLM calls through any gateway other than direct Gemini API for MVP.
- MUST run tests before marking any task complete.
- MUST only inspect files directly related to task given. Do not scan the entier repository, Do not read file in `@.gitignore`. Do not inspect unrelated documents.

## Design System Rules
For any task that modifies or generates UI, styling, layout, or components:
1. Read `@DESIGN.md` completely before implementation.
2. Treat it as the absolute source of truth.
3. Do not invent any colors, spacing, or fonts outside of these rules.

## Project Context 
PaperViz converts academic papers (PDF/pasted text) into simplified-language versions plus re-visualized charts, served at ephemeral (7-day) no-auth share links. Target user: undergraduate students. Full context in PRD.md. Solo-dev project with ~1hr/day human oversight — agent autonomy within a phase is expected, but cross-phase scope changes require explicit human sign-off. --- ## Tech Stack - Backend: Go 1.22+, `chi` router, `modernc.org/sqlite` (no CGO), raw `database/sql`.- Frontend: React 18 + Vite + Tailwind CSS + shadcn/ui, Recharts for chart rendering.- LLM: Google Gemini API, direct HTTP integration.- PDF processing: Go-native text/image extraction libraries (pinned exact versions in `go.mod`).- No Docker/orchestration requirement for MVP — single binary + SQLite file.

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
- **Chart image_fallback has no image-serving endpoint.** `charts` table stores `image_blob BLOB` but `GET /api/documents/:id` response has no image field. Frontend shows annotation only. Fix requires new endpoint or base64-inline decision. Not blocking — satisfies PLAN.md Phase 4 "done means" (demonstrated capture). Revisit before chart re-visualization ships fully.

- **WAL mode required after DB reset.** Enabled via PRAGMA journal_mode=WAL + synchronous=NORMAL in `repository/db.go`. DB file must be deleted when schema changes (single flat migration). Already gitignored.

- **B3 (single-call verification) skipped** — requires live API key + real-document regression testing against B2 before shipping. Plan still describes it as optional.

- **Annotations require authentication.** Unauthenticated users cannot create/edit/delete annotations. Export endpoint also requires auth. Design decision: annotations are per-user research context, not collaborative.
- **Verification is not decoration (12.1 precedent).** `mismatch_detail` is evidence, not UI polish — always surface to user when status is `verification_failed`; pipeline populates claims table from verification output in the same tx, no separate LLM extraction step.

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
   - New `GenerateChapterChart()` in `charts.go`: one Gemini call per chapter, chart type varies (bar/line/pie/scatter)
   - Old `ExtractChartsFromText`, `fullTextChartPrompt` removed. `textChartElem` purged 2026-09-04 (12.1 dead-code slice; C-series cleanup closed).
   - Image fallback path (`ReVisualizeCharts`) unchanged.
7. D1: Annotations enforce per-user ownership — service layer checks `userID` matches before update/delete. 403 returned on ownership mismatch.
8. D2: Export endpoint excludes `OriginalText` and `SimplifiedText` — copyright compliance. Only structured metadata and user annotations are exported.
9. D3: Collections enforce per-user ownership — service Get/Rename/Delete/Add/Remove/ListDocuments take userID, ErrForbidden on mismatch; handler maps forbidden→403. Closes IDOR, mirrors D1.
10. E1: Added auth rate limiting on `POST /api/auth/signup` and `POST /api/auth/login` (5 req/60s, burst 3) via `rateLimitAuth` middleware reusing existing `ipRateLimiter` struct.
11. F1: Fixed silent zero-fill in chart rendering (`frontend/src/components/data-chart.jsx`) — missing values (undefined/null) now excluded from chart data instead of defaulted to 0; prevents misleading visualization where missing data points appear as real zeros.

---

## Coding Conventions 
- MUST use One line comment in every function made.
- MUST follow standard Go formatting (`gofmt`) 
— non-negotiable, run before every commit.
- MUST use explicit error returns (`if err != nil`) 
— no panics for expected error paths (validation failures, API timeouts). Panics reserved for truly unrecoverable states (e.g., failed DB connection at startup).
- MUST NOT use global mutable state for request-scoped data (document ID, request context) 
— pass explicitly.
- Service functions MUST be pure with respect to side effects where possible: `func Simplify(text string, level string) (string, error)` 
— no hidden reads from global config inside business logic functions; pass config explicitly.
- Naming: Go idiomatic (`CamelCase` exported, `camelCase` unexported). No Hungarian notation, no abbreviation-heavy naming.- Logging: structured JSON via standard library or a single chosen logging lib — MUST NOT log full document text content (see ARCHITECTURE.md Logging Policy).

## Testing Rules 
- Every service function MUST have a table-driven unit test: minimum 1 success case, 1 error case.
- Pipeline integration test REQUIRED covering all 4 Acceptance Scenarios and 4 Failure Scenarios listed in ARCHITECTURE.md Section 6.
- MUST run `go test ./...` before considering any task in PLAN.md complete.
- Claim-diff verification MUST be tested against the Phase 0 corrupted-passage case (PLAN.md) to confirm it actually catches injected errors — a verification system that never fails its own test is not proven. 

## Git Rules
- every task do MUST commit and push to remote repository.