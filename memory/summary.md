# PaperViz Summary

- **State:** Roadmap Chunks 0.2 + 1.1 active. Chunk 0.2 (Product Flow Audit) delivered. Chunk 1.1 (Evidence Provenance) backend model done — migration 005, Evidence type, EvidenceRepo. Frontend display + pipeline population pending. Architecture candidates 1–5 complete.
- **Modules:** HTTP handlers (thin adapters); processing services (pipeline, simplify, verify, charts, chapters, intake, extraction, expiry); repository (documents, charts, claim_diffs, chapters, evidence, users, sessions); Gemini/PDF external adapters; React frontend (upload, result, auth, dashboard pages + hooks/components).
- **Recent changes:** Candidate 1–5 architecture refactors (intake service, ExtractJSON[T], PDFDocument seam, useDocumentPoll, AuthForm/useAuthSubmit). Chunk 0.2 audit doc. Chunk 1.1 evidence table + repo. 49/49 Go tests pass, builds clean.
- **Current priority:** Finish Chunk 1.1 (wire evidence into pipeline + GET + frontend + tests), then Chunk 1.2 (Original vs Explained Figure).
- **Known issues:** Chart image fallback lacks image-serving endpoint. Evidence not yet populated/exposed. Live Gemini regression tests outstanding.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `DESIGN.md`, `docs/paperviz-agent-roadmap.md`, `PRODUCT.md`.