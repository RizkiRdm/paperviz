# Reliability

## 1. Reliability Goals

* Document pipeline completes end-to-end or reports a typed failure without crashing server.
* One chart failure must not abort other charts (ARCHITECTURE.md Failure Scenarios 3,4).
* 7-day ephemeral data survives process restarts (WAL + synchronous=NORMAL).

## 2. Reliability Boundaries

```
[CLIENT] → [API: chi + middleware] → [SERVICES: pipeline/intake/charts/verification] → [EXTERNAL: Gemini, pdf extraction, Crossref/Unpaywall] → [DB: SQLite/WAL]
```

Failure in Gemini or PDF extraction is isolated per-stage; failure in DB is global.

## 3. Failure Modes

| Component | Failure | Impact | Detection | Recovery |
|-----------|---------|--------|-----------|----------|
| PDF extraction | Timeout (2.5s text/image), large PDF 20 MiB | Document rejected, `400` | `context.WithTimeout` + `slog.Warn` timeout branch | User retries with smaller PDF |
| Gemini simplify/verify | Timeout, 429, 5xx | Background pipeline stage fails; document is marked `failed` with an error message | `slog.Error` with document ID and stage | User retries with a new upload |
| Chart evidence extraction | No evidence found | `no evidence extracted` `slog.Info`, `degraded=false` | `len(datasets)==0` in `GenerateChapterCharts` | Degraded success (no charts, document still saved) |
| Chart grounding | Inconsistent units, missing evidence, negative pie, NaN/Inf | `Grounded=unsupported`, chart not rendered, `logChartFailure` | `ValidateGrounding` 10 rules | Fallback: text evidence only, other charts unaffected |
| Gemini rate limit (server 429) | `Gemini rate limited by server` | Stage fails | `slog.Error` | No retry loop (client-side backoff only in `gemini.go` retry up to 3 with exponential backoff) |
| DB | `modernc.org/sqlite` error, WAL lock | Request 500 | `slog.Error` | Restart, `rm paperviz.db*` + `make dev` if schema changed |
| Rate limiter | Memory exhaustion via many IPs | Memory growth | In-memory map with 5-min TTL | Eviction `purgeOldEntries` + lazy refresh per `getLimiter` |
| Session | Expired/invalid token | 401 `unauthenticated` | `SessionRepo.Get` checks `expires_at` | User logs in again; `DeleteExpired` on startup |

## 4. Timeouts

### External Requests

* Gemini client: `context` with timeout (default `30s` per call, 3 attempts with backoff `2s << attempt`). Rate-limit 429 triggers 5s sleep + retry.
* PDF extraction: `context.WithTimeout(2.5s)` per text/image page; warns `"pdf ... timed out, background goroutine still running"` and returns partial.

### Internal Operations

* No explicit timeout for DB queries; synchronous `database/sql`.

### Job Execution

* No background worker or job queue. Web ingestion starts the pipeline in an in-process goroutine after intake; `POST /api/documents` returns a document ID while processing continues. No broker or external worker is used.

## 5. Retries

### Retryable Errors

* Provider 429 and 5xx with backoff in `external/llm.go` (`Transport`, up to 3, `backoff = 2s << attempt`). The provider SDK's own retry is disabled with `WithMaxRetries(0)` so the two schedules cannot compound.

### Non-Retryable Errors

* Validation errors (`invalid_email`, `password_too_weak`, `email_taken`, `invalid_credentials`), PDF size >20 MiB, `state` mismatch, grounding `unsupported`.

### Retry Strategy

```text
Gemini: attempt 0 → backoff 2s → attempt 1 → backoff 4s → attempt 2 → fail
PDF: no retry, return partial + Warn
```

### Maximum Attempts

* Gemini: 3 total.
* Pipeline stages: 1 (fail fast, log `pipeline stage failed`).

### Backoff

* Exponential: `time.Sleep(2s * (1 << attempt))` + special 5s for 429.

## 6. Idempotency

### Idempotent Operations

* `GET /api/documents/:id`, `GET /share/doc/:token` (visit++ is `UPDATE ... SET visits=visits+1` but safe to repeat), `POST /api/auth/logout` (deletes session, clears cookie).

### Non-Idempotent Operations

