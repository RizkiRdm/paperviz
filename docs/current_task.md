# Current Task

## Objective

Chunk 2.1 (Paper History) — complete. Next: Chunk 2.2 (Saved Papers).

## Requirements

- **Chunk 2.1 (DONE):** Paper history list — migration 006 adds title column (backfilled from extracted text first line). `deriveTitle()` extracts paper title at intake. `ListSummariesByUser` returns lightweight rows with correlated subqueries for chart/explanation counts. Dashboard rows show title, date, status, summary preview, figure/explanation counts. Dashboard nav links added to upload + result page headers.

- **Chunk 1.4 (DONE):** Structured research summary — Gemini prompt outputs `## `-delimited sections. Frontend parses sections and renders as cards.

- **Chunk 1.1 (DONE):** Evidence provenance — migration 005, `evidence` table, `EvidenceRepo`, pipeline population, GET exposure, frontend Source chips + collapsible source text.

- **Chunk 1.2 (DONE):** Original vs Explained Figure — `/api/documents/:id/charts/:chartId/image` endpoint + 2-zone UI + chart-level provenance.

- **Chunk 1.3 (DONE):** Figure Explanation Quality — enriched prompts (x_axis, y_axis, key_takeaway, limitations, confidence), frontend display, chart_type bugfix.

## Constraints

- Follow `handlers → services → repository/external` layering
- No DB reset needed
- Existing functionality must remain intact

## Relevant Files

- `internal/services/intake.go` — deriveTitle + title at intake
- `internal/repository/documents.go` — ListSummariesByUser, title in Insert/Get/ListByUser
- `frontend/src/pages/dashboard-page.jsx` — paper history rows

## Progress

- Chunk 2.1 complete: enriched document list (title, summary preview, chart/explanation counts, Dashboard nav links). 81 Go tests pass, frontend builds clean, gofmt clean.
- Chunk 1.4 complete: structured research summary. 69 Go tests pass (pre-2.1).
- Chunk 1.3 complete: enriched prompts + frontend display + chart_type bugfix.
- Chunk 1.1 complete: evidence provenance fully displayed.
- Chunk 1.2 complete: original-figure serving endpoint + clear original-vs-interpretation UI.
- Chunk 0.2 audit deliverable: `docs/product/current-user-flow.md`.

## Next Action

1. Start Chunk 2.2 (Saved Papers) — save, rename, delete, reopen.
