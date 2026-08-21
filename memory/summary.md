# PaperViz Summary

- **State:** Chunk 3.3 (Evidence Comparison) complete. Phase 3 (Multi-Paper Intelligence) in progress. 101/101 Go tests pass; frontend builds clean.
- **Modules:** HTTP handlers (thin adapters; documents, collections, auth, chart-image endpoint, compare endpoint); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry, documents, collections, comparison); repository (documents, collections, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard, compare pages + hooks/components).
- **Recent changes:** Chunk 3.3 — Evidence Comparison: EvidenceClaim type, CompareEvidence Gemini service, per-paper stance extraction (supporting/contradicting/unclear), EvidenceComparisonPanel accordion UI with expandable claim details.
- **Current priority:** Phase 4 — Product-Led Distribution (Chunk 4.1: Shareable Figure Explanation).
- **Known issues:** Live Gemini regression tests outstanding (needs API key). Read-path layering deviation (Get/List call repo direct from handler) — flagged, not fixed in this chunk.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.
