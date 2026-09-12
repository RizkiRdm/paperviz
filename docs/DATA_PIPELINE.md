# Data Pipeline

## 1. Purpose

Turn an academic paper (PDF text-layer, pasted text, or DOI/URL import) into simplified language at chosen reading level plus evidence-grounded visualizations. AI may transform evidence but must never manufacture numeric values — every chart value is traceable to extracted evidence. Pipeline runs synchronously in `POST /api/documents` (no queue/broker).

## 2. Pipeline Overview

```
[INPUT: PDF ≤20MiB | pasted text | DOI/URL] 
   ↓ ExtractText + ExtractCharts (pdfcpu/ledongthuc/pdf, in-memory, 2.5s timeout/page)
[SPLIT: DetectChapters() → ≤10 chapters]
   ↓ Simplify (Gemini direct HTTP) 
[VERIFY: dualClaimExtractionPrompt → DiffClaims (2 calls) → mismatch_detail + claims rows in same tx]
   ↓ Evidence
[EXTRACT: ExtractNumericEvidence (5 regex) + ExtractTableData → BuildCandidateDatasets (metric||unit grouping)]
   ↓ Validate
[GROUND: ValidateGrounding (10 rules) → verified/partial/unsupported → DO NOT RENDER if unsupported]
   ↓ Plan
[CHARTS: GenerateChapterCharts per-dataset LLM plan → ChartSpec (bar/line/scatter/pie) | image_fallback | omitted]
   ↓ Persist
[OUTPUT: SQLite documents/charts/claims/tables/methods/results/citations + share 7-day TTL]
```

One chart failure must not abort others (reVisualizeOne isolation).

## 3. Input

### Supported Input

* PDF with text layer (no OCR), pasted raw text, DOI (Crossref+Unpaywall), URL — via `POST /api/documents` (multipart `document` or `text`) and `POST /api/import/doi|url`.
* Levels: `ELI5` / `Simplified`.

### Input Constraints

* PDF ≤20 MiB (`maxUploadBytes = 20 << 20` + `http.MaxBytesReader` 1 MiB slack on JSON; `header.Size > maxUploadBytes` rejected). No disk writes.
* DOI/URL: `https`-only, SSRF block on private IPs (`10/8, 172.16/12, 192.168/16, 127/8`), 100 MiB read cap.
* IP rate limiting: `POST /api/documents` `1 req/30s burst 2`.

### Input Validation

* JSON `1 MiB` limit (`respond.go`). PDF timeout 2.5s per path, warns and returns partial.

## 4. Processing Stages

### Stage 1: Extract & Intake

**Responsibility** — Extract body text + chart images; persist intake.

**Input** — PDF bytes or pasted text; `paperFetcher` for DOI/URL.

**Output** — `ExtractResult{Text, Charts []ExtractedChart{PageNumber, ImageBytes}}`; `documents` row with `processing_stages`.

**Failure Conditions** — `EXTRACTION_ERROR` if no text and no images; timeout returns partial with `Warn`.

---

### Stage 2: Simplify + Chapters

**Responsibility** — LLM simplification then `DetectChapters()` → ≤10.

**Input** — Extracted text, level.

**Output** — `simplifiedText`, `[]Chapter{Title, Summary, Excerpt}`.

**Failure Conditions** — Gemini 429/5xx → backoff retry (3) then `slog.Error stage=simplify`. No chapters → `Info "no chapters detected, skipping chart generation"`.

### Stage 3: Verify

**Responsibility** — Dual-claim extraction + diff.

**Input** — Original vs simplified.

**Output** — `claim_diff` + one `claims` row per `OriginalClaims` in same tx (`intake.go:154–172`). `verification_failed` surfaces `mismatch_detail`.

**Failure Conditions** — Verification failure still stores document; badge gated.

### Stage 4: Evidence Extraction

**Responsibility** — Deterministic, no LLM values.

**Input** — `chapter.Excerpt`.

**Output** — `[]NumericEvidence{Metric, Entity, Value, Unit, Source{Page, Text}, ID}` via 5 patterns (`reFromTo` “X from 72.4% to 81.7%”, `reStat` “beta = 0.42”, `reEntityVal` “Model A achieved 72.4%”, `extractEq` “Model A = 72.4%”, `extractStandalone` metric+number) + `ExtractTableData` pipe/tab.

**Failure Conditions** — `EXTRACTION_ERROR`; empty → no datasets → `Info "no evidence extracted"`.

### Stage 5: Dataset Build

