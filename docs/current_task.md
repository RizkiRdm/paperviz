# Current Task

## Objective

Chunk 3.2 (Paper Comparison UI) — complete. Next: Chunk 3.3 (Evidence Comparison).

## Requirements

- **Chunk 3.2 (DONE):** Paper Comparison UI — dashboard paper selection with checkboxes, Compare Selected button, comparison page redesigned with DESIGN.md tokens, paper header cards with source links, agreements/disagreements sections, dimension comparison grid.

- **Chunk 3.1 (DONE):** Multi-Paper Comparison Model — data model types (PaperSummary, ComparisonDimension, PaperComparison), comparison service (ExtractPaperSummary, ComparePapers), HTTP handler (POST /api/documents/compare), 6 tests.

- **Chunk 2.4 (DONE):** Return Workflow — stats endpoint, dashboard welcome hero, post-analysis "What's Next" panel.

- **Chunk 2.3 (DONE):** Research Collections — migration 008, CollectionRepo, handlers, dashboard UI.

- **Chunk 2.2 (DONE):** Saved Papers — migration 007, ToggleSaved/UpdateTitle/DeleteDocument, handlers, dashboard UI.

- **Chunk 2.1 (DONE):** Paper History — title column, deriveTitle, ListSummariesByUser, dashboard redesign, nav links.

## Constraints

- Follow `handlers → services → repository/external` layering
- No DB reset needed
- Existing functionality must remain intact

## Relevant Files

- `frontend/src/pages/dashboard-page.jsx` — paper selection UI
- `frontend/src/pages/compare-page.jsx` — comparison page
- `internal/services/comparison.go` — comparison service
- `internal/handlers/documents.go` — compare endpoint

## Progress

- Chunk 3.2 complete: dashboard selection, redesigned comparison page, source links, DESIGN.md tokens. 98 Go tests pass, frontend builds clean.
- Chunk 3.1 complete: comparison data model, service, handler, 6 tests.
- Chunk 2.4 complete: stats endpoint, dashboard hero, What's Next panel.
- Chunk 2.3 complete: collections migration, repo, handlers, UI.
- Chunk 2.2 complete: saved papers migration, handlers, UI.
- Chunk 2.1 complete: paper history, title column, dashboard redesign.

## Next Action

1. Start Chunk 3.3 (Evidence Comparison) — structured claim → per-paper evidence view.
