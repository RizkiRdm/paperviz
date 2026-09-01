# PROJECT_STATE.md — PaperViz

> How this is used: this file + AGENTS.md + [one active task file] is the
> complete context dump at the start of every session. Nothing else should
> be needed. Update "Current Focus" at the end of every session (2
> minutes, before closing the laptop). Update other sections ONLY when the
> underlying fact actually changed — if you catch yourself editing the
> same section every day, that content belongs in the task file, not here.
>
> Golden rule: if a line hasn't changed in the last 2 weeks, it's not
> state, it's architecture — move it to AGENTS.md or docs/architecture.md.
> Don't let this file become a second AGENTS.md.
>
> AI reading this: treat everything below as ground truth for where the
> project stands right now. If something here looks stale or contradicts
> what you find in the actual code, say so explicitly instead of silently
> trusting this file — it's maintained by hand and can lag behind reality.

**Last updated:** 2026-09-01 — Chunk 8.2 merged: Evidence Graph (claim_evidence junction, paper_relationships, graph traversal methods, 3 API endpoints)

---

## Current Focus
*(The section that changes most — safe to fully rewrite every session.)*

- **Working on:** Phase 8 — Research Knowledge Layer
- **Active task file:** none
- **Blocked on / pending decision:** none
- **Next action if resuming:** Start Chunk 8.3 — Cross-Paper Research Map

---

