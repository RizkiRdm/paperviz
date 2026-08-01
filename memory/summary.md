# PaperViz Summary

- **State:** Working MVP in a single Go binary with React/Vite frontend and SQLite persistence.
- **Modules:** HTTP handlers; processing services; repository; Gemini/PDF external adapters; upload and result frontend pages.
- **Recent changes:** Poll timeout/retry, copy link, public-link disclaimer, lazy Recharts loading, accessible upload controls, and pre-upload file validation.
- **Current priority:** Land chunk 4 and React Doctor top-three fixes; then plan chunk 5 request timeouts.
- **Known issues:** Result page has 10 deferred React Doctor findings; chart image fallback lacks an image-serving endpoint; live Gemini regression tests remain outstanding.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `docs/PLAN.md`, and `DESIGN.md`.
