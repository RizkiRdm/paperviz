# Plan — Chunk 8.1 Structured Research Objects

## Solution Approach
Formalize research entities by adding database tables, repository methods, and API endpoints for Claim, Table, Method, Result, Citation. Follow existing patterns in codebase.

## Implementation Steps

### Step 1: Database Schema
Create migration `014_structured_research_objects.sql`:
- `claims` table: id, paper_id, claim_text, claim_type, confidence, source_page, source_text, created_at
- `tables` table: id, document_id, page_number, caption, headers (TEXT/JSON), rows (TEXT/JSON), source_text, display_order
- `methods` table: id, paper_id, method_name, description, method_type, source_page, source_text
- `results` table: id, paper_id, result_text, result_type, supporting_evidence_id, source_page, source_text
- `citations` table: id, paper_id, cited_paper_id, authors, title, year, venue, doi, url, source_page

**Verify:** `sqlite3 paperviz.db < migrations/014_structured_research_objects.sql` succeeds

### Step 2: Repository Types
Update `internal/repository/types.go`:
- Add Claim struct with fields matching claims table
- Add PaperTable struct (renamed from Table to avoid conflict with Go's table keyword)
- Add Method struct
- Add Result struct
- Add Citation struct
- Add enum constants for claim_type, confidence, method_type, result_type

**Verify:** `go build ./...` succeeds

### Step 3: Repository Methods
Create new files following existing patterns:
- `internal/repository/claims.go`: ClaimRepo with Insert, ListByPaper, GetByID
- `internal/repository/tables.go`: TableRepo with Insert, ListByPaper, GetByID
- `internal/repository/methods.go`: MethodRepo with Insert, ListByPaper, GetByID
- `internal/repository/results.go`: ResultRepo with Insert, ListByPaper, GetByID
- `internal/repository/citations.go`: CitationRepo with Insert, ListByPaper, GetByID

**Verify:** `go test ./internal/repository/...` passes

### Step 4: Service Layer
Update `internal/services/intake.go` to extract new entities during pipeline:
- Add Method extraction to simplify stage (via Gemini prompt)
- Add Result extraction to simplify stage
- Add Citation extraction to simplify stage
- Add Table extraction to charts stage (alongside figure extraction)
- Add Claim extraction to verify stage (enhance existing ClaimDiff)

**Verify:** Pipeline still processes documents correctly

### Step 5: API Endpoints
Update `internal/handlers/documents.go`:
- Add GET /api/documents/{id}/claims endpoint
- Add GET /api/documents/{id}/tables endpoint
- Add GET /api/documents/{id}/methods endpoint
- Add GET /api/documents/{id}/results endpoint
- Add GET /api/documents/{id}/citations endpoint

**Verify:** `go test ./internal/handlers/...` passes

### Step 6: Update Canonical Contract
Update `docs/canonical-research-output-contract.md`:
- Add Claim, Method, Result, Citation entity schemas
- Update Document entity to reference new entities
- Update identifier conventions table

**Verify:** Document matches implementation

### Step 7: Tests
Add unit tests for each new repository:
- `internal/repository/claims_test.go`
- `internal/repository/tables_test.go`
- `internal/repository/methods_test.go`
- `internal/repository/results_test.go`
- `internal/repository/citations_test.go`

Each test: 1 success case, 1 error case (per AGENTS.md testing rules)

**Verify:** `go test ./...` passes

## Risks
1. Gemini prompt changes may affect extraction quality — mitigate with existing claim verification
2. Table extraction from PDF is complex — may need image fallback like charts
3. Citation parsing from raw text is error-prone — use Gemini for structured extraction
