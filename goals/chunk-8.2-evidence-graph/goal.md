# Goal — Chunk 8.2 Evidence Graph

## Articulated Goal
Build evidence graph infrastructure connecting claims to evidence and papers to each other. This enables traversal of research relationships: which claims are supported by which evidence, and which papers support/contradict each other.

## Reference Documents
- **Facts:** `goals/chunk-8.2-evidence-graph/facts.md`
- **Plan:** `goals/chunk-8.2-evidence-graph/plan.md`
- **Roadmap:** `docs/paperviz-agent-roadmap-revised.md` (Chunk 8.2)

## Done Condition
Evidence graph is functional when:
1. claim_evidence junction table exists with proper schema
2. paper_relationships table exists with proper schema
3. Repository methods for graph traversal are implemented
4. API endpoints return graph data
5. Unit tests pass
6. Documentation updated
