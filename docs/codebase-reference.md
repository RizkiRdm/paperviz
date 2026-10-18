# Codebase Reference

Technical map of PaperViz as implemented on 2026-10-18. For concepts and product context, read [Overview](overview.md). For code changes, also read [`../AGENTS.md`](../AGENTS.md).

## Runtime entrypoints

| Runtime | Entrypoint | Responsibility |
|---|---|---|
| HTTP server | `cmd/server/main.go` | Validate environment (`CREDENTIAL_ENCRYPTION_KEY`, `CLERK_SECRET_KEY`), configure JSONL logging, open migrated SQLite DB, start expiry sweep, serve router |
| MCP server | `cmd/mcp/main.go` | Validate MCP environment, open shared DB, construct MCP server, run stdio transport |
| React SPA | `frontend/src/main.jsx` | Mount React application |
| Route table | `frontend/src/App.jsx` | Define browser routes and compatibility redirects |

All three Go entrypoints call `repository.Open`, which configures the connection and applies migrations. golang-migrate reads the migration directory, so there is no list to keep in sync.

## Top-level structure

| Path | Purpose |
|---|---|
| `.github/workflows/` | One CI path: `go-ci.yml` (all gates + Render deploy), `release.yml` (SemVer tag + deploy); shared composite `.github/actions/render-deploy` |
| `build/` | Backend and frontend container definitions |
| `cmd/` | Server and MCP binaries |
| `docs/` | Product, architecture, API, operations, and project records |
| `e2e/` | Playwright configuration, tests, and package scripts |
| `frontend/` | React/Vite application |
| `goals/` | Goal and chunk plans |
| `internal/` | Go application code |
| `migrations/` | Ordered SQL schema migrations |
| `DESIGN.md` | UI design-system source of truth |
| `PRODUCT.md` | Product direction and audience |
| `AGENTS.md` | Engineering and agent constraints |

## Backend layers

```text
cmd
 ├── handlers ──→ app/documents ──→ services ──→ repository
 └── mcp ───────────────────────────→ services/repository
                                  └──→ external
```

### `internal/app/documents`

Application boundary for document use cases.

| File | Responsibility |
|---|---|
| `service.go` | Create, get, list, and retrieve document research sections |
| `readmodel.go` | Assemble document-centric read responses |

### `internal/handlers`

HTTP transport and cross-cutting web concerns.

| File | Responsibility |
|---|---|
| `router.go` | Route registration, middleware order, static files, SPA fallback |
| `documents.go` | `DocumentHandler` constructor, client resolver, `RequireDocAccess` ownership guard |
| `documents_create.go` | Multipart PDF/text intake and async pipeline start |
| `documents_list.go` | Document list and stats |
| `documents_update.go` | Save toggle, title update, delete |
| `documents_get.go` | Document read model and chart image |
| `documents_sections.go` | Claims, tables, methods, results, citations |
| `documents_readmodel.go` | Evidence graph |
| `documents_relationships.go` | Paper relationships and research map |
| `documents_compare.go` | Cross-document comparison |
| `import.go` | DOI and URL ingestion |
| `auth.go` | Clerk session identity (`Me` only; no local signup/login) |
| `apikey.go` | API-key retrieval and regeneration |
| `credentials.go` | Per-user BYOK model-key CRUD (encrypted at rest) |
| `account.go` | Account summary |
| `annotations.go` | Per-user annotation CRUD |
| `export.go` | Research-context export |
| `collections.go` | Per-user collection CRUD and document membership |
| `share.go` | Public share generation, retrieval, and revocation |
| `usage.go` | Usage summary |
| `analytics.go` | Product analytics endpoints |
| `middleware.go` | Clerk auth (`RequireAuth`/`OptionalAuth`), usage-limit, ingestion toggle |
| `ratelimit.go` | IP and auth rate limiters |
| `fingerprint.go` | Anonymous usage fingerprint |
| `security_headers.go` | Security response headers |
| `health.go` | Database-backed `/healthz` |
| `respond.go` | JSON success/error helpers |
| `validation.go` | Input validation helpers |

### `internal/services`

Business logic and orchestration.

| Area | Main files |
|---|---|
| Intake and persistence | `intake.go` |
| Pipeline orchestration | `pipeline.go`, `pipeline_stages.go` |
| Simplification | `simplification.go` |
| Claim verification | `verification.go` |
| Chapter detection | `chapters.go` |
| Evidence extraction | `evidence_extract.go`, `table_extract.go` |
| Dataset construction | `dataset_build.go` |
| Grounding validation | `grounding.go` |
| Chart planning and fallback | `charts.go`, `charts_llm.go` |
| PDF/document extraction | `extraction.go` |
| DOI/URL import | `import.go`, `import_fetcher.go` |
| Sharing | `share.go` |
| Collections | `collections.go` |
| Annotations | `annotations.go` |
| Export | `export.go` |
| Usage tiers | `tier.go` |
| Research map | `research_map.go` |
| Comparison | `comparison.go` |
| Expiry | `expiry.go` |
| Shared values | `types.go`, `source_policy.go` |

