# PaperViz Summary

- **State:** Chunk 3.1 (Multi-Paper Comparison Model) complete. Phase 3 (Multi-Paper Intelligence) in progress. 98/98 Go tests pass; frontend builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, collections, auth, chart-image endpoint, compare endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry, documents, collections, comparison); repository (documents, collections, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard, compare pages + hooks/components).
- **Recent changes:** Chunk 3.1 — Multi-Paper Comparison Model: data model types (PaperSummary, ComparisonDimension, PaperComparison), comparison service (ExtractPaperSummary, ComparePapers), HTTP handler (POST /api/documents/compare), frontend comparison page, 6 new tests. 98 tests pass.
- **Current priority:** Phase 3 — Multi-Paper Intelligence (Chunk 3.2: Paper Comparison UI).
- **Known issues:** Live Gemini regression tests outstanding (needs API key). Read-path layering deviation (Get/List call repo direct from handler) — flagged, not fixed in this chunk.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.
