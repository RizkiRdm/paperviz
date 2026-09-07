# Chunk 9.2 — Developer SDKs: Execution Plan

## Solution Approach

Build three SDKs wrapping the stable PaperViz REST API (OpenAPI 3.1.0 at `docs/openapi.yaml`):
1. **Python SDK** (`paperviz`) — Primary for research users. Generated core + hand-written ergonomic layer.
2. **TypeScript SDK** (`@paperviz/sdk`) — For web integrations. Generated types + hand-written client with React hooks.
3. **CLI** (`paperviz` Go binary) — Cobra-based, single binary, uses Python SDK or direct HTTP.

**Prerequisite**: Add API key authentication to Go backend (following existing MCP pattern in `internal/mcp/server.go`).

---

## Ordered Steps

### Phase 0: Backend — API Key Authentication (Prerequisite)

| Step | Files Touched | Verification |
|------|---------------|--------------|
| 0.1 Add `api_keys` table migration | `migrations/016_api_keys.sql` | `go run ./cmd/server` creates table |
| 0.2 API key model & repo | `internal/repository/api_keys.go` | Unit tests pass |
| 0.3 API key service (create, validate, hash) | `internal/services/api_keys.go` | Unit tests pass |
| 0.4 API key middleware (Bearer token) | `internal/handlers/api_key_auth.go` | Integration test: `curl -H "Authorization: Bearer $KEY" /api/documents` |
| 0.5 Wire middleware in router | `internal/handlers/router.go` | All protected endpoints accept API key |
| 0.6 Add API key to MCP server pattern reference | `internal/mcp/server.go` (review) | Consistency check |

### Phase 1: Python SDK (`paperviz`)

| Step | Files Touched | Verification |
|------|---------------|--------------|
| 1.1 Generate core client from OpenAPI | `sdks/python/paperviz/_generated/` | `openapi-generator generate -i ../../docs/openapi.yaml -g python -o .` |
| 1.2 Create wrapper package structure | `sdks/python/paperviz/__init__.py`, `client.py`, `models.py`, `errors.py`, `polling.py` | Import works |
| 1.3 Implement `PapervizClient` with: auth, retry, timeout config | `sdks/python/paperviz/client.py` | Unit tests mock HTTP |
| 1.4 Implement typed exceptions (`PapervizError`, `RateLimitError`, `AuthError`, `NotFoundError`, `ValidationError`) | `sdks/python/paperviz/errors.py` | Error mapping tests |
| 1.5 Implement `wait_for_completion()` with callbacks | `sdks/python/paperviz/polling.py` | Integration test against local server |
| 1.6 Convenience methods: `upload_pdf()`, `upload_text()`, `import_doi()`, `import_url()`, `compare()`, `share()`, `list_collections()` | `sdks/python/paperviz/client.py` | Integration tests |
| 1.7 Packaging: `pyproject.toml`, `README.md`, `LICENSE` | `sdks/python/pyproject.toml`, etc. | `pip install -e .` works |
| 1.8 CI: GitHub Actions for test + publish to PyPI | `.github/workflows/python-sdk.yml` | CI passes |

### Phase 2: TypeScript SDK (`@paperviz/sdk`)

| Step | Files Touched | Verification |
|------|---------------|--------------|
| 2.1 Generate TypeScript types from OpenAPI | `sdks/typescript/src/generated/` | `openapi-typescript ../../docs/openapi.yaml -o src/generated/api.ts` |
| 2.2 Generate core client (or hand-write minimal fetch wrapper) | `sdks/typescript/src/client.ts` | Types compile |
| 2.3 Implement `PapervizClient` class: auth (Bearer + cookie), retry, timeout | `sdks/typescript/src/client.ts` | Unit tests with MSW |
| 2.4 Implement typed errors (`PapervizError`, `RateLimitError`, etc.) | `sdks/typescript/src/errors.ts` | Error mapping tests |
| 2.5 Implement `waitForCompletion()` with progress callbacks | `sdks/typescript/src/polling.ts` | Integration test |
| 2.6 Convenience methods mirroring Python SDK | `sdks/typescript/src/client.ts` | Integration tests |
| 2.7 React hooks: `useDocumentPolling()`, `useUpload()` | `sdks/typescript/src/hooks.ts` | Component test |
| 2.8 Packaging: `package.json`, `tsconfig.json`, `README.md` | `sdks/typescript/package.json` | `npm pack` works |
| 2.9 CI: GitHub Actions for test + publish to npm | `.github/workflows/typescript-sdk.yml` | CI passes |

