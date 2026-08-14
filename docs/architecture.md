# Architecture Memory

Canonical specification: [`ARCHITECTURE.md`](./ARCHITECTURE.md). Do not duplicate its full contracts here.

## Current Addendum

- **Auth system (Phase 1):** Users table with email/password_hash, sessions table with token/user_id/expires_at. httpOnly cookies (Secure, SameSite=Lax). AuthMiddleware struct with RequireAuth/OptionalAuth methods. UserIDFromContext helper. Document ownership via nullable user_id on documents table.
- **Migration system:** schema_migrations table tracks applied versions. map[int]string migration system in db.go. Both 001_init.sql and 002_users.sql loaded and applied in order.
- Frontend chart rendering is component-split: `ChartCard` lazy-loads `DataChart`, which dynamically imports Recharts only when a `data_extracted` chart mounts. `DataChart` supports bar/line/pie/scatter via `chartData.type`.
- Processing stage tracking: `documents.processing_stage` column stores current pipeline stage during processing. Pipeline receives `OnStage` callback, calls it at simplifying/verifying/generating_charts transitions. Cleared to NULL on pipeline completion.
- Share dialog: modal overlay in `result-page.jsx` with URL input, copy button, and 7-day expiry note. Triggered from header "Share" button. Escape key dismisses, autoFocus on input, clipboard error feedback.
- Frontend routing: react-router-dom with BrowserRouter. Routes: `/` (upload), `/login`, `/signup`, `/dashboard`, `/:documentId` (result), `*` (404).
- Chapter-tabbed view: migration 004 adds `chapter_id` to charts. Backend links charts to chapters via `chapterIndexToID` map in `saveResult`. Frontend shows horizontal scrollable tabs when 2+ chapters exist, with ARIA roles (`tablist`/`tab`/`tabpanel`) and keyboard navigation (arrow keys). Fallback to linear view for 0-1 chapters.
- Backend architecture remains unchanged: single Go binary, React static frontend, SQLite, and strict `handlers → services → repository/external` flow.
- **Evidence provenance (Chunk 1.1):** `evidence` table (migration 005) with `paper_id` FK cascade, `page`, `figure_id`, `table_id`, `section`, `source_text`, `source_reference`. `EvidenceRepo` (Insert/ListByPaper). Not yet populated by pipeline or exposed in API.
- **Document intake seam (Candidate 1):** `internal/services/intake.go` owns extraction validation + transactional insert; `documents.go` handler is a thin adapter.
- **Generic LLM extraction (Candidate 2):** `external.ExtractJSON[T]` — typed Gemini JSON extraction with fence stripping + recovery. charts/chapters/verification all use it.
- **PDF extraction seam (Candidate 3):** `PDFDocument` + `ParsePDF(pdfBytes, maxCharts)` in `extraction.go` encapsulate text/pages/images with chart cap. Pipeline consumes the bounded document.
- **Frontend hooks:** `useDocumentPoll` (result polling/timeout/retry) and `useAuthSubmit` (auth form submit + error mapping). Shared `AuthForm` component used by login + signup.
