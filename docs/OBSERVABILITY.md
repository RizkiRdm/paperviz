# `docs/OBSERVABILITY.md`

# Observability

## 1. Observability Goals

What questions should operators be able to answer?

Examples:

* Why did this request fail?
* Why is this agent run slow?
* Which processing stage is failing?
* How much does a request cost?

## 2. Signals

### Logs

*

### Metrics

*

### Traces

*

## 3. Request Identity

### Request ID

*

### Agent Run ID

*

### Job ID

*

### User / Tenant ID

*

## 4. Structured Logging

### Log Format

```json
{
  "timestamp": "",
  "level": "",
  "message": "",
  "request_id": "",
  "component": ""
}
```

### Required Fields

| Field | Purpose |
| ----- | ------- |
|       |         |

## 5. Metrics

### Application Metrics

| Metric | Type | Meaning |
| ------ | ---- | ------- |
|        |      |         |

### Agent Metrics

| Metric | Type | Meaning |
| ------ | ---- | ------- |
|        |      |         |

### Infrastructure Metrics

| Metric | Type | Meaning |
| ------ | ---- | ------- |
|        |      |         |

## 6. Latency

Track latency for:

* API request
* Agent execution
* Document parsing
* Evidence extraction
* Validation
* Visualization generation
* External API calls

## 7. Errors

### Error Taxonomy

```text
CLIENT_ERROR
VALIDATION_ERROR
AGENT_ERROR
DEPENDENCY_ERROR
DATABASE_ERROR
INTERNAL_ERROR
```

### Error Tracking

*

## 8. Agent Observability

### Agent Steps

*

### Tool Calls

*

### Model Usage

* model
* input tokens
* output tokens
* latency
* estimated cost

## 9. Health Checks

### Liveness

*

### Readiness

*

### Dependency Health

*

## 10. Debugging Workflow

```text
Request ID
   ↓
Find request logs
   ↓
Find agent run
   ↓
Find failed stage
   ↓
Inspect dependency
   ↓
Identify root cause
```

## 11. Alerts

| Condition | Threshold | Severity | Action |
| --------- | --------: | -------- | ------ |
|           |           |          |        |

## 12. Known Observability Gaps

- OAuth state parameter was previously hardcoded (CSRF vulnerability) — fixed in Chunk 11.7: random nonce generated via `crypto/rand`, stored in httpOnly cookie (10min TTL), validated in callback
- Expired session accumulation — fixed: `DeleteExpired()` called on server startup
- /me endpoint previously had no rate limiting — fixed: wrapped with `rateLimitAuth`

## 13. Planned Improvements

- Add per-user session count metric for abuse detection
- Implement request correlation ID tracing across OAuth flow
- Add password strength histogram for UX analytics
- Monitor OAuth error rates for brute-force detection

---