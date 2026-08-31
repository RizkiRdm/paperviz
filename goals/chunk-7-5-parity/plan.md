# Chunk 7.5 — Human/Agent Capability Parity

## Goal
Prevent the web product and MCP from becoming two unrelated products by mapping MCP operations to core capabilities, verifying research semantics preservation, and documenting intentional unavailability.

## Parity Analysis

### MCP → REST Mapping

| MCP Tool | REST Endpoint | Semantic Match | Gaps |
|----------|---------------|----------------|------|
| `analyze_paper` | `POST /api/documents` | ⚠️ Partial | MCP=text only; REST supports PDF upload |
| `get_summary` | `GET /api/documents/:id` | ✅ Full | — |
| `get_figures` | `GET /api/documents/:id` | ⚠️ Partial | MCP missing `image_url`, `chart_type`, `source_text` |
| `get_claims` | `GET /api/documents/:id` | ✅ Full | — |
| `get_evidence` | `GET /api/documents/:id` | ✅ Full | — |
| `compare_papers` | `POST /api/documents/compare` | ✅ Full | — |

### Operations Intentionally Unavailable to Agents

| REST Endpoint | Why Unavailable | Rationale |
|---------------|-----------------|-----------|
| `GET /api/documents/` (list) | No user context in MCP | MCP is stateless; no session/auth to scope listing |
| `GET /api/documents/stats` | No user context | Same as above |
| `PUT /api/documents/:id/save` | User preference | Agents don't manage user libraries |
| `PATCH /api/documents/:id` | User preference | Agents don't rename user documents |
| `DELETE /api/documents/:id` | Destructive | Agents shouldn't delete user data |
| `POST /api/documents/:id/share` | User preference | Share links are user-initiated |
| `DELETE /api/documents/:id/share` | User preference | Same |
| `PATCH /api/documents/:id/visibility` | User preference | Visibility is user-controlled |
| `POST /share-referrals` | Analytics | Referral tracking is product-level |

### Parity Gaps to Fix

1. **MCP `get_figures` missing fields**: Add `chart_type`, `source_text`, `image_url` (base64 for MCP)
2. **MCP `analyze_paper` text-only**: Document as intentional (agents paste text, not upload PDFs)
3. **MCP response schemas vs REST**: Ensure same research semantics (provenance, uncertainty, source references)

## Implementation Steps

### Step 1: Enhance MCP `get_figures` response
- File: `internal/mcp/tools.go`
- Add `chart_type` field to `ChartInfo` struct
- Add `source_text` field to `ChartInfo` struct
- Add `image_url` field (base64-encoded for MCP transport)
- Verify: `go test ./internal/mcp/...`

### Step 2: Add parity documentation
- File: `docs/mcp-parity.md`
- Document MCP ↔ REST mapping
- Document intentionally unavailable operations with rationale
- Document research semantics preservation

### Step 3: Verify research semantics across interfaces
- Check that provenance fields (page, figure_id, table_id, section, source_reference) are present in both MCP and REST responses
- Check that uncertainty/mismatch information is exposed consistently

### Step 4: Update OpenAPI spec
- File: `docs/openapi.yaml`
- Ensure MCP tool descriptions reference REST equivalents
- Add note about agent-specific limitations

## Verification

1. `go test ./internal/mcp/...` — all tests pass
2. `go test ./...` — full suite passes
3. `go vet ./...` — no issues
4. Manual: MCP tool responses include all research semantics present in REST
5. Documentation: `docs/mcp-parity.md` exists and covers all mappings

## Risks

- Adding image_url as base64 to MCP could increase response size significantly
- Mitigation: Only include when image exists; document size implications
