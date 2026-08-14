# Engineering Decisions

Historical entries are append-only. Deprecate decisions instead of deleting them.

## 2026-08-06 — Auth system: Option A (schema_migrations table)

- **Context:** Need to run multiple SQL migrations in order. Current `db.go` had hardcoded migration logic.
- **Decision:** Create `schema_migrations` table tracking applied versions. `db.go` rewritten with `map[int]string` migration system. Both 001_init.sql and 002_users.sql loaded and applied in order.
- **Alternatives considered:** Embed all migrations in one file; use a third-party migration library.
- **Consequences:** Robust migration tracking for V1.5+. DB schema changed — dev DB needs reset. No external dependencies added.

## 2026-08-06 — Server-side sessions (no JWT/localStorage)

- **Context:** Need session management for auth. JWT in localStorage is insecure (XSS vulnerability).
- **Decision:** Server-side sessions in SQLite (sessions table: token, user_id, expires_at). httpOnly, Secure, SameSite=Lax cookies. No JWT, no localStorage.
- **Alternatives considered:** JWT in httpOnly cookies; JWT in localStorage; third-party session store.
- **Consequences:** More secure than localStorage JWT. Session data server-side (can invalidate). Requires SQLite for session storage.

## 2026-08-06 — AuthMiddleware as struct (not standalone functions)

- **Context:** Need RequireAuth and OptionalAuth middleware. Current middleware.go has standalone functions.
- **Decision:** Create `AuthMiddleware` struct with DB injection. `RequireAuth` and `OptionalAuth` methods. `contextKey` type for context values. `UserIDFromContext` helper.
- **Alternatives considered:** Standalone functions with `dbFromRequest` context approach; middleware per-handler.
- **Consequences:** Cleaner than context-passing approach. Consistent with chi middleware patterns. Easy to test with mock DB.

## 2026-08-06 — Document ownership: nullable user_id

- **Context:** Need to link documents to users while keeping anonymous upload functional.
- **Decision:** Add nullable `user_id TEXT REFERENCES users(id)` to documents. `OptionalAuth` on POST /api/documents: if user_id in context, store it; if absent, leave NULL (anonymous).
- **Alternatives considered:** Required user_id (breaks anonymous flow); separate owned_documents table.
- **Consequences:** Anonymous documents stay valid rows. Authenticated uploads get ownership. List endpoint filters by user_id.

## 2026-08-06 — Frontend routing: react-router-dom

- **Context:** App.jsx had manual popstate routing. Adding login/signup/dashboard requires proper router.
- **Decision:** Install react-router-dom, replace manual routing with BrowserRouter + Routes. Routes: `/` (upload), `/login`, `/signup`, `/dashboard`, `/:documentId` (result), `*` (404).
- **Alternatives considered:** Keep manual routing; use other router libraries.
- **Consequences:** Proper URL handling. Nested routes possible. History API managed by library.

## 2026-08-06 — Error boundary as class component

