# Project Progress

## Current

- **Chapter-tabbed view:** Implemented. Backend links charts to chapters; frontend shows horizontal scrollable tabs.
- **Current work:** Chapter-tabbed view committed. DB reset required (migration 004).
- **Blockers:** Live Gemini regression tests still require a valid API key.
- **Next milestone:** DB reset + rebuild + test chapter-tabbed view. Then: contextual help tooltips, processing progress indicator.

## Changelog

- **2026-08-08:** UX copy clarified — error messages use "We couldn't..." pattern, processing text uses student-friendly language, chart empty states simplified. Impeccable commands run: init (PRODUCT.md), critique (29/36), harden (clipboard + keyboard), shape (chapter tabs), clarify (copy). Chapter-tabbed view implemented: migration 004 (chapter_id on charts), services Chart.ChapterIndex, repository Chart.ChapterID, handler exposes content + chapter_id, frontend horizontal tabs with ARIA roles + keyboard navigation. Fixed DB startup error (stale DB + migration gap), fixed JSONL logger error rendering, added logger regression test.
- **2026-08-07:** P0 Core Features COMPLETE — 7 chunks across 2 batches. Batch A: claim_diff in GET response (A1), expandable claim comparison UI (A2). Batch B: chapters migration (B1), ChapterRepo + type (B2), pipeline carries chapters (B3), save chapters + expose in API (B4), chapter summary card (B5). All committed and pushed.
- **2026-08-06:** Phase 1 SaaS Foundation COMPLETE — All 10 chunks across 7 phases. Backend: users table, sessions table, auth handlers (signup/login/logout/me), RequireAuth/OptionalAuth middleware, document ownership, paginated list endpoint. Frontend: react-router-dom, login/signup pages, dashboard page with auth check, NotFoundPage, copy-link verification. Responsive smoke test passed (no fixed-width issues). Both Go and frontend builds pass.
- **2026-08-06:** CHUNKS 6–9 — React ErrorBoundary, dropzone aria-label, aria-live polling, security headers middleware. Committed + pushed as `d473212`.
- **2026-08-04:** CHUNK 6 — chart types (bar/line/pie/scatter), reading level badge, backend error messages, copy text button, share dialog modal, processing stage tracking (backend column + callback + frontend display). Committed + pushed.
- **2026-08-02:** Client rejects PDFs larger than 20 MB and non-PDF MIME types before upload; Playwright confirmed both errors and valid-file recovery.
- **2026-08-02:** Fixed React Doctor findings for eager Recharts, static-element interaction, and placeholder-only field. Remaining 10 `result-page.jsx` findings are deferred.
- **2026-08-02:** Chunk 3 public-link disclaimer committed and pushed as `9ae1378`.
- **2026-08-02:** Chunk 2 polling timeout and copy-link behavior committed and pushed as `e7515c0`.
