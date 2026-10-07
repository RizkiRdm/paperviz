---
report_type: e2e_qa_audit
schema_version: 1
report_id: PV-QA-2026-09-28
date: 2026-09-28
project: PaperViz
base_url: http://localhost:6767
tester: automated agent (Playwright suite + live browser + API probes)
verdict: NOT_RELEASE_READY
release_gate: BLOCK
critical_count: 2
high_count: 5
medium_count: 5
low_count: 4
finding_ids:
  - QA-001
  - QA-002
  - QA-003
  - QA-004
  - QA-005
  - QA-006
  - QA-007
  - QA-008
  - QA-009
  - QA-010
  - QA-011
  - QA-012
  - QA-013
  - QA-014
  - QA-015
  - QA-016
---

# PaperViz E2E QA Report — 2026-09-28

> **Read this first if you are an AI agent.**
> Findings use stable IDs (`QA-001` … `QA-016`). Every finding has a fixed field
> order: `severity`, `confidence`, `location`, `symptom`, `expected`, `actual`,
> `root_cause`, `reproduce`, `impact`, `fix`, `regression_test`.
> `confidence: verified` means reproduced against a running server in this session.
> `confidence: inferred` means read from code but not executed. `confidence: unverified`
> means a real blocker stopped the test — see "Coverage Gaps".
> Do not re-derive these findings from the code; they are confirmed. Act on the
> `fix` field. Prioritize strictly in `severity` order, then `QA-001` and `QA-002`
> before anything else — they are the release blockers.

---

## 1. Verdict

**NOT release-ready.** Two Critical defects each eliminate a named product
capability outright:

- **QA-001** — cross-user document tampering (security; data isolation lost)
- **QA-002** — charts never render (the feature the landing page sells)

Both fixes are small and localized (`QA-001` = thread a `userID` parameter;
`QA-002` = one `JSON.parse`). The critical insight for whoever picks this up:
**every CI gate is currently green and both bugs still ship.** These are
contract and authorization gaps that no existing gate tests.

---

## 2. Environment

| Field | Value |
|---|---|
| Base URL | `http://localhost:6767` (`PORT` in `.env`) |
| Backend | Go 1.25, `./server` binary rebuilt from `HEAD` for this audit |
| Frontend | `frontend/dist`, freshly built with `npm --prefix frontend run build` |
| Database | `paperviz.db` (SQLite + WAL) |
| Model credential | Google Gemini key supplied BYOK, stored via `/account` |
| Test account A | `qa1@example.com` (deleted after audit) |
| Test account B | `qa2@example.com` (deleted after audit) |

### 2.1 Required to reproduce

```bash
# 1. The checked-in .env is pre-BYOK and lacks the required key. The server
#    refuses to boot without it. Generate one and persist it:
openssl rand -base64 32 > /tmp/pv-enc-key

# 2. Build frontend + backend
npm --prefix frontend run build
CGO_ENABLED=0 go build -o server ./cmd/server

# 3. Run
set -a; source .env; set +a
export CREDENTIAL_ENCRYPTION_KEY="$(cat /tmp/pv-enc-key)"
./server &
curl -s http://localhost:6767/healthz   # expect {"status":"ok"}
```

> **Caution — read before reproducing.** On first boot the server runs a
> startup expiry sweep that permanently deletes documents idle > 7 days
> (see `QA-010`). Copy `paperviz.db` first if you need the existing data.

### 2.2 State changes made by this audit

- A startup expiry sweep deleted 6 pre-existing documents (`documents_deleted: 6`).
- Test fixture document `QAFIXTURE01` and both test accounts were deleted afterwards.
- Orphaned `user_credentials` rows were removed.
- **No source file was modified. No commit was made.**

---

## 3. Findings index

Severity legend: `CRITICAL` = capability/security total loss · `HIGH` = core
flow broken or system-wide · `MEDIUM` = degraded UX or inconsistency ·
`LOW` = cosmetic or hygiene.

