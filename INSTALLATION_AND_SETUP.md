# Installation and Setup

This file preserves the repository's traditional setup entrypoint. The canonical first-run tutorial is [`docs/getting-started.md`](docs/getting-started.md).

## Prerequisites

| Requirement | Version |
|---|---|
| Go | 1.25 |
| Node.js | 24 |
| npm | Included with Node.js |
| Google Gemini API key | Required for real processing |

No external database server, C compiler, message broker, or Docker installation is required for local development.

## Quick setup

```bash
cp .env.example .env
go mod download
npm --prefix frontend ci
```

Edit `.env` before starting the server. The HTTP process validates these variables at startup:

```dotenv
GEMINI_API_KEY=your-real-key
GOOGLE_CLIENT_ID=non-empty-local-value
GOOGLE_CLIENT_SECRET=non-empty-local-value
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback
STRIPE_SECRET_KEY=non-empty-local-value
STRIPE_WEBHOOK_SECRET=non-empty-local-value
FRONTEND_URL=http://localhost:5173
```

Non-empty local placeholders allow server startup, but Google OAuth and Stripe operations require real provider configuration.

## Run in development

The Go process does not load `.env` itself. Export it first.

Terminal 1:

```bash
set -a
source .env
set +a
go run ./cmd/server
```

Terminal 2:

```bash
npm --prefix frontend run dev
```

Open `http://localhost:5173`. Verify backend health at `http://localhost:8080/healthz`.

## Database initialization

No separate migration command is required. Both server and MCP entrypoints:

1. load all 19 registered SQL migrations;
2. open or create SQLite at `DATABASE_PATH`;
3. apply pending migrations in version order;
4. enable WAL, foreign keys, `synchronous=NORMAL`, and busy timeout;
5. use one open SQLite connection.

PaperViz now supports accounts, sessions, API keys, billing, collections, and annotations. Older setup guidance claiming that accounts do not exist is obsolete.

## Build application

```bash
npm --prefix frontend run build
CGO_ENABLED=0 go build -o server ./cmd/server
```

Run it with exported environment variables:

```bash
set -a
source .env
set +a
./server
```

The Go binary serves API routes, static assets, and SPA fallback from `STATIC_DIR`.

Make shortcuts:

```bash
make build
make run
```

## MCP setup caveat

The repository ships `cmd/mcp` as a local stdio server. It requires `GEMINI_API_KEY` and `PAPERVIZ_API_KEY`, and can share the web database through an identical absolute `DATABASE_PATH`.

Current MCP limitations:

- `ingest_document` stores text but does not start the full processing pipeline;
- `/agents` generates remote configuration for `/api/mcp`, but no such route exists in the current HTTP router;
- MCP tools are stateless and global, not scoped to a REST user's library.

See [`docs/getting-started.md`](docs/getting-started.md#7-run-mcp-locally) for local MCP details.

## Verification

```bash
gofmt -l .
go vet ./...
go build ./...
go build ./cmd/mcp
go test ./...
npm --prefix frontend run lint
npm --prefix frontend run build
```

`gofmt -l .` must print nothing. See [`docs/maintainer-guide.md`](docs/maintainer-guide.md) for test requirements and failure handling.

## Project structure

Use [`docs/codebase-reference.md`](docs/codebase-reference.md) for the maintained repository, route, layer, and command map. Older trees in plans or task notes are historical and can list removed files.

## More documentation

- Product and architecture: [`docs/overview.md`](docs/overview.md)
- Architecture constraints: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- Pipeline behavior: [`docs/DATA_PIPELINE.md`](docs/DATA_PIPELINE.md)
- Current state: [`docs/PROJECT_STATE.md`](docs/PROJECT_STATE.md)
- Documentation index: [`docs/README.md`](docs/README.md)
