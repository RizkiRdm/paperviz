# PaperViz Summary

- **State:** Working MVP — single Go binary + React/Vite frontend + SQLite. All chunks (0–9) complete.
- **Modules:** HTTP handlers; processing services (pipeline, simplification, verification, charts, chapters, expiry); repository; Gemini/PDF external adapters; upload and result frontend pages.
- **Recent changes (CHUNKS 6–9):** Error boundary (class component wrapping App), keyboard-accessible dropzone (aria-label), aria-live polling status (screen reader announcements), security headers (X-Content-Type-Options, X-Frame-Options, CSP allowing Google Fonts).
- **Current priority:** Await user direction — all PAPERVIZ_AGENT_TASKS.md chunks done.
- **Known issues:** Result page has deferred React Doctor findings; chart image fallback lacks an image-serving endpoint; live Gemini regression tests remain outstanding. DB schema changed — dev DB needs reset.
- **Source of truth:** `docs/ARCHITECTURE.md`, `docs/PRD.md`, `docs/PLAN.md`, and `DESIGN.md`.