**Responsibility** — Group by metric+unit.

**Input** — Evidence.

**Output** — `[]CandidateDataset{ID ds_<metric>_<unit>, Title, Metric, Unit, Points []DatasetPoint{Label, Value, EvidenceID}}` — skip empty `Entity`, key `metric||unit`, sort by entity/metric.

**Failure Conditions** — `DATASET_ERROR` if zero groups.

### Stage 6: Chart Plan & Grounding

**Responsibility** — Per-dataset LLM chart choice + deterministic gate.

**Input** — Dataset + chapter `Title/Summary`.

**Output** — `GenerateChapterCharts` per dataset: `perDatasetChartPrompt` → `HasChart?` + `ChartType (bar|line|scatter|pie)` → `Chart{SourceMethod=data_extracted, ChartData{labels,values,title}, Provenance{EvidenceIDs, DatasetID, verified}}`. Validated by 10 rules (type supported, structure, dimensions match, units consistent, evidence traceable, pie non-negative, scatter numeric, NaN/Inf).

**Failure Conditions** — `CHART_SELECTION_ERROR`, `GROUNDING_ERROR` (unsupported → not rendered), `SCHEMA_ERROR`, logged via `logChartFailure`.

## 5. Evidence & Provenance

### Evidence Model

```json
{"metric":"accuracy","entity":"Model A","value":72.4,"unit":"%","source":{"page":1,"text":"Model A achieved 72.4% accuracy"}}
```

### Provenance Rules

* Every `DatasetPoint.EvidenceID` must exist in evidence index; `ValidateGrounding` returns `missing evidence for X` or `evidence ID not found`.
* `ChartProvenance` persisted; frontend shows grounding badge.

### Unsupported / Unverified Data

* `unsupported` → not rendered; `partial` if untraceable categories.

## 6. Validation

### Validation Rules (10 deterministic)

1. Type in `bar,line,scatter,pie` 2. Required fields per type 3. Valid floats (no NaN/Inf) 4. Dimensions match 5. Units consistent 6. Categories traceable 7. Evidence ID exists 8. Pie non-negative 9. Pie parts-of-whole 10. Scatter numeric X/Y.

### Validation Failures

* Any rule → `unsupported` + `Errors[]`; untraceable → `partial`.

### Validation Strategy

* **Deterministic** (`grounding.go`), no LLM override. Frontend `data-chart.jsx` badge, `chart-card.jsx` unsupported state.

## 7. Transformation

* `Evidence → Datasets` (deterministic grouping) → `ChartSpec` (LLM picks dataset/type/title/takeaway, values from `Labels()/Values()` only).
* Image fallback: `tryExtractChartData` → `data_extracted`, else `image_fallback` + `annotateImage`, else `omitted`.

## 8. Output

### Output Format

```json
{"chart_type":"bar","title":"...","labels":["A","B"],"values":[72.4,81.7],"provenance":{"evidence_ids":["..."],"grounding_status":"verified"}}
```

Stored `charts.chart_data JSON`, `source_method ∈ {data_extracted,image_fallback,omitted}`, `image_blob BLOB` (≤5 images/doc, not served yet).

### Output Guarantees

* Values subset of extracted evidence (no synthetic numbers).

## 9. Error Handling

| Stage | Failure | Detection | Recovery |
|-------|---------|-----------|----------|
| Extract | timeout | context + Warn | degraded partial |
| Simplify/Verify | 429/5xx | slog.Error | 500, retry by user |
| Evidence | none | len==0 | no charts |
| Grounding | unsupported | 10 rules | drop chart |
| Schema | marshal fail | logChartFailure | skip dataset |

## 10. Performance Considerations

### Bottlenecks

* Gemini calls dominate (simplify + verify 2 + per-dataset plan 3–8s/chart). `maxImageChartsPerDocument=5` bounds cost.

### Current Measurements

| Operation | Latency | Notes |
|-----------|---------|-------|
| PDF extract | 0.5–2.5s | per page timeout |
| Simplify | 2–5s | direct HTTP |
| Verify | 2–4s | 2 calls |
| GenerateChapterCharts | 3–8s/chart | per dataset |

No persistent metrics (see `OBSERVABILITY.md`).

## 11. Known Limitations

* `image_blob` not served (annotation only).
* `WAL` required after DB reset (single migration, gitignored).
* `B3` single-call verification skipped.

## 12. Future Changes

* SSE/polling for long documents.
* Periodic `DeleteExpired` sweeper (now only startup).
* Serve `image_blob` via endpoint.
