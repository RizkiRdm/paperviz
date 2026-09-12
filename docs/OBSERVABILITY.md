# Observability

## 1. Observability Goals

* Why did this request fail (auth, validation, pipeline stage, Gemini)?
* Which pipeline stage is failing (simplify/verify/chapters/charts)?
* Why is chart missing or degraded?
* How much does a request cost (per-doc Gemini token usage)?

## 2. Signals

### Logs

Structured `log/slog` JSONL via `internal/external/logger.go` (`NewJSONLHandler`). All request/pipeline/gemini/pdf logs go through `slog`.

### Metrics

*Not implemented.* No Prometheus/metrics endpoint. Usage counted via `share_visits/conversions` and `user_tiers` fingerprint counters only.

### Traces

*Not implemented.* No distributed tracing. Request correlation via `RequestID` only.

## 3. Request Identity

### Request ID

`chi/middleware.RequestID` injects `X-Request-ID` header + `ctx`. `slogRequestLogger` logs `id` alongside `method, path, status, duration`.

### Agent Run ID

*Not implemented.* No agent runs (MCP tools are stateless read-only).

### Job ID

Document `id` (`repository.NewID()`). Used in `processing_stages` and pipeline logs (`document_id`).

### User / Tenant ID

`user_id` from `session_token` cookie via `RequireAuth` → `context.WithValue(userIDKey)`. Logged via `UserIDFromContext` where applicable; anonymous docs use fingerprint.

## 4. Structured Logging

### Log Format

```json
{"time":"...","level":"INFO|WARN|ERROR","msg":"request|pipeline stage failed|chart failure|gemini call","id":"req-...","method":"POST","path":"/api/documents","status":200,"duration":"1.2s","stage":"simplify","error":"..."}
```

`internal/external/logger.go` JSONLHandler; sensitive fields (document text, password_hash, tokens) never logged (ARCHITECTURE Logging Policy).

### Required Fields

| Field | Purpose |
|-------|---------|
| `id` | Request correlation (`middleware.RequestID`) |
| `method, path, status, duration` | HTTP observability (`slogRequestLogger`, 400→Warn, 500→Error) |
| `stage, chapter, dataset_id, error_category` | Pipeline/chart failure (`logChartFailure`, `pipeline stage failed`) |
| `document_id` | Intake correlation |

## 5. Metrics

*Currently none exposed.*

### Application Metrics

| Metric | Type | Meaning | Implemented |
|--------|------|---------|-------------|
| `share_visits` | counter | per-doc/chart visits | Yes (DB) |
| `papers_used` | gauge | fingerprint monthly count | Yes (usage) |
| HTTP latency/duration | log field | per-request | Yes (slog) |
| Pipeline stage duration | — | — | Not measured |

### Agent Metrics

*Not implemented* (no agent execution metrics).

### Infrastructure Metrics

*Not implemented* (no CPU/mem/DB pool metrics).

## 6. Latency

`duration.String()` logged per request. No histogram.

Track latency for (logged but not aggregated):

* API request (via `slogRequestLogger`)
* Document parsing (pdf 2.5s timeout, Warn on exceed)
* Evidence extraction (regex, negligible)
* Validation (grounding, negligible)
* Visualization generation (Gemini per-chart 3–8s, `chart generation complete` log)
* External API calls (Gemini client `gemini call` with retry backoff)

## 7. Errors

### Error Taxonomy

```text
CLIENT_ERROR (400 invalid_request/invalid_email/password_too_weak/email_taken, 401 unauthenticated/invalid_credentials, 429 rate_limited)
VALIDATION_ERROR (400 state mismatch missing_oauth_state/invalid_oauth_state)
DEPENDENCY_ERROR (Gemini 429/5xx, Crossref/Unpaywall fetch)
DATABASE_ERROR (sqlite Exec/Query, 500 internal_error)
INTERNAL_ERROR (500 internal_error, hash/generate failures)
CHART_ERROR (EXTRACTION_ERROR|DATASET_ERROR|CHART_SELECTION_ERROR|GROUNDING_ERROR|SCHEMA_ERROR|RENDER_ERROR)
```

### Error Tracking

`writeError(w, status, code)` generic codes to client; `slog.Error/Warn` server-side with `error, stage, chapter, dataset_id`. No error aggregation service.

## 8. Agent Observability

Stateless MCP research tools (no agent run persistence).

### Agent Steps

Not an autonomous agent loop; pipeline steps logged as `stage=simplify|verify|chapters|chart` in `pipeline.go`/`charts.go`.

### Tool Calls

MCP tools share service layer with REST (`docs/mcp-parity.md`); tool usage logged via `slog` when applicable (Gemini calls).

### Model Usage

* model: `gemini-...` via `external.GeminiClient`
* input/output tokens: not explicitly logged per call (estimated via `cost-model.md` only)
* latency: per-call `Gemini call` log
* estimated cost: tier margin analysis in `docs/cost-model.md` (not runtime)

## 9. Health Checks

### Liveness

`GET /` serves landing; `middleware.Recoverer` keeps process up. No `/healthz`.

### Readiness

DB opened at startup; `WAL` mode enabled. No readiness probe.

### Dependency Health

Gemini reachability assumed via direct HTTP; failures observed as stage errors. No dependency health endpoint.

## 10. Debugging Workflow

```
Request ID (X-Request-ID header, id field)
   ↓ Find request logs (slogRequestLogger)
   ↓ Find stage logs (slog.Error stage=simplify|verify|chapters|chart)
   ↓ Find chart failure (slog.ErrorContext error_category, dataset_id)
   ↓ Inspect dependency (Gemini timeout/429, pdf Warn, DB error)
   ↓ Identify root cause (invalid input, grounding unsupported, CSRF state mismatch)
```

## 11. Alerts

| Condition | Threshold | Severity | Action |
|-----------|----------:|----------|--------|
| *No alerts configured* | — | — | Check logs |

No Alertmanager/PagerDuty.

## 12. Known Observability Gaps

* No metrics endpoint (latency histograms, error rates, token usage not collected).
* No tracing beyond RequestID.
* No health endpoint.
* Expired session accumulation — mitigated: `DeleteExpired()` on startup (no periodic sweep).
* OAuth state previously hardcoded — fixed: random nonce via `crypto/rand` + httpOnly cookie 10m TTL.

## 13. Planned Improvements

* Metrics: per-stage latency histogram, Gemini token cost per document, rate-limit hit counters.
* Health: `/healthz` + DB pool check.
* Tracing: propagate `id` through pipeline `ctx`.
* Periodic session sweeper.
