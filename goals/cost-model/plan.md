# Cost Model Plan

## Solution Approach
Create `docs/cost-model.md` documenting all cost drivers, per-operation cost models, and tier margin analysis.

## Steps

### 1. Research Gemini Pricing
- Check current Gemini Flash pricing (input/output tokens)
- File: external research (web search)
- Verify: pricing table in cost-model.md

### 2. Count Gemini Calls Per Operation
- Analyze pipeline.go, charts.go, comparison.go
- Document: Simplify (1), DiffClaims (2), DetectChapters (1), GenerateChapterChart (1/chapter), ReVisualizeCharts (2/image)
- File: cost-model.md "Gemini Calls per Operation" section
- Verify: call count table matches code

### 3. Estimate Token Usage Per Call
- Sample prompts in services/*.go
- Estimate input/output tokens per operation type
- File: cost-model.md "Token Estimates" section
- Verify: estimates are reasonable (not order-of-magnitude off)

### 4. Calculate Per-Operation Costs
- Multiply calls × tokens × pricing
- Include: paper analysis, figure analysis, comparison
- File: cost-model.md "Cost per Operation" section
- Verify: math is correct, assumptions documented

### 5. Model Storage & Bandwidth
- SQLite growth per paper (text + simplified + charts)
- Share page bandwidth per view
- File: cost-model.md "Storage & Bandwidth" section

### 6. Tier Margin Analysis
- Free/Pro/Research limits from tier.go
- Cost-to-serve vs assumed revenue
- Maximum free-tier abuse scenario
- File: cost-model.md "Tier Economics" section

### 7. Sensitivity Analysis
- Key variables: Gemini pricing changes, paper length variance, chart density
- File: cost-model.md "Sensitivity" section

## Verification
- `go test ./...` (no code changes, but ensure no breakage)
- Manual review of calculations
