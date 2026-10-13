# ADR-003: Clerk owns authentication

Date: 2026-10-13
Status: Accepted (reverses the auth half of ADR-002's "no third-party account" rule)
Supersedes: the 2026-09-26 removal of Google OAuth and Stripe

## Context

The 2026-09-26 pivot removed Google OAuth and Stripe. The stated reason was that a
third-party account is a single point of catastrophic failure for a solo
maintainer, so PaperViz would hold no such account. Auth became local email and
password with `cmd/admin` recovery; billing was deleted outright.

That left two real costs:

- Password reset needed an email provider, so it could not exist. Recovery was a
  CLI an operator had to run.
- Users maintained BYOK model keys against a hand-rolled bcrypt store with no
  verification, no rate limiting beyond a thin IP bucket, and no bot protection.

The owner subsequently decided to adopt Clerk. This ADR records that decision and
the risk it re-introduces rather than burying it in a commit message.

## Decision

Clerk owns signup, sign-in, session lifecycle, and password recovery. PaperViz
verifies the resulting session token on every authenticated request.

### What Clerk does not own

The `users` table stays. It is PaperViz's own profile store, holding the Clerk
subject as `id` plus the email, service API key digest, and creation time.
Documents, collections, annotations, and BYOK credentials keep referencing
`users.id` unchanged, so no ownership foreign key needed rewriting.

BYOK model keys are still encrypted at rest with `CREDENTIAL_ENCRYPTION_KEY` and
never leave PaperViz. Clerk never sees a model key. Billing stays deleted.

### Identity mapping

`users.id` holds the Clerk subject (`user_2abc...`). A local row is created the
first time an identity makes an authenticated request, not at signup — Clerk owns
the signup flow, so PaperViz does not learn an account exists until it is used.
Steady-state requests are one indexed lookup with no call to Clerk.

### Transport and verification

The browser sends `Authorization: Bearer <token>`. The Go backend verifies it with
`clerkhttp.WithHeaderAuthorization()`, which handles JWKS retrieval, caching, and
key rotation. A transport failure or an unverifiable token produces 401, never a
fallback to any local session.

### Local credential paths were deleted

- `POST /api/auth/signup` and `POST /api/auth/login` no longer exist.
- The `sessions` table is dropped (migration 022). It was an at-rest bearer-token
  store that Clerk could not revoke.
- `session_token` cookies, bcrypt password handling, and `rateLimitAuth` are gone.
- `cmd/admin reset-password` is gone. `cmd/admin revoke-sessions <email>` replaces
  it and revokes every live Clerk session for an account.

A test asserts `/api/auth/signup` and `/api/auth/login` return 404, so the
credential path cannot quietly come back.

### UI

PaperViz renders its own sign-in and sign-up forms using Clerk's headless hooks
(`useSignIn`, `useSignUp`). None of Clerk's prebuilt components are used, and
`prefetchUI={false}` keeps the prebuilt UI bundle out of the payload.

## Consequences

### Accepted risks

**Clerk is closed-source SaaS with no self-hosting option.** If the Clerk account
is suspended, deleted, or its owner loses access, nobody can sign in and nobody
can reach their stored model keys. This is a worse version of the exact failure
ADR-002 removed: OAuth was only a login handshake, whereas Clerk now owns the
user directory, sessions, and password recovery.

Mitigations: `users` and all user data remain local, so PaperViz retains the data
even if the identity provider disappears; `cmd/admin` remains a break-glass path.

**A dependency can be pulled out from under the product.** Clerk can change or
remove free-tier behaviour. The free tier is 10k MAU and requires no payment
method, which is what makes it viable here at all — R2 was rejected on exactly
that constraint.

### Improvements

- Self-service password recovery now exists, removing the reason the CLI was
  built. The user chooses their own secret instead of an operator printing a
  generated one to a terminal.
- Credential guessing and bot traffic happen against Clerk, which has abuse
  protection, instead of against PaperViz's IP bucket.

### Unchanged

- MCP is untouched. It is stateless and carries no user identity, so it needed no
  Clerk integration.
- Anonymous document reads, share links, and the local ingestion toggle are
  unchanged.
- Model spend is still entirely the user's.

## Testing

Playwright cannot complete a Clerk signup unattended: it needs a Clerk **test
instance** (`pk_test_...`) to mint a testing token. Until one exists, the specs
that require a signed-in session are skipped rather than worked around with a
test-only auth backdoor, which would reintroduce the bypassable credential path
this ADR exists to close.