### `internal/repository`

Raw SQLite access using `database/sql`.

Repositories cover:

- users (Clerk subject profile) and per-user model credentials;
- documents, chapters, charts, and evidence;
- claim diffs, claims, and claim-evidence links;
- tables, methods, results, citations, and paper relationships;
- annotations and collections;
- sharing, referrals, usage, analytics, and tiers.

Infrastructure files:

| File | Responsibility |
|---|---|
| `db.go` | Open SQLite, configure pragmas, apply migrations |
| `migrate.go` | golang-migrate wiring: Migrate, Down, DownTo, DownAll, Version |
| `id.go` | Generate document and related IDs |

SQLite pragmas:

```text
journal_mode = WAL
synchronous = NORMAL
foreign_keys = ON
busy_timeout = 5000
max_open_conns = 1
```

### `internal/external`

| File | Responsibility |
|---|---|
| `llm.go` | Provider-agnostic `Transport`: HTTP client, retry schedule, 90s budget, process-wide concurrency slot, `Provider` type |
| `llm_goai.go` | Gemini/Anthropic/OpenAI backends via `github.com/zendev-sh/goai` (SDK retries disabled) |
| `pdf.go` | PDF text and embedded-image extraction |
| `crossref.go` | DOI metadata and source resolution |
| `unpaywall.go` | Open-access source lookup |
| `logger.go` | JSONL logging integration |
| `chartstorage.go` | Optional Supabase Storage for chart image bytes |
| `crypto.go` | AES-256-GCM credential cipher |

### `internal/mcp`

| File | Responsibility |
|---|---|
| `server.go` | MCP server construction and tool registration |
| `tools.go` | Five tool definitions and handlers |
| `schemas.go` | Tool input and output types |
| `errors.go` | Stable MCP error codes |
| `ratelimit.go` | Per-key token buckets |
| `jobs.go` | Per-key concurrent-job cap |
| `mcp_contract_test.go` | Registered tool and contract regression tests |

### Other backend packages

| Package | Purpose |
|---|---|
| `internal/models/` | Chart specifications, datasets, evidence, and document value types |
| `internal/apperr/` | Shared typed application errors |
| `internal/infra/ratelimit/` | Infrastructure rate-limit primitives |

## Frontend structure

| Path | Responsibility |
|---|---|
| `frontend/src/App.jsx` | Route table and top-level providers |
| `frontend/src/main.jsx` | React mount entry |
| `frontend/src/index.css` | Design tokens and global styles |
| `frontend/src/lib/api.js` | REST client functions and API errors |
| `frontend/src/lib/utils.js` | Shared class-name utility |
| `frontend/src/hooks/useApi.js` | Execute/loading/error/retry state |
| `frontend/src/hooks/use-document-poll.js` | Two-second document polling, soft warning, timeout, retry |
| `frontend/src/hooks/use-current-user.js` | Current Clerk user state |
| `frontend/src/hooks/useResultState.js` | Result-page state machine (`loading|partial|ready|error|retryable-error|empty`) and handlers |
| `frontend/src/hooks/useIngest.js` | Unified PDF/paste/DOI/URL ingestion |
| `frontend/src/hooks/useAccount.js` | Account page state |
| `frontend/src/hooks/useAnnotations.js` | Annotation CRUD state |
| `frontend/src/components/result/` | Header, understanding, evidence, figures, source, actions, and status-view sections |
| `frontend/src/components/ingest.jsx` | Unified ingestion form with `source_type` badge |
| `frontend/src/components/account/` | Account cards (agent key, stored credentials) |
| `frontend/src/components/annotations/` | Annotation row and add form |
| `frontend/src/components/ui/` | Reusable UI primitives and reading-level selector |
| `frontend/src/pages/` | Route-level components |

## Frontend routes

