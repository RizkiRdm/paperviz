# Project Progress

## Current

- **Completed:** Agent-task chunks 0–3; chunk 4 implementation and verification.
- **Current work:** Commit and push React Doctor top-three fixes plus chunk 4.
- **Blockers:** Independent read-only reviews are pending. Live Gemini paper regression tests still require a valid API key.
- **Next milestone:** Plan chunk 5 (`frontend/src/lib/api.js` request timeouts) after current commits land.

## Changelog

- **2026-08-02:** Client rejects PDFs larger than 20 MB and non-PDF MIME types before upload; Playwright confirmed both errors and valid-file recovery.
- **2026-08-02:** Fixed React Doctor findings for eager Recharts, static-element interaction, and placeholder-only field. Remaining 10 `result-page.jsx` findings are deferred.
- **2026-08-02:** Chunk 3 public-link disclaimer committed and pushed as `9ae1378`.
- **2026-08-02:** Chunk 2 polling timeout and copy-link behavior committed and pushed as `e7515c0`.
