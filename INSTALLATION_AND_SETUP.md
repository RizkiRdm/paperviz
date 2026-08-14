# INSTALLATION_AND_SETUP.md — PaperViz

This document explains how to install, configure, run, and build PaperViz
from a clean checkout. It also records what was cleaned up during
scaffolding and one important caveat about how this codebase was verified.

---

## 0. Important: Verification Status

The Go backend in this repository was written and statically reviewed
(syntax-checked with `gofmt`, cross-referenced by hand for type
correctness across all files) but **could not be compiled or tested**
inside the sandbox used to build it. That sandbox's network egress
allowlist does not include `modernc.org` or `golang.org`, and
`modernc.org/sqlite`'s core transitive dependency (`modernc.org/libc`) has
no GitHub mirror to route around that restriction.

**Before treating this MVP as done, run on a normal machine with internet
access:**

```bash
go mod tidy
go build ./...
go vet ./...
```

The frontend, by contrast, **was fully compiled and lint-checked** in the
build sandbox (`vite build` succeeded with 0 errors, `oxlint` reported 0
warnings) — it does not carry the same caveat.

See `AGENTS.md`'s "Known Issues" section for one functional gap found
during review: the chart image-fallback path stores an image blob that
the API/frontend cannot currently display (only its generated annotation
shows). This was flagged rather than fixed unilaterally, per this
project's own escalation rules around adding new API surface.

---

## 1. Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| Go | 1.22+ (1.25+ preferred per ARCHITECTURE.md) | `modernc.org/sqlite` is a pure-Go, no-CGO SQLite driver — no C compiler or `CGO_ENABLED` setup needed. |
| Node.js | 18+ (24+ preferred) | For the frontend build only; not needed at runtime. |
| npm | 9+ | Ships with Node. |
| Google Gemini API key | — | Get one at [aistudio.google.com/apikey](https://aistudio.google.com/apikey). The free tier is assumed sufficient for MVP validation traffic (PRD.md Technical Assumptions). |

No Docker, no external database server, no message broker — this is a
single Go binary plus a SQLite file, by design (ARCHITECTURE.md
Non-goals).

---

## 2. Installation

### Clone and enter the repository

```bash
git clone <your-repo-url> paperviz
cd paperviz
```

### Backend: install Go dependencies

```bash
go mod tidy
```

This resolves and pins the exact dependency versions declared in `go.mod`:
`github.com/go-chi/chi/v5`, `github.com/ledongthuc/pdf`,
`github.com/pdfcpu/pdfcpu`, and `modernc.org/sqlite`, plus their
transitive dependencies. It also generates `go.sum` (not committed — see
`.gitignore` — regenerate it locally on first setup).

### Configure environment variables

```bash
cp .env.example .env
```

Then edit `.env` and set at minimum:

```
GEMINI_API_KEY=your-real-key-here
```

All other variables in `.env.example` have sane defaults (see that file's
comments) and only need overriding if you want non-default behavior.

**The Go server reads environment variables directly via `os.Getenv`** —
it does not load `.env` files itself (no dependency was added for that,
per YAGNI). Export the variables into your shell before running the
server, or use a tool like `direnv`, or prefix the run command:

```bash
export $(grep -v '^#' .env | xargs) && go run ./cmd/server
```

### Initialize the database

No separate migration step is required. The server applies
`migrations/001_init.sql` automatically on first startup if the
`documents` table doesn't already exist (see `repository/db.go`'s `Open`
function). The SQLite file is created at the path in `DATABASE_PATH`
(default `paperviz.db`) the first time the server runs.

There is no seed data — PaperViz has no concept of accounts or
pre-populated content (PRD.md: no user accounts, ephemeral links only).

### Frontend: install dependencies

```bash
cd frontend
npm install
cd ..
```

---

## 3. Running in Development

Run the backend and frontend as two separate processes during development
(the frontend dev server proxies `/api/*` calls to the Go backend — see
`frontend/vite.config.js`).

**Terminal 1 — backend:**
```bash
export $(grep -v '^#' .env | xargs)
go run ./cmd/server
```
Server starts on `http://localhost:8080` (or your configured `PORT`).