| Route | Component | Purpose |
|---|---|---|
| `/` | `UploadPage` | PDF, paste, DOI, and URL input |
| `/upload` | redirect | Redirects to `/` |
| `/dashboard` | redirect | Redirects to `/account` |
| `/login` | `LoginPage` | Clerk-backed sign-in form |
| `/signup` | `SignupPage` | Clerk-backed sign-up form |
| `/account` | `AccountPage` | Account, API key, usage, and stored model keys |
| `/agents` | `AgentsPage` | Generated agent-client configuration |
| `/share/fig/:shareToken` | `ShareFigurePage` | Public figure view |
| `/share/doc/:shareToken` | `SharePaperPage` | Public paper view |
| `/:documentId` | `ResultPage` | Processing and result sections |
| `*` | `NotFoundPage` | Not-found page |

Static SEO files under `frontend/public/` can outlive React route cuts. They are not proof of current navigation.

## REST routes

All routes below are registered in `internal/handlers/router.go`.

### Operational

| Method and path | Auth | Purpose |
|---|---|---|
| `GET /healthz` | Public | Database-backed health check |

### Documents and research objects

| Method and path | Auth | Purpose |
|---|---|---|
| `POST /api/documents/` | Optional + usage limit | Ingest PDF or pasted text; start full pipeline |
| `GET /api/documents/{id}` | Public | Get document read model; refresh access time |
| `GET /api/documents/{id}/charts/{chartId}/image` | Public | Get stored chart image |
| `GET /api/documents/{id}/claims` | Optional + doc access | List claims |
| `GET /api/documents/{id}/tables` | Optional + doc access | List tables |
| `GET /api/documents/{id}/methods` | Optional + doc access | List methods |
| `GET /api/documents/{id}/results` | Optional + doc access | List results |
| `GET /api/documents/{id}/citations` | Optional + doc access | List citations |
| `GET /api/documents/{id}/evidence-graph` | Optional + doc access | Get evidence graph |
| `GET /api/documents/{id}/research-map` | Optional + doc access | Get grouped research relationships |
| `GET /api/documents/stats` | Required | User document statistics |
| `GET /api/documents/` | Required | User document list |
| `PUT /api/documents/{id}/save` | Required | Toggle saved state |
| `PATCH /api/documents/{id}` | Required | Update title |
| `DELETE /api/documents/{id}` | Required | Delete owned document |
| `POST /api/documents/compare` | Required | Compare two documents |
| `GET /api/papers/{id}/relationships` | Optional + doc access | Get paper relationships |
| `POST /api/papers/{id}/relationships` | Required + doc access | Create paper relationship |

"Optional + doc access" = `OptionalAuth` + `RequireDocAccess` (wave G guard): anonymous readers reach public documents, owners reach their own, everyone else gets 403/404. New per-document routes must copy this pattern (regression test `TestSectionRoutesRespectOwnershipOverRouter`).

### Import

| Method and path | Auth | Purpose |
|---|---|---|
| `POST /api/import/doi` | Optional + usage limit | Resolve and ingest DOI |
| `POST /api/import/url` | Optional + usage limit | Fetch and ingest URL |

### Authentication and API keys

Clerk owns signup, sign-in, and sessions (ADR-003); `POST /api/auth/signup` and `POST /api/auth/login` are intentionally absent (404). Bearer token = Clerk session token.

| Method and path | Auth | Purpose |
|---|---|---|
| `GET /api/auth/me` | Required | Current identity |
| `GET /api/auth/apikey` | Required | `configured` flag only |
| `POST /api/auth/apikey` | Required | Issue service API key (plaintext returned once) |
| `POST /api/auth/apikey/regenerate` | Required | Replace service API key |

### Model credentials (BYOK)

| Method and path | Auth | Purpose |
|---|---|---|
| `POST/GET /api/credentials/` | Required | Store (encrypted) or list model keys |
| `POST /api/credentials/{id}/default` | Required | Set default key |
| `DELETE /api/credentials/{id}` | Required | Delete stored key |

### Account

| Method and path | Auth | Purpose |
|---|---|---|
| `GET /api/account/summary` | Required | Account summary |

### User-owned research context

| Method and path | Auth | Purpose |
|---|---|---|
| `GET/POST /api/documents/{id}/annotations` | Required | List or create annotations |
| `PUT/DELETE /api/documents/{id}/annotations/{annotationId}` | Required | Update or delete owned annotation |
| `GET /api/documents/{id}/export` | Required | Export research context |
| `POST/GET /api/collections/` | Required | Create or list collections |
| `GET/PATCH/DELETE /api/collections/{id}` | Required | Read, rename, or delete owned collection |
| `POST/DELETE /api/collections/{id}/documents` | Required | Add or remove document |

### Sharing

