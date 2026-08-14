# Project Progress

## Current

- **Chunk 1.2 (Original vs Explained Figure):** Complete — image-serving endpoint + 2-zone UI + chart-level provenance.
- **Chunk 1.1 (Evidence Provenance):** Backend complete — model, pipeline population, GET exposure, repo/services tests. Frontend evidence display PENDING.
- **Chunk 0.2 (Product Flow Audit):** Complete — `docs/product/current-user-flow.md`.
- **Current work:** Chunk 1.1 frontend evidence display, then Chunk 1.3 (Figure Explanation Quality).
- **Blockers:** Live Gemini regression tests still require a valid API key.
- **Next milestone:** Frontend evidence display; then Chunk 1.3.

## Changelog

- **2026-08-15:** Chunk 1.1 tail — evidence populated in `savePipelineResult` (image-origin charts, full page text, `Figure on page N` ref), exposed as `evidence[]` in GET, repo + services tests. 67/67 tests pass.
- **2026-08-15:** Chunk 1.2 (Original vs Explained Figure) — `GET /api/documents/:id/charts/:chartId/image` (MIME-sniffed, document-scoped), `image_url` in GET chart response, ChartCard 2-zone Original Figure vs PaperViz AI Interpretation with provenance badges. 62/62 tests pass. Committed `feb2d71`.
- **2026-08-14:** Chunk 1.1 (Evidence Provenance) backend — migration 005_evidence.sql, Evidence struct, EvidenceRepo, registered migration 5, updated migration-count test. 49/49 tests pass. (Frontend display + pipeline population pending.)
- **2026-08-14:** Chunk 0.2 (Product Flow Audit) — `docs/product/current-user-flow.md` delivered. No behavior changes. Mapped upload → intake → pipeline → polling → result → share journey, ranked abandonment risks, identified aha moments.
- **2026-08-14:** Architecture candidates 3–5 — Candidate 3: `PDFDocument`/`ParsePDF` seam in `extraction.go`, pipeline uses it. Candidate 4: `useDocumentPoll` hook extracted from ResultPage. Candidate 5: `AuthForm` + `useAuthSubmit` consolidate login/signup (~40 lines/page).
- **2026-08-14:** Architecture candidates 1–2 (uncommitted until now) — Candidate 1: `internal/services/intake.go` transactional intake, thin `documents.go` handler. Candidate 2: `external.ExtractJSON[T]` for structured LLM extraction (fence stripping + JSON recovery), charts/chapters/verification refactored onto it.
- **2026-08-08:** UX copy clarified — error messages use "We couldn't..." pattern, processing text uses student-friendly language, chart empty states simplified. Impeccable commands run: init (PRODUCT.md), critique (29/36), harden (clipboard + keyboard), shape (chapter tabs), clarify (copy). Chapter-tabbed view implemented: migration 004 (chapter_id on charts), services Chart.ChapterIndex, repository Chart.ChapterID, handler exposes content + chapter_id, frontend horizontal tabs with ARIA roles + keyboard navigation. Fixed DB startup error (stale DB + migration gap), fixed JSONL logger error rendering, added logger regression test.
- **2026-08-07:** P0 Core Features COMPLETE — 7 chunks across 2 batches. Batch A: claim_diff in GET response (A1), expandable claim comparison UI (A2). Batch B: chapters migration (B1), ChapterRepo + type (B2), pipeline carries chapters (B3), save chapters + expose in API (B4), chapter summary card (B5). All committed and pushed.
- **2026-08-06:** Phase 1 SaaS Foundation COMPLETE — All 10 chunks across 7 phases. Backend: users table, sessions table, auth handlers (signup/login/logout/me), RequireAuth/OptionalAuth middleware, document ownership, paginated list endpoint. Frontend: react-router-dom, login/signup pages, dashboard page with auth check, NotFoundPage, copy-link verification. Responsive smoke test passed (no fixed-width issues). Both Go and frontend builds pass.
- **2026-08-06:** CHUNKS 6–9 — React ErrorBoundary, dropzone aria-label, aria-live polling, security headers middleware. Committed + pushed as `d473212`.
- **2026-08-04:** CHUNK 6 — chart types (bar/line/pie/scatter), reading level badge, backend error messages, copy text button, share dialog modal, processing stage tracking (backend column + callback + frontend display). Committed + pushed.
- **2026-08-02:** Client rejects PDFs larger than 20 MB and non-PDF MIME types before upload; Playwright confirmed both errors and valid-file recovery.
- **2026-08-02:** Fixed React Doctor findings for eager Recharts, static-element interaction, and placeholder-only field. Remaining 10 `result-page.jsx` findings are deferred.
- **2026-08-02:** Chunk 3 public-link disclaimer committed and pushed as `9ae1378`.
- **2026-08-02:** Chunk 2 polling timeout and copy-link behavior committed and pushed as `e7515c0`.
