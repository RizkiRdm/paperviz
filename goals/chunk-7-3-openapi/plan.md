# Chunk 7.3 — OpenAPI Spec

**Goal:** Create a versioned OpenAPI 3.1.0 spec covering all PaperViz API endpoints, schemas, authentication, rate limits, and error formats.

## Files to Create
- `docs/openapi.yaml` — Main OpenAPI spec

## Solution Approach

1. Create OpenAPI 3.1.0 YAML spec with:
   - Info section (title, version, description)
   - Server config (production URL)
   - Security schemes (session cookie auth)
   - All endpoint paths from `router.go`
   - Request/response schemas matching `canonical-research-output-contract.md`
   - Error response schema
   - Rate limit documentation

2. Schemas to define (from canonical contract):
   - Document, Chart, Chapter, ClaimDiff, Evidence
   - PaperSummary, ComparisonDimension, Comparison, EvidenceClaim
   - DocumentListItem, Collection, CollectionListItem
   - Error response

3. Endpoints to document (from router.go):
   - Auth: signup, login, logout, me
   - Documents: create, get, list, stats, update, delete, toggle save
   - Charts: get image, generate/revoke share
   - Sharing: generate/revoke doc share, update visibility
   - Public share: get figure, get paper
   - Collections: CRUD, add/remove documents
   - Usage: get usage
   - Analytics: pricing view, upgrade intent
   - Comparison: compare papers

## Verification
- YAML syntax valid
- All endpoints from router.go documented
- Schemas match canonical-research-output-contract.md
- Error codes match structured-research-api.md