| ID | Severity | Confidence | Area | One-line summary |
|---|---|---|---|---|
| QA-001 | CRITICAL | verified | security / authz | Any authenticated user can rename, delete, and annotate another user's document (IDOR) |
| QA-002 | CRITICAL | verified | frontend / data-contract | Charts never render — `chart_data` is a JSON string, frontend treats it as an object |
| QA-003 | HIGH | verified | product / agent onboarding | `/agents` publishes MCP config for a nonexistent `/api/mcp` endpoint |
| QA-004 | HIGH | verified | sharing | Share links render raw JSON; both share React pages are unreachable dead code |
| QA-005 | HIGH | verified | config / LLM | Default Gemini model `gemini-2.5-flash-lite` is retired; default flow hard-fails |
| QA-006 | HIGH | verified | security headers / design system | CSP blocks the brand font on every page |
| QA-007 | HIGH | verified | onboarding / copy | Anonymous path is a dead end and leaks the raw `missing credential` error code |
| QA-008 | MEDIUM | verified | responsive | Result page overflows horizontally at 375px (568px content) |
| QA-009 | MEDIUM | verified | error handling | Bad DOI returns `502 fetch_failed` instead of a 4xx; credential-check ordering inconsistent |
| QA-010 | MEDIUM | verified | data durability | Startup expiry sweep deletes documents with no confirmation or disable flag |
| QA-011 | MEDIUM | verified | account UX | No edit path for an existing model credential |
| QA-012 | MEDIUM | verified | UX | ~66s of silent retrying behind a static "Reading document…" label |
| QA-013 | LOW | verified | ops / health | Stale worktree binaries can silently break `/healthz` |
| QA-014 | LOW | verified | dead content | Deleted SEO pages still served from `frontend/public/` |
| QA-015 | LOW | verified | config hygiene | `.env` is pre-BYOK: contains removed `GOOGLE_*`/`STRIPE_*`, lacks required encryption key |
| QA-016 | LOW | verified | error recovery | A credential AAD mismatch yields `500 credential unreadable` with no recovery path |

---

## 4. Detailed findings

### QA-001 — Cross-user IDOR: rename, delete, and annotate any document

- **severity:** CRITICAL
- **confidence:** verified
- **location:** `internal/handlers/documents.go:195-225` (`UpdateTitle`, `Delete`); `internal/handlers/annotations` create path
- **symptom:** Authenticated user B mutates user A's document by ID. No ownership check on write.
- **expected:** `403` for any cross-user document write.
- **actual:** `PATCH` → `200` and title overwritten; `DELETE` → `200` and document destroyed. `POST .../annotations` writes into A's document, but A's annotation list stays `[]` — the write is silent and invisible to the owner.
- **root_cause:** `services.DeleteDocument(h.db, id)` is called with no user identity at all:
  ```go
  // internal/handlers/documents.go:222
  if err := services.DeleteDocument(h.db, id); err != nil {
  ```
  The router applies `authMiddleware.RequireAuth` (authentication) but no ownership
  check (authorization). `UpdateTitle` has the same gap.
- **reproduce:**
  ```bash
  # B = attacker account, A = victim
  curl -X POST localhost:6767/api/auth/signup -H 'Content-Type: application/json' \
    -d '{"email":"b@example.com","password":"Str0ngPassw0rd!"}' -c B.jar

  # B renames A's document
  curl -X PATCH localhost:6767/api/documents/<A_DOC_ID> \
    -H 'Content-Type: application/json' -d '{"title":"HIJACKED"}' -b B.jar
  # observed: 200 {"title":"HIJACKED"}   — expected: 403

  # B deletes A's document
  curl -X DELETE localhost:6767/api/documents/<A_DOC_ID> -b B.jar
  # observed: 200, row gone — expected: 403

  # B writes into A's document
  curl -X POST localhost:6767/api/documents/<A_DOC_ID>/annotations \
    -H 'Content-Type: application/json' \
    -d '{"target_type":"paper","target_id":"<A_DOC_ID>","content":"injected"}' -b B.jar
  # observed: 201 — expected: 403
  ```
- **impact:** Any account can destroy or deface any other account's data. Document IDs are 15-char opaque strings, so this is not brute-force resistant — they leak in URLs, share links, and logs. Combined with the invisible annotation write, an attacker can plant content in another user's research notes.
- **contrast:** Collections **do** enforce ownership — `PATCH /api/collections/{id}` as user B returns `{"error":"forbidden"}`. Annotations and collections hardening (D1/D3 in `AGENTS.md`) is real; the documents service was simply never covered. This is the precise gap.
- **fix:** Resolve `userID` from the session and pass it into `DeleteDocument`, `UpdateTitle`, and the annotation write. Reject with `403` when `documents.user_id` does not match. Reuse the pattern already working in `internal/handlers/collections.go`.
- **regression_test:** For every authenticated write route under `/api/documents`, assert a second user gets `403` and that the victim row is unchanged. Also assert the owner still succeeds — a blanket-403 mistake would otherwise pass.

### QA-002 — Charts can never render (headline feature dead)

