# Chart Engine Rework — Facts

## Scope
- C1 (audit) already done — start at C2
- C2-C17 in order per dependency graph
- Parallel where safe: C3+C4, C9+C10, C11+C12

## Evidence Model (C2)
- NumericEvidence struct: id, metric, entity, value, unit, source_page, source_text
- Optional: context, group, experiment, time, confidence
- Every item has source provenance
- Generated numbers are NOT numeric evidence

## Text Extraction (C3)
- Extract numbers from body text, results, comparisons, stats
- Preserve original meaning — ambiguous → ambiguous
- Do not assume baseline=0 for "12% improvement"

## Table Extraction (C4)
- Reuse existing table extraction
- Preserve units, headers, page provenance
- Missing cells stay missing, not zero

## Candidate Dataset (C5)
- Dataset: id, title, dimensions, unit, points[label, value, evidence_id]
- Values come from evidence only
- One chapter → 0..N datasets

## Chart Generation (C6)
- LLM decides: useful? which dataset? type? title? axis? takeaway?
- LLM must NOT generate numeric values
- No meaningful dataset → no chart

## Chart Spec (C7)
- Type-specific schemas: bar(categories, series), line(x, series), scatter(x, y), pie
- Scatter = numeric X + numeric Y
- Pie only for parts-of-whole

## Grounding Validator (C8)
- 10 validation rules
- Status: verified/partial/unsupported
- unsupported → DO NOT RENDER
- Deterministic, no LLM override

## Missing Values (C9)
- Never null→0
- Missing required value → validation failure

## Multiple Charts (C10)
- Chapter → evidence → datasets → 0..N charts
- No arbitrary one-chart limit

## Image Fallback (C11)
- Keep existing image extraction
- Image numbers must be grounded
- Unsupported → omit chart

## Provenance (C12)
- paper_id, page, evidence_ids, source_method, dataset_id, grounding_status
- Survives API serialization

## Frontend Safety (C13)
- Render validated spec only
- No inferring/repairing/inventing
- Show grounding status

## Failure Categories (C14)
- EXTRACTION_ERROR, DATASET_ERROR, CHART_SELECTION_ERROR, GROUNDING_ERROR, SCHEMA_ERROR, RENDER_ERROR
- Log: paper_id, page, stage, category, message, fallback_used

## Regression Tests (C15)
- 8 test cases: text comparison, table, ambiguous, missing, scatter, pie, unsupported, image fallback

## Real Paper Validation (C16)
- Test corpus: numeric text, tables, image-only charts, stats, ambiguous, multi-dataset, incomplete
- Metric: grounded/total charts

## Cleanup (C17)
- Remove old prompts, unused code
- Keep image fallback
- Keep backward compat
