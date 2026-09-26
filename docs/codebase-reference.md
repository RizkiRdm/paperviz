# Codebase Reference

Technical map of PaperViz as implemented on 2026-09-25. For concepts and product context, read [Overview](overview.md). For code changes, also read [`../AGENTS.md`](../AGENTS.md).

## Runtime entrypoints

| Runtime | Entrypoint | Responsibility |
|---|---|---|
| HTTP server | `cmd/server/main.go` | Validate environment, configure JSONL logging, open migrated SQLite DB, clean sessions, start expiry sweep, serve router |
| MCP server | `cmd/mcp/main.go` | Validate MCP environment, open shared DB, construct MCP server, run stdio transport |
| React SPA | `frontend/src/main.jsx` | Mount React application |
| Route table | `frontend/src/App.jsx` | Define browser routes and compatibility redirects |

Both Go entrypoints call `repository.LoadMigrations` and `repository.Open`, preventing server/MCP migration drift.

## Top-level structure

| Path | Purpose |
|---|---|
| `.github/workflows/` | Go, frontend, Docker-image, and Playwright CI |
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
| `documents_create.go` | Multipart PDF/text intake and async pipeline start |
| `documents.go` | Document read model, research objects, comparison, relationships |
| `import.go` | DOI and URL ingestion |
| `auth.go` | Signup, login, logout, session, Google OAuth |
| `apikey.go` | API-key retrieval and regeneration |
| `billing.go` | Stripe checkout, portal, webhook |
| `account.go` | Account summary |
| `annotations.go` | Per-user annotation CRUD |
| `export.go` | Research-context export |
| `collections.go` | Per-user collection CRUD and document membership |
| `share.go` | Public share generation, retrieval, and revocation |
| `usage.go` | Usage summary |
| `analytics.go` | Product analytics endpoints |
| `middleware.go` | Auth and usage-limit middleware |
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
| Billing | `billing.go` |
| Usage tiers | `tier.go` |
| Research map | `research_map.go` |
| Comparison | `comparison.go` |
| Expiry | `expiry.go` |
| Shared values | `types.go`, `source_policy.go` |

### `internal/repository`

Raw SQLite access using `database/sql`.

Repositories cover:

- users and sessions;
- documents, chapters, charts, and evidence;
- claim diffs, claims, and claim-evidence links;
- tables, methods, results, citations, and paper relationships;
- annotations and collections;
- sharing, referrals, usage, analytics, and tiers.

Infrastructure files:

| File | Responsibility |
|---|---|
| `db.go` | Open SQLite, configure pragmas, apply pending migrations |
| `migrations.go` | Register and load all 19 migration files |
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
| `gemini.go` | Direct Gemini HTTP client and concurrency control |
| `pdf.go` | PDF text and embedded-image extraction |
| `crossref.go` | DOI metadata and source resolution |
| `unpaywall.go` | Open-access source lookup |
| `logger.go` | JSONL logging integration |

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
| `frontend/src/hooks/use-auth-submit.js` | Authentication submission state |
| `frontend/src/hooks/useResultState.js` | Result-page state composition |
| `frontend/src/components/result/` | Understanding, evidence, figures, source, header, and actions sections |
| `frontend/src/components/ui/` | Reusable UI primitives and reading-level selector |
| `frontend/src/pages/` | Route-level components |

## Frontend routes

