# Chunk 9.2 — Developer SDKs: Facts

## Accepted Facts

1. **SDK Priority Order**: Python SDK first (research users primary audience), then TypeScript SDK, then CLI
2. **Auth Model**: Add API key authentication to backend (following MCP pattern: `PAPERVIZ_API_KEY` env var). SDKs use `Authorization: Bearer <api_key>` header. Browser-based TypeScript SDK can also support session cookies.
3. **Code Generation Strategy**: Generate core client from OpenAPI spec (docs/openapi.yaml) using openapi-generator/oapi-codegen, then hand-write ergonomic wrapper layer for polling, retry logic, typed errors, and convenience methods
4. **Python Packaging**: PyPI package name `paperviz` (verify availability). Dependencies: `httpx`, `pydantic`, `pydantic-settings`. Semver matching API version. MIT license. Python 3.10+
5. **TypeScript Packaging**: npm package `@paperviz/sdk` (verify availability). ESM + CJS dual package. Export TypeScript types generated from OpenAPI. Optional React hooks for polling. Peer dependency on `fetch` (global).
6. **CLI Framework**: Cobra (Go) for single binary distribution matching backend stack. Commands: `upload`, `status`, `get`, `share`, `list`, `compare`, `import-doi`, `import-url`, `collections`. Config file (`~/.paperviz.yaml`) for API key + base URL.
7. **Error Handling**: Typed exception hierarchy per SDK. Automatic retry with exponential backoff for 429 (rate limit) and 5xx (transient). Configurable retry policy (max retries, base delay, max delay). Error codes as constants/enums matching API snake_case codes.
8. **Polling Helpers**: Built-in `waitForCompletion(documentId, options?)` with configurable poll interval (default 2s), timeout (default 5min), and progress callbacks. Returns final Document or throws on failure/timeout.
9. **Testing Strategy**: Unit tests with mocked HTTP (httpx/msw). Integration tests against local dev server. Contract tests generated from OpenAPI spec. CI runs both.
10. **API Key Backend**: Add API key middleware to Go backend. Store hashed API keys in DB. Scope: read/write endpoints. Rate limits per key. Follows MCP server pattern already implemented.
11. **Versioning**: SDK versions track API version. Breaking API changes = major SDK version. Use OpenAPI `info.version` as baseline.
12. **Documentation**: Each SDK includes README with quickstart, authentication, all methods documented with examples. Generate reference docs from code (pdoc/sphinx for Python, typedoc for TypeScript).

## Verification Commands

- Python: `pip install -e . && python -c "import paperviz; print(paperviz.__version__)"`
- TypeScript: `npm pack && npm install ./paperviz-sdk-*.tgz && node -e "const s = require('@paperviz/sdk'); console.log(s.version)"`
- CLI: `go build -o paperviz-cli ./cmd/cli && ./paperviz-cli --help`
- All: Run integration tests against local server