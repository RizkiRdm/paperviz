# Goal — Chunk 8.1 Structured Research Objects

## Articulated Goal
Formalize research entities (Paper, Claim, Evidence, Figure, Table, Method, Result, Citation) with stable identifiers by adding database tables, repository methods, and API endpoints. This establishes the structured research knowledge layer for cross-paper comparison and evidence linking.

## Reference Documents
- **Facts:** `goals/chunk-8.1-structured-research-objects/facts.md`
- **Plan:** `goals/chunk-8.1-structured-research-objects/plan.md`
- **Canonical Contract:** `docs/canonical-research-output-contract.md`
- **Roadmap:** `docs/paperviz-agent-roadmap-revised.md` (Chunk 8.1)

## Done Condition
All new entities (Claim, Table, Method, Result, Citation) have:
1. Database tables with proper schema
2. Repository methods (Insert, ListByPaper, GetByID)
3. API endpoints returning structured data
4. Unit tests passing
5. Canonical contract updated
6. Pipeline extraction working for new analyses