| Route | Component | Purpose |
|---|---|---|
| `/` | `UploadPage` | PDF, paste, DOI, and URL input |
| `/upload` | redirect | Redirects to `/` |
| `/dashboard` | redirect | Redirects to `/account` |
| `/login` | `LoginPage` | Email/password and Google login |
| `/signup` | `SignupPage` | Account creation and Google signup |
| `/account` | `AccountPage` | Account, API key, usage, and billing |
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
| `GET /api/documents/{id}/claims` | Public | List claims |
| `GET /api/documents/{id}/tables` | Public | List tables |
| `GET /api/documents/{id}/methods` | Public | List methods |
| `GET /api/documents/{id}/results` | Public | List results |
| `GET /api/documents/{id}/citations` | Public | List citations |
| `GET /api/documents/{id}/evidence-graph` | Public | Get evidence graph |
| `GET /api/documents/{id}/research-map` | Public | Get grouped research relationships |
| `GET /api/documents/stats` | Required | User document statistics |
| `GET /api/documents/` | Required | User document list |
| `PUT /api/documents/{id}/save` | Required | Toggle saved state |
| `PATCH /api/documents/{id}` | Required | Update title |
| `DELETE /api/documents/{id}` | Required | Delete owned document |
| `POST /api/documents/compare` | Public | Compare two documents |
| `GET /api/papers/{id}/relationships` | Public | Get paper relationships |
| `POST /api/papers/{id}/relationships` | Public | Create paper relationship |

### Import

| Method and path | Auth | Purpose |
|---|---|---|
| `POST /api/import/doi` | Optional + usage limit | Resolve and ingest DOI |
| `POST /api/import/url` | Optional + usage limit | Fetch and ingest URL |

### Authentication and API keys

| Method and path | Auth | Purpose |
|---|---|---|
| `POST /api/auth/signup` | Public + rate limit | Create account and session |
| `POST /api/auth/login` | Public + rate limit | Authenticate and create session |
| `POST /api/auth/logout` | Public | Clear session |
| `GET /api/auth/me` | Public + rate limit | Current user/session state |
| `GET /api/auth/google/login` | Public | Begin Google OAuth |
| `GET /api/auth/google/callback` | Public | Complete Google OAuth |
| `GET /api/auth/apikey` | Required | Get current API key |
| `POST /api/auth/apikey/regenerate` | Required | Regenerate API key |

### Billing and account

| Method and path | Auth | Purpose |
|---|---|---|
| `POST /api/billing/checkout` | Required | Create Stripe checkout session |
| `POST /api/billing/portal` | Required | Create customer portal session |
| `POST /api/billing/webhook` | Stripe signature | Process billing events |
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
- `GEMINI_API_KEY` and `PAPERVIZ_API_KEY` required at startup;
- no tool calls Gemini;
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
| `GEMINI_API_KEY` | MCP only | Yes for MCP | Local stdio server calls the provider directly; the web server never reads it |
| `GEMINI_MODEL` | MCP only | No | `gemini-2.5-flash-lite` |
| `PAPERVIZ_API_KEY` | MCP | Yes for MCP | Process-level MCP key |
| `MIGRATIONS_DIR` | MCP, admin | No | `migrations` |

The server holds no model vendor credential. Web model calls run on the requesting user's own key, decrypted per request by `internal/app/credentials.Resolver`; there is no `GEMINI_API_KEY` and no `FRONTEND_URL` in the server path. `CREDENTIAL_ENCRYPTION_KEY` is the only secret `cmd/server` refuses to boot without.

The HTTP process does not load `.env`. Export values before running it.

## Commands

### Setup

```bash
go mod download
npm --prefix frontend ci
npm --prefix e2e ci
```

### Development

```bash
set -a; source .env; set +a
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
| `make run` | Build and run with `.env` exported |
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
| 019 | `019_billing_columns.sql` | Stripe customer and subscription state |

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
| `.github/workflows/go-ci.yml` | gofmt, vet, build, tests, container builds, Playwright |
| `.github/workflows/frontend-ci.yml` | npm install and production build |
| `.github/workflows/e2e.yml` | frontend build, server startup, health check, Playwright |

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
| Migration | `migrations/` and `internal/repository/migrations.go` |
| Pipeline | `internal/services/pipeline*.go`, stages, and data-pipeline doc |
| Design token | `DESIGN.md` and frontend styles |
| Current status | `docs/PROJECT_STATE.md` |