## Stack & Constraints
*(Changes rarely. If this duplicates AGENTS.md, delete it from here — don't maintain two copies.)*

- see AGENTS.md

---

## Shipped / Stable
*(Features that are DONE and shouldn't be reconsidered. This exists so an
AI starting fresh doesn't propose redoing something finished, or touching
something that should stay frozen.)*

- Phases 1–3 (Trust & Core Value, Activation & Retention, Multi-Paper Intelligence) — done, omitted from active roadmap
- Chunk 4.1 Shareable Figure Explanation — done 2026-08-24, known limitation: no interactive Recharts on share page (image + text only)
- Chunk 4.2 Shareable Paper Explanation — done 2026-08-25 (`POST/DELETE /api/documents/{id}/share`, `PATCH /api/documents/{id}/visibility`, public `GET /share/doc/{token}`, `/share/doc/:shareToken` page)
- Chunk 4.3 Product-Led Referral Loop — done 2026-08-25 (`share_visits`/`share_conversions` counters on documents+charts, visit++ on public share GETs, `POST /api/share-referrals` conversion beacon, CTA links carry `?ref=`, upload page persists ref via `localStorage.paperviz_ref`)
- Chunk 5.1 SEO Architecture — done 2026-08-25 (`docs/seo-architecture.md` = IA source of truth; router NotFound is now a real dispatcher: `/api/*`→404 JSON, static files served, GET/HEAD→SPA fallback; `X-Robots-Tag: noindex, nofollow` on share GET/HEAD)
- Chunk 5.2 Product Pages — done 2026-08-26 (3 static HTML landing pages in `frontend/public/`: research-paper-summarizer, figure-explanation, compare-research-papers; `robots.txt`; React Router routes for SPA fallback; inline styles matching DESIGN.md; structured data + OG tags)
- Chunk 5.3 Programmatic SEO foundation — done 2026-08-26 (`frontend/seo/explain-pages.json` publish registry, `/explain/:slug` route + `ExplainPage`, crawlable `frontend/public/explain/sleep-quality-executive-function.html`, `/sitemap.xml`, `robots.txt` Sitemap line, default `X-Robots-Tag: noindex, nofollow` on `/explain/*`)
- Chunk 6.2 Usage Limits — done 2026-08-27 (Free/Pro/Research tiers, `user_tiers` table, fingerprint-based tracking, `UsageLimitMiddleware` on POST /api/documents, `GET /api/usage` endpoint, frontend `UsageDisplay` + `UpgradeCta` components)
- Chunk 6.3 Cost Model — done 2026-08-28 (`docs/cost-model.md` — Gemini API pricing analysis, per-operation cost breakdown, tier margin analysis, storage/bandwidth estimates, sensitivity analysis)
- Chunk 6.4 Pricing & Packaging Experiment — done 2026-08-29 (`docs/pricing-strategy.md` — 3-tier pricing, experiment design; `frontend/src/pages/pricing-page.jsx` — 3-column pricing page; conversion tracking endpoints)
- Chunk 7.1 Structured Research API — done 2026-08-29 (`docs/structured-research-api.md` — comprehensive API documentation for all endpoints)
- Chunk 7.2 Canonical Research Output Contract — done 2026-08-29 (`docs/canonical-research-output-contract.md` — 15 entity schemas with provenance and uncertainty models)
- Chunk 7.3 OpenAPI — done 2026-08-31 (`docs/openapi.yaml` — 1748-line OpenAPI 3.1.0 spec, 34 endpoints, 31 component schemas, session cookie auth, rate limits, error codes)
- Chunk 7.4 MCP — done 2026-08-31 (`cmd/mcp/main.go`, `internal/mcp/` — 6 research tools: analyze_paper, get_summary, get_figures, get_claims, get_evidence, compare_papers; official Go SDK, stdio transport)
- Chunk 7.5 Human/Agent Capability Parity — done 2026-08-31 (`docs/mcp-parity.md` — MCP↔REST parity map; `internal/mcp/tools.go` — added image_url base64 to get_figures; documented intentionally unavailable agent operations)
- Chunk 7.6 MCP Usage, Security & Cost Controls — done 2026-08-31 (`internal/mcp/errors.go` — MCPError type + sentinels; `internal/mcp/ratelimit.go` — per-key token bucket (analyze 5/min, read 30/min, compare 2/min); `internal/mcp/jobs.go` — concurrent job limiter; `internal/mcp/server.go` — API key auth; `internal/mcp/tools.go` — 500KB size cap, 5-min timeout, rate/job checks; `cmd/mcp/main.go` — PAPERVIZ_API_KEY required)
- Chunk 8.1 Structured Research Objects — done 2026-09-01 (`migrations/014_structured_research_objects.sql` — 5 new tables: claims, paper_tables, methods, results, citations; `internal/repository/` — 5 new repos: ClaimRepo, PaperTableRepo, MethodRepo, ResultRepo, CitationRepo; `internal/handlers/documents.go` — 5 GET endpoints; `docs/canonical-research-output-contract.md` — updated with new entity schemas; 203 tests passing)
- Chunk 8.2 Evidence Graph — done 2026-09-01 (`migrations/015_evidence_graph.sql` — 2 new tables: claim_evidence, paper_relationships; `internal/repository/` — 2 new repos: ClaimEvidenceRepo, PaperRelationshipRepo + traversal methods on Claims/Evidence; `internal/handlers/documents.go` — 3 endpoints: GET evidence-graph, GET/POST relationships; 219 tests passing)

---

## Known Gaps (not built yet, on purpose)
*(Different from a bug — this is something that intentionally doesn't
exist yet, and that's fine for now. Prevents an AI from "discovering" a
gap you already know about and already decided to defer, and re-litigating
it.)*

- Interactive Recharts render on share pages — why deferred: avoid scope creep in 4.1; image + text explanation suffices for MVP
- Revoke UI for document share links — why deferred: `revokeDocumentShare` exists in api.js; no button requested by chunk scope
- Visibility in GET /api/documents/:id payload — why deferred: frontend selector initializes locally to "private"; add field when needed

---

## Key Decisions Log
*(ONE line per decision that would otherwise get re-argued every session.
Not a full history — link out to a separate doc if the full reasoning
matters. This is the guardrail that stops a new session from re-proposing
something you already rejected for a clear reason.)*

| Date | Decision | Why |
|---|---|---|
| 2026-08-24 | Lazy share-token generation | Avoids wasting tokens on documents never shared |
| 2026-08-24 | Figures inherit document visibility | Simpler, no per-figure toggle for MVP |
| 2026-08-24 | No extra auth on GetChartImage | Share service gates data layer; URL only exposed for non-private docs |
| 2026-08-25 | Sharing auto-bumps private→unlisted | One-click UX; fixes 4.1 gap where all-private docs 404'd figure share pages |
| 2026-08-25 | Revoke downgrades unlisted→private, never public→private | Revoking a link shouldn't silently change explicit public choice |
| 2026-08-25 | Switching to private clears share_token | Private must kill existing links |
| 2026-08-25 | Shared paper payload excludes original_text/user_id/tokens | Copyright risk on public URLs |
| 2026-08-25 | Funnel = atomic counters, no events table | Rule 2 minimal; per-token counts answer share→visit→analysis without time series |
| 2026-08-25 | Conversion attribution via localStorage beacon | No-auth product; server-side cookie infra unjustified for MVP signal |
| 2026-08-25 | Visit counted only after visibility check passes | Private/expired tokens must not inflate funnel |
| 2026-08-25 | Prune `/scientific-figure-analysis` + `/research-paper-analysis` routes | One page per intent; near-duplicate keywords cannibalize |
| 2026-08-25 | Crawlable pages must be server-delivered full HTML | Zero-JS AI crawlers + social scrapers; WRS render lag |
| 2026-08-25 | Share URLs: X-Robots-Tag header, never robots.txt Disallow | Disallow hides noindex from crawler; header works on CSR pages |
| 2026-08-25 | `/explain/` reserved, noindex until all 5 quality gates pass | Scaled-content-abuse policy has no volume threshold |
| 2026-08-26 | Static-in-dist for landing pages | Zero Go changes; existing file server serves them; Vite copies public/ → dist/ on build |
| 2026-08-26 | React Router routes for static landing pages | SPA fallback when users navigate from within app; static HTML for crawlers |
| 2026-08-27 | Fingerprint = IP + User-Agent + Accept-Language (SHA256) | No-auth product; browser fingerprint sufficient for MVP usage tracking |
| 2026-08-29 | Pro tier $29/month for 50 papers | 52% margin at $0.30/paper cost; validates willingness to pay |
| 2026-08-29 | Research tier = "Contact us" (no fixed price) | Avoids unprofitable fixed pricing; custom for high-volume users |
| 2026-08-29 | Upgrade flow = waitlist/email capture | No payment processing yet; measures intent before building billing |

---

## Where Things Live
*(Not a full folder tree — that's AGENTS.md/architecture.md's job. This is
just a fast map: "if I need to change X, which file do I open".)*

- Share services (figure + paper): `internal/services/share.go`
- Share handlers + routes wiring: `internal/handlers/share.go`, `internal/handlers/router.go`
- Document share repo methods: `internal/repository/documents.go`
- Public share pages: `frontend/src/pages/share-figure-page.jsx`, `frontend/src/pages/share-paper-page.jsx`
- Share API client fns: `frontend/src/lib/api.js`
- Referral counters: repo methods in `internal/repository/documents.go` + `charts.go`, service in `internal/services/share.go` (`TrackReferralConversion`), route `POST /api/share-referrals`
- Ref attribution: `frontend/src/pages/upload-page.jsx` (localStorage capture + beacon), CTA links on both share pages
- SEO IA + URL-space contract: `docs/seo-architecture.md`; router dispatcher `internal/handlers/router.go` (`spaNotFound`, `noindexMiddleware`)
- Landing pages (static HTML): `frontend/public/research-paper-summarizer.html`, `frontend/public/figure-explanation.html`, `frontend/public/compare-research-papers.html`
- Explain publish registry + pages: `frontend/seo/explain-pages.json`, `frontend/src/pages/explain-page.jsx`, `frontend/public/explain/*.html`
- Sitemap: `frontend/public/sitemap.xml`
- Crawl policy: `frontend/public/robots.txt`
- Migrations: `migrations/009_share_tokens.sql`, `migrations/010_document_share.sql`, `migrations/011_share_referrals.sql`, `migrations/012_usage_analytics.sql`, `migrations/013_usage_tiers.sql`
- Analytics repo + handler: `internal/repository/analytics.go`, `internal/handlers/analytics.go`
- Tier service: `internal/services/tier.go`
- Usage middleware + fingerprint: `internal/handlers/middleware.go` (`UsageLimitMiddleware`), `internal/handlers/fingerprint.go`
- Usage API: `internal/handlers/usage.go`
- Frontend usage components: `frontend/src/components/usage-display.jsx`, `frontend/src/components/upgrade-cta.jsx`
- Task/chunk docs: `docs/`
- Cost model: `docs/cost-model.md`
- Pricing strategy: `docs/pricing-strategy.md`
- Pricing page: `frontend/src/pages/pricing-page.jsx`
- Conversion tracking: `internal/handlers/analytics.go` (`TrackPricingView`, `TrackUpgradeIntent`)
- Structured Research API docs: `docs/structured-research-api.md`
- Canonical Research Output Contract: `docs/canonical-research-output-contract.md`
- OpenAPI spec: `docs/openapi.yaml`
- MCP server: `cmd/mcp/main.go`, `internal/mcp/server.go`, `internal/mcp/tools.go`, `internal/mcp/errors.go`, `internal/mcp/ratelimit.go`, `internal/mcp/jobs.go`
- MCP↔REST parity map: `docs/mcp-parity.md`
- Structured research objects migration: `migrations/014_structured_research_objects.sql`
- Structured research repos: `internal/repository/claims.go`, `internal/repository/tables.go`, `internal/repository/methods.go`, `internal/repository/results.go`, `internal/repository/citations.go`
- Structured research API endpoints: `internal/handlers/documents.go` (GET /api/documents/:id/{claims,tables,methods,results,citations})
- Evidence graph migration: `migrations/015_evidence_graph.sql`
- Evidence graph repos: `internal/repository/claim_evidence.go`, `internal/repository/paper_relationships.go`
- Evidence graph traversal: `internal/repository/claims.go` (GetEvidence), `internal/repository/evidence.go` (GetClaims)
- Evidence graph API endpoints: `internal/handlers/documents.go` (GET evidence-graph, GET/POST relationships)
