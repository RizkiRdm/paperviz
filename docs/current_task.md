# Current Task

## Objective

Chunk 1.1 (Evidence Provenance) — frontend evidence display alongside figure explanations. Backend complete (model + pipeline population + GET exposure + tests). Frontend display pending.

## Requirements

- **Chunk 1.1 backend (DONE):**
  - Migration 005: `evidence` table (`id`, `paper_id` FK cascade, `page`, `figure_id`, `table_id`, `section`, `source_text`, `source_reference`)
  - `repository.Evidence` type in `types.go`
  - `repository/evidence.go`: `EvidenceRepo` `Insert` + `ListByPaper`
  - Pipeline population: `savePipelineResult` writes one evidence row per image-origin chart (`Services.Chart.SourceText`, set in `reVisualizeOne` from page text) — `paper_id`, `page`, `figure_id`=chart id, `source_text`=full page text, `source_reference`="Figure on page N"
  - Exposed in `GET /api/documents/:id` as `evidence` array
  - Table-driven repo tests (`evidence_test.go`): Insert+List success, empty list, FK-violation error; services test verifies persistence path

- **Chunk 1.1 (PENDING):**
  - Frontend evidence display alongside figure explanations (provenance: page, section, source ref)

- **Chunk 1.2 (DONE):** Original vs Explained Figure — `/api/documents/:id/charts/:chartId/image` endpoint + 2-zone UI + chart-level provenance.

- **Chunk 1.3 (DONE):** Figure Explanation Quality — enriched prompts (x_axis, y_axis, key_takeaway, limitations, confidence), frontend display, chart_type bugfix.

- **Provenance scoping (decided):** Evidence rows only where original source text exists — image-origin charts. Chapter-derived charts have no original page mapping → no fabricated provenance (Rule 4).

## Constraints

- Follow `handlers → services → repository/external` layering
- No DB reset needed (chunk 1.1 tail added no schema change)
- Existing functionality must remain intact

## Relevant Files

- `migrations/005_evidence.sql` — evidence table
- `internal/repository/types.go` — Evidence struct
- `internal/repository/evidence.go` — EvidenceRepo
- `internal/repository/evidence_test.go` — repo tests
- `internal/services/types.go` — Chart.SourceText
- `internal/services/charts.go` — SourceText set from page context, enriched chapterChartPrompt
- `internal/services/intake.go` — `savePipelineResult` evidence insert
- `internal/services/save_pipeline_result_test.go` — persistence test
- `internal/handlers/documents.go` — `evidence` in GET
- `frontend/src/pages/result-page.jsx` — where evidence will display (pending)
- `frontend/src/components/data-chart.jsx` — chart explanation display

## Progress

- Chunk 1.3 complete: enriched prompts + frontend display + chart_type bugfix. 67 Go tests pass, build + gofmt clean.
- Chunk 1.1 backend tail complete: pipeline population + GET exposure + repo/services tests. 67 Go tests pass.
- Chunk 1.2 complete: original-figure serving endpoint + clear original-vs-interpretation UI + tests.
- Chunk 0.2 audit deliverable: `docs/product/current-user-flow.md`.

## Next Action

1. Render `evidence` from GET response in result page (provenance under each figure explanation)
2. Then start Chunk 1.4 (Research-Oriented Summary)