**Terminal 2 — frontend:**
```bash
cd frontend
npm run dev
```
Dev server starts on `http://localhost:5173` (Vite's default) and proxies
API calls to port 8080. Open this URL in your browser during development.

---

## 4. Building for Production

```bash
# Build the frontend static assets
cd frontend
npm run build
cd ..

# Build the Go binary
go build -o server ./cmd/server
```

`npm run build` outputs to `frontend/dist/`, which the Go binary serves
directly at runtime (see `handlers/router.go`'s `NewRouter`, and
`STATIC_DIR` in `.env.example` — defaults to exactly this path). This is
what "single binary" means in ARCHITECTURE.md: one Go process serves both
the API and the built React app, no separate frontend server needed in
production.

**Run the production build:**
```bash
export $(grep -v '^#' .env | xargs)
./server
```

---

## 5. Project Structure

```
paperviz/
├── .env.example              — environment variable template
├── .gitignore                — excludes secrets, DB files, build artifacts
├── AGENTS.md                 — agent operating rules + Known Issues log
├── ARCHITECTURE.md            — architecture source of truth
├── DESIGN.md                  — design system source of truth
├── PLAN.md                    — phased implementation checklist (kept in sync with actual progress)
├── PRD.md                     — product requirements source of truth
├── go.mod                     — Go module + pinned dependency versions
│
├── cmd/server/
│   └── main.go                — entrypoint: config, DB open, router wiring, expiry sweep start
│
├── handlers/                  — HTTP layer only (parse, validate shape, call services, serialize)
│   ├── documents.go           — POST /api/documents, GET /api/documents/:id
│   ├── respond.go             — shared JSON response helpers
│   ├── router.go               — chi router wiring + static file serving
│   └── validation.go           — PDF magic-byte content check
│
├── services/                  — business logic (pure functions + orchestration)
│   ├── pipeline.go             — the sequential extract→simplify→verify→chart orchestrator
│   ├── extraction.go           — PDF extraction service wrapper
│   ├── simplification.go       — Gemini prompts for Simplified/ELI5 reading levels
│   ├── verification.go         — claim-diff extraction + comparison
│   ├── charts.go               — chart re-visualization: data-path, image-fallback, omitted
│   ├── expiry.go                — 7-day inactivity expiry sweep loop
│   └── types.go                 — shared service-layer types
│
├── repository/                — SQLite CRUD only, no business logic
│   ├── db.go                   — connection open + migration bootstrap
│   ├── documents.go             — documents table CRUD
│   ├── charts.go                 — charts table CRUD
│   ├── claim_diffs.go            — claim_diffs table CRUD
│   ├── types.go                  — domain types + enum constants
│   └── id.go                      — cryptographically random ID generator (nanoid-style)
│
├── external/                  — third-party API/library wrappers
│   ├── gemini.go                — direct HTTP client for Gemini API (no SDK, no gateway)
│   └── pdf.go                    — PDF text + image extraction (in-memory only, no disk writes)
│
├── migrations/
│   └── 001_init.sql             — full schema: documents, charts, claim_diffs tables
│
└── frontend/                   — Vite + React SPA, built to static assets
    ├── index.html
    ├── vite.config.js           — Tailwind v4 plugin + dev API proxy
    ├── package.json
    └── src/
        ├── main.jsx              — React entry point
        ├── App.jsx                — top-level page switch (Upload ↔ Result), URL-based routing
        ├── index.css              — DESIGN.md's full token system as Tailwind v4 @theme
        ├── lib/
        │   ├── api.js             — fetch wrapper for the two backend endpoints
        │   └── utils.js            — cn() class-merging helper
        ├── components/
        │   ├── upload-dropzone.jsx
        │   ├── chart-card.jsx
        │   └── ui/
        │       ├── button.jsx
        │       ├── reading-level-selector.jsx
        │       └── status-banners.jsx
        └── pages/
            ├── upload-page.jsx
            └── result-page.jsx
```

---

## 6. Repository Cleanup

The following default framework files were generated by `npm create vite`
and then removed, since they were unrelated to PaperViz:

| File | Reason removed |
|---|---|
| `frontend/src/App.css` | Default Vite demo styling — replaced entirely by `index.css`'s DESIGN.md token system. |
| `frontend/src/index.css` (original) | Default Vite demo styles — overwritten with the real design system. |
| `frontend/src/assets/react.svg`, `vite.svg`, `hero.png` | Demo/placeholder images not used anywhere in PaperViz's UI. |
| `frontend/public/vite.svg`, `favicon.svg`, `icons.svg` | Same — demo assets. `public/` itself was removed once empty. |
| `frontend/README.md` | Generic Vite template README — this document replaces it. |
| `frontend/.gitignore` | Redundant with the root `.gitignore`, which already covers `frontend/node_modules/` and `frontend/dist/`. |

`frontend/src/App.jsx` and `frontend/src/main.jsx` were **kept** (Vite
requires these exact filenames as build entry points) but their content
was fully replaced with PaperViz's actual application code — none of the
original demo counter/logo code remains.

Additionally, during dependency review: `class-variance-authority` was
installed anticipating a shadcn-style variant pattern, then removed after
`Button` ended up using a simpler plain object map instead — kept the
dependency list to only what's actually imported (see quality checklist
in the project's engineering guidelines).

---

## 7. Git Workflow

Once you've run the verification steps in Section 0 successfully:

```bash
# Review what's changed
git status
git diff

# Confirm formatting and linting pass
gofmt -l .            # should print nothing
cd frontend && npx oxlint && cd ..

# Confirm the build succeeds
go build ./...
cd frontend && npm run build && cd ..

# Commit
git add .
git commit -m "feat: implement PaperViz MVP per PRD/ARCHITECTURE/PLAN"

# Push
git push origin <branch-name>
```

Follow `AGENTS.md`'s Git Conventions for ongoing work: branch names like
`phase-N-<short-description>`, atomic commits, no PR requirement for this
solo-dev workflow.

---

## 8. Known Limitations (MVP Scope)

These are intentional, documented exclusions — not oversights. See
`PRD.md`'s "Non-MVP Features" and `AGENTS.md`'s "Out of Scope" for the
full list. Notably:

- No user accounts, no authentication — access is link-possession-only by design.
- No OCR for scanned PDFs — text-layer PDFs only.
- No interactive charts — static re-rendered charts only.
- Single instance, no horizontal scaling, no job queue — by design for this traffic volume.
- Chart image-fallback display gap — see Section 0 above and `AGENTS.md`'s Known Issues.
