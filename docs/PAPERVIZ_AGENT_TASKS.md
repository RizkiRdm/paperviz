# PaperViz V1 — Agent Task Chunks (Phase 1: SaaS Foundation)

Target agent: low-effort model. Written for machine execution — no ambiguity, one concern per
chunk, sequential dependency order.

## Ground Rules (apply to every chunk below)

- **NEVER** modify test assertions, use `t.Skip()`, or change expected
  values to make a test pass. If a test fails, the fix is in the
  implementation, not the test.
- **ALWAYS** run `grep -rn` for the relevant pattern/symbol BEFORE
  modifying any file, to confirm you understand current usage across the
  codebase. Paste the grep output as part of your proof.
- **ALWAYS** end each chunk with a `PROOF OF COMPLETION` section containing raw,
  unedited terminal output proving the task works (test run, build
  output, curl response, etc). No claim of completion without this.
- Do not start chunk N+1 until chunk N's PROOF OF COMPLETION is verified.
- Assumption baked into this plan: anonymous document creation stays
  functional (no login required to try the product once); only
  persistence/library/ownership requires an account. If this assumption
  is wrong, stop and ask before proceeding past Chunk 4.

---

## Chunk 1 — DB Migration: users table + document ownership

**Context:** Current schema (`migrations/001_init.sql`) has no `users`
table. `documents` has no owner column.

**Task:**
1. Run `grep -rn "documents" migrations/` and `grep -rn "CREATE TABLE" migrations/` first, paste output.
2. Create `migrations/002_users.sql`:
   - `CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, created_at INTEGER NOT NULL)`
   - Add nullable `user_id TEXT REFERENCES users(id) ON DELETE SET NULL` column to `documents` via `ALTER TABLE documents ADD COLUMN user_id TEXT REFERENCES users(id)`
   - Add index `idx_documents_user_id ON documents(user_id)`
3. Do NOT touch `001_init.sql` — this must be an additive migration only.
4. Confirm migration runner picks it up (check how `internal/repository/db.go` applies migrations first via grep).

**Constraints:** Nullable `user_id` is mandatory — existing anonymous
documents must remain valid rows.

**PROOF OF COMPLETION:** Terminal output of the migration running successfully
against a fresh SQLite file, plus `.schema documents` and `.schema users`
output from `sqlite3` confirming the new column/table exist.

---

## Chunk 2 — Signup/Login handlers (backend)

**Context:** No auth exists. Depends on Chunk 1.

**Task:**
1. `grep -rn "bcrypt\|argon2"` in `go.sum` first — check if a password
   hashing lib is already a dependency before adding a new one.
2. Add `POST /api/auth/signup` (email + password → hash password, insert
   user, create session) and `POST /api/auth/login` (verify password,
   create session).
3. Session: httpOnly, Secure, SameSite=Lax cookie holding a random
   session token stored server-side (new `sessions` table: `token TEXT
   PRIMARY KEY, user_id TEXT, expires_at INTEGER`). Do NOT use a JWT
   stored in localStorage — that is a deliberate security decision, not
   optional.
4. Follow existing error contract pattern from `internal/handlers/respond.go`
   (`{"error": "<snake_case_code>"}`) — grep that file first to match style.

**PROOF OF COMPLETION:** `curl` output of a successful signup, a successful
login returning a `Set-Cookie` header, and a failed login with wrong
password returning the correct error code.

---

## Chunk 3 — Auth middleware

**Context:** Depends on Chunk 2.

**Task:**
1. Grep `internal/handlers/middleware.go` first to match existing
   middleware style (how `RequestID`/`Recoverer` are structured in
   `router.go`).
2. Add `RequireAuth` middleware: reads session cookie, looks up session,
   attaches `user_id` to request context if valid, otherwise returns
   `401 {"error": "unauthenticated"}`.
3. Add `OptionalAuth` middleware (does not reject if missing, just
   attaches `user_id` if present) — needed because anonymous upload stays
   supported per the ground-rules assumption.

**PROOF OF COMPLETION:** `curl` output showing a protected test route returning
401 without cookie, and 200 with a valid session cookie from Chunk 2.

---

## Chunk 4 — Wire ownership into document handlers

**Context:** Depends on Chunk 1 + 3.

**Task:**
1. Grep `internal/handlers/documents.go` for the current `Create` and
   `Get` handler bodies first.
2. Apply `OptionalAuth` to `POST /api/documents`: if `user_id` present in
   context, store it on the new document row; if absent, leave `user_id`
   NULL (anonymous, same as today).
3. Add new `GET /api/documents` (list) endpoint, protected by
   `RequireAuth`, returning only documents where `user_id` matches the
   authenticated user, paginated (limit/offset query params, default
   limit 20).

**Constraints:** Do not change behavior of `GET /api/documents/{id}` for
anonymous/shared-link access — that must keep working exactly as before
for any document regardless of owner (this is the shareable-link feature,
do not gate it behind auth).

**PROOF OF COMPLETION:** `curl` output of: (a) anonymous create still working,
(b) authenticated create attaching user_id (verify via direct sqlite
query), (c) `GET /api/documents` returning only that user's documents.

---

## Chunk 5 — Frontend: introduce react-router-dom

