# Chunk 7.6 — MCP Usage, Security & Cost Controls

## Facts

1. MCP authentication uses environment variable `PAPERVIZ_API_KEY` — agents set this in their MCP config
2. Rate limits: analyze_paper 5/min, read ops 30/min, compare 2/min per API key
3. Max text size for analyze_paper: 500KB (~100K words)
4. Max concurrent analyze_paper jobs per key: 2
5. Analysis timeout: 5 minutes
6. MCP reuses existing free/pro/research tier system — API keys map to tiers
7. Cost attribution: log operation, tokens used, estimated cost to analytics_events
8. Abuse protection: prevent retry loops via idempotency keys
9. Clear error semantics: structured JSON errors with error codes
10. Request size limits enforced before Gemini calls
