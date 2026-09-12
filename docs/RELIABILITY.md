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
| Gemini simplify/verify | Timeout, 429, 5xx | Pipeline stage `simplify`/`verify` fails, `pipeline stage failed` logged | `slog.Error` + `writeError 500` | No automatic retry; user re-uploads |
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

* No background worker/job queue. Pipeline runs synchronously in `POST /api/documents` handler (no ORM, no broker per non-goals). Request holds connection.

## 5. Retries

### Retryable Errors

* Gemini 429 (`Gemini rate limited`), 5xx with backoff in `external/gemini.go: GenerateWithRetry` (up to 3, `backoff = 2s << attempt`).

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

## 7. Job / Worker Reliability

No async jobs. Synchronous handler model.

### Job States

```
pipeline: simplify → verify → chapters → charts → persist
         ↓ fail → log "pipeline stage failed" → 500
```

`processing_stages` table tracks per-stage status (updated via `intake.go: update processing stage failed`).

### Duplicate Jobs

* No duplicate detection beyond IP rate limiting (`1 req/30s burst 2`).

### Worker Crash

* Server `Recoverer` middleware converts panic to 500; DB WAL ensures committed writes survive.

### Stuck Jobs

* No stuck-job detector; pipeline timeout is request timeout. Leaked PDF goroutines remain observable via `Warn` log.

## 8. Dependency Failure

### Model Provider (Gemini)

* 429/5xx → backoff retry (3). After failure, pipeline returns 500, document not stored. `logChartFailure` categories: `CHART_SELECTION_ERROR`, `GROUNDING_ERROR`.

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
* Synchronous pipeline holds HTTP connection; large PDFs under Gemini latency may hit client timeout.
* Leaked PDF goroutines on timeout remain running until completion (observable via Warn, not bounded).
* Single SQLite file is SPOF; no replica/healthcheck beyond `WAL` + `synchronous=NORMAL`.

## 13. Reliability Improvements

* `ratelimit.go` TTL eviction (5 min) prevents unbounded map memory growth.
* `maxImageChartsPerDocument=5` bounds quota burn free-tier.
* `dualClaimExtractionPrompt` reduces Gemini calls 3→2 for verification.
* `GenerateChapterCharts` per-dataset evaluation isolates failures (`degraded` flag).
