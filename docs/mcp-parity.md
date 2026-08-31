# MCP ↔ REST API Parity Map

**Chunk 7.5 — Human/Agent Capability Parity**

## MCP Tools → REST Endpoints

| MCP Tool | REST Endpoint | Parity Status | Notes |
|----------|---------------|---------------|-------|
| `analyze_paper` | `POST /api/documents` | ⚠️ Partial | MCP=text only; REST supports PDF+text. Intentional: agents paste text, not upload PDFs |
| `get_summary` | `GET /api/documents/:id` | ✅ Full | Same research semantics |
| `get_figures` | `GET /api/documents/:id` | ✅ Full | MCP now includes base64 `image_url` for image_fallback charts |
| `get_claims` | `GET /api/documents/:id` | ✅ Full | Same claim diff data |
| `get_evidence` | `GET /api/documents/:id` | ✅ Full | Same evidence references |
| `compare_papers` | `POST /api/documents/compare` | ✅ Full | Same comparison output |

## Operations Intentionally Unavailable to Agents

| REST Endpoint | Why Unavailable | Rationale |
|---------------|-----------------|-----------|
| `GET /api/documents/` (list) | No user context | MCP is stateless; no session/auth to scope listing |
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
