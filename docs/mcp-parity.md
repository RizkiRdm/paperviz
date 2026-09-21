# MCP ↔ REST API Parity Map

**Chunk 7.5 — Human/Agent Capability Parity (updated Chunk 12.5 — 2026-09-21, matches 5 actual tools in `internal/mcp/tools.go`)**

## MCP Tools → REST Endpoints

| MCP Tool | REST Endpoint | Parity Status | Notes |
|----------|---------------|---------------|-------|
| `ingest_document` | `POST /api/documents` | ⚠️ Partial | MCP=text only, deterministic extraction only (no LLM, no pipeline); REST supports PDF+text. Intentional: agents paste text, not upload PDFs |
| `search_documents` | `GET /api/documents/` (list) | ⚠️ Partial | MCP=global title search, stateless, no auth scoping; REST=per-user library (requires auth). See Scope note below |
| `get_document` | `GET /api/documents/:id` | ✅ Full | Same document data; MCP supports selective `include` (metadata,sections,evidence,figures) |
| `get_figures` | `GET /api/documents/:id` | ✅ Full | MCP includes base64 `image_url` for image_fallback charts; same provenance `source_text`/`page_number`/`chapter_id` |
| `get_evidence` | `GET /api/documents/:id/evidence-graph` + `GET /api/documents/:id/claims` | ✅ Full | MCP merges evidence+claims with `claim_evidence` links and provenance; same as REST evidence graph |

**Scope note for `search_documents`:** stateless, global search across all ingested documents (ephemeral, 7-day expiry). Not scoped to API-key-owned documents, not per-user. This is intentional for MVP: MCP has no session/auth to scope listing, and document data is ephemeral anyway. See Key Decisions Log in `docs/PROJECT_STATE.md` (2026-09-21 entry). If per-user scoping is needed later, `search_documents` will need to accept auth context and filter by `user_id`.

### Tools Not Implemented (by design)

| Tool | Status | Rationale |
|------|--------|-----------|
| `get_tables` | Not implemented | No product reason yet; tables extracted on-demand per paper |
| `compare_papers` (prev `compare_papers`) | Not implemented | Removed in Chunk 11 MCP lock to 5 tools; no demand signal |
| `analyze_paper` / `get_summary` / `get_claims` (legacy names) | Not implemented | Renamed/merged into `ingest_document` / `get_document` / `get_evidence` in Chunk 11; old names in this doc were drift |

Former doc listed `analyze_paper`, `get_summary`, `get_claims`, `compare_papers` — none exist in `internal/mcp/tools.go` as of 2026-09-21. This table now matches the 5 registered tools (`grep -c "Name:" internal/mcp/tools.go` == 5).

## Operations Intentionally Unavailable to Agents

| REST Endpoint | Why Unavailable | Rationale |
|---------------|-----------------|-----------|
| `GET /api/documents/` (list) | No user context | MCP is stateless; no session/auth to scope listing — `search_documents` provides global title search instead |
| `GET /api/documents/stats` | No user context | Same as above |
| `PUT /api/documents/:id/save` | User preference | Agents don't manage user libraries |
| `PATCH /api/documents/:id` | User preference | Agents don't rename user documents |
| `DELETE /api/documents/:id` | Destructive | Agents shouldn't delete user data |
| `POST /api/documents/:id/share` | User preference | Share links are user-initiated |
| `DELETE /api/documents/:id/share` | User preference | Same |
| `PATCH /api/documents/:id/visibility` | User preference | Visibility is user-controlled |
| `POST /share-referrals` | Analytics | Referral tracking is product-level |

## Research Semantics Preservation

All MCP responses preserve these research semantics (matching REST):

- **Provenance**: `page_number`, `figure_id`, `table_id`, `section`, `source_reference`
- **Uncertainty**: `mismatch_detected`, `mismatch_detail` in claim verification
- **Source attribution**: `source_text` linking claims to original passages
- **Chapter linkage**: `chapter_id` connecting figures to document sections

## Architecture Rule

> Business logic and research semantics live in the core/application layer, not separately inside the web UI, API, or MCP adapter.

Both REST and MCP handlers read from the same repository layer and return the same underlying data structures. No duplicated business logic.

## P35 Dependency Direction (Verified 2026-09-15)

**Allowed:** `MCP → Data/Tools → Repo/Infra`

```
internal/mcp/tools.go imports:
  paperviz/internal/repository   ← data layer (Repo)
  paperviz/internal/services     ← app services (Tools)
```

**Forbidden (verified zero hits):**
- `MCP → Handler → Repo` — `internal/mcp` does NOT import `internal/handlers`
- `MCP → Gemini → result` — tool handlers never call `GeminiClient`; the field exists on `MCPServer` struct for future use but is never invoked in any tool path

**Evidence:** `grep -r "handlers" internal/mcp/` returns only comment strings ("tool handlers"), zero import paths. `grep -r "gemini\|Gemini" internal/mcp/tools.go` returns only comments (lines 19, 23, 59), zero function calls.
