# goals/chunk-11-agent-first-pivot/plan.md

# Chunk 11 — Agent-First Pivot (Full Implementation)

**Context:** Final direction locked: PaperViz becomes agent-first, Context7-
style. Human touches the product for 3 things only — auth, billing, and an
optional manual-upload escape hatch. The MCP server (`internal/mcp`) is the
actual product surface. This supersedes the earlier draft of Chunk 10 —
10.1 is redefined below as the PNP install + SEO surface; the rest of
Chunk 10 (doc drift, route consolidation) is folded in as 11.6/11.7.

**Sequencing matters.** OAuth must ship before Stripe (checkout needs a
real `user_id`) and before the `/agents` page (key prefill needs a session).
Do not parallelize 11.1 with 11.2/11.3.

---

## 11.1 [P0] Google OAuth — login + auto-register

**Goal:** Replace manual signup friction with one-click Google login.
First-time Google login auto-creates the account — no separate signup step.

**What to change:**
- Add `golang.org/x/oauth2` + Google provider config (client ID/secret via env).
- New route `GET /api/auth/google/login` (redirect to Google consent screen),
  `GET /api/auth/google/callback` (exchange code, fetch profile, upsert user
  by email, issue the existing session cookie — reuse `internal/handlers/auth.go`
  session logic, do not build a second session system).
- DB: add `oauth_provider`, `oauth_id` columns to `users` table (new migration,
  nullable — existing password-based accounts unaffected).
- Frontend: replace signup+login page pair with a single `/login` page,
  one "Continue with Google" button. Keep existing email/password path as
  fallback only if you already have password-based accounts to support;
  otherwise drop it entirely per the page-cut list (11.5).

**Files affected:**
- `internal/handlers/auth.go` (add OAuth handlers, reuse session issuance)
- `internal/repository/users.go` (upsert-by-email logic)
- `migrations/017_oauth_columns.sql` (new)
- `frontend/src/pages/login-page.jsx` (rewrite, single button)
- `.env.example` (add `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`)

**Acceptance criteria:**
- [ ] First-time Google login creates a `users` row with no separate signup step
- [ ] Returning Google user logs in without creating a duplicate row (matched by email)
- [ ] Session cookie issued matches existing `HttpOnly+Secure+SameSite=Lax` config
- [ ] `go test ./internal/handlers/... ./internal/repository/...` passes

**Dependencies:** none — this is the foundation everything else needs.

---

## 11.2 [P0] Stripe — subscription billing

**Goal:** Real payment path. Replace the "coming soon" waitlist placeholder
with an actual checkout.

**What to change:**
- Add `stripe-go` (v79+). Use Stripe Checkout Session (hosted page) — do not
  build a custom card form, do not touch raw card data.
- New endpoint `POST /api/billing/checkout` — creates a Checkout Session for
  the logged-in `user_id`, redirects to Stripe.
- New endpoint `POST /api/billing/webhook` — handles `checkout.session.completed`
  and `customer.subscription.deleted`, updates `users.subscription_tier` /
  `subscription_status`. Verify Stripe signature — do not trust unsigned webhook bodies.
