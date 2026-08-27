# Cost Model — PaperViz

> **Last updated:** 2026-08-28
> **Model used:** Gemini 3.5 Flash (per .env.example)
> **Purpose:** Estimate cost-per-paper, gross margins, and free-tier abuse scenarios

---

## 1. Gemini API Pricing

| Model | Input (per 1M tokens) | Output (per 1M tokens) | Free Tier |
|-------|----------------------|------------------------|-----------|
| Gemini 3.5 Flash | $1.50 | $9.00 | Yes (rate-limited) |
| Gemini 2.5 Flash-Lite | $0.10 | $0.40 | Yes |

**PaperViz uses Gemini 3.5 Flash** for quality reasons (simplification, verification, chart generation require capable reasoning).

---

## 2. Gemini Calls per Document

### Pipeline Stages (sequential)

| Stage | Function | Calls | Purpose |
|-------|----------|-------|---------|
| 1. Simplify | `Simplify()` | 1 | Convert to reading level |
| 2. Verify | `DiffClaims()` | 2 | Claim extraction + diff |
| 3. Detect Chapters | `DetectChapters()` | 1 | Split into ≤10 chapters |
| 4. Generate Charts | `GenerateChapterChart()` | 1 per chapter | One chart per chapter |
| 5. Image Charts | `ReVisualizeCharts()` | 2 per image | Data extraction + annotation |

### Call Count by Document Type

| Document Type | Min Calls | Max Calls | Typical |
|---------------|-----------|-----------|---------|
| Pasted text (no charts) | 4 | 14 | 6-8 |
| PDF without images | 4 | 14 | 6-8 |
| PDF with images (≤5) | 4 | 24 | 10-14 |

**Constraints:**
- `maxImageChartsPerDocument = 5` caps image processing
- `DetectChapters` returns ≤10 chapters
- 3-second sleep between Gemini-heavy stages (rate limit recovery)

---

## 3. Token Estimates per Call

Based on prompt analysis in `services/*.go`:

| Operation | Input Tokens | Output Tokens | Notes |
|-----------|--------------|---------------|-------|
| Simplify | 8,000-15,000 | 6,000-12,000 | Proportional to paper length |
| DiffClaims (extraction) | 12,000-20,000 | 1,500-3,000 | Original + simplified text |
| DiffClaims (comparison) | 2,000-4,000 | 500-1,000 | Two claim sets |
| DetectChapters | 6,000-12,000 | 800-1,500 | Simplified text input |
| GenerateChapterChart | 4,000-8,000 | 1,500-3,000 | Chapter text + schema |
| Image data extraction | 800-1,500 | 400-800 | Image tokens + context |
| Image annotation | 800-1,500 | 300-600 | Image + page context |

**Assumptions:**
- Average paper: ~8,000 words ≈ 10,000 tokens
- Simplified text: ~60% of original length
- Chapter text: ~1,000-2,000 tokens each

---

## 4. Cost per Operation

### Using Gemini 3.5 Flash ($1.50 input / $9.00 output per 1M tokens)

| Operation | Input Tokens | Output Tokens | Cost |
|-----------|--------------|---------------|------|
| **Single-paper analysis** | | | |
| - Simplify | 12,000 | 9,000 | $0.018 + $0.081 = **$0.099** |
| - DiffClaims (2 calls) | 18,000 | 2,500 | $0.027 + $0.023 = **$0.050** |
| - DetectChapters | 9,000 | 1,200 | $0.014 + $0.011 = **$0.025** |
| - GenerateChapterChart (avg 3 chapters) | 18,000 | 6,750 | $0.027 + $0.061 = **$0.088** |
| **Subtotal (text-only paper)** | **57,000** | **19,450** | **$0.262** |
| | | | |
| **Figure analysis (per image)** | | | |
| - Data extraction | 1,200 | 600 | $0.002 + $0.005 = **$0.007** |
| - Annotation | 1,200 | 450 | $0.002 + $0.004 = **$0.006** |
| **Subtotal per image** | **2,400** | **1,050** | **$0.013** |
| | | | |
| **Paper comparison (2 papers)** | | | |
| - Per paper (simplified) | 57,000 | 19,450 | $0.262 |
| - Comparison call | 25,000 | 4,000 | $0.038 + $0.036 = **$0.074** |
| **Subtotal** | **139,000** | **42,900** | **$0.598** |

### Summary Table

