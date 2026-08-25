# SEO Architecture — PaperViz

> Chunk 5.1 deliverable. This document is the source of truth for URL
> taxonomy, crawlability, and indexing policy. Chunks 5.2 (product pages)
> and 5.3 (programmatic SEO) implement against it — they may add detail,
> but must not contradict a decision recorded here.

---

## 1. Rendering / Crawlability Strategy

PaperViz is a single Go binary serving a React SPA. Two surface classes
with different rules:

| Surface | Rendering | Rationale |
|---|---|---|
| App surfaces (`/login`, `/signup`, `/dashboard`, `/compare`, result pages) | CSR React SPA | Not search targets; crawlers don't need dashboards |
| Crawlable surfaces (5.2 landing pages, future `/explain/{slug}`) | **Fully-formed HTML delivered by the Go binary at request time** | AI answer-engine crawlers execute zero JavaScript; Bing/social scrapers don't render; Googlebot's second-wave render lags days for low-authority domains |

**CSR-only is banned for crawlable surfaces.** The concrete mechanism for
5.2+ is either pre-rendered HTML files emitted into `frontend/dist`
(preferred — the existing static file serving picks them up with zero Go
changes) or Go `html/template`. The choice is a 5.2 ADR; recommendation:
static-in-dist.

Share pages remain CSR. Optional OG/meta injection middleware for nicer
link previews is a follow-up, not scheduled.

## 2. URL Taxonomy (chunk 5.1 route decisions)

Roadmap candidate routes, pruned by the "genuine user value" constraint —
one page per distinct intent:

| Candidate | Verdict | Reason |
|---|---|---|
| `/research-paper-summarizer` | **KEEP** → build in 5.2 | Primary high-intent query; maps to core capability |
| `/figure-explanation` | **KEEP** → build in 5.2 | Figure explanation is the product differentiator |
| `/scientific-figure-analysis` | REJECT | Near-duplicate intent of figure-explanation; one page per intent avoids keyword cannibalization |
| `/research-paper-analysis` | REJECT | Overlaps summarizer; fold this angle into its page copy |
| `/compare-research-papers` | **KEEP** → build in 5.2 | Maps to the existing `/compare` capability |
| `/explain/…` | RESERVE namespace → build in 5.3 | Parametric `/explain/{slug}` behind quality gates (§6) |

Rejected candidates are never created; no redirects needed since nothing
ever shipped at those paths.

## 3. Route Precedence Contract

The catch-all result route `/:documentId` coexists with present and future
static paths under these rules:

**Server (chi)** — unmatched-path dispatcher precedence:
1. `/api/*` prefix → `404 application/json` (`not_found`). API typos never
   receive the SPA shell.
2. Path resolves to an existing file under `frontend/dist` → serve it
   (hashed assets are served through this same path — an unconditional
   index.html branch would break all JS/CSS).
3. GET/HEAD on anything else → `index.html` (true SPA fallback: deep links
   like `/share/doc/x` and `/dashboard` work on direct visit in the
   production binary).
4. Other methods on unknown paths → plain 404.

**Client (React Router v6+):** static segments outrank dynamic params by
path-ranking automatically, provided marketing routes are registered in
the same route tree (not conditionally).

**Prerequisite for 5.2:** `/:documentId` gets a shape guard (UUID-like
pattern) so a typo'd static path renders NotFound instead of firing a
doomed fetch.

## 4. Indexing & Robots Policy

- **`/share/fig/*` and `/share/doc/*` are never indexed.** Enforced with
  `X-Robots-Tag: noindex, nofollow` response headers (works on CSR pages,
  unlike a meta tag which requires rendering). Share URLs never appear in
  sitemaps.
- Do **not** `Disallow: /share/` in robots.txt — blocking the crawl
  prevents the crawler from seeing the noindex signal; the URL could then
  still appear bare in results. Secret-link ≠ private (see the 2026
  Google Docs "anyone-with-the-link" indexing incident); noindex here is
  defense-in-depth on top of visibility rules, not permission.
- `robots.txt` ships with the first indexable page in 5.2: `Allow: /`,
  `Disallow: /api/`, plus a Sitemap reference once sitemaps exist (5.3).
- Soft-404 hygiene: the dispatcher above keeps API and asset misses out of
  the "200 shell for everything" failure mode.

## 5. Canonical & Host Policy

- Canonical form: `https://<apex>/<path>` — apex host, https, no trailing
  slash except namespace roots that define one (`/explain/`).
- The base URL comes from environment configuration when canonical tags or
  sitemap generation ship; no hardcoded domain anywhere in code or docs.
- Cache headers (prerequisite hardening in 5.2): hashed assets =
  long-lived immutable; `index.html` must be no-cache so deploys don't
  strand clients on stale shells referencing deleted bundles.

## 6. `/explain/` Semantics + Programmatic Quality Gates (5.3)

Reserved now, built in 5.3:

- Shape: `/explain/{slug}`, slug = kebab-case paper-derived identifier,
  stable once published.
- Default state: `X-Robots-Tag: noindex`. A page earns indexing only when
  **all** gates pass:
  1. Answers one distinct real question (not keyword-shaped)
  2. Unique value derived from actual source-paper data
  3. Provenance visible (title/authors/source link) — research-integrity rule
  4. Human-reviewable pipeline entry
  5. Count defensible page-by-page; stale pages refreshed or deleted
- Copyright gate: only material with an explicit share-permission basis
  becomes a public explain page. (The shared-paper payload already
  excludes raw original text.)
- Sitemaps include only gate-passed URLs; sitemap.xml generation belongs
  wholly to 5.3.

Context: Google's scaled-content-abuse policy has no volume threshold —
the test is intent + outcome ("created to manipulate rankings" vs
"genuinely useful"). AI generation itself isn't penalized; unoriginal
unhelpful-at-scale is.

## 7. Structured Data (minimal viable set)

| Where | Schema | Chunk |
|---|---|---|
| Home + landing pages | `WebSite`, `WebApplication`, `Organization` | 5.2 |
| Explain pages | `ScholarlyArticle` (+ `BreadcrumbList` where hierarchy exists) | 5.3 |

Skip `FAQPage`: rich results have been restricted to gov/health sites
since Aug 2023 and tooling support was removed in 2026 — zero reward.
Markup must mirror visible content only.

## 8. Decision Log

| Date | Decision | Why |
|---|---|---|
| 2026-08-25 | Prune `/scientific-figure-analysis`, `/research-paper-analysis` | One page per intent; cannibalization risk |
| 2026-08-25 | Crawlable pages = server-delivered full HTML | Zero-JS AI crawlers; social scrapers; WRS lag |
| 2026-08-25 | X-Robots-Tag (not meta, not robots.txt Disallow) for share URLs | Works on CSR; Disallow hides noindex from crawler |
| 2026-08-25 | True SPA fallback + API 404 JSON in dispatcher | Deep links broken in prod binary; API contract correctness |
| 2026-08-25 | Static-in-dist preferred over Go templates for 5.2 pages | Zero new infra; existing FileServer serves it naturally |
