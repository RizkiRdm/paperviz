# Actionable Todos

## Critical

- Run live Gemini regression tests on corrupted and clean passages before trusting claim verification in production.

## High

- DB reset + rebuild + test chapter-tabbed view with real PDF (migration 004: chapter_id on charts).
- Decide and implement chart image-fallback serving contract before full chart release.

## Medium

- Add contextual help tooltips for VerificationBadge and ClaimComparisonPanel (P2 from critique).
- Add processing progress indicator or elapsed time (P3 from critique).
- Run full five-paper manual flow before MVP release.

## Low

- Investigate missing asset 404 and chart-list key warning observed during runtime QA.

## Completed

- UX copy clarified — error messages, processing text, chart empty states (2026-08-08).
- Chapter-tabbed view implemented — backend migration 004, chapter_id on charts, frontend horizontal tabs with ARIA + keyboard nav (2026-08-08).
- Result page hardened — clipboard fallback feedback, share dialog Escape dismiss + autoFocus (2026-08-08).
- Critique on result page — 29/36 score, 2 P1s fixed, 2 P2s + 1 P3 identified (2026-08-08).
- PRODUCT.md created — platform, users, positioning, principles (2026-08-08).
- Fixed DB startup error — stale DB + migration gap + logger defect (2026-08-08).
- Fixed JSONL logger error rendering — errors now show as strings not `{}` (2026-08-08).
- Added logger regression test — red→green (2026-08-08).
- P0 Core Features: ALL 7 CHUNKS COMPLETE (2026-08-07).
- Phase 1 SaaS Foundation: ALL 10 CHUNKS COMPLETE (2026-08-06).
