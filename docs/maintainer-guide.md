# Maintainer Guide

This guide provides task-oriented procedures for changing PaperViz. It assumes you have completed [Getting Started](getting-started.md).

## Before any change

1. Read [`../AGENTS.md`](../AGENTS.md).
2. Read the focused contract for the subsystem:
   - architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md);
   - pipeline: [`DATA_PIPELINE.md`](DATA_PIPELINE.md);
   - API: [`structured-research-api.md`](structured-research-api.md) and [`openapi.yaml`](openapi.yaml);
   - MCP: [`mcp-parity.md`](mcp-parity.md);
   - UI: [`../DESIGN.md`](../DESIGN.md).
3. Query current code before trusting old plans or documentation.
4. Define success and failure behavior before editing.
5. Keep changes within existing layer boundaries.

## Architecture invariants

```text
Frontend → API client → handlers → app/domain services → repository
MCP tools → app/domain services and repository
Services → external providers and infrastructure
```

Do not introduce:

- an ORM;
- a job queue or message broker;
- a microservice split;
- an LLM gateway;
- hidden Gemini calls inside MCP;
- user-owned list/save/delete/share operations in MCP without an explicit product decision;
- disk persistence for uploaded PDF bytes.

Handlers parse and serialize transport. Services own business rules. Repositories own SQL. MCP tools adapt shared data and domain behavior to tool schemas.

## Choose the correct change path

### Change REST behavior

1. Put shared business behavior in `internal/app/documents/` or `internal/services/`.
2. Keep `internal/handlers/` responsible for request parsing, auth context, status mapping, and response encoding.
3. Register or adjust routes in `internal/handlers/router.go`.
4. Add handler and service tests.
5. Update `docs/openapi.yaml` and `docs/structured-research-api.md` when the external contract changes.
6. Update route inventories in `docs/codebase-reference.md` and `docs/product/current-user-flow.md`.

### Change document processing

1. Keep orchestration in `internal/services/pipeline.go` and `pipeline_stages.go`.
2. Keep intake and transactional persistence in `internal/services/intake.go`.
3. Keep evidence extraction and grounding deterministic.
4. Let Gemini transform or explain extracted content, not create unsupported numeric values.
5. Add table-driven success and error cases for each changed service function.
6. Extend pipeline acceptance and failure regression tests where behavior changes.
7. Update [`DATA_PIPELINE.md`](DATA_PIPELINE.md).

### Change chart behavior

Chart code spans deterministic evidence handling and presentation:

```text
source text
  → numeric/table evidence
  → candidate datasets
  → Gemini chart plan
  → deterministic grounding validation
  → ChartSpec persistence
  → Recharts rendering
```

When changing this path:

- never introduce model-generated values;
- preserve evidence and dataset identifiers in provenance;
- cover missing values with real omission, not zero-fill;
- cover unsupported grounding with a non-rendering state;
- test chart-type-specific constraints;
- update data-pipeline and output-contract documentation.

### Change MCP

1. Confirm a user or product need exists. Do not add speculative tools.
2. Define whether operation is research data or user-owned preference.
3. Keep tool logic in `internal/mcp/tools.go` thin and deterministic.
4. Reuse shared services or repositories; do not import HTTP handlers.
5. Define input and output schemas in `internal/mcp/schemas.go`.
6. Define explicit errors in `internal/mcp/errors.go`.
7. Apply rate and job limits appropriate to operation cost.
8. Add contract tests in `internal/mcp/mcp_contract_test.go`.
9. Update `docs/mcp-parity.md` in the same change.
10. Run `go build ./cmd/mcp` and `go test ./...`.

Current special case: `ingest_document` does not start the full processing pipeline. Do not document it as doing so unless code and tests are changed together.

### Change frontend route or navigation

1. Read `DESIGN.md` completely.
2. Update `frontend/src/App.jsx` as the route source of truth.
3. Avoid duplicate canonical routes; use redirects only for compatibility.
4. Update API client functions in `frontend/src/lib/api.js`.
5. Preserve loading, timeout, retry, and input state.
6. Every `catch` must provide user-facing error feedback, retry behavior when retry makes sense, development `console.error`, and input preservation.
7. Update route tables in documentation.
8. Run frontend lint and build.

### Change schema

PaperViz uses a single flat migration sequence. There is no migration runner for incremental schema evolution.

1. Add the next numbered SQL file under `migrations/`.
2. Register it in `internal/repository/migrations.go`.
3. Update repository code and tests.
4. Update schema descriptions in relevant docs.
5. Stop the running server.
6. Delete local SQLite files only after confirming data can be discarded:

```bash
rm -f paperviz.db paperviz.db-wal paperviz.db-shm paperviz.db-journal
```

7. Restart the server and verify `/healthz`.
8. Exercise the changed query against the fresh schema.

