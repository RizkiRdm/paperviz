# Facts — Chunk 8.2 Evidence Graph

## Scope
1. Create `claim_evidence` junction table linking claims to supporting evidence
2. Create `paper_relationships` table for cross-paper links (supporting, contradicting, citing, similar_methodology)
3. Add EvidenceID field to Claim entity (alternative to junction table for simpler queries)
4. Add relationship traversal methods to repos
5. Add API endpoints to query evidence graph

## Entity Relationships
6. Claim → Evidence: one-to-many (claim can have multiple supporting evidence)
7. Evidence → Figure/Table: already exists (figure_id, table_id FKs)
8. Paper ↔ Paper: many-to-many via paper_relationships table

## Database Schema
9. claim_evidence table: id, claim_id, evidence_id, relationship_type (supports|contradicts|clarifies), created_at
10. paper_relationships table: id, source_paper_id, target_paper_id, relationship_type (supporting|contradicting|citing|similar_methodology|different_findings), evidence_text, created_at

## Repository Methods
11. ClaimRepo: GetEvidence(claimID) returns []Evidence
12. EvidenceRepo: GetClaims(evidenceID) returns []Claim
13. New PaperRelationshipRepo: Insert, GetBySourcePaper, GetByTargetPaper, GetByBothPapers

## API Endpoints
14. GET /api/documents/{id}/evidence-graph returns full graph for a paper
15. GET /api/papers/{id}/relationships returns cross-paper relationships
16. POST /api/papers/{id}/relationships creates a relationship (for manual linking)

## Traversal Queries
17. GetPaperEvidenceGraph(paperID) returns: claims with their evidence, evidence with their figures/tables
18. GetCrossPaperLinks(paperID) returns: all papers linked to this paper with relationship types
