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

## DB Reset Protocol

When schema changes (new column/table):
1. `kill` server process
2. `rm paperviz.db paperviz.db-wal paperviz.db-shm paperviz.db-journal` (all present)
3. `go run ./cmd/server` — fresh DB, fresh schema, fresh WAL
4. Data ephemeral (7-day expiry). No migration runner yet.

## Chart Pipeline Fix — Execution Order (docs/CHART_FIX.md)

Chunks execute sequentially. Each produces proof before next starts.
1. Chunk 1: fix `context.Background()` → `ctx` in `gemini.go:87`
2. Chunk 2: remove per-stage `stageTimeout` wraps in `pipeline.go`
3. Chunk 3: raise `backgroundPipelineTimeout` 5min → 20min
4. STOP → live verification with real paper upload
5. Chunk 4: add `chart_extraction_degraded` flag (7 files)
6. Chunk 6: add WAL pragmas in `db.go` (after DB reset for new column)