### Phase 3: CLI (Go, Cobra)

| Step | Files Touched | Verification |
|------|---------------|--------------|
| 3.1 Init Cobra project | `cmd/cli/main.go`, `cmd/cli/cmd/*.go` | `go run ./cmd/cli --help` |
| 3.2 Config: `~/.paperviz.yaml` (api_key, base_url, tier) | `cmd/cli/cmd/root.go`, `internal/cli/config.go` | Config loads |
| 3.3 Command: `upload` (PDF or text, --reading-level) | `cmd/cli/cmd/upload.go` | Uploads file, prints document_id |
| 3.4 Command: `status` (poll until complete) | `cmd/cli/cmd/status.go` | Polls and shows progress |
| 3.5 Command: `get` (fetch full document JSON) | `cmd/cli/cmd/get.go` | Outputs JSON |
| 3.6 Command: `share` (generate share link) | `cmd/cli/cmd/share.go` | Prints share URL |
| 3.7 Command: `list` (list documents with filters) | `cmd/cli/cmd/list.go` | Table output |
| 3.8 Command: `compare` (multi-paper comparison) | `cmd/cli/cmd/compare.go` | Outputs comparison JSON |
| 3.9 Command: `import-doi` / `import-url` | `cmd/cli/cmd/import.go` | Imports paper |
| 3.10 Command: `collections` (CRUD) | `cmd/cli/cmd/collections.go` | Full CRUD |
| 3.11 Build: `make cli` → `paperviz` binary | `Makefile` | Binary runs on Linux/macOS/Windows |
| 3.12 CI: Goreleaser for multi-platform releases | `.goreleaser.yaml`, `.github/workflows/cli.yml` | Release artifacts |

### Phase 4: Documentation & Examples

| Step | Files Touched | Verification |
|------|---------------|--------------|
| 4.1 Python SDK docs (pdoc) | `sdks/python/docs/` | `pdoc -o docs paperviz` |
| 4.2 TypeScript SDK docs (typedoc) | `sdks/typescript/docs/` | `typedoc src/` |
| 4.3 CLI docs (Cobra auto-generates) | `cmd/cli/cmd/*.go` | `paperviz --help` |
| 4.4 Example scripts: `examples/python/analyze_paper.py`, `examples/typescript/upload.ts` | `examples/` | Examples run |

---

## Verification Checklist

- [ ] Backend API key auth works (unit + integration tests)
- [ ] Python SDK: `pip install -e .` → import → upload → poll → get results
- [ ] Python SDK: retry on 429 works, typed errors raised
- [ ] TypeScript SDK: `npm install` → import → upload → poll → get results
- [ ] TypeScript SDK: React hooks work in test component
- [ ] CLI: `paperviz upload paper.pdf` → `paperviz status <id>` → `paperviz get <id>`
- [ ] All SDKs: version matches API version in OpenAPI spec
- [ ] CI passes for all three SDKs
- [ ] No breaking changes to existing API

---

## Risks & Open Questions

| Risk | Mitigation |
|------|------------|
| PyPI/npm name taken | Check availability early; fallback: `paperviz-sdk`, `paperviz-client` |
| OpenAPI codegen produces poor ergonomics | Hand-written wrapper layer is mandatory; generated code is internal only |
| API key auth not ready in time | Can start SDKs with cookie auth for browser TypeScript; add Bearer later |
| CLI binary size | Go binary ~10MB acceptable; UPX if needed |
| Rate limit handling differs per SDK | Shared retry policy spec in `docs/sdk-retry-policy.md` |
| Session cookie vs API key for TypeScript | Support both: `PapervizClient({apiKey})` for Node, `PapervizClient({useCookies: true})` for browser |

---

## Definition of Done

- [ ] Backend API key auth merged and deployed
- [ ] Python SDK published to PyPI (or test PyPI)
- [ ] TypeScript SDK published to npm (or test npm)
- [ ] CLI binary released via GitHub Releases
- [ ] All integration tests pass against staging
- [ ] Documentation published for all three
- [ ] Example scripts work end-to-end