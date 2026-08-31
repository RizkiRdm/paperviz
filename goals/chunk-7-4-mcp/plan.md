# Chunk 7.4 — MCP Server

**Goal:** Make PaperViz usable from AI-agent workflows via Model Context Protocol.

## Files to Create
- `internal/mcp/server.go` — MCP server setup and tool registration
- `internal/mcp/tools.go` — Tool definitions and handlers
- `cmd/mcp/main.go` — Entry point for MCP server (stdio transport)

## Dependencies
- `github.com/modelcontextprotocol/go-sdk/mcp` — Official Go SDK

## Tools to Expose (from roadmap)
1. `analyze_paper` — Upload and analyze a paper (wraps POST /api/documents)
2. `get_summary` — Get simplified text for a paper
3. `get_figures` — Get charts/figures for a paper
4. `get_tables` — Get tables (placeholder, not yet implemented)
5. `get_claims` — Get claim verification data
6. `get_evidence` — Get evidence references
7. `compare_papers` — Compare 2-10 papers
8. `search_papers` — Search papers (only if capability exists)

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
