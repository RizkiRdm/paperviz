# ADR-002: Platform Foundation

**Status:** accepted
**Date:** 2026-10-09
**Supersedes:** no
**Context:** QA audit before the agent-first product launch. One hosted Go
process, Cloudflare in front, no CORS, no second runtime, no third-party
model vendor.

## Decision

| Concern | Decision |
|---|---|
| Compute | Render, Docker web service, free plan. Repo-connected Git auto-deploy from `main`. |
| Edge / TLS | Render-managed TLS on `*.onrender.com`. Single origin — no CORS, no reverse proxy. |
| Asset delivery | Served by the Go binary from the same origin as the API. No static-site split. |
| Storage | Local SQLite on the container filesystem. 21 golang-migrate schema migrations. |
| Migrations | `github.com/golang-migrate/migrate/v4` + its `database/sqlite` driver. Filesystem-discovered up/down pairs. |
| Analytics | PostHog, client-side, three events. |
| Alerting | Not yet chosen. Deferred — no budget or on-call. |
| Availability | Best-effort; no SLO. Free plan permits spin-down after 15 min idle. |

## Consequences

Data is **ephemeral and disposable** on the free plan: SQLite lives on Render's
ephemeral filesystem, which is recreated each deploy and spin-down. Documents
are user-created and not production-recoverable, so no backup burden is
acceptable to a beta audience.

The durable data lives in two places that are *not* SQLite:

- **Analytics.** PostHog receives `signup_completed`, `api_key_added`, and
  `paper_analysed` from the browser and holds them on its own cluster, so
  signup counts and feature usage survive every Render restart. This is why
  the analytics surface is not persisted locally.
- **Model credentials.** Each user's key is stored only in their own session;
  the server holds no shared credential and no secret is persisted to disk.

## What the free plan costs

Every deploy wipes the local SQLite file (accounts, sessions, processed
documents). The `CREDENTIAL_ENCRYPTION_KEY` in Render's env survives a
restart, so nothing silently unreadable is produced — but the content it
would decrypt is ephemeral by design. For a launch-scale audience with a
7-day document expiry, that trade is accepted.

Disks are available on paid plans only ($0.25/GB-month). Choose one *if* a
stakeholder asks for durability before re-architecting.

## Rejected

### Convex — dropped

No Go SDK, so integration would run through a generated OpenAPI layer or a
TypeScript worker. That contradicts "one long-lived Go process, one writer,
no ORM" and would introduce a second runtime on the request path for a state
made disposable by the hosted SQLite choice anyway.

### Better Auth — dropped

TypeScript-only. It would force a Node service on a repo with no JS server
code, purely for a problem Go's existing password + session flow already
handles (ADR-001).

### Workers — dropped

Workers are V8 isolates with no SQLite and no long-lived process. The orphan
background goroutines in the pipeline assume a stateful process; nothing in
this production path runs per-request.

### Static frontend on Cloudflare Pages — dropped

Pages splits the origin. The session cookie has no `Domain` and is
`SameSite=Lax`, so a cookie owned by `onrender` cannot authenticate the API
for a page served from `pages.dev`. A working split would need `SameSite=None`
plus shared CORS-proxied auth, which widens CSRF surface, adds a cross-origin
credential leak, and supplies no benefit a render-managed CDN cannot already
give a low-traffic beta.

### Cloudflare Tunnel — dropped

Tunnel would host the app on the owner's machine: paid bandwidth, a public IP
attached to a network, and uptime tied to a home machine — while offering
nothing over Render's free plan for a repo-backed web service.

## Verification

- `https://paperviz-z4ne.onrender.com/healthz` — `200 {"status":"ok"}`
- `POST /api/auth/signup` — `201`; user created and logged in
- Password rules — `password_too_short` (8 chars) and `password_too_weak`
  (needs upper, lower, digit, symbol) both verified
- Rate limit — `POST /api/documents` returns `429` on the third request
- Security headers — CSP, `X-Frame-Options: DENY`, `X-Content-Type-Options`
  verified live; CSP PostHog entries verified by test
- Render — `chore/golang-migrate` fast-forward merged to `main`, branch
  deleted, origin updated

## Follow-up

- Durability upgrade if and only if paying users arrive: paid Render disk, or an
  external durable store, or a snapshot/export job. Not before.
- PostHog events are the durable signal; local documents are not.
- `connect-src` must include `https://us.i.posthog.com` — analytics surface
  silently otherwise.
- Add a custom domain only if DNS routing needs to survive a Render move.
