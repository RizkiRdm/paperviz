# PaperViz Summary

- **State:** Roadmap Chunks 0.2, 1.1 (backend), 1.2 complete. Chunk 1.1 frontend evidence display pending, then Chunk 1.3 (Figure Explanation Quality). 67/67 Go tests pass; builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, auth, chart-image endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry); repository (documents, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard pages + hooks/components).
- **Recent changes:** Chunk 1.2 Original vs Explained Figure — `GET /api/documents/:id/charts/:chartId/image` endpoint (MIME-sniffed, doc-scoped) + 2-zone ChartCard (Original Figure vs PaperViz AI Interpretation). Chunk 1.1 tail — evidence rows populated for image-origin charts in `savePipelineResult`, exposed as `evidence[]` in GET; repo + services tests.
- **Current priority:** Frontend evidence display (provenance under figures), then Chunk 1.3 (Figure Explanation Quality).
- **Known issues:** Evidence not yet rendered in frontend. Live Gemini regression tests outstanding (needs API key).
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.