- **severity:** CRITICAL
- **confidence:** verified
- **location:** `internal/services/charts_llm.go:248` · `internal/handlers/documents.go:254` · `frontend/src/components/chart-card.jsx:244` · `frontend/src/components/data-chart.jsx:22-26`
- **symptom:** Every chart on a completed document renders the red text `Chart data is missing labels or values.` Zero Recharts instances mount.
- **expected:** Bar / line / scatter / pie chart renders with the extracted values.
- **actual:** Red error text; `document.querySelectorAll('.recharts-wrapper').length === 0`.
- **root_cause:** A backend/frontend contract mismatch. The backend persists `chart_data` as a **JSON-encoded string**; the frontend treats it as an object.
  ```go
  // charts_llm.go:248 — marshalled to bytes, stored in a TEXT column
  chartData := chartDataJSON{Labels: ds.Labels(), Values: ds.Values(), Title: plan.Title}
  dataRaw, err := chartData.marshalJSON()
  ```
  ```go
  // handlers/documents.go:254 — declared as a string
  ChartData *string `json:"chart_data"`
  ```
  ```jsx
  // chart-card.jsx:244 — string handed straight to the renderer
  <LazyDataChart chartData={chart.chart_data} ... />

  // data-chart.jsx:22 — read as an object
  const labels = chartData.labels   // undefined, because chartData is "..."
  ```
  `frontend/src/components/ui/status-banners.jsx` contains the only `JSON.parse`
  in the codebase. The chart render path has none.
- **reproduce:**
  ```bash
  # Store the EXACT shape the Go backend produces (a JSON string):
  sqlite3 paperviz.db "UPDATE charts SET chart_data='{\"labels\":[\"Baseline\",\"Restricted\"],\"values\":[245,412],\"title\":\"Reaction time\",\"chart_type\":\"bar\"}' WHERE id='<chart_id>';"

  # Confirm the API returns a string, not an object:
  curl -s -b A.jar localhost:6767/api/documents/<A_DOC_ID> \
    | python3 -c "import json,sys; print(type(json.load(sys.stdin)['charts'][0]['chart_data']).__name__)"
  # observed: str

  # Open the page — error text is present, no chart renders.
  ```
- **impact:** 100% of charts, on both the main result page and the shared-figure page. The landing meta description advertises "re-visualized charts"; the product's differentiator per `README.md` ("evidence-grounded figure re-visualization") produces nothing but an error string.
- **verification_note:** This was confirmed with the correct payload, not assumed. Two shapes were initially suspected: `models.ChartSpec` (`type` / `data.categories` / `data.series`) and the actually-persisted `chartDataJSON` (`labels` / `values` / `title`). `ChartSpec` is **not** what gets stored and is a decoy for future readers. The finding holds against the correct shape.
- **fix:** Parse at the boundary — `const parsed = typeof chartData === "string" ? JSON.parse(chartData) : chartData` in `data-chart.jsx`, with a `try/catch` returning the existing degraded message on malformed JSON. Alternatively change the handler to emit `json.RawMessage`, but that is a breaking API change; the frontend fix is smaller.
- **regression_test:** Seed one chart with each of the four types using the string-encoded shape and assert a `.recharts-wrapper` mounts. Also assert a genuinely malformed payload degrades to the message rather than throwing.

### QA-003 — Primary CTA publishes a config for a nonexistent endpoint

- **severity:** HIGH · **confidence:** verified
- **location:** `frontend/src/pages/agents-page.jsx` · `internal/handlers/router.go`
- **symptom:** `/agents` tells users to paste a config pointing at `https://paperviz.com/api/mcp?key=…`. No `/api/mcp` route exists — `router.go` registers only a stdio MCP binary (`cmd/mcp`).
- **expected:** Either a working HTTP MCP transport, or directions for the stdio config the repo actually ships.
- **actual:** Copy reads "One config block. Paste it in your client's MCP settings and you're done." Copying it yields an agent that cannot connect.
- **root_cause:** Product surface and shipped implementation diverged during the agent-first pivot. `README.md` documents the gap under "Current Limitations"; the UI does not.
- **reproduce:** `curl -s -o /dev/null -w "%{http_code}" http://localhost:6767/api/mcp` → `404` (the SPA fallback serves `index.html` for non-`/api` paths, and this path is absent from `router.go`). Confirm the URL emitted by `/agents` matches.
- **impact:** The landing page's primary CTA ("Add to Claude Code", the only above-the-fold action alongside "Sign in") routes to a dead end. The agent-first thesis has no working entry point.
- **fix:** Ship `/api/mcp` or repoint `/agents` at the stdio config. Add a note that the remote endpoint is unavailable.
- **regression_test:** Assert the URL emitted by `/agents` is a route present in the router; fail if it 404s.

### QA-004 — Share links render raw JSON; share pages are dead code

- **severity:** HIGH · **confidence:** verified
- **location:** `internal/handlers/router.go:141-144` vs `frontend/src/App.jsx:29-30`
- **symptom:** Opening `/share/doc/<token>` or `/share/fig/<token>` in a browser displays the raw JSON payload as page text.
- **expected:** The React `SharePaperPage` / `ShareFigurePage` render.
- **actual:** Body text begins `{"document_id":"QAFIXTURE01","title":…`. The React components can never mount — the Go routes shadow the SPA at the same paths.
- **root_cause:** Path collision between the JSON API routes and the client routes, resolved server-side first.
- **reproduce:**
  ```bash
  curl -s -X POST -b A.jar localhost:6767/api/documents/<doc_id>/share   # -> {"share_url":"/share/doc/<token>"}
  playwright-cli goto http://localhost:6767/share/doc/<token>
  playwright-cli eval "document.body.innerText.slice(0,80)"
  # observed: {"document_id":"...","title":"..."   — raw JSON, not the React page
  ```
