# Plan — Chunk 8.2 Evidence Graph

## Solution Approach
Build evidence graph using junction tables for relationships. Simplest architecture: foreign keys + junction tables, no graph database.

## Implementation Steps

### Step 1: Database Schema
Create migration `015_evidence_graph.sql`:
- `claim_evidence` table: id, claim_id FK, evidence_id FK, relationship_type, created_at
- `paper_relationships` table: id, source_paper_id FK, target_paper_id FK, relationship_type, evidence_text, created_at
- Indexes on foreign keys

**Verify:** `sqlite3 paperviz.db < migrations/015_evidence_graph.sql` succeeds

### Step 2: Repository Types
Update `internal/repository/types.go`:
- Add ClaimEvidence struct (junction)
- Add PaperRelationship struct
- Add relationship type constants

**Verify:** `go build ./...` succeeds

### Step 3: Repository Methods
Create new files:
- `internal/repository/claim_evidence.go`: ClaimEvidenceRepo with Insert, GetByClaim, GetByEvidence
- `internal/repository/paper_relationships.go`: PaperRelationshipRepo with Insert, GetBySourcePaper, GetByTargetPaper, GetRelationships

Add traversal methods to existing repos:
- `internal/repository/claims.go`: Add GetEvidence method
- `internal/repository/evidence.go`: Add GetClaims method

**Verify:** `go test ./internal/repository/...` passes

### Step 4: API Endpoints
Update `internal/handlers/documents.go`:
- Add GET /api/documents/{id}/evidence-graph endpoint
- Add GET /api/papers/{id}/relationships endpoint
- Add POST /api/papers/{id}/relationships endpoint

**Verify:** `go test ./internal/handlers/...` passes

### Step 5: Tests
Add unit tests for new repos and methods:
- `internal/repository/claim_evidence_test.go`
- `internal/repository/paper_relationships_test.go`
- Update claims_test.go and evidence_test.go with new methods

**Verify:** `go test ./...` passes

### Step 6: Update Documentation
Update `docs/canonical-research-output-contract.md`:
- Add ClaimEvidence and PaperRelationship schemas
- Document graph traversal capabilities

**Verify:** Document matches implementation

## Risks
1. Cross-paper relationships may be sparse initially (only from comparison feature)
2. Graph traversal queries could be slow with many relationships — mitigate with indexes
3. Manual relationship creation requires understanding of research context
