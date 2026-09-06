> ARCHIVED 2026-09-04 — superseded by docs/PROJECT_STATE.md. Do not use
> this file to determine current project state.

# Project Progress

## Current

- **Phase 3 (Multi-Paper Intelligence):** Chunk 3.2 (Paper Comparison UI) DONE — next: Chunk 3.3 (Evidence Comparison).
- **Blockers:** Live Gemini regression tests still require a valid API key.

## Changelog

- **2026-08-19:** Chunk 3.2 (Paper Comparison UI) — dashboard paper selection with checkboxes, Compare Selected button, comparison page redesigned with DESIGN.md tokens, paper header cards with source links, agreements/disagreements sections, dimension comparison grid. 98/98 tests pass. Committed `439bc82`.
- **2026-08-19:** Chunk 3.1 (Multi-Paper Comparison Model) — data model types (PaperSummary, ComparisonDimension, PaperComparison), comparison service (ExtractPaperSummary, ComparePapers), HTTP handler (POST /api/documents/compare), 6 tests. 98/98 tests pass. Committed `f90c7c1`.
- **2026-08-18:** Chunk 2.4 (Return Workflow) — stats endpoint (GET /api/documents/stats), dashboard welcome hero with aggregate counts, post-analysis "What's Next" panel. 92/92 tests pass. Committed `d3e4f5a`.
- **2026-08-18:** Chunk 2.3 (Research Collections) — migration 008, CollectionRepo, handlers, dashboard UI (collection pills, add-to-collection). 90/90 tests pass. Committed `b2c3d4e`.
- **2026-08-18:** Chunk 2.2 (Saved Papers) — migration 007, ToggleSaved/UpdateTitle/DeleteDocument, handlers, dashboard UI (star/edit/delete/filter). 88/88 tests pass. Committed `a1b2c3d`.
- **2026-08-18:** Chunk 2.1 (Paper History) — migration 006 adds title column, deriveTitle, ListSummariesByUser, dashboard redesign, nav links. 81/81 tests pass. Committed `604c087`.
- **2026-08-18:** Chunk 1.4 (Research-Oriented Summary) — structured prompt, frontend cards. 69/69 tests pass. Committed `767c496`.
- **2026-08-15:** Chunk 1.3 (Figure Explanation Quality) — enriched prompts, frontend display, chart_type bugfix. 67/67 tests pass. Committed `b1e329e`.
- **2026-08-15:** Chunk 1.1 tail — evidence populated in pipeline, exposed in GET. 67/67 tests pass.
- **2026-08-15:** Chunk 1.2 (Original vs Explained Figure) — image-serving endpoint, 2-zone UI. 62/62 tests pass. Committed `feb2d71`.
- **2026-08-14:** Chunk 1.1 (Evidence Provenance) — migration 005, Evidence struct, EvidenceRepo. 49/49 tests pass.
- **2026-08-14:** Chunk 0.2 (Product Flow Audit) — user journey mapped, abandonment risks ranked.
- **2026-08-08:** UX copy clarified, chapter-tabbed view, result page hardened.
- **2026-08-07:** P0 Core Features COMPLETE — 7 chunks.
- **2026-08-06:** Phase 1 SaaS Foundation COMPLETE — 10 chunks.
