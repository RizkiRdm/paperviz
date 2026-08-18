# PaperViz Summary

- **State:** Roadmap Chunks 0.2, 1.1, 1.2, 1.3, 1.4 complete. Phase 1 (Trust & Core Value) fully complete. Next: Chunk 2.1 (Paper History). 69/69 Go tests pass; frontend builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, auth, chart-image endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry); repository (documents, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard pages + hooks/components).
- **Recent changes:** Chunk 1.4 — structured research summary (backend prompt outputs `## `-delimited sections, frontend parses + renders as cards). Pre-existing gofmt drift in charts.go normalized.
- **Current priority:** Chunk 2.1 (Paper History).
- **Known issues:** Live Gemini regression tests outstanding (needs API key).
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.
