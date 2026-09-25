# PaperViz Documentation

Documentation for people and AI agents building, operating, or reasoning about PaperViz.

## Start Here

| If you want to... | Read |
|---|---|
| Get a 10-minute product overview | [Overview](overview.md) |
| Run PaperViz locally | [Getting Started](getting-started.md) |
| Understand architecture constraints | [Architecture](ARCHITECTURE.md) |
| Understand paper processing | [Data Pipeline](DATA_PIPELINE.md) |
| Make a code change | [Maintainer Guide](maintainer-guide.md) |
| Find a file, route, setting, or command | [Codebase Reference](codebase-reference.md) |
| Understand current work and known gaps | [Project State](PROJECT_STATE.md) |
| Use or evaluate REST | [Structured Research API](structured-research-api.md) and [OpenAPI](openapi.yaml) |
| Understand MCP scope | [MCP ↔ REST Parity](mcp-parity.md) |

## Diátaxis Reading Paths

### Tutorial: learn by doing

[Getting Started](getting-started.md) guides a new maintainer from clean checkout to a working web application, first paper analysis, health check, and local MCP process.

### Explanation: build a mental model

[Overview](overview.md) explains the product problem, users, agent-first direction, web fallback, trust model, processing stages, and architectural boundaries.

### How-to guides: accomplish a task

[Maintainer Guide](maintainer-guide.md) provides recipes for running tests, debugging failures, changing layers, modifying schema, updating MCP, and keeping docs synchronized.

### Reference: look up facts

[Codebase Reference](codebase-reference.md) describes files, routes, tools, dependencies, environment variables, commands, and ownership boundaries.

## Source-of-Truth Hierarchy

When documents disagree, use this order:

1. **Current source code and runtime configuration** define implemented behavior.
2. **Domain contracts** define allowed architecture and product boundaries.
3. **Focused references** define one subsystem in detail.
4. **Project state and plans** describe current or historical intent; they can lag code.
5. **Archived reports and task notes** provide context only, not present-tense guarantees.

Key contracts:

| Concern | Source of truth |
|---|---|
| Implemented routes and tools | `internal/handlers/router.go`, `frontend/src/App.jsx`, `internal/mcp/tools.go` |
| Runtime requirements | `go.mod`, `frontend/package.json`, `.env.example`, `Makefile` |
| Architecture boundaries | [`AGENTS.md`](../AGENTS.md) and [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) |
| Product direction | [`PRODUCT.md`](../PRODUCT.md) and [`docs/PRD.md`](PRD.md) |
| User flow | [`docs/product/current-user-flow.md`](product/current-user-flow.md) |
| UI behavior and tokens | [`DESIGN.md`](../DESIGN.md) and current frontend code |
| Processing semantics | [`docs/DATA_PIPELINE.md`](DATA_PIPELINE.md) and service code |
| REST contract | [`docs/openapi.yaml`](openapi.yaml) |
| MCP scope | [`docs/mcp-parity.md`](mcp-parity.md) and `internal/mcp/tools.go` |
| Security controls | [`docs/SECURITY.md`](SECURITY.md) and handler middleware |
| Current engineering state | [`docs/PROJECT_STATE.md`](PROJECT_STATE.md) |

## Architecture Boundaries

Use these dependency directions when changing code:

```text
Frontend UI → API client → REST handlers → app/domain services → repository
MCP client → MCP tools → app/domain services and repository
Domain services → external providers and infrastructure
```

Non-negotiable constraints:

- No ORM.
- No job queue or message broker.
- No microservice split.
- No Gemini gateway; use direct Gemini API integration.
- Do not persist uploaded PDF bytes to disk.
- Keep business logic out of HTTP and MCP adapters.
- Do not expose user-owned or destructive operations through MCP without an explicit product decision.
- Do not add new MCP tools or marketing routes without demonstrated demand.

## Documentation Categories

### Stable orientation

- [`docs/overview.md`](overview.md)
- [`docs/getting-started.md`](getting-started.md)
- [`docs/maintainer-guide.md`](maintainer-guide.md)
- [`docs/codebase-reference.md`](codebase-reference.md)

### Architecture and product

- [`PRODUCT.md`](../PRODUCT.md)
- [`docs/PRD.md`](PRD.md)
- [`docs/ARCHITECTURE.md`](ARCHITECTURE.md)
- [`docs/DATA_PIPELINE.md`](DATA_PIPELINE.md)
- [`docs/product/current-user-flow.md`](product/current-user-flow.md)
- [`docs/canonical-research-output-contract.md`](canonical-research-output-contract.md)

### Interfaces

- [`docs/structured-research-api.md`](structured-research-api.md)
- [`docs/openapi.yaml`](openapi.yaml)
- [`docs/mcp-parity.md`](mcp-parity.md)

### Operations

- [`docs/SECURITY.md`](SECURITY.md)
- [`docs/RELIABILITY.md`](RELIABILITY.md)
- [`docs/OBSERVABILITY.md`](OBSERVABILITY.md)

### Dynamic project records

- [`docs/PROJECT_STATE.md`](PROJECT_STATE.md)
- [`docs/PLAN.md`](PLAN.md)
- [`docs/ENGINEERING_LOG.md`](ENGINEERING_LOG.md)
- [`docs/decisions.md`](decisions.md)
- [`docs/DECISIONS/`](DECISIONS/)

### Historical material

- [`docs/Task Agent/`](Task%20Agent/)
- [`report/`](../report/)
- Goal files under `goals/`

Archived `docs/archive/` files were removed after their content was superseded by `docs/PROJECT_STATE.md` and `docs/PRD.md`.


Historical material can explain why a decision existed. It must not be cited as proof that current code still behaves that way.

## Instructions for AI Agents

Before changing PaperViz:

1. Read [`AGENTS.md`](../AGENTS.md).
2. Read the focused document for the subsystem being changed.
3. Query current code before relying on route lists, tool lists, file locations, or status claims.
4. Preserve documented layer boundaries.
5. Update affected documentation in the same change.
6. Run required tests and report failures without silently changing unrelated code.

Useful verification points:

- REST routes: `internal/handlers/router.go`
- Frontend routes: `frontend/src/App.jsx`
- MCP tools: `internal/mcp/tools.go`
- Environment variables: `.env.example` and both `cmd/*/main.go` entrypoints
- Migration inventory: `internal/repository/migrations.go`
- Architecture state: `docs/PROJECT_STATE.md`

## Documentation Maintenance Rule

When behavior changes, update documentation at the same time:

| Change | Update |
|---|---|
| Product direction or audience | `PRODUCT.md`, `docs/PRD.md`, `docs/overview.md` |
| Route or navigation | `README.md`, `docs/codebase-reference.md`, `docs/product/current-user-flow.md` |
| MCP tool or boundary | `docs/mcp-parity.md`, `docs/codebase-reference.md`, MCP code |
| Environment variable or command | `.env.example`, `docs/getting-started.md`, `docs/codebase-reference.md` |
| Architecture boundary | `AGENTS.md`, `docs/ARCHITECTURE.md`, affected guides |
| Pipeline or data semantics | `docs/DATA_PIPELINE.md`, pipeline code and tests |
| UI behavior or tokens | `DESIGN.md` and frontend code |
| Current milestone or known gap | `docs/PROJECT_STATE.md` |
