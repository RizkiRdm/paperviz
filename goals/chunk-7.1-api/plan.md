# Chunk 7.1 — Structured Research API

## Goal
Expose clean, documented API for programmatic access. Formalize existing endpoints as the "Structured Research API" for developers and AI agents.

## Current State
- Existing routes under `/api/documents` — CRUD + compare
- DocumentHandler, ShareHandler, AnalyticsHandler, UsageHandler
- JSON responses, fingerprint auth, chi router
- No API versioning, no OpenAPI spec, no developer-facing docs

## Scope

### Approach: Extend, Don't Duplicate
Existing `/api/documents` already exposes core functionality. Chunk 7.1:
1. Adds missing granular endpoints (summary, figures, tables, claims separately)
2. Ensures consistent response schemas
3. Documents the API as "Structured Research API"
4. No new `/api/papers` prefix — keep existing conventions

### Deliverables
1. `docs/structured-research-api.md` — API documentation
2. `internal/handlers/papers.go` — New handler for granular research endpoints
3. Route additions in `internal/handlers/router.go`

### New Endpoints
```
GET /api/documents/{id}/summary     — Get simplified text summary
GET /api/documents/{id}/figures     — Get all charts/figures
GET /api/documents/{id}/tables      — Get extracted tables (if any)
GET /api/documents/{id}/claims      — Get extracted claims
POST /api/documents/{id}/compare    — Already exists (keep as-is)
```

### Response Schemas
All responses follow consistent shape:
```json
{
  "id": "doc_abc123",
  "title": "Paper Title",
  "summary": { "simplified_text": "...", "reading_level": "simplified" },
  "figures": [{ "chart_id": "...", "annotation": "...", "page_number": 1 }],
  "tables": [{ "table_id": "...", "data": [...], "page_number": 2 }],
  "claims": [{ "claim_id": "...", "text": "...", "evidence": [...] }]
}
```

### API Documentation
`docs/structured-research-api.md` covers:
- Authentication (fingerprint-based for MVP)
- Rate limits
- Response formats
- Error codes
- Example requests/responses

## Implementation Steps

### Step 1: API Documentation
**File:** `docs/structured-research-api.md`
- Endpoint list with methods, paths, descriptions
- Request/response schemas
- Authentication and rate limits
- Error handling
- Example curl commands

### Step 2: Papers Handler
**File:** `internal/handlers/papers.go`
- `PapersHandler` struct (similar to DocumentHandler)
- `GetSummary` — returns simplified_text + reading_level
- `GetFigures` — returns charts array
- `GetTables` — returns tables (placeholder for future)
- `GetClaims` — returns claims with evidence

### Step 3: Route Registration
**File:** `internal/handlers/router.go`
- Add `papersHandler := NewPapersHandler(db)`
- Register routes under `/api/documents/{id}/...`

## Verification
- [ ] All endpoints return consistent JSON
- [ ] Documentation covers all endpoints
- [ ] Build passes (`go build ./...`)
- [ ] Tests pass (`go test ./...`)
- [ ] Existing functionality unchanged

## Constraints
- No new auth system (fingerprint for MVP)
- No API versioning yet (v1 comes later)
- No breaking changes to existing endpoints
- Follow existing handler patterns