- **impact:** Sharing — a stated core loop ("shareable in seconds") — produces links recipients cannot read.
- **positively_verified:** The share payload is clean: no `original_text`, no `user_id` (copyright/privacy controls hold). Revocation correctly returns `404`. Bogus tokens return `404`. Only presentation is broken.
- **fix:** Move JSON endpoints under `/api/share/...` and let the SPA own `/share/...`, or have the share routes serve `index.html` and let the client fetch from an `/api` path.
- **regression_test:** Assert the share URL returns `text/html` for a browser and that a share page renders its heading.

### QA-005 — Default Gemini model is retired; the default flow hard-fails

- **severity:** HIGH · **confidence:** verified
- **location:** `internal/external/llm.go:47`
- **symptom:** A user who adds a Gemini key without naming a model gets `simplification_failed`. Provider error: `This model models/gemini-2.5-flash-lite is no longer available to new users.`
- **expected:** A working current model.
- **actual:** `func (p Provider) DefaultModel()` returns `gemini-2.5-flash-lite` for the default case. The pipeline fails on the very first call.
- **root_cause:** A model default went stale with no test or live probe guarding it. Unit tests use `gemini-3.5-flash`, so `DefaultModel()` was never exercised.
- **reproduce:**
  ```bash
  # Add a Gemini key via /account, leaving "Model (optional)" empty, then ingest any text.
  # Server log:
  #   {"error":"simplify text: This model models/gemini-2.5-flash-lite is no longer available to new users.",
  #    "message":"pipeline stage failed","stage":"simplify"}
  # Document row: status=failed, error_message=simplification_failed
  ```
  Recovering requires manually re-adding the key with an explicit model (see `QA-011`).
- **impact:** The out-of-the-box path for the most likely provider is broken. Recovering requires the user to know to type a model name — which the UI does not tell them (see `QA-007`).
- **related_risk:** The sibling defaults `claude-sonnet-4-6` (Anthropic) and `gpt-5.5` (OpenAI) are equally unverified strings. They were **not** executed — no keys were available. Treat all three as unverified.
- **fix:** Update all three defaults to currently-available models and add a test asserting `DefaultModel()` returns a model that is not a known-retired string.
- **regression_test:** Requires a live probe or a recorded-fixture test per provider; a pure unit test cannot catch provider-side retirement.

### QA-006 — CSP blocks the brand font on every page

- **severity:** HIGH · **confidence:** verified
- **location:** `internal/handlers/security_headers.go:14` vs `frontend/index.html:12`
- **symptom:** Three CSP console errors on every page load. The Satoshi brand font never renders; all pages fall back to a system font.
- **expected:** `DESIGN.md` typography loads as specified.
- **actual:** Browser reports: `Loading the font 'https://cdn.fontshare.com/…woff2' violates … directive: "font-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com https://api.fontshare.com". The action has been blocked.`
- **root_cause:** CSP allows `api.fontshare.com` (which serves the CSS) but not `cdn.fontshare.com` (which serves the font binaries).
- **reproduce:** Open any route and read the console. Three errors, one per font format:
  ```bash
  playwright-cli goto http://localhost:6767/
  playwright-cli eval "document.fonts.size"   # 0 loaded, or only fallback
  # console: [ERROR] Loading the font 'https://cdn.fontshare.com/.../7AHD....woff2'
  #          violates the following Content Security Policy directive: "font-src 'self'
  #          https://fonts.googleapis.com https://fonts.gstatic.com https://api.fontshare.com".
  ```
  Confirm with `curl -sI http://localhost:6767/ | grep -i content-security-policy` and compare against the blocked host.
- **impact:** The entire design system's typography is dead in production. The defect is invisible in review because the `<link>` tag points at the allowed host.
- **fix:** Add `https://cdn.fontshare.com` to the `font-src` directive.
- **regression_test:** Assert no console error containing `violates the following Content Security Policy` on any page.

### QA-007 — Anonymous path is a dead end and leaks a raw error code