- **Context:** Zero error boundaries in app. Any uncaught render error blanks entire page with no message.
- **Decision:** Create `ErrorBoundary` class component with `getDerivedStateFromError` + `componentDidCatch`. Wrap both `ResultPage` and `UploadPage` branches in `App.jsx`. Fallback UI uses existing `ErrorBanner` + Reload button.
- **Alternatives considered:** Hooks-only version (can't catch render errors); error boundary per page; Sentry integration.
- **Consequences:** Class component required (React limitation). Catches all render errors in both pages. No external dependencies added.

## 2026-08-06 — Security headers via chi middleware

- **Context:** Static file serving sets no security headers (no CSP, X-Frame-Options, X-Content-Type-Options).
- **Decision:** New `SecurityHeaders` middleware in `internal/handlers/security_headers.go`. Sets `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, and CSP allowing Google Fonts + Fontshare. Wired after `Recoverer`, before `slogRequestLogger`.
- **Alternatives considered:** Headers in `cmd/server/main.go`; per-handler headers; nginx/reverse proxy.
- **Consequences:** All responses get security headers. CSP allows `fonts.googleapis.com`, `fonts.gstatic.com`, `api.fontshare.com`. `connect-src 'self'` blocks cross-origin API calls.

## 2026-08-07 — PipelineOutput moved from pipeline.go to types.go

- **Context:** Adding `Chapters` field to `PipelineOutput` required modifying the struct. Discovered `PipelineOutput` was declared in both `pipeline.go` and `types.go` (redeclaration error).
- **Decision:** Keep `PipelineOutput` in `types.go` only. Remove duplicate from `pipeline.go`.
- **Alternatives considered:** Keep in `pipeline.go` and remove from `types.go`.
- **Consequences:** Single source of truth for pipeline output type. All service-layer types now co-located in `types.go`.

## 2026-08-07 — MismatchDetail as *string in claimDiffResponse

- **Context:** `ClaimDiff.MismatchDetail` is `*string` in repository. Initially declared as `string` in `claimDiffResponse` wire struct.
- **Decision:** Keep `*string` in response struct to match repository type exactly. Avoids nil-pointer dereference on documents without mismatch detail.
- **Alternatives considered:** Convert to `string` with empty-default; use `omitempty` with `string`.
- **Consequences:** API consumers see `null` for documents without mismatch detail (honest representation). No silent empty-string conversion.

## 2026-08-06 — aria-live on processing branch only

- **Context:** Screen reader users get no announcement when document status changes from "processing" to "complete"/"failed".
- **Decision:** Add `role="status"` + `aria-live="polite"` to processing branch `<div>` wrapper. ErrorBanner/WarningBanner already have `role="alert"` (self-announcing).
- **Alternatives considered:** aria-live on complete branch; aria-live on entire ResultPage; live region outside processing branch.
- **Consequences:** Stage changes announced ("Simplifying language..." → "Verifying claims..."). No double-wrapping of error/warning banners.

## 2026-08-04 — Chart type switching via chartData.type

- **Context:** `data-chart.jsx` only rendered BarCharts. Pipeline generates charts with `type` field (bar/line/pie/scatter) but frontend ignored it.
- **Decision:** Switch on `chartData.type` and render the correct Recharts component (BarChart, LineChart, PieChart, ScatterChart). Pie uses `Cell` with color array. Scatter uses category X-axis.
- **Alternatives considered:** Always render bar; add a chart-type selector UI.
- **Consequences:** Charts now render in their intended type. Pie charts show percentage labels. No new dependencies.

## 2026-08-04 — Processing stage via DB column + pipeline callback

- **Context:** Processing spinner showed generic "Simplifying & Verifying..." for entire 30-60s pipeline. User had no visibility into progress.
- **Decision:** Add `processing_stage TEXT` column to documents table. Pipeline receives `OnStage` callback, calls it at each stage transition (simplifying/verifying/generating_charts). Handler wires callback to `docRepo.UpdateStage`. GET response returns `processing_stage`. Frontend maps stages to labels.
- **Alternatives considered:** Return stage from GET without persisting (in-memory only); use existing `error_message` field for stage.
- **Consequences:** DB schema change requires reset. Stage is cleared on pipeline completion (`processing_stage = NULL` in UpdateStatus). Minimal overhead — one UPDATE per stage transition.

## 2026-08-04 — Share dialog as modal overlay

- **Context:** CopyLinkButton was inline in header. For a shareable-link product, a proper share dialog with URL, copy, and expiry info is expected.
- **Decision:** New `ShareDialog` component rendered as fixed overlay. Shows URL input (read-only, auto-select on focus), copy button, 7-day expiry note. Triggered from header "Share" button.
- **Alternatives considered:** Keep inline CopyLinkButton; add a dropdown popover.
- **Consequences:** One new component inside result-page.jsx (not a separate file — avoids unnecessary file proliferation for MVP). Modal dismisses on backdrop click or X button.

## 2026-08-02 — Load Recharts on demand

- **Context:** React Doctor reported the static Recharts import as expensive initial-page JavaScript.
- **Decision:** `ChartCard` lazy-loads `DataChart`; `DataChart` uses runtime `import("recharts")`, handles load failure, and ignores completion after unmount.
- **Alternatives considered:** Keep the eager import; move a static import into a lazy wrapper; dynamic import in the existing component.
- **Consequences:** Initial app JS falls from about 598 kB to 247 kB; Recharts becomes a separate async chunk. Chart display briefly shows existing-token loading text.

## 2026-08-02 — Validate selected files in UploadPage

- **Context:** `UploadDropzone` receives files, while `UploadPage` owns selected-file and error state plus server-aligned messages.
- **Decision:** Validate size and MIME type in `UploadPage.onFileChange`; invalid files clear selection and reuse `file_too_large` or `invalid_file_type` messages.
- **Alternatives considered:** Export error messages into the dropzone; pass validation errors through new props; duplicate checks in selection and submit paths.
- **Consequences:** Picker and drag/drop share one validation boundary without prop drilling. Strict MIME checking may reject valid PDFs whose browser reports an empty MIME type.

## 2026-08-02 — Use native semantics for upload interaction

- **Context:** Clickable dropzone markup was a static `div`, inaccessible by keyboard and unnamed as an interactive control.
- **Decision:** Use a native `button type="button"` beside the hidden file input and add a visible associated label to pasted-text textarea.
- **Alternatives considered:** Add `role`, `tabIndex`, and custom keyboard handlers to the `div`.
- **Consequences:** Browser-native focus, keyboard activation, and screen-reader semantics replace custom behavior.

## 2026-08-08 — Chapter-tabbed view: backend links charts to chapters

- **Context:** Chapter list was informational but not navigable. Students couldn't jump to specific sections. True tabbed view requires chapter-segmented text and chart-chapter linking.
- **Decision:** Add `chapter_id` to charts table (migration 004). Services `Chart` gets `ChapterIndex` field. Repository `Chart` gets `ChapterID` field. Handler builds `chapterIndexToID` map in `saveResult` to link charts to chapters. Frontend shows horizontal scrollable tabs when 2+ chapters exist.
- **Alternatives considered:** Anchor scroll (simpler but less focused); sticky sidebar (more chrome); tabbed view without backend changes (fragile text splitting).
- **Consequences:** DB schema change requires reset. Backend changes are minimal (struct fields + mapping logic). Frontend gets proper tabbed interface with ARIA accessibility. Charts are now scoped to chapters in the API response.

## 2026-08-08 — UX copy: "We couldn't..." pattern for errors

- **Context:** Error messages used internal jargon ("simplification step", "text extraction", "chart extraction") that students don't understand. Some messages were vague ("Something went wrong").
- **Decision:** Rewrite all error messages to use "We couldn't..." pattern with student-friendly language. Processing text uses "checking it against the original" instead of "verifying key statements against source". Chart empty states simplified.
- **Alternatives considered:** Keep technical terms but add explanations; use generic "Error occurred" messages.
- **Consequences:** All error messages are now comprehensible to first-time students. Consistent voice across the product. Recovery hints added where helpful ("It may be image-only or corrupted").

## 2026-08-08 — Result page hardening: clipboard + keyboard

- **Context:** Clipboard fallback was silent `catch {}` — user clicks copy, nothing happens, no feedback. Share dialog had no Escape dismiss or autoFocus.
- **Decision:** Add clipboard error feedback ("Couldn't copy automatically. Select the link and press Ctrl+C / Cmd+C."). Add Escape key handler to share dialog. Add autoFocus to share input. Add `aria-label` to share input.
- **Alternatives considered:** Keep silent fallback; add toast notifications; use `document.execCommand('copy')` as fallback.
- **Consequences:** User gets clear feedback when clipboard fails. Keyboard-only users can dismiss share dialog. Screen readers get proper labels.

## 2026-08-08 — PipelineOutput chapters-first insertion order

- **Context:** Original `saveResult` inserted charts before chapters. New requirement: link charts to chapters via `chapter_id`. Chapters must be inserted first to get their IDs.
- **Decision:** Restructure `saveResult` to insert chapters first, build `chapterIndexToID` map, then insert charts with `chapter_id` linked.
- **Alternatives considered:** Generate chapter IDs in pipeline (before chart insertion); use display order as chapter reference.
- **Consequences:** Charts are properly linked to chapters. Insertion order is now chapters → charts (was charts → chapters). No functional change to existing behavior.

## 2026-08-14 — Candidate 1: transactional document intake in services

- **Context:** Handler `documents.go` did validation + DB insert inline; mixed HTTP and business logic in one layer.
- **Decision:** New `internal/services/intake.go` `ValidateAndInsert` owns extraction validation (no-text-layer, size, MIME) + transactional row insert. Handler is now a thin adapter.
- **Alternatives considered:** Keep logic in handler; move only validation to service.
- **Consequences:** Handler shrinks; business rules live in services layer per ARCHITECTURE.md. One place to test intake.

## 2026-08-14 — Candidate 2: external.ExtractJSON[T] generic LLM extraction

- **Context:** charts/chapters/verification each re-implemented prompt→JSON parsing with ad-hoc fence stripping and error recovery.
- **Decision:** Add `ExtractJSON[T]` to `internal/external/gemini.go` — generic typed extraction with fence stripping + JSON recovery. All three services refactored onto it.
- **Alternatives considered:** Keep per-service parsing; add a JSON schema lib.
- **Consequences:** Single parsing path, less drift. DiffClaims call count unchanged (2). No new deps.

## 2026-08-14 — Candidate 3: PDFDocument/ParsePDF extraction seam

- **Context:** `pipeline.go` mixed extraction (ExtractFromPDF + manual chart capping) inline; PDF details leaked into orchestration.
- **Decision:** `PDFDocument` domain object + `ParsePDF(pdfBytes, maxCharts)` in `extraction.go` encapsulate text/pages/charts + cap. Pipeline calls ParsePDF.
- **Alternatives considered:** Keep capping in pipeline; move only the cap into a helper.
- **Consequences:** Pipeline only sees the bounded document object. Chart cap is a service-internal concern.

## 2026-08-14 — Candidate 4: useDocumentPoll hook

- **Context:** ResultPage owned 6 pieces of polling state (doc, error, notFound, timedOut, takingLong, retryNonce) inline; effect + timers tangled with render.
- **Decision:** Extract `frontend/src/hooks/use-document-poll.js` returning `{ doc, error, notFound, timedOut, takingLong, retry }`. ResultPage keeps only UI state.
- **Alternatives considered:** Keep inline; context provider.
- **Consequences:** Polling lifecycle (interval, soft warn, timeout, cleanup) testable in isolation. ResultPage shrinks.

## 2026-08-14 — Candidate 5: AuthForm + useAuthSubmit consolidation

- **Context:** Login and Signup pages duplicated the same form shell, input styling, error mapping, and submit fetch logic (~113 lines each).
- **Decision:** Shared `AuthForm` component (uncontrolled inputs) + `useAuthSubmit(endpoint, errorMessages)` hook returning `{ error, isSubmitting, submit, setError }`. Pages become ~40-line wrappers.
- **Alternatives considered:** Form library; keep duplication.
- **Consequences:** One place for auth form styling per DESIGN.md. Uncontrolled inputs avoid state noise. Submit returns boolean so caller controls redirect.

## 2026-08-14 — Evidence model: dedicated table over inline fields

- **Context:** Chunk 1.1 requires reusable provenance — paper_id, page, figure_id, table_id, section, source_text, source_reference.
- **Decision:** New `evidence` table (migration 005) FK-cascading to documents. EvidenceRepo with Insert/ListByPaper. Follows existing repo patterns.
- **Alternatives considered:** JSON column on documents; embed in simplified_text markdown.
- **Consequences:** Queryable, indexable provenance per paper. Clean seam for Chunk 8.2 Evidence Graph later. Schema change requires migration 005 (already registered).

