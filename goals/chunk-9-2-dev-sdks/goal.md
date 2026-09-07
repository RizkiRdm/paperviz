# Goal: Chunk 9.2 — Developer SDKs

## Articulated Goal

Build three Developer SDKs wrapping the stable PaperViz REST API:
1. **Python SDK** (`paperviz` on PyPI) — Primary for research users
2. **TypeScript SDK** (`@paperviz/sdk` on npm) — For web integrations
3. **CLI** (`paperviz` Go binary via GitHub Releases) — For power users and automation

All SDKs wrap the existing OpenAPI 3.1.0 spec (`docs/openapi.yaml`, 34 endpoints, 31 schemas) without introducing separate product semantics.

## Shared Understanding

See `facts.md` for 12 accepted facts covering:
- SDK priority order (Python → TypeScript → CLI)
- API key authentication model (Bearer token, following MCP pattern)
- Code generation strategy (OpenAPI codegen + hand-written ergonomic wrapper)
- Packaging details for each SDK
- Error handling, polling helpers, testing strategy

## Execution Plan

See `plan.md` for 4-phase ordered plan:
- **Phase 0**: Backend API key authentication (prerequisite)
- **Phase 1**: Python SDK (generated core + wrapper)
- **Phase 2**: TypeScript SDK (generated types + wrapper + React hooks)
- **Phase 3**: CLI (Cobra, Go single binary)
- **Phase 4**: Documentation & examples

Each step lists files touched and verification commands.

## Done Condition

- Backend API key auth merged and working
- Python SDK installable via `pip install paperviz` with full API coverage
- TypeScript SDK installable via `npm install @paperviz/sdk` with full API coverage + React hooks
- CLI binary downloadable from GitHub Releases, `paperviz upload paper.pdf` works end-to-end
- All integration tests pass against local dev server
- Documentation published for all three SDKs
- No breaking changes to existing API or frontend