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

**Last updated:** 2026-09-05 — TASK-2 chart missing value fix merged (exclude undefined/null instead of silent zero-fill), 331 tests passing

---

## Current Focus
*(The section that changes most — safe to fully rewrite every session.)*

- **Working on:** Phase 12 Cleanup (unused-export audit or tooltip/a11y polish)
- **Active task file:** none
- **Blocked on / pending decision:** none
- **Next action if resuming:** unused-export audit (grep exported symbols with zero callers) OR tooltip/a11y polish pass

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
- Chunk 8.3 Cross-Paper Research Map — done 2026-09-01 (`internal/repository/types.go` — 5 relationship type constants; `internal/services/research_map.go` — GetResearchMap service grouping relationships by type with paper titles; `internal/handlers/documents.go` — GET /api/documents/:id/research-map endpoint; `frontend/src/components/research-map.jsx` — collapsible type groups with color-coded cards; `frontend/src/pages/result-page.jsx` — Research Map toggle button; 6 new tests, 225 total)
- Chunk 9.1 DOI & URL Import — done 2026-09-02 (`internal/external/crossref.go` + `unpaywall.go` — DOI resolution clients; `internal/services/import.go` — FetchByDOI + FetchByURL with SSRF protection; `internal/handlers/import.go` — POST /api/import/doi + /api/import/url; `internal/services/import_fetcher.go` — paperFetcher adapter; `internal/handlers/router.go` — wired /api/import routes; `frontend/src/lib/api.js` — importByDOI + importByURL functions; `frontend/src/components/doi-import.jsx` + `url-import.jsx` — import UI components; `frontend/src/pages/upload-page.jsx` — 4-tab mode selector; 297 tests passing)
- Chunk 9.3 Research Knowledge Accumulation — done 2026-09-03 (`migrations/016_annotations.sql` — annotations table; `internal/repository/annotations.go` — AnnotationRepo CRUD; `internal/services/annotations.go` — annotation service with ownership enforcement; `internal/services/export.go` — ExportResearchContext assembles full research context JSON; `internal/handlers/annotations.go` — 4 CRUD endpoints; `internal/handlers/export.go` — GET export endpoint; `frontend/src/components/annotation-panel.jsx` — collapsible annotation UI; `frontend/src/pages/result-page.jsx` — annotation panel + export button; 315 tests passing)
- Chunk 10.2 Workflow Lock-In — done 2026-09-03 (`internal/services/collections.go` + `collections_test.go` — Get/Rename/Delete/Add/Remove/ListDocuments take userID, ErrForbidden on mismatch; `internal/handlers/collections.go` — UserIDFromContext 401, forbidden→403, notFound→404; `frontend/src/lib/api.js` — 5 fns listCollections/createCollection/getCollection/addDocumentToCollection/removeDocumentFromCollection; `frontend/src/components/collections-panel.jsx` — collapsible panel; 328 tests passing in 7 packages)
- Chunk 12.1 silent-error hardening (frontend-only, no backend change) — done 2026-09-04 (`frontend/src/components/annotation-panel.jsx` — row saveError/deleteError + loadError/createError Retry; `frontend/src/components/chart-card.jsx` — shareError + Retry; `frontend/src/components/collections-panel.jsx` — loadError/actionError + Retry; zero empty catch; inputs preserved on failure; DESIGN.md tokens only; grep catch audit: remaining bare catches only with handling bodies — clipboard fallback select + cancelled-guarded error; go test 328 passed 7 pkgs)
- Chunk 12.1 dead-code slice (`textChartElem` purge) — done 2026-09-04 (`internal/services/charts.go` 86-95 deleted, `internal/services/charts_test.go` `TestChartValuesInStruct` retargeted to `chapterChartJSON` with `has_chart` wrap; `chartValues` + `UnmarshalJSON` untouched/live; `GenerateChapterChart`/`tryExtractChartData` unchanged; grep `textChartElem` zero in `internal/`; `gofmt` clean; `go vet` clean; `go test` 328 passed 7 pkgs)
- Chunk 12.1 ponytail slice (`comparison.go` ceiling comments) — done 2026-09-04 (`internal/services/comparison.go` ~136/163/189/221/252: fixed-8-dimensions; single-prompt join; first-2-papers stance; joined-evidence prompt; exact-overlap keywords; each `// ponytail: ... — ceiling: ... ; upgrade: ...`; YAGNI kept: buildComparisonDimensions, synthesizeDimensions, identifyAgreementsAndDisagreements, findCommonKeywords+stopWords, ExtractPaperSummary/ComparePapers/CompareEvidence; zero logic change, no dead lines; grep ponytail 5 hits; `gofmt` clean; `go vet` clean; `go test` 328 passed 7 pkgs)
- Auth rate limiting (TASK-1) — done 2026-09-04 (`internal/handlers/ratelimit.go` — `rateLimitAuth` middleware, 5 req/60s/burst 3; `internal/handlers/router.go` — `/signup` and `/login` wrapped; 331 tests passing)
- Chart missing value fix (TASK-2) — done 2026-09-05 (`frontend/src/components/data-chart.jsx` — replaced `?? 0` silent zero-fill with `.filter()` exclude undefined/null; missing values now dropped from chart render instead of rendered as zero; explanatory comment added; `npm run build` clean)
- Verification-polish chunk — done 2026-09-04 (`frontend/src/pages/result-page.jsx` ~313-328 badge gating + ~364-394 banner detail/claims opener/compare; `frontend/src/components/status-banners.jsx` 29+/12− hardened panel + badge; `internal/services/intake.go` ~154-172 claims fan-out tx; `save_pipeline_result_test.go` 3 new test cases; behavior: verification_failed now shows real `mismatch_detail` + claims opener + Compare-with-Original; Verified badge disabled when no claim_diff + aria-expanded on opener; ClaimComparisonPanel try/catch + empty state + count badge; pipeline writes one claims row per `OriginalClaims` in the same tx; `go test` 331 passed 7 pkgs; `npm run build` clean; screenshots snap-16/17 confirmed)

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
| 2026-09-02 | URL import: https-only + block private IPs | SSRF protection; prevents user-controlled URLs from hitting internal network |
| 2026-09-03 | Annotations are per-user, not per-document | Ownership enforcement ensures users can only edit/delete their own annotations |
| 2026-09-03 | Export excludes OriginalText/SimplifiedText | Copyright compliance — users export structured metadata and their own annotations, not source text |
| 2026-09-03 | Collections enforce owner, 403 on mismatch | Mirrors annotations D1; per-user ownership closes IDOR on saved collections |
| 2026-09-04 | Frontend inline-error standard | User msg friendly no stack, dev console.error, retry preserves input; research-map block is canonical |
| 2026-09-04 | Test-only shims live in _test.go or retarget to live types, never in production files | Dead-shim purge rule; `textChartElem` removed from `charts.go`, test retargeted to `chapterChartJSON` |
| 2026-09-04 | Ponytail comment standard: `// ponytail: <simplification> — ceiling: <limit> ; upgrade: <path>` | Comments only, never logic with the marking; records simplification ceiling + upgrade path without changing behavior |
| 2026-09-04 | mismatch_detail is evidence, not decoration | Always surface to user on verification_failed — pipeline populates claims table from verification output (YAGNI, no separate LLM extraction step) |

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
- Research map service: `internal/services/research_map.go`
- Research map API endpoint: `internal/handlers/documents.go` (GET /api/documents/:id/research-map)
- Research map frontend: `frontend/src/components/research-map.jsx`
- Relationship type constants: `internal/repository/types.go` (PaperRelationshipSupporting, etc.)
- DOI/URL import external clients: `internal/external/crossref.go`, `internal/external/unpaywall.go`
- DOI/URL import service: `internal/services/import.go`, `internal/services/import_fetcher.go`
- DOI/URL import handler: `internal/handlers/import.go` (POST /api/import/doi, POST /api/import/url)
- Import API client functions: `frontend/src/lib/api.js` (importByDOI, importByURL)
- Import UI components: `frontend/src/components/doi-import.jsx`, `frontend/src/components/url-import.jsx`
- Upload page with import tabs: `frontend/src/pages/upload-page.jsx`
- Annotations migration: `migrations/016_annotations.sql`
- Annotations repo: `internal/repository/annotations.go`
- Annotations service: `internal/services/annotations.go`
- Annotations handler: `internal/handlers/annotations.go` (POST/PUT/DELETE/GET /api/documents/:id/annotations)
- Export service: `internal/services/export.go`
- Export handler: `internal/handlers/export.go` (GET /api/documents/:id/export)
- Annotation frontend: `frontend/src/components/annotation-panel.jsx` (row saveError/deleteError, loadError/createError Retry inline blocks)
- Chart card frontend: `frontend/src/components/chart-card.jsx` (shareError Retry inline block)
- Collections panel frontend: `frontend/src/components/collections-panel.jsx` (loadError/actionError Retry inline blocks)
- Collections backend: `internal/services/collections.go`, `internal/handlers/collections.go`
- Collections frontend: `frontend/src/lib/api.js` (listCollections, createCollection, getCollection, addDocumentToCollection, removeDocumentFromCollection), `frontend/src/components/collections-panel.jsx`
- Export collections join: `internal/services/export.go` (research context joins collections)
- Charts service: `internal/services/charts.go` (`GenerateChapterChart`, `tryExtractChartData`; `textChartElem` purged 2026-09-04, live type `chapterChartJSON`)
- Comparison service: `internal/services/comparison.go` (ponytail ceiling comments 2026-09-04, 5 hits, zero logic change)
- Verification UI: `frontend/src/pages/result-page.jsx` (~313-328 badge gating, ~364-394 banner detail + claims opener + compare)
- Verification banners: `frontend/src/components/status-banners.jsx` (hardened panel + disabled badge)
- Claims fan-out from verification: `internal/services/intake.go` (~154-172, writes one claims row per OriginalClaims in pipeline tx)
