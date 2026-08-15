# PaperViz Summary

- **State:** Roadmap Chunks 0.2, 1.1 (backend), 1.2, 1.3 complete. Chunk 1.1 frontend evidence display pending, then Chunk 1.4 (Research-Oriented Summary). 67/67 Go tests pass; builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, auth, chart-image endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry); repository (documents, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard pages + hooks/components).
- **Recent changes:** Chunk 1.3 Figure Explanation Quality — enriched chapterChartPrompt with x_axis, y_axis, key_takeaway, limitations, confidence; frontend displays axes labels, key takeaway card, limitations note, confidence badge; fixed pre-existing chartData.type→chart_data.chart_type bug. Fixed Rules of Hooks crash in result-page.jsx (useState after conditional returns).
- **Current priority:** Frontend evidence display (Chunk 1.1 tail), then Chunk 1.4 (Research-Oriented Summary).
- **Known issues:** Evidence not yet rendered in frontend. Live Gemini regression tests outstanding (needs API key).
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.
