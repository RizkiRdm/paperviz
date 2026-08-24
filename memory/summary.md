# PaperViz Summary

- **State:** Chunk 4.1 (Shareable Figure Explanation) complete. Phase 4 (Product-Led Distribution) in progress. 101/101 Go tests pass; frontend builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, collections, auth, chart-image endpoint, compare endpoint, share endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry, documents, collections, comparison, share); repository (documents, collections, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard, compare, share-figure pages + hooks/components).
- **Recent changes:** Chunk 4.1 — Shareable Figure Explanation: migration 009 (share_token on charts, visibility on documents), ChartRepo share token methods (SetShareToken, GetByShareToken, RevokeShareToken), services/share.go (GenerateShareToken, RevokeShareToken, GetSharedFigure), handlers/share.go (POST/DELETE token endpoints, GET public share page), ShareFigurePage frontend, per-chart share button in ChartCard.
- **Current priority:** Phase 4 — Product-Led Distribution (Chunk 4.2: Shareable Paper Explanation).
- **Known issues:** Live Gemini regression tests outstanding (needs API key). Read-path layering deviation (Get/List call repo direct from handler) — flagged, not fixed in this chunk. DB must be reset after migration 009 (rm paperviz.db* && make dev).
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.
