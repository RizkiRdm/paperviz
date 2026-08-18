# PaperViz Summary

- **State:** Chunk 2.4 (Return Workflow) complete. Phase 2 (Activation & Retention) complete. 92/92 Go tests pass; frontend builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, collections, auth, chart-image endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry, documents, collections); repository (documents, collections, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard pages + hooks/components).
- **Recent changes:** Chunk 2.4 — stats endpoint (GET /api/documents/stats), dashboard welcome hero with aggregate counts, post-analysis "What's Next" panel (add to collection, upload another, view all). 92 tests pass.
- **Current priority:** Phase 3 — Multi-Paper Intelligence (Chunk 3.1).
- **Known issues:** Live Gemini regression tests outstanding (needs API key). Read-path layering deviation (Get/List call repo direct from handler) — flagged, not fixed in this chunk.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.
