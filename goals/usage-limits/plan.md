# Plan — Chunk 6.2: Usage Limits

## Solution Approach
Add tier-based usage limits enforced at document creation. Track usage via browser fingerprint + IP. Show usage on frontend with upgrade CTA.

## Steps

### Step 1: Database Schema
**Files:** `migrations/013_usage_tiers.sql`
- Create `user_tiers` table: fingerprint TEXT PRIMARY KEY, tier TEXT DEFAULT 'free', papers_used INTEGER DEFAULT 0, last_reset TIMESTAMP
- Add index on fingerprint

**Verification:** `sqlite3 paperviz.db ".schema user_tiers"`

### Step 2: Tier Service
**Files:** `internal/services/tier.go`
- `GetTier(fingerprint string) (string, error)` — returns current tier
- `CheckUsage(fingerprint string) (bool, int, error)` — returns (canCreate, papersUsed, error)
- `IncrementUsage(fingerprint string) error` — bumps papers_used
- `ResetMonthlyUsage() error` — resets all counts on first of month
- Tier limits as constants: free=5, pro=50, research=500

**Verification:** `go test ./internal/services/... -run Tier`

### Step 3: Usage Middleware
**Files:** `internal/handlers/middleware.go`
- `UsageLimitMiddleware(next http.Handler) http.Handler`
- Extracts fingerprint from request (header or generated)
- Calls `CheckUsage` before document creation
- Returns 429 with upgrade CTA JSON if limit reached

**Verification:** Unit test with mock fingerprint hitting limit

### Step 4: Fingerprint Generation
**Files:** `internal/handlers/fingerprint.go`
- `GenerateFingerprint(r *http.Request) string` — combines IP + User-Agent + Accept-Language
- `GetFingerprint(r *http.Request) string` — checks header first, generates if missing

**Verification:** Unit test fingerprint consistency

### Step 5: Frontend Usage Display
**Files:** `frontend/src/components/usage-display.jsx`
- Shows current tier, papers used this month, limit
- Progress bar (green < 80%, yellow 80-99%, red 100%)
- Fetches from `GET /api/usage` endpoint

**Verification:** Visual check on upload page

### Step 6: Upgrade CTA
**Files:** `frontend/src/components/upgrade-cta.jsx`
- Shown when at limit
- "Upgrade to Pro" / "Upgrade to Research" buttons
- Links to placeholder pricing page (can be TODO for now)

**Verification:** Trigger limit, see CTA appear

### Step 7: Usage API Endpoint
**Files:** `internal/handlers/usage.go`
- `GET /api/usage` — returns {tier, papers_used, limit, reset_date}
- Protected by auth (or fingerprint if no auth)

**Verification:** `curl localhost:8080/api/usage`

### Step 8: Wire Everything
**Files:** `internal/handlers/router.go`
- Add usage middleware to POST /api/documents
- Register GET /api/usage route

**Verification:** Full flow: create 5 papers as free user, 6th blocked

## Risks
- Fingerprint spoofing — acceptable for MVP, upgrade to auth later
- Monthly reset — simple cron or check on each request
- No payment integration — tiers exist but no way to upgrade yet (placeholder)
