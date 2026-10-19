# Getting Started

This tutorial takes a clean PaperViz checkout to a working local application and first paper analysis. It also shows how to start the repository's local MCP process without claiming that the current MCP ingestion path runs the full analysis pipeline.

## 1. Install prerequisites

Install:

- Go 1.25
- Node.js 24
- npm
- `curl`

The server validates two required variables at startup: `CREDENTIAL_ENCRYPTION_KEY` and `CLERK_SECRET_KEY` (Clerk owns authentication — see `docs/DECISIONS/ADR-003-clerk-auth.md`). The frontend build additionally needs `VITE_CLERK_PUBLISHABLE_KEY`. Get a free Clerk application for local development; no other third-party account is required.

## 2. Get source and install dependencies

```bash
git clone <repository-url> paperviz
cd paperviz

go mod download
npm --prefix frontend ci
```

Install E2E dependencies only when you plan to run browser tests:

```bash
npm --prefix e2e ci
npm --prefix e2e exec -- playwright install chromium
```

## 3. Create local configuration

```bash
cp .env.example .env.development
```

A minimal bootable local file looks like this:

```dotenv
CREDENTIAL_ENCRYPTION_KEY=base64-encoded-32-random-bytes
CLERK_SECRET_KEY=sk_test_...
VITE_CLERK_PUBLISHABLE_KEY=pk_test_...

INGESTION_ENABLED=true
PIPELINE_MAX_CONCURRENCY=2

DATABASE_PATH=paperviz.db
PORT=8080
```

Two processes run locally: the Go API on `PORT`, and the SPA in Vite on
`http://localhost:5173` (its `/api` requests are proxied to the Go port, so
`VITE_API_BASE_URL` can stay unset locally). Start them with `make dev` and
`make frontend-dev`.

Generate the one required secret with:

```bash
openssl rand -base64 32
```

Notes:

- `CREDENTIAL_ENCRYPTION_KEY` must be present and valid; the server exits at startup otherwise. Losing or rotating it makes every stored user model key unreadable.
- `CLERK_SECRET_KEY` must be present too; the server exits without it because a wrong key makes every login fail while `/healthz` still reports ok.
- Users bring their own model key through the web UI, so no model credential belongs in this file.
- The MCP process reads only `PAPERVIZ_API_KEY` (plus `DATABASE_PATH`/`MIGRATIONS_DIR`); it is deterministic and calls no model (P28), so no model key is needed for it either.
- The server reads process environment variables directly; it does not load `.env` itself.
- Never commit `.env.development` or real credentials. `.env.example` is the only tracked env file.
- Production has no env file. Render's env vars are the single production source of truth.

## 4. Start backend

Export `.env.development` into the current shell:

```bash
set -a
source .env.development
set +a
```

Start server:

```bash
go run ./cmd/server
```

Expected startup behavior:

- migrations load from `migrations/`;
- SQLite is created at `DATABASE_PATH`;
- WAL, foreign keys, and `synchronous=NORMAL` are enabled;
- the document expiry sweep starts;
- HTTP server listens on `PORT`.

Verify database-backed health:

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```

A `503` response with `db_unreachable` means the process started but SQLite is not healthy.

## 5. Start frontend

Open a second terminal from repository root:

```bash
npm --prefix frontend run dev
```

Open `http://localhost:5173`.

Vite proxies `/api` to `http://localhost:${PORT}`. Vite reads `VITE_*` variables from the environment, so if the page reports "publishable key is not set", run `set -a; source .env.development; set +a` in the frontend terminal before `npm run dev`. Keep backend and frontend terminals running independently during development.

## 6. Analyze your first paper

Use a paper with selectable text. PaperViz does not perform OCR for image-only scans.

1. Open `http://localhost:5173`.
2. Choose an input mode:
   - **PDF** for a local text-layer PDF;
   - **Paste** for at least 50 characters of text;
   - **DOI** for a DOI beginning with `10.`;
   - **URL** for an HTTP or HTTPS source URL.
3. For PDF or pasted text, choose `Simplified` or `ELI5`.
4. Select **Simplify paper** or **Import**.
5. Wait for navigation to `/:documentId`.
6. Follow visible processing stages. The frontend polls every two seconds, shows a long-running notice after two minutes, and stops after ten minutes.
7. Inspect the result sections:
   - understanding;
   - evidence and claims;
   - figures;
   - original source;
   - management actions.

If processing fails, use the inline error and retry controls. Inputs remain in the form so the same request can be retried.

### What success looks like

A successful run should show:

