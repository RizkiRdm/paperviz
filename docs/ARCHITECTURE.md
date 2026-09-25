# PaperViz Architecture

> Canonical architecture contract. Current source snapshot verified against local code on 2026-09-25. Run repository validation commands before treating this document as a release gate.
>
> The lowercase [`architecture.md`](architecture.md) file is a compatibility pointer, not a second source of truth.

## 1. System Context

PaperViz is a modular monolith for transforming academic papers into simplified explanations, structured claims, evidence, and grounded figures.

It has two runtime targets:

1. **HTTP server** — React application, REST API, authentication, billing, manual ingestion, research management, and static asset serving.
2. **MCP server** — local stdio tools for AI-agent data access.

Both runtimes can share one SQLite database.

Current product direction is agent-first, but the web ingestion path remains the only path that starts the complete Gemini-backed processing pipeline.

## 2. Architecture Style

- Go 1.25 modular monolith
- React 19 single-page application
- One production HTTP process
- Separate local MCP process
- One SQLite database file
- No ORM
- No job queue or message broker
- No microservice decomposition
- No uploaded PDF persistence
- Direct Gemini API integration

The architecture optimizes for a solo-maintainer project and low-concurrency deployment. It is not designed for horizontal scaling.

## 3. Dependency Direction

Target dependency graph:

```text
cmd/server
  → internal/handlers
    → internal/app/documents
      → internal/services
        → internal/repository
        → internal/external

cmd/mcp
  → internal/mcp
    → internal/services and/or internal/repository

cmd/*
  → internal/repository
```

Rules:

- handlers parse HTTP input, map auth context, call application/domain behavior, and serialize responses;
- application/domain services own use cases and business rules;
- repositories own raw SQL and transaction mechanics;
- external packages wrap Gemini, PDF, DOI, URL, and logging providers;
- MCP tools adapt shared domain/data behavior to MCP schemas;
- repositories must not import handlers, MCP, or services;
- MCP must not import HTTP handlers;
- MCP must not contain hidden Gemini calls or independent reasoning logic.

### Current migration state

`internal/app/documents` now owns document creation and read operations, and create handling has been split from document reads. Some remaining HTTP handlers still instantiate repositories directly for specialized research objects. This is a known refactor gap. New document behavior should use the application/service boundary rather than extending handler-to-repository coupling.

## 4. HTTP Runtime

`cmd/server/main.go` performs startup orchestration:

1. open JSONL log file;
2. configure structured `slog` output to stdout and file;
3. require Gemini, Google OAuth, Stripe, and frontend URL environment variables;
4. resolve Gemini model, database path, static directory, and port;
5. load all registered migrations;
6. open SQLite and apply pending migrations;
7. delete expired sessions;
8. construct Gemini client;
9. start document expiry sweep;
10. start HTTP router.

The server does not load `.env` itself. Environment export happens outside the process.

### Middleware and routing

`internal/handlers/router.go` configures:

- request IDs;
- panic recovery;
- security headers;
- structured request logging;
- route-level auth, rate, and usage middleware;
- public share responses with `noindex` headers;
- static file serving;
- API-specific JSON 404;
- SPA fallback for GET/HEAD deep links.

Current route inventory is maintained in [`codebase-reference.md`](codebase-reference.md). Machine-readable HTTP contract lives in [`openapi.yaml`](openapi.yaml).

## 5. MCP Runtime

`cmd/mcp/main.go`:

1. requires `GEMINI_API_KEY`;
2. requires `PAPERVIZ_API_KEY`;
3. resolves model, database path, and migration directory;
4. opens the same migrated SQLite model as HTTP;
5. constructs MCP server;
6. runs stdio transport.

The current MCP surface contains five tools:

- `ingest_document`;
- `search_documents`;
- `get_document`;
- `get_figures`;
- `get_evidence`.

MCP tools are deterministic. They do not call Gemini. `ingest_document` validates and inserts text only; it does not call `RunPipelineAndPersist`. No queue currently advances MCP-created documents.

`docs/mcp-parity.md` defines capability scope and must stay synchronized with `internal/mcp/tools.go`.

### Remote transport mismatch

`frontend/src/pages/agents-page.jsx` generates configuration for `https://paperviz.com/api/mcp` through `mcp-remote`. The HTTP router does not register `/api/mcp`. The repository's implemented agent transport is local stdio. Treat the `/agents` page as unverified for production remote MCP use until a compatible transport endpoint exists and is tested.

## 6. Document Processing

Web document creation validates intake, inserts a document, and starts `RunPipelineAndPersist` in a goroutine. The background pipeline has a 20-minute context timeout.

```text
intake
  → simplify with Gemini
  → verify original and simplified claims
  → detect chapters
  → extract numeric and table evidence
  → build candidate datasets
  → plan figures
  → validate grounding
  → persist result transaction
```

Key contracts:

- uploaded PDF bytes are passed in memory and are not persisted;
- original text and derived research objects are persisted;
- verification mismatch detail remains visible;
- Gemini does not generate chart values;
- deterministic grounding controls unsupported chart state;
- image fallback is bounded to five embedded charts per document;
- one chart failure must not abort unrelated figures;
- frontend polls every two seconds and exposes progress, long-running, timeout, and retry states.

Detailed processing semantics live in [`DATA_PIPELINE.md`](DATA_PIPELINE.md).

## 7. Persistence Architecture

### Migrations

PaperViz uses 19 ordered SQL files under `migrations/`. `internal/repository/migrations.go` is the single registration point used by both server and MCP.

Migration policy:

