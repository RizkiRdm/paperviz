# Chunk 7.4 — MCP Server

**Goal:** Make PaperViz usable from AI-agent workflows via Model Context Protocol.

## Files to Create
- `internal/mcp/server.go` — MCP server setup and tool registration
- `internal/mcp/tools.go` — Tool definitions and handlers
- `cmd/mcp/main.go` — Entry point for MCP server (stdio transport)

## Dependencies
- `github.com/modelcontextprotocol/go-sdk/mcp` — Official Go SDK

## Tools to Expose (locked to 5 — Chunk 11 MCP boundary)
1. `ingest_document` — deterministic text ingestion (replaces `analyze_paper`; no LLM, no pipeline)
2. `search_documents` — global title search, stateless
3. `get_document` — selective retrieval via `include` (replaces `get_summary`)
4. `get_evidence` — evidence + claims with provenance (merges `get_claims`)
5. `get_figures` — structured chart data + provenance

Legacy names `analyze_paper`, `get_summary`, `get_claims`, `compare_papers` are deleted; compare stays in-app only.

### Tools Not Implemented (by design)
- `get_tables` — No product reason yet; tables extracted on-demand per paper
- `search_papers` — No per-user library exposed to MCP; stateless by design

## Architecture
- MCP server uses stdio transport (standard for MCP)
- Tools call existing service functions directly (no HTTP roundtrip)
- Each tool returns concise, machine-readable JSON with provenance
- Server reuses existing DB connection and service layer

## Constraints
- MCP is an interface, not a second application architecture
- Do not expose dozens of low-level CRUD tools
- Preserve research integrity and source attribution
- No generic chatbot features

## Verification
- `go build ./cmd/mcp` succeeds
- MCP Inspector can connect and list tools
- Tool calls return valid JSON matching canonical contract
