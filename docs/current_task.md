# Current Task

## Objective

Chunk 1.1 (Evidence Provenance) — make every AI-generated explanation traceable to the original paper. Backend data model complete; frontend display + pipeline population + repo tests pending.

## Requirements

- **Backend (DONE):**
  - Migration 005: `evidence` table (`id`, `paper_id` FK cascade, `page`, `figure_id`, `table_id`, `section`, `source_text`, `source_reference`)
  - `repository.Evidence` type in `types.go`
  - `repository/evidence.go`: `EvidenceRepo` with `Insert` + `ListByPaper`
  - Migration 5 registered in `cmd/server/main.go` `loadMigrations`; migration-count test updated

- **Backend (PENDING):**
  - Populate evidence during pipeline (source_text/page/section extracted where available)
  - Expose evidence in `GET /api/documents/:id` response
  - Table-driven repo tests (min 1 success + 1 error case per AGENTS.md)

- **Frontend (PENDING):**
  - Display evidence alongside summaries/charts with provenance (page, section, source ref)

## Constraints

- Reusable Evidence model per roadmap Chunk 1.1 — don't over-engineer schema
- Follow `handlers → services → repository/external` layering
- DB reset not required if migration runner applies 005 cleanly (verify)
- Existing functionality must remain intact

## Relevant Files

- `migrations/005_evidence.sql` — new evidence table
- `internal/repository/types.go` — Evidence struct
- `internal/repository/evidence.go` — EvidenceRepo
- `cmd/server/main.go` — registered migration 5
- `cmd/server/main_test.go` — migration count updated
- `internal/services/pipeline.go` — where evidence population will hook in
- `internal/handlers/documents.go` — where evidence will be exposed in GET
- `frontend/src/pages/result-page.jsx` — where evidence will display

## Progress

- Migration 005 created + registered. Evidence type + repo methods added.
- 49/49 tests pass, build succeeds.
- Chunk 0.2 audit deliverable: `docs/product/current-user-flow.md`.

## Next Action

1. Decide evidence population source (chapter text? chart annotations? per-page PDF text via `ExtractTextByPage`)
2. Wire evidence into pipeline output + persistence in `saveResult`
3. Expose evidence in GET response; render in result page
4. Add EvidenceRepo + integration tests
5. Then start Chunk 1.2 (Original vs Explained Figure)