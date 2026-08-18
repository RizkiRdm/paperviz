# PaperViz Summary

- **State:** Chunk 2.1 (Paper History) complete. Phase 2 (Activation & Retention) started — next: Chunk 2.2 (Saved Papers). 81/81 Go tests pass; frontend builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, auth, chart-image endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry); repository (documents, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard pages + hooks/components).
- **Recent changes:** Chunk 2.1 — migration 006 adds title column (backfilled from extracted text first line), deriveTitle() at intake, ListSummariesByUser with correlated subqueries for chart/explanation counts, dashboard rows redesigned (title/date/status/summary/counts), Dashboard nav links on upload + result pages. New test files: documents_test.go, deriveTitle tests in intake_test.go.
- **Current priority:** Chunk 2.2 (Saved Papers).
- **Known issues:** Live Gemini regression tests outstanding (needs API key). Read-path layering deviation (Get/List call repo direct from handler) — flagged, not fixed in this chunk.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.
