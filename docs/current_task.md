# Current Task

## Objective

Chunk 1.2 (Original vs Explained Figure) — users clearly distinguish original evidence from generated interpretation, with provenance. Backend + frontend + tests complete. Chunk 1.1 (Evidence Provenance) backend model done; pipeline population + GET exposure + frontend display still pending.

## Requirements

- **Chunk 1.2 (DONE):**
  - `GET /api/documents/:id/charts/:chartId/image` serves original figure bytes (MIME-sniffed). Scoped `WHERE id=? AND document_id=?` — bare chart ID can't read another document's figure
  - `chartResponse` += `image_url` (only `image_fallback` charts)
  - `ChartRepo.GetByDocumentAndID` + `handlers.detectImageMIME` (PNG/JPEG/GIF/WebP)
  - Frontend `ChartCard` → 2-zone card: Original Figure (image + neutral badge "Original Figure · Page N") vs "PaperViz AI Interpretation" (blue badge; Recharts or annotation). `data-chart.jsx` badge relabeled
  - Tests: `handlers/image_test.go` (MIME sniff), `repository/charts_test.go` (GetByDocumentAndID: success/wrong-doc/not-found)
  - docs/ARCHITECTURE.md API contracts updated

- **Chunk 1.1 (PENDING tail):**
  - Populate evidence during pipeline (source_text/page/section extracted where available)
  - Expose evidence in `GET /api/documents/:id` response
  - Table-driven repo tests for `EvidenceRepo` (min 1 success + 1 error case per AGENTS.md)

- **Chunk 1.2 (DONE, provenance note):** Provenance shipped chart-level only (`page_number` + chapter title + `source_method`). Evidence-table provenance deferred to 1.1 tail (decided Opt 1).

## Constraints

- Reusable Evidence model per roadmap Chunk 1.1 — don't over-engineer schema
- Follow `handlers → services → repository/external` layering
- DB reset not required (chunk 1.2 added no schema change — verify)
- Existing functionality must remain intact

## Relevant Files

- `migrations/005_evidence.sql` — evidence table
- `internal/repository/types.go` — Evidence struct
- `internal/repository/evidence.go` — EvidenceRepo (tests pending)
- `internal/repository/charts.go` — `GetByDocumentAndID`
- `internal/handlers/image.go` — `detectImageMIME`
- `internal/handlers/documents.go` — `GetChartImage`, `chartResponse.image_url`
- `internal/handlers/router.go` — `/api/documents/{id}/charts/{chartId}/image`
- `frontend/src/components/chart-card.jsx` — 2-zone Original vs Interpretation card
- `frontend/src/components/data-chart.jsx` — interpretation badge
- `frontend/src/pages/result-page.jsx` — chart render site

## Progress

- Chunk 1.2 complete: original-figure serving endpoint + clear original-vs-interpretation UI + chart-level provenance + tests. 62 Go tests pass, build + frontend build succeed.
- Chunk 1.1: migration 005 + Evidence type + EvidenceRepo.Insert/ListByPaper done; repo tests started (charts_test.go). Evidence population + GET exposure pending.
- Chunk 0.2 audit deliverable: `docs/product/current-user-flow.md`.

## Next Action

1. Finish 1.1 tail: populate evidence in pipeline (chapter/image chart source text), expose `evidence` in GET response, render provenance in result page, add EvidenceRepo tests
2. Then start Chunk 1.3 (Figure Explanation Quality)