- a terminal document state other than `processing`;
- a simplified explanation;
- verification state and mismatch detail when applicable;
- charts with grounding and provenance state;
- source-derived evidence where extraction succeeded.

A paper without extractable numeric evidence can still produce a valid explanation with zero charts. That is expected, not a rendering failure.

## 7. Run MCP locally

The repository ships an stdio MCP server, not an HTTP MCP route.

Build it:

```bash
go build -o paperviz-mcp ./cmd/mcp
```

MCP startup requires:

```dotenv
DATABASE_PATH=/absolute/path/to/paperviz/paperviz.db
MIGRATIONS_DIR=/absolute/path/to/paperviz/migrations
PAPERVIZ_API_KEY=local-mcp-key
```

No model key: as of P28 the MCP server is deterministic end to end and calls no provider.

Use the same absolute `DATABASE_PATH` as the web server when both processes should read the same documents.

Start process:

```bash
export PAPERVIZ_API_KEY=local-mcp-key
./paperviz-mcp
```

A correctly connected MCP client should see five tools:

```text
ingest_document
search_documents
get_document
get_figures
get_evidence
```

### Important MCP behavior

`ingest_document` currently performs deterministic intake only:

1. validates non-empty text;
2. enforces the 500 KiB tool input limit;
3. validates reading level;
4. inserts a document with `status = processing`;
5. returns its ID.

It does **not** call `RunPipelineAndPersist`, Gemini, claim verification, or chart generation. No queue exists to advance that document later. Use the web ingestion path when you need the complete processing pipeline.

### Why `/agents` is not the local MCP setup

`frontend/src/pages/agents-page.jsx` currently generates remote configuration using `@anthropic-ai/mcp-remote` and `https://paperviz.com/api/mcp`. The current `internal/handlers/router.go` does not register `/api/mcp`. For this repository, configure an MCP client to launch the local stdio binary instead, or deploy a compatible remote transport before using that page's generated configuration.

## 8. Build production-style local binary

```bash
npm --prefix frontend run build
CGO_ENABLED=0 go build -o server ./cmd/server
```

Run it with the same environment used for development:

```bash
set -a
source .env.development
set +a
./server
```

Open `http://localhost:8080` — that is the API, so `GET /` answers with JSON
status and unknown paths answer 404. For the web app, run `make frontend-dev` (or
`cd frontend && npm run dev`) and open `http://localhost:5173`.

Repository shortcuts are also available:

```bash
make build          # Go binary only
make run            # build + run the API with .env.development exported
make frontend-dev   # Vite dev server for the SPA
```

## 9. Run verification commands

Backend:

```bash
go test ./...
go vet ./...
gofmt -l .
```

Frontend:

```bash
npm --prefix frontend run lint
npm --prefix frontend run build
```

E2E, with server already running and frontend built:

```bash
BASE_URL=http://localhost:8080 npm --prefix e2e test
```

## Common startup failures

| Symptom | Meaning | Action |
|---|---|---|
| `CREDENTIAL_ENCRYPTION_KEY environment variable is required` | Required secret missing or malformed | `openssl rand -base64 32` and export it before starting |
| `CLERK_SECRET_KEY environment variable is required` | Clerk secret missing | Export a `sk_test_...`/`sk_live_...` value from your Clerk application |
| Frontend shows "publishable key is not set" | `VITE_CLERK_PUBLISHABLE_KEY` absent at Vite start/build | Source `.env.development` in that shell, or use `make frontend/dist` for builds |
| `failed to load migrations` | Server started outside expected layout or migration file missing | Run from repository root or correct working directory |
| `failed to open database` | SQLite path or permissions invalid | Check `DATABASE_PATH` and directory permissions |
| `/healthz` returns `db_unreachable` | Database ping or query failed | Check database file, filesystem, and server logs |
| Frontend loads but API calls fail | Backend not running or `PORT` mismatch | Confirm backend health and Vite `PORT` value |
| PDF returns `no_text_layer` | No selectable PDF text | Use OCR externally, then paste or import extracted text |
| MCP document remains `processing` | Expected current behavior for MCP-only intake | Use web ingestion for full pipeline |
| Generated `/agents` config cannot connect | Remote MCP transport is not implemented in current router | Use local stdio MCP configuration |

## Next steps

- Understand system boundaries: [Overview](overview.md)
- Learn safe change workflows: [Maintainer Guide](maintainer-guide.md)
- Look up exact routes and files: [Codebase Reference](codebase-reference.md)
- Study processing internals: [Data Pipeline](DATA_PIPELINE.md)
