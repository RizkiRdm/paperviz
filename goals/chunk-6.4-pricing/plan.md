# Chunk 6.4 — Pricing & Packaging Experiment (Revised)

## Goal
Validate willingness to pay without building complex billing. Define packages, test pricing, track conversion metrics.

## Scope

### Deliverables
1. `docs/pricing-strategy.md` — Package definitions, pricing, experiment design
2. `frontend/src/pages/pricing-page.jsx` — 3-column pricing page
3. Router integration — `/pricing` route
4. `frontend/src/components/upgrade-cta.jsx` — Link to `/pricing`
5. Conversion tracking — pricing views, engaged time, CTA clicks

### Price Points (Revised)
- **Free:** $0 — 5 papers/month, basic analysis
- **Pro:** $29/month — 50 papers/month, 52% margin at $0.30/paper
- **Research:** "Contact us" — no fixed price, custom for high-volume

### Feature Differentiation
Volume only for MVP experiment. No comparison/exports/API gates yet.

### Upgrade Flow
Waitlist/email capture. No payment processing. Transparent "Coming Soon" messaging.

### Conversion Metrics
1. Pricing page view
2. Time on page >10s (engaged)
3. CTA click
4. Email capture (waitlist)

### Experiment Timeline
30 days. Success = 5% of free users click upgrade at least once.

## Implementation Steps

### Step 1: Pricing Strategy Document
**File:** `docs/pricing-strategy.md`
- Package definitions with features
- Price points justified by cost model
- Success metrics and experiment design
- Timeline for evaluation

### Step 2: Pricing Page
**File:** `frontend/src/pages/pricing-page.jsx`
- 3-column layout (Free / Pro / Research)
- Feature comparison table
- Price display
- CTA buttons (upgrade or waitlist)

### Step 3: Router Integration
**File:** `internal/handlers/router.go` — add `/pricing` route (static file serving)
**File:** `frontend/src/App.jsx` — add React Router route

### Step 4: Upgrade CTA Update
**File:** `frontend/src/components/upgrade-cta.jsx`
- Link to `/pricing` instead of "#"
- Pass current tier context

### Step 5: Conversion Tracking
**File:** `internal/handlers/analytics.go` — add `TrackPricingView`, `TrackUpgradeIntent`
**File:** `frontend/src/pages/pricing-page.jsx` — fire tracking events on page view, 10s engagement, CTA click

## Verification
- [ ] Pricing page loads at `/pricing`
- [ ] Feature comparison is clear and accurate
- [ ] Upgrade CTAs link to pricing page
- [ ] Conversion events fire on page view and click
- [ ] Build passes (`go build ./...`, `cd frontend && npm run build`)
- [ ] Tests pass (`go test ./...`)

## Constraints
- No payment processing
- No enterprise billing
- No complex entitlements
- Follow existing design system (DESIGN.md)
- Match existing React patterns
- Follow ARCHITECTURE.md layers
