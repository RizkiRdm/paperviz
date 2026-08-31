# Chunk 7.6 — MCP Usage, Security & Cost Controls

## Goal
Make MCP agent access safe and economically viable with authentication, rate limits, usage tracking, and cost attribution.

## Reference
- Facts: `goals/chunk-7-6-security/facts.md`
- Plan: `goals/chunk-7-6-security/plan.md`

## Done Condition
- MCP server requires valid API key
- Rate limits enforced per API key
- Request size limits enforced
- Concurrent job limits enforced
- Analysis timeout implemented
- Cost attribution logged
- All tests pass (159+)
- `go vet ./...` clean