- New endpoint `GET /api/billing/portal` — generates a Stripe Customer Portal
  link so users manage/cancel their own subscription (don't build cancel/upgrade
  UI yourself, Stripe's portal already does this).
- DB: add `stripe_customer_id`, `subscription_tier`, `subscription_status` to `users`.

**Files affected:**
- `internal/handlers/billing.go` (new)
- `internal/repository/users.go` (tier/status fields)
- `migrations/018_billing_columns.sql` (new)
- `.env.example` (add `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, price IDs)

**Acceptance criteria:**
- [ ] Checkout Session redirect works end-to-end with a real Stripe test-mode key
- [ ] Webhook signature verified — reject unsigned/forged requests (test this explicitly)
- [ ] `users.subscription_tier` updates correctly on `checkout.session.completed`
- [ ] Customer Portal link opens and allows self-serve cancel

**Dependencies:** 11.1 (needs real `user_id` to attach subscription to)

---

## 11.3 [P0] `/agents` — PNP install + SEO discovery surface

**Goal:** Single page that replaces the 4 removed SEO pages as the primary
discovery surface, AND is the zero-friction install flow. This is the page
that makes the MCP server actually usable by a stranger.

**What to change:**
- Tab selector: Claude Code / Claude Desktop / Cursor / ChatGPT.
- Each tab renders a copy-paste-ready config block. If session exists, inject
  the real API key server-side; if not, show a placeholder + "Sign in with
  Google to get your key" CTA (links to 11.1's login).
- One test prompt shown below the config ("Try: analyze this paper [url]")
  so the user gets instant confirmation without reading further docs.
- SEO metadata targets: "MCP server", "Claude Code paper analysis", "academic
  research MCP tool" — this page carries the SEO weight the 4 removed pages
  used to.
- Mirror the same config blocks in `README.md` for GitHub-based discovery.

**Files affected:**
- `frontend/src/pages/agents-page.jsx` (new)
- `frontend/src/App.jsx` (new route `/agents`)
- `internal/handlers/auth.go` or new `internal/handlers/apikey.go` — endpoint
  to fetch/regenerate the logged-in user's API key (reuse existing key
  generation if it exists; check before writing new logic)
- `README.md`

**Acceptance criteria:**
- [ ] Logged-in user sees their real key pre-filled, zero manual copy-paste of the key itself
- [ ] Logged-out user sees a clear sign-in CTA, no broken/empty config block
- [ ] All 4 client tabs produce a config that works with zero manual edits
- [ ] Test prompt is present and copyable
- [ ] Page metadata (title/description) targets agent-integration keywords, not old SEO keywords

**Dependencies:** 11.1 (key prefill needs a session)

---

## 11.4 [High] `/account` — single-page account management

**Goal:** Replace `/dashboard` (sidebar + library UI) with one flat page:
API key, usage this month, subscription status. No nav, no sidebar.

**What to change:**
- New page: API key (regenerate button, reuses 11.3's endpoint), usage count
  this month (single number, reuse existing usage-tracking/fingerprint infra
  — don't build a new metering system), subscription tier + "Manage billing"
  link (opens 11.2's Customer Portal link).
- Remove sidebar navigation entirely — there is nothing left to navigate to.

**Files affected:**
- `frontend/src/pages/account-page.jsx` (new, replaces `dashboard-page.jsx`)
- `frontend/src/App.jsx` (route swap)
- `internal/handlers/billing.go` (usage summary endpoint, reuse existing usage tables)

**Acceptance criteria:**
- [ ] One page, no sidebar, no nested nav
- [ ] Shows real usage number pulled from existing tracking (not fabricated)
- [ ] "Manage billing" opens working Stripe Portal link

**Dependencies:** 11.1, 11.2

---

## 11.5 [High] Page cuts — remove 6, keep 5

**Goal:** Execute the page-inventory decision. This is deletion, not archival —
don't leave dead routes half-wired.

**What to change — REMOVE:**
- `/dashboard` (replaced by 11.4)
- `/pricing` (3-tier page → direct Stripe Checkout link from `/account` or landing CTA, no standalone page)
- `/compare` + `/compare-research-papers` web UI (keep `compare_papers` as an MCP tool only — do not delete the underlying service function, only the page/route)
- `/research-paper-summarizer`, `/figure-explanation`, `/explain/:slug` (3 SEO pages, zero measured traffic — see review from 2026-09-09)
- separate `/signup` page (merged into `/login` via OAuth, 11.1)

**KEEP as-is:** `/`, `/agents`, `/login`, `/account`, `/upload` (result page and share pages stay only if an agent-produced result still needs a shareable link — confirm with 11.3 usage before deciding, don't remove preemptively without checking if MCP tool outputs reference share URLs)

**Files affected:**
- `frontend/src/App.jsx` (remove routes)
- Delete: `dashboard-page.jsx`, `pricing-page.jsx`, `compare-page.jsx` (web entry only, keep service), `explain-page.jsx`, and the two SEO-only page components
- `internal/handlers/router.go` — remove any backend routes that only existed to serve now-deleted pages (check before deleting — some may still be used by `/upload` or MCP)

**Acceptance criteria:**
- [ ] No dead links anywhere in the remaining 5 pages pointing to removed routes
- [ ] `grep -rn "dashboard-page\|pricing-page" frontend/src` returns nothing
- [ ] Underlying `compare_papers` service function still passes its existing tests — only the web route is gone

**Dependencies:** 11.3, 11.4 (need replacements live before deleting what they replace)

---

## 11.6 [Medium] Minimal landing page rebuild

**Goal:** Landing page becomes a 1-line pitch + two CTAs, not a feature tour.

**What to change:**
- Single above-the-fold section: what PaperViz does in one sentence, "Add to
  Claude Code" button (→ `/agents`), "Sign in" link (→ `/login`).
- No feature carousel, no demo video, no pricing table on this page.

**Files affected:**
- `frontend/src/pages/upload-page.jsx` or wherever `/` currently renders — repurpose, don't build from scratch if an existing component is close

**Acceptance criteria:**
- [ ] Page has exactly 2 CTAs above the fold
- [ ] No reference to removed pages (pricing table, dashboard preview, etc.)

**Dependencies:** 11.3 (CTA target must exist)

---

## 11.7 [Medium] Fix MCP doc/code drift (carried from Chunk 10.2)

**Goal:** `goals/chunk-7-4-mcp/plan.md` and `docs/mcp-parity.md` still list
`get_tables` and `search_papers`, neither implemented in `internal/mcp/tools.go`.

**What to change:** Strike both from docs, or implement `get_tables` by
wiring the existing `GetTables` handler logic into MCP (cheap reuse).
Leave `search_papers` parked — no product reason yet under the agent-first
model either, since there's still no per-user library exposed to MCP.

**Files affected:** `docs/mcp-parity.md`, `goals/chunk-7-4-mcp/plan.md`, `internal/mcp/tools.go` (only if implementing)

**Acceptance criteria:**
- [ ] Tool count in docs matches `internal/mcp/tools.go` exactly

**Dependencies:** none

---

## 11.8 [Low, post-launch] Discovery submissions

**Goal:** Get `/agents` listed where agent-tool users actually look.

**What to change:** Submit to MCP registry/directory (Anthropic official if
available), mcp.so, Smithery, Glama; open one PR to an awesome-mcp-servers
list; tag the GitHub repo with `mcp-server`, `model-context-protocol` topics.

**Files affected:** none (external submissions), `README.md` topics/badges

**Acceptance criteria:**
- [ ] Listed on at least 2 of the directories above within a week of `/agents` going live

**Dependencies:** 11.3 must be live first