| Method and path | Auth | Purpose |
|---|---|---|
| `POST/DELETE /api/documents/{id}/share` | Required | Generate or revoke document share |
| `PATCH /api/documents/{id}/visibility` | Required | Set private, unlisted, or public visibility |
| `POST/DELETE /api/documents/{id}/charts/{chartId}/share` | Required | Generate or revoke figure share |
| `GET/HEAD /share/doc/{shareToken}` | Public token | Get shared paper |
| `GET/HEAD /share/fig/{shareToken}` | Public token | Get shared figure |
| `POST /share-referrals` | Public | Track referral conversion |

### Usage and analytics

| Method and path | Auth | Purpose |
|---|---|---|
| `GET /api/usage` | Public | Current anonymous or user usage |
| `GET /analytics` | Required | Account analytics summary |
| `POST /api/analytics/pricing-view` | Public | Track pricing view |
| `POST /api/analytics/upgrade-intent` | Public | Track upgrade intent |

HTTP contract details belong in [`openapi.yaml`](openapi.yaml). This table is a navigation map, not a replacement for schemas and response definitions.

## MCP tools

| Tool | Input summary | Result summary | Current side effect |
|---|---|---|---|
| `ingest_document` | Text, optional reading level | Document ID, title, status, source metadata | Inserts processing document only |
| `search_documents` | Title query, limit | Matching metadata and sections | Read-only global search |
| `get_document` | Document ID, include sections | Metadata and selected data | Read-only |
| `get_figures` | Document ID | Chart data, source, provenance | Read-only |
| `get_evidence` | Document ID | Evidence, claims, links, provenance | Read-only |

MCP constraints:

- stdio transport;
- `PAPERVIZ_API_KEY` required at startup (`DATABASE_PATH`, `MIGRATIONS_DIR` optional);
- no tool calls a model — all five are deterministic (P28); no `GEMINI_API_KEY` is read by `cmd/mcp`;
- no PDF upload;
- no user session or library ownership scope;
- no list/save/rename/delete/share operations;
- per-key rate and job limits.

## Environment variables

| Variable | Consumer | Required | Default or notes |
|---|---|---:|---|
| `CREDENTIAL_ENCRYPTION_KEY` | server | Yes | base64 32 bytes; AES-256-GCM for stored user keys. Startup fails if missing or malformed |
| `INGESTION_ENABLED` | server | No | `true`. `false` returns 503 on document creation; read/share/library keep working |
| `PIPELINE_MAX_CONCURRENCY` | server | No | `2`. Bounds concurrent pipelines, not model spend |
| `DATABASE_PATH` | server, MCP, admin | No | `paperviz.db` |
| `STATIC_DIR` | server | No | `frontend/dist` |
| `PORT` | server and Vite proxy | No | `8080` |
| `LOG_FILE` | server | No | `paperviz.log.jsonl` |
| `CLERK_SECRET_KEY` | server | Yes | Clerk session verification; startup fails if missing |
| `VITE_CLERK_PUBLISHABLE_KEY` | frontend build | Yes | Inlined by Vite at build time |
| `PAPERVIZ_API_KEY` | MCP | Yes for MCP | Process-level MCP key |
| `MIGRATIONS_DIR` | MCP, admin | No | `migrations` |

`GEMINI_API_KEY`/`GEMINI_MODEL` are no longer read by any binary as of P28 (2026-10-17) — `.env.example` still lists them under a stale "Read by cmd/mcp" heading; treat them as dead until that file is cleaned.

The server holds no model vendor credential. Web model calls run on the requesting user's own key, decrypted per request by `internal/app/credentials.Resolver`; there is no `GEMINI_API_KEY` and no `FRONTEND_URL` in the server path. `CREDENTIAL_ENCRYPTION_KEY` and `CLERK_SECRET_KEY` are the two secrets `cmd/server` refuses to boot without.

The HTTP process does not load any `.env` file. Export values before running it. `.env.example` is the template; `.env.development` is the local (gitignored) copy that `make run`/`make dev` export, overridable with `ENV_FILE=`. Production config lives only in Render env vars — there is no `.env.production` file.

## Commands

### Setup

```bash
go mod download
npm --prefix frontend ci
npm --prefix e2e ci
```

### Development

```bash
set -a; source .env.development; set +a
go run ./cmd/server
npm --prefix frontend run dev
```

### Production-style build

```bash
npm --prefix frontend run build
CGO_ENABLED=0 go build -o server ./cmd/server
./server
```

### Make targets

| Target | Effect |
|---|---|
| `make build` | Build frontend if needed, then `./server` |
| `make run` | Build and run with `.env.development` exported (`ENV_FILE=` to override) |
| `make dev` | Run backend with Air |
| `make install-air` | Install Air |
| `make clean` | Remove binary, frontend dependencies, and frontend build |
| `make container` | Build Podman image |
| `make container-run` | Run container with persistent data volume |
| `make container-stop` | Stop and remove container |

