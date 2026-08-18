# PaperViz Summary

- **State:** Chunk 2.3 (Research Collections) complete. Phase 2 (Activation & Retention) ongoing — next: Chunk 2.4 (Return Workflow). 90/90 Go tests pass; frontend builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, collections, auth, chart-image endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry, documents, collections); repository (documents, collections, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard pages + hooks/components).
- **Recent changes:** Chunk 2.3 — migration 008 adds collections + document_collections tables, CollectionRepo with CRUD + document management, service layer (collections.go), POST/GET/PATCH/DELETE /api/collections + document endpoints, dashboard UI (collection pills, create/rename/delete, add-to-collection dropdown). 90 tests pass.
- **Current priority:** Chunk 2.4 (Return Workflow).
- **Known issues:** Live Gemini regression tests outstanding (needs API key). Read-path layering deviation (Get/List call repo direct from handler) — flagged, not fixed in this chunk.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.