Do not commit database files or real user data.

## Run tests

### Backend

```bash
gofmt -l .
go vet ./...
go build ./...
go test ./...
```

`gofmt -l .` must print nothing.

Testing requirements:

- service functions use table-driven tests;
- every changed service has at least one success and one error case;
- pipeline integration coverage preserves all acceptance and failure scenarios;
- claim-diff verification still detects the corrupted-passage regression;
- ownership tests cover annotations and collections;
- URL import tests preserve SSRF protections;
- MCP contract tests match registered tool names and schemas.

### Frontend

```bash
npm --prefix frontend run lint
npm --prefix frontend run build
```

### End to end

Start a built application and run:

```bash
BASE_URL=http://localhost:8080 npm --prefix e2e test
```

E2E tests require a real reachable application. They do not replace focused Go or frontend tests.

## Stop on validation failure

When any required check fails:

1. stop work at the failing gate;
2. report the exact command and failure;
3. identify likely cause without changing unrelated code;
4. propose a fix;
5. request approval before applying that fix;
6. rerun the failed check and relevant broader checks after approval.

Do not turn a validation failure into an unrelated cleanup task.

## Debugging procedures

### Server does not start

Check, in order:

1. required environment variables;
2. current working directory;
3. presence of all registered migration files;
4. database path permissions;
5. structured logs in `LOG_FILE` and stdout;
6. port availability.

### `/healthz` fails

`/healthz` checks both `db.Ping()` and `SELECT 1`. A failure points to database reachability or query execution, not frontend routing.

### Document stays `processing`

Check:

- processing stage in document response;
- JSONL logs for the document ID;
- whether intake came from web or MCP;
- Gemini timeout or rate-limit errors;
- whether process stayed alive for the 20-minute background timeout.

MCP-only intake currently creates a processing row without starting the pipeline, so it remains processing unless another process advances it.

### Simplified result looks wrong

Inspect:

- original text;
- generated simplified text;
- claim-diff mismatch detail;
- claim records;
- verification status and UI gating.

Do not remove mismatch details to make the UI appear successful.

### Chart is missing

Check pipeline failure category:

- `EXTRACTION_ERROR`;
- `DATASET_ERROR`;
- `CHART_SELECTION_ERROR`;
- `GROUNDING_ERROR`;
- `SCHEMA_ERROR`;
- `RENDER_ERROR`.

A missing chart can be correct when source evidence is absent or fails deterministic grounding. Do not replace missing values with zeros or ask Gemini to invent data.

### Frontend request fails

Inspect:

- browser network request;
- Vite proxy and `PORT`;
- backend route registration;
- error code returned by API;
- frontend `catch` handling and retry path.

User-visible error text may simplify server codes, but logs and tests must retain the underlying failure.

## Logging and observability

- Server uses structured `log/slog` JSONL.
- Default log file: `paperviz.log.jsonl`.
- Logs also go to stdout.
- Request IDs correlate request logs.
- Never log full document text, source text, secrets, session tokens, or API keys.
- Use document IDs and stage names for correlation.

See [`OBSERVABILITY.md`](OBSERVABILITY.md) for current coverage and gaps.

## Performance constraints

Current design intentionally uses:

- one SQLite open connection;
- WAL mode;
- synchronous `NORMAL`;
- bounded background pipeline timeout;
- bounded image-chart fallback count;
- rate limits on expensive or sensitive routes.

Before raising concurrency, benchmark SQLite behavior, Gemini quota, memory, and MCP request patterns. A single-instance low-concurrency design is not a horizontal-scaling design.

## Add or change an environment variable

1. Update `.env.example` with purpose, default, and whether required.
2. Read and validate it in the correct `cmd/*/main.go` or focused service.
3. Add tests for missing and valid values where behavior is explicit.
4. Update [`getting-started.md`](getting-started.md) and [`codebase-reference.md`](codebase-reference.md).
5. Check deployment and container configuration.
6. Never place a real secret in committed configuration.

## Change documentation only

Documentation changes still require accuracy checks:

1. Verify claims against current code and config.
2. Prefer local project sources; do not invent external behavior.
3. Check every relative path and anchor.
4. Avoid copying historical prose into current docs.
5. Search for superseded terms and stale counts.
6. Run required repository validation commands before declaring completion.

## Change checklist

Before considering work complete:

- [ ] Layer boundaries preserved
- [ ] No secrets or full document text logged
- [ ] Success and failure behavior tested
- [ ] Claim verification and chart grounding remain explicit
- [ ] Frontend errors show feedback and retry where appropriate
- [ ] `gofmt -l .` prints nothing
- [ ] `go vet ./...` passes
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] frontend lint passes
- [ ] frontend build passes
- [ ] relevant docs updated
- [ ] MCP and REST parity docs updated when applicable
- [ ] database reset performed only for schema work and disposable data
