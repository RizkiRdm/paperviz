# Actionable Todos

## Critical

- Run live Gemini regression tests on corrupted and clean passages before trusting claim verification in production.

## High

- Chunk 1.1: render `evidence` from GET response in result page (provenance under each figure explanation).

## Medium

- Chunk 1.4: Research-Oriented Summary (structured research understanding replacing generic summarization).
- Add contextual help tooltips for VerificationBadge and ClaimComparisonPanel (P2 from critique).
- Add processing progress indicator or elapsed time (P3 from critique).
- Run full five-paper manual flow before MVP release.

## Low

- Investigate missing asset 404 and chart-list key warning observed during runtime QA.

## Completed

- Chunk 1.3 Figure Explanation Quality — enriched prompts, frontend display, chart_type bugfix (2026-08-15).
- Fixed Rules of Hooks crash in result-page.jsx (2026-08-15).
- Chunk 1.1 Evidence backend tail — pipeline population for image-origin charts + `evidence` in GET + repo/services tests (2026-08-15).
- Chunk 1.2 Original vs Explained Figure — image-serving endpoint, 2-zone UI, chart-level provenance (2026-08-15).
- Chunk 1.1 Evidence backend — migration 005, Evidence type, EvidenceRepo, registered + test updated (2026-08-14).
- Chart image-fallback serving contract decided + implemented as dedicated `/image` endpoint (2026-08-15).
- Chunk 0.2 Product Flow Audit — `docs/product/current-user-flow.md` (2026-08-14).
- Architecture candidates 1–5 — intake service, ExtractJSON[T], PDFDocument seam, useDocumentPoll, AuthForm/useAuthSubmit (2026-08-14).
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
