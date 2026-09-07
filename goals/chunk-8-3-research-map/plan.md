# Chunk 8.3 — Cross-Paper Research Map

## Goal
Allow users to understand how their analyzed papers relate to each other — supporting, contradicting, similar methodology, different datasets, different findings. This is the foundation of the long-term product moat.

## What Exists (from 8.2)
- `paper_relationships` table: id, source_paper_id, target_paper_id, relationship_type (free text), evidence_text, created_at
- PaperRelationshipRepo: Insert, GetBySourcePaper, GetByTargetPaper, ListByPaper, GetRelationships (bidirectional)
- API: GET/POST /api/papers/{id}/relationships
- claim_evidence junction table with relationship_type

## What Needs to Be Built

### 1. Relationship Type Constants
**File:** `internal/repository/types.go`

Add constants for supported relationship types:
```go
const (
    RelTypeSupporting        = "supporting"
    RelTypeContradicting     = "contradicting"
    RelTypeSimilarMethod    = "similar_methodology"
    RelTypeDifferentDataset = "different_dataset"
    RelTypeDifferentFindings = "different_findings"
    RelTypeExtends          = "extends"
)
```

### 2. Service Layer — Cross-Paper Map
**File:** `internal/services/research_map.go` (new)

Service that aggregates relationships by type with paper metadata:
- `GetResearchMap(documentID string) (*ResearchMap, error)`
- Returns relationships grouped by type, enriched with paper titles/status
- Uses PaperRelationshipRepo + DocumentRepo

### 3. API Endpoint — GET Research Map
**File:** `internal/handlers/documents.go`

Add handler:
- `GetResearchMap(w http.ResponseWriter, r *http.Request)`
- Route: GET /api/documents/:id/research-map
- Returns: { supporting: [...], contradicting: [...], similar_methodology: [...], ... }

**File:** `internal/handlers/router.go`

Register new route in the documents route group.

### 4. Frontend — Research Map Panel
**File:** `frontend/src/components/research-map.jsx` (new)

Component that displays related papers grouped by relationship type:
- Card-based list matching existing UI patterns
- Each card shows: title, relationship type badge, evidence text, link to view
- Empty state when no relationships exist

**File:** `frontend/src/pages/result-page.jsx`

Add Research Map as a tab/section in the result page action bar.

### 5. Tests
**Files:**
- `internal/services/research_map_test.go` (new)
- `internal/handlers/documents_test.go` (update)

Table-driven tests for service and handler.

## Verification
1. `go build ./...` — compiles
2. `go test ./...` — all tests pass
3. Manual: Create relationships via POST, verify GET returns grouped data
4. Frontend: Research Map panel renders on result page

## Risks
- Relationship type is free text in DB — constants enforce validity at app level
- No auto-detection yet — manual creation only for MVP
