# PaperViz Summary

- **State:** Chunk 2.2 (Saved Papers) complete. Phase 2 (Activation & Retention) ongoing — next: Chunk 2.3 (Research Collections). 88/88 Go tests pass; frontend builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, auth, chart-image endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry, documents); repository (documents, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard pages + hooks/components).
- **Recent changes:** Chunk 2.2 — migration 007 adds saved column + index, ToggleSaved/UpdateTitle/DeleteDocument repository methods, service layer (documents.go), PUT /:id/save + PATCH /:id + DELETE /:id handlers, dashboard UI (star toggle, inline title edit, delete confirmation, filter tabs). 88 tests pass.
- **Current priority:** Chunk 2.3 (Research Collections).
- **Known issues:** Live Gemini regression tests outstanding (needs API key). Read-path layering deviation (Get/List call repo direct from handler) — flagged, not fixed in this chunk.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.
