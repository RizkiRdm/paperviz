# Facts — Chunk 8.1 Structured Research Objects

## Scope
1. Formalize existing entities (Document, Chart, Evidence) with stable IDs — they already have stable IDs via `repository.NewID()`, so this is verify-only
2. Create new `claims` table and `Claim` entity for structured research claims (separate from ClaimDiff which is for verification)
3. Create new `tables` table and `Table` entity for extracted tables (schema already defined in canonical contract)
4. Create new `methods` table and `Method` entity for research methodologies
5. Create new `results` table and `Result` entity for research findings
6. Create new `citations` table and `Citation` entity for references

## Entity Definitions
7. Claim entity: id (claim_), paper_id, claim_text, claim_type (hypothesis|finding|conclusion|method|limitation), confidence (high|medium|low|null), source_page, source_text, created_at
8. Table entity: id (table_), document_id, page_number, caption, headers (JSON), rows (JSON), source_text, display_order
9. Method entity: id (method_), paper_id, method_name, description, type (experimental|survey|qualitative|quantitative|computational|other), source_page, source_text
10. Result entity: id (result_), paper_id, result_text, result_type (primary|secondary|negative|null), supporting_evidence_id (nullable), source_page, source_text
11. Citation entity: id (citation_), paper_id, cited_paper_id (nullable), authors, title, year, venue, doi (nullable), url (nullable), source_page

## Pipeline Integration
12. Method, Result, Citation extraction added to simplify stage of pipeline
13. Claim extraction happens during verify stage (already partially exists via ClaimDiff)
14. Table extraction happens during charts stage (alongside figure extraction)

## Backward Compatibility
15. No backfill for existing documents — new entities only for new analyses
16. All new fields nullable in existing tables
17. ClaimDiff coexists with new Claim entity (different purposes)

## Repository Layer
18. New repos: ClaimRepo, TableRepo, MethodRepo, ResultRepo, CitationRepo
19. Each repo follows existing pattern: dbExecutor interface, Insert/ListByPaper methods
20. IDs generated via repository.NewID() with appropriate prefix

## API Layer
21. New endpoints: GET /papers/{id}/claims, GET /papers/{id}/tables, GET /papers/{id}/methods, GET /papers/{id}/results, GET /papers/{id}/citations
22. All endpoints return arrays of entities with stable IDs
23. Provenance fields always present (source_page, source_text)