* `POST /api/documents` (creates document + pipeline run), `POST /api/auth/signup` (unique email), `POST /api/documents/:id/share` (generates new `share_token`).

### Idempotency Strategy

* No deduplication key. Retry of `POST /api/documents` creates duplicate document. `share_tokens` lazily generated; revoking `private` clears token.

## 7. In-Process Processing Reliability

PaperViz has no durable job queue or external worker. Web ingestion starts `RunPipelineAndPersist` in a goroutine after the intake transaction, and the client polls the document while it runs.

### Processing States

```text
intake → simplify → verify → chapters/evidence → figures → persist
                                      ↓
                         failure recorded on document row
```

`processing_stage` records the current user-visible pipeline stage. A document remains `processing` until the pipeline writes a terminal state or fails.

### Duplicate Work

* No durable deduplication key exists. A client retry of `POST /api/documents` creates another document and another in-process run.
* IP rate limiting limits burst creation but does not provide idempotency.

### Process Failure

* HTTP `Recoverer` converts handler panics to 500 responses.
* A process exit can interrupt an in-flight goroutine; there is no durable replay mechanism.
* SQLite WAL preserves committed writes that reach disk before exit.

### Stuck Processing

* The background pipeline has a 20-minute context timeout.
* There is no separate stuck-job detector because no durable job registry exists.
* If a document remains `processing`, inspect pipeline logs, Gemini failures, PDF timeout warnings, and whether the intake originated through MCP-only flow.


## 8. Dependency Failure

### Model Provider (Gemini)

* 429/5xx → backoff retry (3). After exhaustion, the in-process pipeline records a failed document state and error message; there is no HTTP 500 because processing is asynchronous. `logChartFailure` categories: `CHART_SELECTION_ERROR`, `GROUNDING_ERROR`.

### Storage (SQLite)

* Single file, no replica. `WAL` mode. Failure surfaces as 500. No retry.

### External APIs (Crossref/Unpaywall)

* DOI import: SSRF guard + 100 MiB cap; failure → `500` with `internal_error`. No retry.

## 9. Graceful Degradation

* Charts: one chart failure → `omitted` or `image_fallback` with annotation; rest of document unaffected (`reVisualizeOne` isolated per chart).
* Verification: if verification fails, save with `verification_failed` + `mismatch_detail` + claims; badge gated.
* Imports: if DOI resolution fails, user fallback to paste/upload.
* Share pages work without auth, noindexed, tolerant of missing charts.

## 10. Recovery

### Recovery Procedure

1. Check `slog` `request` line for `status, duration, id`; then pipeline `stage` logs.
2. For DB inconsistency: `kill server; rm paperviz.db* ; make dev` (ephemeral data).

### Data Recovery

* No backup; 7-day TTL ephemeral. Re-upload document.

### Retry / Replay

* Client retries `POST /api/documents` (creates new document ID).

## 11. Failure Testing

| Scenario | Expected Result | Tested? |
|----------|-----------------|---------|
| Dependency timeout (Gemini) | Backoff 3, then 500 | Manual via `gemini.go` retry logs |
| Invalid model output (chart JSON) | `tryExtractChartData` → fallback → `omitted` | Covered by `chart_regression_test.go` (8), `chart_validation_test.go` (5), `charts_evidence_test.go` |
| Database unavailable | 500 `internal_error` | Not automated |
| Worker crash (panic) | 500 via `Recoverer` | Not automated |
| Duplicate request (rate limited) | 429 `rate_limited` | `ratelimit` limiter tests (IP-based) |

## 12. Known Reliability Risks

* No periodic sweeping of expired sessions (only on startup).
* In-process background processing shares the single SQLite connection; long Gemini latency can keep a document in `processing`, and there is no durable replay mechanism.
* Leaked PDF goroutines on timeout remain running until completion (observable via Warn, not bounded).
* Single SQLite file is SPOF; no replica/healthcheck beyond `WAL` + `synchronous=NORMAL`.

## 13. Reliability Improvements

* `ratelimit.go` TTL eviction (5 min) prevents unbounded map memory growth.
* `maxImageChartsPerDocument=5` bounds quota burn free-tier.
* `dualClaimExtractionPrompt` reduces Gemini calls 3→2 for verification.
* `GenerateChapterCharts` per-dataset evaluation isolates failures (`degraded` flag).
