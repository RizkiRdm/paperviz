# PRD.md — PaperViz (Agent-First)

> **Supersedes:** the pre-pivot version of this file (no-auth ephemeral MVP, undergrad-student-only persona) and the scope described in `PRODUCT.md`.
> **Last verified against code:** commit `68d213b`, 2026-09-12.
> **Status legend used throughout this doc:** `[SHIPPED]` = verified working in code · `[IN PROGRESS]` = code exists but incomplete/broken · `[PLANNED]` = not started. Do not upgrade a status without re-verifying against the actual repo — this file drifted from reality once already; the goal of the status tags is to make that harder to repeat.

---

## Project Summary

**Overview**
PaperViz is an academic-paper analysis tool whose primary product surface is now an **MCP server**, not a web dashboard. An AI agent (Claude Code, Cursor, Claude Desktop, etc.) installs PaperViz via MCP and calls tools to convert a paper into a simplified-language explanation, re-visualized charts with provenance, extracted claims, and evidence — all grounded against the source text. The web app is a thin human-facing shell around the three things a human still has to do by hand: sign in, pay, and (optionally) manually upload a paper the agent doesn't have.

**Objective**
Reduce the friction of extracting structured, verifiable understanding from academic papers — for an agent acting on a person's behalf, not for a person reading a dashboard.

**Value Proposition**
An agent (and, through it, the person directing the agent) gets a paper's claims, evidence, figures, and comparisons back as structured, grounded data — not a hallucinated summary. The differentiator vs. "just ask the LLM to summarize this" is the same as before the pivot: deterministic grounding validation (chart data must trace to source evidence, not be invented) and claim-diff verification (simplified text is checked against the original, not trusted blindly). What changed is *who* consumes that output first.

---

## Target Users

**Primary: AI agents acting on a user's behalf `[SHIPPED — MCP tools live]`**
- Consumes: `analyze_paper`, `get_summary`, `get_figures`, `get_claims`, `get_evidence`, `compare_papers` (see `docs/mcp-parity.md` for the current tool list — re-check that file against `internal/mcp/tools.go` before trusting it, per its own architecture rule).
- Input: pasted text only via MCP (not PDF — see Non-Goals). Stateless, no session, no access to any user's saved library by design.
- This is the interface almost all product usage should flow through going forward.

**Secondary: the human who installs and pays for it `[PARTIAL — auth broken, billing not started]`**
- Touches the product for exactly three things:
  1. **Auth** — sign in so the agent's calls can be attributed/rate-limited to an account. `[IN PROGRESS — see Known Gaps]`
  2. **Billing** — pay for usage above the free tier. `[PLANNED — Chunk 11.2, not started]`
  3. **Manual upload escape hatch** — paste/upload a paper directly through the web UI when not going through an agent. `[SHIPPED — /upload page functional independent of the pivot's broken pieces]`
- Everything else a human might have wanted from the old consumer-SaaS version (dashboard, saved-paper library, collections, comparison UI, pricing page) is **out of scope for this persona now** — see Non-MVP / Cut Features below.

**Not a target (explicitly, to prevent scope creep back into consumer SaaS):**
- A human browsing a dashboard of their past papers as the primary workflow. If `/account` ends up needing a papers list, that's a *utility* for the escape-hatch flow, not a product pillar to build features around.

---

## Problem Statement

Same underlying problem as the pre-pivot version — academic papers are written for peer researchers, imposing a comprehension tax on anyone else trying to use them (now: an agent trying to extract structured facts for its user, not a student reading directly) — but the delivery problem has changed:

1. **Grounding, not just simplification.** An agent handing a paper's claims to its user needs those claims verifiable against the source, or it's just moving the hallucination risk one hop downstream. This is *more* important for the agent-first audience than the original human-reader audience, because a human catches an obviously wrong chart by eye; an agent (and the person trusting it) may not.
2. **Zero-friction installation.** An agent user will not tolerate a multi-step manual setup. If installing PaperViz as an MCP server takes more than copy-pasting one config block with a pre-filled key, it loses to whatever the agent already has built in.

**Why it matters:** Same academic-integrity argument as before — agents are already being asked to "explain this paper" with no grounding and no persistent, checkable artifact. PaperViz's bet is that grounded, structured output is worth an agent (and its user) switching to a dedicated tool instead of asking the base model.

---

## Success Metrics

*(Carried forward from the pre-pivot PRD where still applicable; a few added for the agent-first surface. None of these have been re-validated against the current code — treat as targets, not measured results, until stated otherwise.)*

