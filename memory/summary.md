# PaperViz Summary

- **State:** Working MVP — single Go binary + React/Vite frontend + SQLite. Chunks 0–6 complete.
- **Modules:** HTTP handlers; processing services (pipeline, simplification, verification, charts, chapters, expiry); repository; Gemini/PDF external adapters; upload and result frontend pages.
- **Recent changes (CHUNK 6):** Chart type variety (bar/line/pie/scatter), reading level badge, backend error messages, copy text button, share dialog modal, processing stage tracking with DB column and pipeline callback.
- **Current priority:** CHUNK 7 (chart re-visualization frontend integration) or next user direction.
- **Known issues:** Result page has 10 deferred React Doctor findings; chart image fallback lacks an image-serving endpoint; live Gemini regression tests remain outstanding. DB schema changed — dev DB needs reset.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `docs/PLAN.md`, and `DESIGN.md`.
