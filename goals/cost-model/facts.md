# Cost Model Facts

## Core Deliverable
1. Create `docs/cost-model.md` with structured cost analysis

## Cost Drivers to Model
2. Gemini API costs: input/output tokens per operation type
3. PDF processing costs: extraction time, CPU/memory
4. Storage costs: SQLite database growth per paper
5. Bandwidth costs: serving share pages, API responses
6. Infrastructure costs: server hosting, domain, SSL
7. Support costs: estimated overhead

## Operations to Cost
8. Single-paper analysis (upload → simplified text + claims)
9. Figure analysis (image extraction → chart data + annotation)
10. Paper comparison (2 papers → comparison output)
11. Share page view (public GET /share/doc/{token})

## Analysis Required
12. Cost per paper (average across operation types)
13. Gross margin calculation (cost vs assumed revenue)
14. Maximum free-tier abuse scenario (worst-case cost)
15. Tier-specific margin analysis (Free/Pro/Research)

## Output Format
16. Markdown document with tables
17. Version-controlled in docs/
18. Include assumptions section
19. Include sensitivity analysis for key variables