- **severity:** HIGH · **confidence:** verified
- **location:** `frontend/src/pages/upload-page.jsx` (copy + error surface)
- **symptom:** Landing says "Free for researchers. No account required to try." Submitting pasted text as an anonymous user shows a red alert reading literally `missing credential`, with no sign-up CTA and no explanation.
- **expected:** Either a working anonymous path, or copy that states an account plus a model key is required, with a link to create one.
- **actual:** Raw snake_case API code rendered to end users. `README.md` confirms uploads without a usable key always fail with `missing_credential`.
- **root_cause:** Landing copy predates the BYOK pivot; the error surface renders the API code rather than mapped copy. Compare `result-page.jsx:18-22`, which maps codes to human sentences — the upload page has no such map.
- **reproduce:**
  ```bash
  # Log out (or use a fresh context), then:
  playwright-cli goto http://localhost:6767/
  playwright-cli click "[role=tab]:has-text('Paste')"
  playwright-cli fill textarea "Sleep deprivation study. Reaction time rose from 245 ms to 412 ms."
  playwright-cli click "button:has-text('Simplify paper')"
  playwright-cli eval "document.querySelector('[role=alert]').textContent"
  # observed: missing credential
  ```
  Server side: `curl -s -X POST localhost:6767/api/documents -d '{"source_type":"pasted_text",...}'` → `400` with `missing_credential`.
- **impact:** First-contact users are told the product is free and account-free, then hit an uninterpretable dead end. Both the landing CTA and the `/account` "Get started with PaperViz →" link point at `/agents` (`QA-003`), so there is no working path to first value.
- **fix:** Correct the copy; map the error code to human copy plus a "Create an account" action; consider failing fast before upload when no credential is configured.
- **regression_test:** Assert anonymous submit produces copy containing a next step, and that it matches no `/^[a-z_]+$/` API code.

### QA-008 — Result page overflows horizontally at 375px

- **severity:** MEDIUM · **confidence:** verified
- **location:** `frontend/src/components/result/result-header.jsx:27-28`
- **symptom:** `document.documentElement.scrollWidth === 568` at `innerWidth === 375` — 193px of horizontal scroll on the most content-heavy page.
- **expected:** No horizontal overflow at 375px.
- **actual:** Culprit is a 280px `div.flex.flex-col.items-end.gap-1` containing the Verified badge, a visibility `<select>`, and the Share button, inside a `div.flex.items-center.gap-3`. Zero `flex-wrap` in the file; no responsive breakpoints on that row.
- **root_cause:** The result header's action row was laid out for a desktop width with no `flex-wrap` and no responsive variant, so its intrinsic width (280px) plus page padding exceeds a 375px viewport.
- **reproduce:**
  ```bash
  playwright-cli resize 375 812
  playwright-cli goto http://localhost:6767/<doc_id>
  playwright-cli eval "document.documentElement.scrollWidth + ' / ' + innerWidth"
  # observed: 568 / 375
  ```
- **impact:** Every phone user must scroll horizontally to reach the visibility selector and Share button — the page's primary actions. The result page is the product's main output surface.
- **fix:** Add `flex-wrap` and responsive stacking, or collapse the actions into an overflow menu below `sm`.
- **regression_test:** Assert `scrollWidth <= innerWidth` at 375px on every route.

### QA-009 — Bad DOI returns 502; credential-check ordering is inconsistent

- **severity:** MEDIUM · **confidence:** verified
- **location:** `internal/handlers/import.go:114,165`
- **symptom:** A nonexistent DOI returns `502 {"error":"fetch_failed"}`. The server log shows the true cause: `DOI not found in external registry` — a client input error, not a bad gateway.
- **expected:** A 4xx with clear copy, e.g. "We couldn't find that DOI."
- **actual:** `502`, indistinguishable from an upstream outage.
- **root_cause:** `import.go:114` maps every failure of the DOI fetch to `StatusBadGateway` without distinguishing "not found in the registry" (a client error) from "registry unreachable" (a genuine 502). A second, independent ordering defect: the DOI path performs the external fetch before the credential check, while the URL path checks the credential first.
- **reproduce:**
  ```bash
  curl -s -X POST localhost:6767/api/import/doi -H 'Content-Type: application/json' \
    -d '{"doi":"10.1000/xyz"}'
  # observed: 502 {"error":"fetch_failed"}
  # log:      {"message":"fetch by DOI failed","error":"...DOI not found in external registry"}
  ```
  Ordering: compare the same call authenticated-without-credential on each path — DOI reaches the network and returns 502; URL returns `missing_credential` without a fetch.
- **impact:** A user who mistypes a DOI is told the service is broken rather than that the identifier is wrong, and will retry against a fault that cannot clear. The inconsistent ordering also wastes a registry request for users who cannot proceed anyway.
- **secondary:** The DOI path performs the external fetch **before** the credential check, while the URL path checks the credential first (it returns `missing_credential`). Inconsistent ordering means one path burns a registry request for a user who cannot proceed anyway.
- **fix:** Map "not found in registry" to `404`/`400` with dedicated copy. Perform the credential check before the network call in both paths.
- **regression_test:** Assert a well-formed but unknown DOI yields 4xx, not 502.

### QA-010 — Startup expiry sweep deletes data with no confirmation