**Context:** `frontend/src/App.jsx` currently has an explicit YAGNI
comment saying no router is needed for 2 screens. That assumption is now
false — we're adding login/signup/dashboard/settings.

**Task:**
1. Grep `frontend/src/App.jsx` and `frontend/src/pages/` first to see
   current routing-by-hand logic (window.history / popstate handling).
2. `npm install react-router-dom` in `frontend/`.
3. Replace the manual `documentIdFromLocation`/`popstate` logic with
   proper routes: `/` (upload or landing — see Chunk 9), `/login`,
   `/signup`, `/dashboard`, `/:documentId` (result page, unchanged
   behavior for shared links).
4. Do not change the visual output of `upload-page.jsx` or
   `result-page.jsx` in this chunk — routing only, no UI changes yet.

**PROOF OF COMPLETION:** `npm run build` succeeding with no errors, plus a
screenshot-free terminal confirmation that `npm run dev` starts clean
(paste console output, no errors/warnings about routing).

---

## Chunk 6 — Frontend: Login and Signup pages

**Context:** Depends on Chunk 5 (routes exist) and Chunk 2 (backend
endpoints exist).

**Task:**
1. Grep `frontend/src/components/ui/button.jsx` and `frontend/src/index.css`
   first to reuse existing design tokens (from DESIGN.md) — do not
   introduce new colors/spacing values outside the existing token set.
2. Build `frontend/src/pages/login-page.jsx` and
   `frontend/src/pages/signup-page.jsx`: email + password form, inline
   validation, calls the Chunk 2 endpoints, redirects to `/dashboard` on
   success, shows the server's error code as a human-readable message on
   failure.

**PROOF OF COMPLETION:** Terminal output of `npm run build` succeeding, plus a
manual curl-based confirmation (or note if manual browser testing was
done) that signup → redirect → login flow works end to end.

---

## Chunk 7 — Backend: paginated list endpoint proof + Frontend: Dashboard page

**Context:** Depends on Chunk 4 (list endpoint) and Chunk 5 (routing).

**Task:**
1. Build `frontend/src/pages/dashboard-page.jsx`: fetch
   `GET /api/documents`, render list with status badge per document
   (processing/complete/failed — reuse existing status logic from
   `result-page.jsx`, grep it first), empty state if zero documents with
   a clear "Upload your first PDF" CTA, "New Upload" button linking to
   `/`.
2. Protect the `/dashboard` route client-side: redirect to `/login` if no
   valid session (check via a lightweight `GET /api/auth/me` endpoint —
   add this small endpoint to the backend if it doesn't exist yet, grep
   first to confirm).

**PROOF OF COMPLETION:** `npm run build` output plus terminal proof of
`/api/auth/me` returning user data with valid session cookie and 401
without.

---

## Chunk 8 — Copy-link button fix on result page

**Context:** Flagged as missing in a prior SaaS-readiness review, still
not present per current `result-page.jsx`. Independent of auth work,
safe to do any time after Chunk 5.

**Task:**
1. Grep `frontend/src/pages/result-page.jsx` first to see current layout
   and where a share/copy action would fit.
2. Add a "Copy link" button using `navigator.clipboard.writeText`,
   copying the current document's full shareable URL. Show a brief
   confirmation state (e.g. icon swaps to checkmark for ~2s) after
   successful copy — do not use a browser `alert()`.

**PROOF OF COMPLETION:** `npm run build` output confirming no build errors, and
the exact JSX diff showing the button and clipboard call.

---

## Chunk 9 — 404 page

**Context:** Independent, safe to do any time after Chunk 5.

**Task:**
1. Grep `internal/handlers/router.go` for the current `NotFound` handler
   (it currently falls back to serving the SPA for any unmatched path).
2. Add a `NotFoundPage` React route (catch-all `*` route in
   react-router) that renders when a `:documentId` route resolves with a
   `document_not_found` error from the API (grep `result-page.jsx`'s
   error handling first to see where this check belongs) — distinct from
   a generic 404, since most "not found" cases here are expired/invalid
   document links, not bad routes.

**PROOF OF COMPLETION:** `npm run build` output plus confirmation (terminal) of
visiting a nonexistent document ID and getting the new not-found UI
state instead of a blank/broken screen.

---

## Chunk 10 — Responsive smoke test

**Context:** Final chunk of Phase 1. Depends on all UI chunks above.

**Task:**
1. Grep for any fixed-pixel widths in `frontend/src/pages/*.jsx` and
   `frontend/src/components/*.jsx` (`grep -rn "px]" frontend/src` for
   Tailwind arbitrary values, and `grep -rn "width:" frontend/src`).
2. Fix any layout that breaks below 375px width, with particular
   attention to the recharts chart container in `data-chart.jsx` (grep it
   first) and the upload dropzone.
3. Do not redesign anything — only fix actual breakage found via the grep
   above and manual resize testing.

**PROOF OF COMPLETION:** List of every fixed-width issue found (file:line) and
the diff that fixed each one. If none found, paste the grep output
showing why (e.g. all widths already use relative units).

---

## Not included in this batch (deliberately)

Billing/Stripe integration, dark mode, password reset, and email
verification are V1.5 — do not start them until Phase 1 above is fully
verified and merged. Chunking speculative future work before V1 ships is
wasted planning effort.