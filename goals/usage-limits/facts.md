# Facts — Chunk 6.2: Usage Limits

## Tier Structure
1. Three tiers: Free, Pro, Research
2. Free tier: 5 papers/month, basic figure extraction only, no comparison
3. Pro tier: 50 papers/month, advanced figure analysis, saved library, 10 comparisons/month
4. Research tier: 500 papers/month, all features, unlimited comparisons, export

## Enforcement
5. Hard block when limit reached — cannot process new papers
6. Clear upgrade CTA shown on block
7. Grace period not needed for MVP

## User Identification
8. Browser fingerprint + IP for usage tracking (no auth required)
9. localStorage stores usage count as fallback
10. Server-side tracks via analytics_events table (already exists)

## Frontend
11. Usage display on upload page showing current month usage
12. Progress bar showing papers used / limit
13. Upgrade CTA button when near or at limit
14. No modal, no redirect — inline message

## Backend
15. New middleware: usage-check before document creation
16. Tier stored in user fingerprint record (new table or column)
17. Monthly reset logic (first of month)
18. Rate limit already exists (1 req/30s) — usage limits are separate concept

## Scope
19. Backend: tier logic, enforcement, tracking
20. Frontend: usage display, upgrade CTA
21. No admin dashboard (deferred)
22. No payment integration (deferred — just tiers and limits)