- **severity:** MEDIUM · **confidence:** verified
- **location:** `internal/services/expiry.go:58`
- **symptom:** First server boot logged `{"message":"expiry sweep","documents_deleted":6}`. Six documents were permanently destroyed during a routine audit startup.
- **expected:** Either a confirmation gate, or a documented and configurable window, for an action that is irreversible at boot.
- **actual:** `RunExpirySweepLoop` calls `sweepOnce` immediately, deleting everything idle > 7 days. No env var disables it, no dry-run exists.
- **root_cause:** The sweep is unconditional and runs before the HTTP server is ready, so it is unavoidable for anyone starting the binary. The deletion is logged only at `INFO` with a count — no cutoff timestamp, no document IDs, no WARN level.
- **reproduce:**
  ```bash
  # Seed any document with last_accessed_at older than 7 days, then start the server:
  sqlite3 paperviz.db "UPDATE documents SET last_accessed_at = strftime('%s','now') - 8*86400;"
  ./server
  # first log line: {"message":"expiry sweep","stage":"expiry_sweep","documents_deleted":<N>}
  sqlite3 paperviz.db "SELECT count(*) FROM documents;"   # rows gone, no prompt
  ```
  This is exactly what happened on the first boot of this audit: 6 documents destroyed.
- **impact:** Irreversible data loss triggered by an ordinary startup. Directly relevant to a solo-maintainer project.
- **fix:** Add `EXPIRY_SWEEP_ENABLED` (default true) and log at WARN with the count and cutoff, so a mass deletion is never silent.
- **regression_test:** Assert the sweep respects the flag and that a disabled sweep deletes nothing.

### QA-011 — No edit path for an existing model credential

- **severity:** MEDIUM · **confidence:** verified
- **location:** `frontend/src/pages/account-page.jsx`
- **symptom:** An existing credential offers only "Remove" and "Default". Changing a model requires deleting the key and re-pasting the secret.
- **expected:** Edit the model in place.
- **actual:** Remove + re-add. The API key field must be re-entered from scratch, so the user must keep the provider secret at hand.
- **root_cause:** `GET /api/credentials` returns only `provider`, `model`, and `key_hint` (correctly — never the plaintext), and there is no update route, so `model` is immutable after creation. The account page renders only Default and Remove per credential.
- **reproduce:**
  ```bash
  # Add a key, then revisit /account and inspect the rendered credential row:
  playwright-cli goto http://localhost:6767/account
  playwright-cli snapshot    # shows "Default" and "Remove" only; no edit control
  curl -s -X PATCH localhost:6767/api/credentials/<id> -b A.jar -d '{"model":"other"}'
  # observed: no such route (405/404) — PATCH is not registered
  ```
- **impact:** Directly worsened `QA-005`: a user who hits the retired-model error has no discoverable way to fix it without destroying and re-entering their key.
- **fix:** Add a `PATCH /api/credentials/{id}` that updates `model` only, never returning the plaintext key.
- **regression_test:** Assert the model can be changed without resubmitting the key.

### QA-012 — Silent retrying for ~66s

- **severity:** MEDIUM · **confidence:** verified
- **location:** `frontend/src/pages/result-page.jsx:145-166` · `internal/external/llm.go` retry schedule
- **symptom:** On a provider outage the UI shows a static "Reading document…" for ~66 seconds across 5 retry attempts, then fails. The retry logic itself is correct.
- **expected:** Indication that a retry is in progress, or elapsed time.
- **actual:** No change in copy or state until terminal failure. The user cannot tell a working pipeline from a stalled one.
- **root_cause:** The document status is a single coarse `processing` value with no sub-state for retry attempts, so the frontend has nothing to render during backoff. The retry schedule itself (5 attempts, ~66s total) is correct.
- **reproduce:** Point the credential at a provider that returns 503, then submit a document:
  ```bash
  playwright-cli goto http://localhost:6767/<doc_id>
  # observe: heading stays "Reading document..." unchanged for ~66s,
  #          then flips to "We couldn't simplify this paper. Please try again."
  # server log: 5 "llm call" lines with attempt 1..5, all success=false
  ```
- **impact:** The longest common failure path looks like a hang. Users may cancel or resubmit, compounding provider load during exactly the outage that made it slow.
- **fix:** Expose attempt/elapsed state, or vary the stage label during retries.
- **regression_test:** Assert the label changes across a multi-attempt failure.

### QA-013 — Stale worktree binaries silently break `/healthz`

- **severity:** LOW · **confidence:** verified
- **location:** worktree root (`server`, `paperviz`)
- **symptom:** A stale `./server` binary served the SPA `index.html` for `/healthz` with `200 OK`, instead of `{"status":"ok"}`.
- **expected:** `GET /healthz` returns `{"status":"ok"}`.
- **actual:** `Content-Type: text/html`, `Content-Length: 1042`. Health checks read as healthy while `healthzHandler` was never reached.
- **root_cause:** `router.go:78` correctly registers `/healthz`. The checked-in binary predated that route. Both `server` and `paperviz` are gitignored and not committed, so this is a local-hygiene trap rather than a repo defect.
- **reproduce:**
  ```bash
  # With a prebuilt ./server older than the healthz route:
  curl -si http://localhost:6767/healthz | head -6
  # observed: HTTP/1.1 200 OK / Content-Type: text/html / Content-Length: 1042
  # after `go build -o server ./cmd/server`:
  # observed: {"status":"ok"}
  ```