| Operation | Cost per Call | Typical Volume | Cost per Unit |
|-----------|---------------|----------------|---------------|
| Single-paper analysis | $0.262 | 1 paper | **$0.26** |
| Figure analysis | $0.013 | 3 images/paper | **$0.04** |
| Paper comparison | $0.598 | 1 comparison | **$0.60** |
| Share page view | $0.00 | N/A | **$0.00** |

---

## 5. Storage & Bandwidth

### SQLite Storage per Document

| Data Type | Size Estimate |
|-----------|---------------|
| Original text | ~50 KB |
| Simplified text | ~30 KB |
| Claims JSON | ~5 KB |
| Charts JSON | ~10 KB |
| Metadata | ~1 KB |
| **Total per document** | **~96 KB** |

**Monthly storage at scale:**
- 1,000 documents/month → ~96 MB
- 10,000 documents/month → ~960 MB

### Bandwidth per Share Page View

| Component | Size |
|-----------|------|
| HTML (SPA) | ~50 KB |
| JS bundle | ~200 KB |
| CSS | ~30 KB |
| API response (document) | ~10 KB |
| **Total per view** | **~290 KB** |

**Monthly bandwidth at scale:**
- 10,000 views/month → ~2.9 GB
- 100,000 views/month → ~29 GB

### Infrastructure Costs (Estimated)

| Item | Monthly Cost |
|------|--------------|
| VPS (2 vCPU, 4GB RAM) | $20-40 |
| Domain + SSL | $1-2 |
| SQLite backup storage | $1-2 |
| **Total infra** | **$22-44** |

---

## 6. Tier Economics

### Current Tier Limits

| Tier | Papers/Month | Assumed Price | Cost-to-Serve | Gross Margin |
|------|--------------|---------------|---------------|--------------|
| Free | 5 | $0 | $1.31 | N/A |
| Pro | 50 | $20 (est) | $13.10 | **34.5%** |
| Research | 500 | $100 (est) | $131.00 | **-31%** |

**Assumptions:**
- Average 3 images per paper
- Cost per paper: $0.262 + (3 × $0.013) = **$0.301**
- Pro/Research pricing is placeholder (roadmap 6.4 validates)

### Free-Tier Abuse Scenario

**Worst case:** User creates 5 papers/month, each with 5 images, max chapter count

| Scenario | Calls | Tokens | Cost |
|----------|-------|--------|------|
| 5 papers × 24 calls | 120 | ~360K input, ~120K output | **$1.51** |
| Max chapter charts (10/paper) | +20 | +60K input, +30K output | +$0.36 |
| **Absolute worst case** | **140** | **~420K in, ~150K out** | **$1.87** |

**Free-tier cost cap: ~$1.87/user/month**

---

## 7. Sensitivity Analysis

### Key Variables

| Variable | Base Case | Optimistic | Pessimistic |
|----------|-----------|------------|-------------|
| Paper length | 8,000 words | 4,000 words | 15,000 words |
| Images per paper | 3 | 0 | 5 |
| Chapters per paper | 3 | 1 | 10 |
| Gemini pricing | $1.50/$9.00 | $0.75/$3.75 (promo) | $3.00/$18.00 (2x) |

### Cost Impact

| Scenario | Cost per Paper |
|----------|----------------|
| Base case | $0.30 |
| Short paper, no images | $0.15 |
| Long paper, 5 images, 10 chapters | $0.52 |
| Gemini price increase (2x) | $0.60 |
| Gemini promo pricing | $0.15 |

---

## 8. Recommendations

1. **Free tier is viable:** Max abuse cost ~$1.87/user/month is acceptable for acquisition
2. **Pro tier needs validation:** At $20/month with 50 papers, margin is thin (34.5%). Consider:
   - Raising Pro price to $25-30
   - Reducing Pro limit to 30-40 papers
3. **Research tier is unprofitable at current limits:** 500 papers at $100 = -$31 margin. Options:
   - Cap Research at 200 papers
   - Price at $150-200
   - Require custom pricing for high-volume
4. **Monitor actual token usage:** Estimates above are conservative. Real usage may be 20-40% lower
5. **Consider Gemini 2.5 Flash-Lite** for simpler operations (chart annotation) to reduce costs by ~10x

---

## 9. Open Questions

- [ ] Actual token usage in production (need logging)
- [ ] Real paper length distribution
- [ ] Image density per paper type
- [ ] Pro/Research willingness to pay (roadmap 6.4)
- [ ] Whether to use Flash-Lite for cost-insensitive operations