- Simplification job completes in **<60s** end-to-end (unchanged target; **not re-benchmarked** since the Gemini/SQLite concurrency limits were noted as a bottleneck — see `docs/ARCHITECTURE.md` §4a).
- Claim-diff verification runs on **100%** of jobs; a job that fails verification is never published silently.
- **New:** An agent can go from "reads the `/agents` install page" to "first successful tool call" with **zero manual MCP-terminology exposure** — no explaining what MCP is, no manual key copy-paste if the person is signed in. **This metric cannot currently be measured — `/agents` has no implementation yet.**
- **New:** MCP tool call p95 latency under concurrent agent load. **Not measured. Current backend serializes all Gemini calls (`internal/external/gemini.go`) and all SQLite access (`internal/repository/db.go`) — this metric will look bad until that's addressed, see `docs/ARCHITECTURE.md`.**

---

## Core Capabilities

### Shipped and grounded in code today
1. **Document ingestion** — PDF upload or pasted text via REST (`POST /api/documents`); pasted text only via MCP (`analyze_paper`).
2. **Simplification engine** — Gemini-based, reading levels (Simplified / ELI5), claim-diff verification against source.
3. **Chart re-visualization + grounding validator** — evidence extraction → dataset build → deterministic grounding rules (`internal/services/grounding.go`) → chart spec. Falls back to annotated original image when data extraction fails.
4. **Structured research objects** — claims, evidence, tables, methods, results, citations, evidence graph, research map (`GET /api/documents/:id/...` family — see `docs/openapi.yaml`).
5. **MCP server** (`internal/mcp`, `cmd/mcp`) — 6 tools, stateless, additive to REST, same service layer (`docs/mcp-parity.md`).
6. **Sharing** — ephemeral figure/paper share links (`/share/fig/:token`, `/share/doc/:token`), 7-day-style expiry model retained from the original product.
7. **Email/password auth** — signup/login/logout/session cookie (`internal/handlers/auth.go`: `Signup`/`Login`/`Logout`/`Me`), password complexity + rate limiting applied.

### In progress / broken as committed — do not build on top of these without fixing first
8. **Google OAuth auto-register (Chunk 11.1).** Handler code exists (`GoogleLogin`, `GoogleCallback`) but is **not registered in the router** and **references a `repository.User` struct/`UpsertByOAuth` method that don't exist**, and **no migration adds the columns it needs**. This does not compile as committed. See `PROJECT_REVIEW_REPORT.md` §3 for exact file:line evidence. **This blocks Chunk 11.2 and 11.3 — neither should start until this is actually fixed and verified end-to-end, not just claimed fixed.**
9. **`/agents` PNP install page (Chunk 11.3).** Route is declared in `App.jsx`; **the component file does not exist.** Zero percent implemented.
10. **`/account` page (Chunk 11.4).** Same situation — route declared, file missing.

### Explicitly planned, not started
11. **Stripe Checkout + webhook + Customer Portal (Chunk 11.2).** No Stripe code found anywhere in the repo (grep-confirmed). Correctly sequenced *after* OAuth — don't start early.
12. **MCP directory submissions (Chunk 11.8).** Do not submit to mcp.so / Smithery / Glama / awesome-mcp-servers until items 8–10 are actually working — a broken first impression on a discovery directory is expensive to undo.

---

## Non-MVP / Cut Features (agent-first scope reduction)

Per `AGENTS.md`, these routes were removed from `App.jsx`: `/dashboard`, `/pricing`, `/compare`, `/compare-research-papers`, `/research-paper-summarizer`, `/figure-explanation`, `/explain/:slug`. Kept: `/`, `/agents`, `/login`, `/signup`, `/account`, `/upload`, plus the two share routes.

**Frontend cleanup still owed (not yet done as of this audit):**
- `frontend/src/pages/{dashboard-page,compare-page,pricing-page,explain-page}.jsx` are still sitting in the repo, fully unreferenced. Delete them; don't just leave them unlinked.
- `upload-page.jsx`, `result-page.jsx`, and `upgrade-cta.jsx` still contain dead links to `/dashboard` and `/pricing`. These are live pages — this is a real, user-visible 404 bug today.