- add a new numbered SQL file;
- register it in `migrationPaths`;
- update repository code and tests;
- delete disposable local SQLite files;
- restart against a fresh schema when required by current project policy.

There is no incremental migration runner or migration rollback system.

### SQLite

`internal/repository/db.go` configures:

- one open connection;
- WAL journal mode;
- `synchronous=NORMAL`;
- foreign keys;
- five-second busy timeout;
- `schema_migrations` tracking;
- ordered pending migration application.

One connection favors low-concurrency correctness. It limits read concurrency and agent-scale throughput.

### Data domains

Current migrations and repositories cover:

- documents, titles, saved state, and access time;
- users, sessions, OAuth identity, API keys, and billing state;
- chapters;
- charts, chart data, image blobs, and annotations;
- claim diffs, claims, evidence, and claim-evidence links;
- tables, methods, results, citations, and paper relationships;
- collections;
- document and figure sharing;
- referrals, usage, and analytics.

Exact migration inventory is in [`codebase-reference.md`](codebase-reference.md).

## 8. Trust and Provenance Boundaries

PaperViz separates two trust mechanisms.

### Claim verification

Claim verification compares original and simplified claims. Mismatch state and detail are persisted. `verification_failed` is evidence, not presentation polish.

### Figure grounding

Grounding uses deterministic evidence and validation:

```text
source text
  → NumericEvidence/TableData
  → CandidateDataset
  → Gemini chart plan
  → deterministic ValidateGrounding
  → ChartSpec with provenance
```

Rules prohibit synthetic numeric values. Missing values remain missing rather than becoming zero. Unsupported charts do not render as verified data.

## 9. Interface Boundaries

### REST

REST owns browser-facing and integration-facing operations, including authentication, ingestion, research-object retrieval, user-owned management, sharing, analytics, and billing.

### MCP

MCP owns deterministic agent-facing data operations. It is additive to REST. User library and destructive operations remain outside MCP unless a deliberate product decision changes that boundary.

### Frontend

The React app calls REST through `frontend/src/lib/api.js`. Route-level components compose hooks and components. UI design values come from [`../DESIGN.md`](../DESIGN.md).

## 10. Security Architecture

Implemented controls include:

- environment-only secrets;
- startup fail-loud validation for required integrations;
- secure, HTTP-only, same-site session cookies;
- Google OAuth state nonce validation;
- password complexity enforcement;
- auth and document-ingestion rate limits;
- 20 MiB PDF limit and PDF content validation;
- HTTPS URL import and private-network SSRF blocking;
- ownership checks for annotations and collections;
- Stripe webhook verification boundary;
- API-key requirement for MCP process;
- per-key MCP rate and job limits;
- security headers and panic recovery;
- no full document text in logs.

Focused control documentation lives in [`SECURITY.md`](SECURITY.md) and [`RELIABILITY.md`](RELIABILITY.md).

## 11. Observability

- structured JSONL logging through `log/slog`;
- stdout and `LOG_FILE` output;
- request IDs;
- request method, path, status, and duration;
- document IDs and pipeline stages in relevant logs;
- database-backed `/healthz` endpoint.

Do not log original text, simplified text, source passages, credentials, cookies, session tokens, or API keys.

See [`OBSERVABILITY.md`](OBSERVABILITY.md).

## 12. Concurrency and Scaling

Current constraints:

- one SQLite connection;
- process-local Gemini concurrency control;
- request-scoped background goroutines for web document processing;
- process-local MCP rate-limit and job state;
- local SQLite file as coordination point.

Before increasing traffic:

1. benchmark SQLite reads and writes;
2. measure Gemini quota and latency;
3. profile memory for concurrent document jobs;
4. define whether one process or multiple processes may process documents;
5. add a queue only through an explicit architecture decision, not as an incidental refactor;
6. preserve no-PDF-disk-write and no-microservice constraints until those non-goals are deliberately changed.

## 13. Non-Goals

The following remain prohibited unless product and architecture owners explicitly change them:

- ORM introduction;
- job queue or message broker introduction;
- microservice split;
- Gemini gateway;
- OCR for scanned PDFs;
- uploaded PDF persistence to disk;
- speculative MCP tools;
- duplicate marketing routes;
- user-owned destructive operations through MCP;
- silent verification failure;
- model-generated chart values.

## 14. Current Architecture Gaps

| Gap | Current evidence | Architectural impact |
|---|---|---|
| MCP ingestion does not process documents | `internal/mcp/tools.go` calls `ValidateAndInsert` only | Agent cannot obtain simplified output from MCP-only ingestion |
| Remote MCP config lacks route | `/agents` page versus `internal/handlers/router.go` | Generated agent configuration cannot reach current server |
| Handler/repository coupling remains | specialized handlers instantiate repositories | Business and data access boundaries are not fully isolated |
| Single SQLite connection | `internal/repository/db.go` | Limits concurrent reads and agent-scale traffic |
| Process-local controls | Gemini, auth rate, and MCP limit state | Multiple instances would not share runtime state |
| Static SEO artifacts lag route cuts | `frontend/public/` and `assets/` | Can misrepresent current product navigation |

## 15. Architecture Change Checklist

Before changing architecture:

- update this document;
- update [`../AGENTS.md`](../AGENTS.md) if an invariant changes;
- update focused subsystem documentation;
- add tests for new boundaries and failure paths;
- run `gofmt -l .`, `go vet ./...`, `go build ./...`, and `go test ./...`;
- build and lint frontend when interface behavior changes;
- keep REST and MCP parity documentation synchronized;
- do not claim a gap is closed until code and validation evidence agree.