### Verification

```bash
gofmt -l .
go vet ./...
go build ./...
go build ./cmd/mcp
go test ./...
npm --prefix frontend run lint
npm --prefix frontend run build
BASE_URL=http://localhost:8080 npm --prefix e2e test
```

## Migration inventory

Filenames on disk carry a 6-digit prefix (`000001_init.up.sql`); short names below match this table's original convention.

| Version | File | Area |
|---:|---|---|
| 001 | `001_init.sql` | Documents, charts, claim diffs |
| 002 | `002_users.sql` | Users and sessions |
| 003 | `003_chapters.sql` | Chapters |
| 004 | `004_chapter_charts.sql` | Chapter-chart linkage |
| 005 | `005_evidence.sql` | Evidence and provenance |
| 006 | `006_document_title.sql` | Derived document titles |
| 007 | `007_saved_papers.sql` | Saved-paper state |
| 008 | `008_research_collections.sql` | Collections and membership |
| 009 | `009_share_tokens.sql` | Figure share tokens |
| 010 | `010_document_share.sql` | Document sharing and visibility |
| 011 | `011_share_referrals.sql` | Referral counters |
| 012 | `012_usage_analytics.sql` | Usage analytics |
| 013 | `013_usage_tiers.sql` | Usage tiers |
| 014 | `014_structured_research_objects.sql` | Tables, methods, results, citations, claims |
| 015 | `015_evidence_graph.sql` | Claim-evidence and paper relationships |
| 016 | `016_annotations.sql` | User annotations |
| 017 | `017_oauth_columns.sql` | OAuth identity |
| 018 | `018_api_key_column.sql` | User API key |
| 019 | `019_billing_columns.sql` | Stripe customer and subscription state (dropped again by 021) |
| 020 | `020_user_credentials.sql` | Per-user encrypted model keys |
| 021 | `021_drop_oauth_billing.sql` | Removes OAuth columns and Stripe columns |
| 022 | `022_clerk_auth.sql` | Drops `sessions` table and password columns (Clerk) |

## Test locations

| Area | Location |
|---|---|
| Repository and migration behavior | `internal/repository/*_test.go` |
| Pipeline, evidence, grounding, charts | `internal/services/*_test.go` |
| Auth, image serving, import, router | `internal/handlers/*_test.go` |
| MCP contract | `internal/mcp/mcp_contract_test.go` |
| Server configuration | `cmd/server/main_test.go` |
| Browser behavior | `e2e/tests/e2e/*.spec.ts` |

## CI gates

| Workflow | Gates |
|---|---|
| `.github/workflows/go-ci.yml` | gofmt, vet, build, `go test`, MCP stdio handshake, frontend lint + build, container builds, Playwright e2e; `deploy` job on push to `main` or `v*` tags → Render prod via `.github/actions/render-deploy` (asserts `/healthz` 200, `/api/collections` 401, `/api/auth/me` 401) |
| `.github/workflows/release.yml` | `workflow_dispatch` SemVer tag cut (validates `x.y.z`, green main HEAD, annotated tag) + direct deploy (GITHUB_TOKEN recursion guard) |

Secrets: `RENDER_API_KEY`, `VITE_CLERK_PUBLISHABLE_KEY`. Former `frontend-ci.yml` and `e2e.yml` were deleted 2026-10-17 (folded into `go-ci.yml`).

## Known contract mismatches

| Mismatch | Evidence | Consequence |
|---|---|---|
| `/agents` config expects remote `/api/mcp` | `frontend/src/pages/agents-page.jsx` versus `internal/handlers/router.go` | Generated remote config cannot connect to current server |
| MCP ingest does not run pipeline | `internal/mcp/tools.go` calls `ValidateAndInsert` only | MCP-created document can remain `processing` |
| Lowercase architecture file is a compatibility pointer | `docs/architecture.md` | Keep redirects current; use uppercase document for contracts |
| SEO files and screenshots include removed routes | `frontend/public/`, `assets/` | Must not be used as current navigation proof |

## Change verification points

| Change | Verify in |
|---|---|
| Frontend route | `frontend/src/App.jsx` |
| REST route | `internal/handlers/router.go` |
| MCP tool | `internal/mcp/tools.go` and contract test |
| Runtime env | `.env.example` and both `cmd/*/main.go` |
| Migration | `migrations/` and `internal/repository/migrate.go` |
| Pipeline | `internal/services/pipeline*.go`, stages, and data-pipeline doc |
| Design token | `DESIGN.md` and frontend styles |
| Current status | `docs/PROJECT_STATE.md` |
