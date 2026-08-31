# Chunk 7.6 — MCP Usage, Security & Cost Controls

## Goal
Make MCP agent access safe and economically viable with authentication, rate limits, usage tracking, and cost attribution.

## Implementation Steps

### Step 1: Add API key authentication to MCP server
- File: `internal/mcp/server.go`
- Add `apiKey` parameter to `NewMCPServer`
- Validate `PAPERVIZ_API_KEY` env var at startup
- Pass API key through context to tool handlers
- Verify: `go test ./internal/mcp/...`

### Step 2: Add rate limiter for MCP operations
- File: `internal/mcp/ratelimit.go` (new)
- Token bucket rate limiter per API key
- Separate limits: analyze_paper (5/min), read ops (30/min), compare (2/min)
- Verify: unit tests for rate limiter

### Step 3: Add request size validation
- File: `internal/mcp/tools.go`
- Max text size constant: 500KB
- Validate in `handleAnalyzePaper` before processing
- Verify: test with oversized input

### Step 4: Add concurrent job limiter
- File: `internal/mcp/jobs.go` (new)
- Semaphore per API key (max 2 concurrent analyses)
- Track active jobs, reject when at limit
- Verify: unit tests for job limiter

### Step 5: Add analysis timeout
- File: `internal/mcp/tools.go`
- Context with 5-minute timeout for `handleAnalyzePaper`
- Return clear error on timeout
- Verify: test timeout path

### Step 6: Add tier mapping for API keys
- File: `internal/mcp/auth.go` (new)
- Map API key to tier (free/pro/research)
- Reuse existing `TierService` logic
- Store API key → tier mapping in DB or env config
- Verify: unit tests for tier resolution

### Step 7: Add cost attribution logging
- File: `internal/mcp/cost.go` (new)
- Log to `analytics_events`: operation, api_key, tokens_used, estimated_cost
- Hook into each tool handler
- Verify: check analytics_events table after tool calls

### Step 8: Add abuse protection
- File: `internal/mcp/tools.go`
- Idempotency key support (optional header)
- Prevent duplicate analysis within 5-minute window
- Verify: test duplicate detection

### Step 9: Add structured error responses
- File: `internal/mcp/errors.go` (new)
- Error types: rate_limited, auth_failed, timeout, size_limit, job_limit
- Consistent JSON error format across all tools
- Verify: test each error type

### Step 10: Update MCP server entrypoint
- File: `cmd/mcp/main.go`
- Read `PAPERVIZ_API_KEY` from env
- Pass to `NewMCPServer`
- Log startup with configured limits
- Verify: manual test with valid/invalid keys

## Verification

1. `go test ./internal/mcp/...` — all tests pass
2. `go test ./...` — full suite passes (159+ tests)
3. `go vet ./...` — no issues
4. Manual: MCP server rejects invalid API key
5. Manual: Rate limit triggers after threshold
6. Manual: Oversized text rejected with clear error
7. Manual: Concurrent job limit enforced

## Risks

- API key storage: env var is simplest but not most secure; documented as acceptable for MVP
- Tier mapping: need to decide between DB table or env config for key→tier mapping
- Cost attribution: Gemini API doesn't expose per-call token counts in current integration