- **impact:** Low in production (the build is fresh); moderate for anyone running against a prebuilt artifact, since a monitor would report healthy during an outage.
- **fix:** None required in code. Note in `docs/maintainer-guide.md` that `make build` must precede `make run`.
- **regression_test:** Assert `/healthz` returns `application/json` containing `"status":"ok"`, so a SPA fallback can never be mistaken for a healthy backend.

### QA-014 — Deleted SEO pages are still served

- **severity:** LOW · **confidence:** verified
- **location:** `frontend/public/explain/`, `compare-research-papers.html`, `figure-explanation.html`, `research-paper-summarizer.html`
- **symptom:** These routes were removed from `App.jsx` but all still return `200` in production.
- **expected:** 404, since the routes are gone.
- **actual:** Live content for four removed pages.
- **context:** `README.md` acknowledges this under "Current Limitations". They are static files bypassed by the SPA router, so the `spaNotFound` handler serves them directly.
- **root_cause:** `frontend/public/` is copied verbatim into `frontend/dist/` on every build. Removing a route from `App.jsx` does not remove a matching static file, and `spaNotFound` serves any existing file under `STATIC_DIR` before falling back to the SPA shell.
- **reproduce:**
  ```bash
  for p in /explain/sleep-quality-executive-function.html /compare-research-papers.html \
           /figure-explanation.html /research-paper-summarizer.html; do
    printf "%s %s\n" "$(curl -s -o /dev/null -w '%{http_code}' http://localhost:6767$p)" "$p"
  done
  # observed: 200 for all four — none are routes in App.jsx
  ```
- **impact:** Stale marketing content conflicting with current product claims; SEO surface for deleted pages.
- **fix:** Delete the files and rebuild.
- **regression_test:** Assert no file in `frontend/public/` maps to a route absent from `App.jsx`.

### QA-015 — `.env` is pre-BYOK

- **severity:** LOW · **confidence:** verified
- **location:** `.env`
- **symptom:** Contains `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`, `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_PRO`, `STRIPE_PRICE_RESEARCH` — all for features deleted on 2026-09-26 — and **no** `CREDENTIAL_ENCRYPTION_KEY`.
- **expected:** `.env` matches `.env.example`, which documents `CREDENTIAL_ENCRYPTION_KEY` as the one required secret.
- **actual:** The server refuses to boot: `CREDENTIAL_ENCRYPTION_KEY is required`. This is correct fail-loud behavior, but the local env cannot start the app as-is.
- **root_cause:** `.env` was never migrated after the 2026-09-26 BYOK pivot removed OAuth and Stripe. `cmd/server/main.go:83-90` validates `CREDENTIAL_ENCRYPTION_KEY` as the one hard requirement, and `.env.example` was updated but `.env` was not.
- **reproduce:**
  ```bash
  grep -oE '^[A-Z_]+' .env | tr '\n' ' '
  # observed: GEMINI_API_KEY GEMINI_MODEL DATABASE_PATH STATIC_DIR PORT
  #           GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET GOOGLE_REDIRECT_URL
  #           STRIPE_SECRET_KEY STRIPE_WEBHOOK_SECRET STRIPE_PRICE_PRO STRIPE_PRICE_RESEARCH FRONTEND_URL
  # missing:    CREDENTIAL_ENCRYPTION_KEY
  ```
- **impact:** Onboarding friction only — a developer following `README.md` from a clean clone cannot start the server. Dead Google/Stripe keys remain in a developer environment for features that no longer exist.
- **fix:** Refresh `.env` from `.env.example` and remove the dead Google/Stripe keys. `.env` is gitignored, so nothing leaked to history.
- **regression_test:** None practical — this is local config. Consider a startup self-check that warns when known-removed variables (`GOOGLE_*`, `STRIPE_*`) are still set.

### QA-016 — A credential AAD mismatch has no recovery path

- **severity:** LOW · **confidence:** verified
- **location:** `internal/app/credentials/resolver.go:28` · `internal/handlers/documents.go`
- **symptom:** When the `model` on a credential row diverges from the value bound into the AAD, every call fails with `500 credential unreadable` and log `decrypt: cipher: message authentication failed`.
- **expected:** A recoverable, self-explanatory error.
- **actual:** The end user sees an opaque 500. The only fix is manual Remove + re-add.
- **root_cause:** AAD binds each ciphertext to `user_id|provider|model` (`internal/external/crypto.go`). Changing `model` on the row invalidates the AAD, so AES-GCM authentication fails and the resolver cannot distinguish a tampered blob from a wrong encryption key.
- **reproduce:**
  ```bash
  # Modifying model directly breaks the AAD binding user_id|provider|model
  sqlite3 paperviz.db "UPDATE user_credentials SET model='other' WHERE id='<id>';"
  # next ingest → 500 {"error":"credential unreadable"}
  ```