**Backend surface whose fate is undecided, not yet cut:**
These still exist, fully wired, with no current frontend consumer since their originating pages were removed:
- Research collections (`/api/collections/*`) — organize-saved-papers feature, tied to the old dashboard UX.
- Paper list/stats/save/rename (`GET /api/documents`, `/stats`, `PUT .../save`, `PATCH .../:id`) — dashboard-only by original design.
- Pricing/upgrade analytics tracking (`POST /api/analytics/pricing-view`, `/upgrade-intent`) — currently unreachable, since the page that fired these events no longer exists.
- Share-referral tracking (`POST /share-referrals`) — tied to the old viral-share growth loop, not an agent-first concern.

**Decision needed, not made by this document:** whether `/account` (once built) needs a "my papers" view, which would un-orphan the list/stats/save endpoints — or whether `/account` is auth + billing + API key display only, in which case collections, list, stats, and referral tracking should be deleted outright, not left running. This PRD does not make that call; it flags it as the open product decision blocking a clean cut.

---

## User Flows

### Primary Flow (agent-first): Install → Call → Trust
1. Person signs in at `/login` or `/signup` (or via Google OAuth — **not currently functional**).
2. Person visits `/agents` (**not currently implemented**) — sees a plug-and-play config block per client (Claude Code, Claude Desktop, Cursor, etc.) with their API key already filled in from their session, one test prompt, zero MCP jargon.
3. Agent calls `analyze_paper` with pasted text. Same simplification + grounding + verification pipeline as the web path runs underneath.
4. Agent calls `get_summary`/`get_figures`/`get_claims`/`get_evidence` as needed; `compare_papers` for multi-paper work.
5. All responses are grounded — figures carry `image_url`/provenance, claims carry `source_reference`, chart data traces to extracted evidence or is omitted rather than invented.

### Secondary Flow (human escape hatch): Manual Upload
1. Person visits `/upload` directly (works today, independent of the OAuth/`/agents`/`/account` breakage).
2. Uploads PDF or pastes text, selects reading level.
3. Gets a result page at `/:documentId` with simplified text, charts, claims, evidence — same as before the pivot.
4. Can share a figure or the whole document via ephemeral link.

### Failure Scenarios (unchanged from pre-pivot PRD, still enforced in code)
- Chart data extraction fails + image fallback fails → chart omitted with inline note, rest of document still delivered.
- LLM call fails/times out → retry once, then surface a clear error; no partial/corrupted result is published.
- Claim-diff mismatch detected → job flagged, not published silently as a clean result.

---

## High-Level Tech Stack

- **Backend:** Go, `chi` router, `modernc.org/sqlite` (no CGO), raw `database/sql`. `go.mod` currently pins `go 1.25.0`.
- **Frontend:** React 19 + Vite + Tailwind CSS v4 + shadcn/ui.
- **LLM:** Google Gemini API, direct HTTP integration (no gateway).
- **Agent interface:** `internal/mcp` (stdio MCP server, `cmd/mcp`), stateless, shares the service layer with REST.
- **Deployment:** single Go binary serving the built React SPA; no Docker/orchestration requirement.

---

## Technical Assumptions (updated for agent-first)

- **Concurrency profile has changed.** The original "solo-dev, low-traffic MVP, single-instance sufficient" assumption was written for sporadic human visitors. Agent traffic is programmatic and repetitive within a session (an agent may call 3-5 tools back-to-back analyzing one paper, or loop `compare_papers` across several). The current hardcoded Gemini semaphore of 1 and SQLite `SetMaxOpenConns(1)` were reasonable under the old assumption and are now a stated architectural risk — see `docs/ARCHITECTURE.md` §4a. Do not scale marketing/distribution (Chunk 11.8) ahead of fixing this, or the first real agent traffic will look broken.
- **No PII handling** — unchanged; papers assumed public/non-sensitive academic documents.
- **English-language source papers** — unchanged assumption.
- **MCP tools remain read-only and stateless by design** — per `AGENTS.md`'s agent-integration rules, do not add user-owned operations (list, save, delete, share) to MCP without a deliberate, documented decision to reverse that stance.

---

## Open Questions This PRD Does Not Resolve

1. Does `/account` need a papers list/history, or is it auth+billing+API-key only? (Determines fate of collections/list/stats backend surface.)
2. What's the actual API-key issuance flow for `/agents` prefill? `AGENTS.md` currently claims this exists (`/api/auth/apikey`); it does not, as of this audit. This needs a real design decision, not just a bug fix.
3. Rate-limit / usage-tier model for agent traffic specifically — the existing tier service (`internal/services` usage/tier code) was built for human-triggered document creation; confirm it makes sense for agent call patterns before billing (Chunk 11.2) goes live.