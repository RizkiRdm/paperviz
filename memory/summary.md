# PaperViz Summary

- **State:** P0 Core Features COMPLETE. Claim-diff verification data exposed; chapter structure persisted and displayed. All 7 chunks committed and pushed.
- **Modules:** HTTP handlers; processing services (pipeline, simplification, verification, charts, chapters, expiry); repository (documents, charts, claim_diffs, chapters, users, sessions); Gemini/PDF external adapters; upload and result frontend pages; auth system.
- **Recent changes (P0):** A1: claim_diff wired through GET response. A2: expandable claim comparison panel under Verified badge. B1: chapters migration. B2: ChapterRepo + type. B3: pipeline carries chapters. B4: save chapters + expose in API. B5: "Sections in this paper" summary card.
- **Current priority:** Await user direction for next phase.
- **Known issues:** DB schema changed (003_chapters.sql) — dev DB needs reset. Chart image fallback lacks image-serving endpoint. Live Gemini regression tests outstanding. Result page has deferred React Doctor findings.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `docs/PLAN.md`, `DESIGN.md`.