- **impact:** A user who edits a credential outside the UI permanently loses access to it and gets no instruction to re-enter the key. Not reachable through the normal UI today, but `QA-011` (no edit path) is the likely origin of a future in-app route to this state.
- **context — this is a positive result:** `AGENTS.md` claims an AAD mismatch makes credentials permanently unreadable with no way to distinguish "wrong key" from "wrong AAD". **Confirmed working exactly as designed.** The gap is only the user-facing error and its recovery path.
- **fix:** Map this condition to an actionable `409` explaining that the stored key must be re-entered, rather than a generic 500.
- **regression_test:** Assert an AAD mismatch yields the actionable status, not 500.

---

## 5. What passed

Verified correct — do not regress these:

- **Auth:** signup, login, logout, session invalidation (`/api/auth/me` → 401 after logout)
- **Input validation:** invalid email → `invalid_email`; weak password → `password_too_weak`
- **Rate limits:** auth `401,401,401,429,429,429…`; document create `…,429,429,429`
- **Credential storage:** AES-256-GCM; only `key_hint` (last 4 chars) ever returned; plaintext never returned
- **Copyright/privacy:** export excludes `original_text` and `simplified_text`; share payload excludes `original_text` and `user_id`
- **Share revocation:** correctly `404` after revoke; bogus tokens `404`
- **Collection ownership:** cross-user write returns `forbidden` (D3 holds)
- **SSRF protection:** `https`-only scheme plus private-IP DNS resolution check in `internal/services/import.go:130,245`
- **AAD binding:** model mismatch correctly invalidates the credential (see `QA-016`)
- **Retry + backoff:** 5 attempts over ~66s, correctly ordered
- **Routing:** `/healthz`, 404 page, `/dashboard` → `/account` redirect, `/api/*` 404 returns JSON
- **Gates:** `go build` · `go test ./...` (627 pass, 13 packages) · `go vet` · `gofmt -l` empty · `oxlint` · `vite build` · MCP parity (5 tools / 5 doc rows) all green
- **E2E suite:** 37/38 passed; the single failure is a stale assertion inside `QA-003`

---

## 6. Coverage gaps

Not tested, with the reason. Do not interpret absence of findings here as
absence of defects.

| Area | Blocker |
|---|---|
| Anthropic / OpenAI providers | No keys available. Their `DefaultModel()` values are unverified (`QA-005`) |
| LLM happy path end-to-end | All 3 Gemini models returned `503 high demand` across 5 attempts. No verified paper was ever produced; every pipeline run ended `failed` |
| PDF ingestion | No text-layer PDF available; scanned PDFs are unsupported by design |
| Research Map, Notes overlay, copy/share dialogs | Require a genuinely completed document, which could not be produced |
| Account agent-API-key creation flow | Reachable but not exercised end-to-end |
| Real MCP client handshake against `cmd/mcp` | Out of scope; only the HTTP `/api/mcp` absence was confirmed |
| Horizontal overflow at 375px on `/`, `/login`, `/signup`, `/account`, `/agents` | Only `/` (clean) and the result page (overflow) were measured |
| Concurrency: `PIPELINE_MAX_CONCURRENCY` | No load applied |
| `INGESTION_ENABLED=false` kill switch | Not exercised |

**Method note:** a deterministic fixture document was used to reach the result
page, since the LLM path was provider-blocked. That is a limitation of coverage,
not a substitute — no conclusion about the LLM pipeline's output quality can be
drawn from this audit.

---

## 7. Remediation order

Strict sequence. `QA-001` and `QA-002` are the release gate.

1. **QA-001** — IDOR. Treat as a security incident. Add cross-user regression tests for every authenticated write under `/api/documents`.
2. **QA-002** — chart rendering. One-line fix; restores the headline feature.
3. **QA-005** — fix `DefaultModel()` for all three providers, then verify each against a live key.
4. **QA-003** — ship `/api/mcp` or repoint `/agents` at the stdio config.
5. **QA-007** — correct the false landing claims; map `missing_credential` to actionable copy.
6. **QA-004** — re-path the share routes.
7. **QA-006** — add `cdn.fontshare.com` to `font-src`; restores the design system.
8. `QA-008` … `QA-016` as capacity allows.

**Do not cut a release tag until 1–7 are closed and covered by tests.**
Given the 2026-09-28 audit is the second in two weeks to find Critical defects
behind a fully green CI, the gate most likely to fail again is untested
authorization and untested data contracts — prioritize adding those test
categories over any additional feature work.
