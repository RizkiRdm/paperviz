## graphify

Code graph at `graphify-out/`. Query before grep/read.

Rules:
- `graphify query "<question>"` for scoped subgraph
- `graphify path "<A>" "<B>"` for relationships
- `graphify explain "<concept>"` for focused concepts
- GRAPH_REPORT.md only for broad architecture
- After code changes, run `graphify update .`

## Known Issues

- **Chart image_fallback has no image-serving endpoint.** `charts` table stores `image_blob BLOB` but `GET /api/documents/:id` response has no image field. Frontend shows annotation only. Fix requires new endpoint or base64-inline decision. Not blocking — satisfies PLAN.md Phase 4 "done means" (demonstrated capture). Revisit before chart re-visualization ships fully.

- **WAL mode required after DB reset.** Enabled via PRAGMA journal_mode=WAL + synchronous=NORMAL in `repository/db.go`. DB file must be deleted when schema changes (single flat migration). Already gitignored.

- **B3 (single-call verification) skipped** — requires live API key + real-document regression testing against B2 before shipping. Plan still describes it as optional.

## DB Reset Protocol

When schema changes (new column/table):
1. `kill` server process
2. `rm paperviz.db paperviz.db-wal paperviz.db-shm paperviz.db-journal` (all present)
3. `go run ./cmd/server` — fresh DB, fresh schema, fresh WAL
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
   - Old `ExtractChartsFromText`, `fullTextChartPrompt` removed. `textChartElem` kept (test compat).
   - Image fallback path (`ReVisualizeCharts`) unchanged.
