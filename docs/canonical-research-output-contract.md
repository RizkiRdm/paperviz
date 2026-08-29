# Canonical Research Output Contract — PaperViz

> **Version:** 0.1.0 (pre-release)
> **Last updated:** 2026-08-29
> **Purpose:** Stable JSON schemas consumed by Web UI, REST API, MCP, and future SDKs

## Design Principles

1. **One source of truth:** Business logic lives in core layer, not in adapters
2. **Stable identifiers:** Every entity has a unique, persistent ID
3. **Provenance always present:** Claims and explanations reference their source
4. **Uncertainty expressed:** Confidence levels where applicable
5. **Presentation separate from data:** Display text ≠ canonical model

## Core Entities

### Document
The root entity representing an analyzed research document.

```json
{
  "id": "string — unique identifier (doc_abc123)",
  "title": "string — paper title (extracted or user-provided)",
  "status": "processing | complete | failed | verification_failed",
  "source_type": "pdf | pasted_text",
  "reading_level": "simplified | eli5",
  "created_at": "integer — Unix timestamp (seconds)",
  "last_accessed_at": "integer — Unix timestamp (seconds)",
  "original_text": "string — full original text (private, not exposed in public API)",
  "simplified_text": "string | null — markdown-formatted simplified version",
  "error_message": "string | null — error details if status is failed",
  "chart_extraction_degraded": "boolean — true if image extraction had issues",
  "processing_stage": "string | null — current pipeline stage during processing",
  "processing_time_ms": "integer | null — total processing time in milliseconds",
  "user_id": "string | null — owner (null for anonymous)",
  "saved": "boolean — user has bookmarked this paper",
  "visibility": "public | unlisted | private",
  "share_token": "string | null — lazy-generated share link token",
  "charts": ["Chart[] — re-visualized figures"],
  "chapters": ["Chapter[] — detected sections"],
  "claim_diff": "ClaimDiff | null — claim verification data",
  "evidence": ["Evidence[] — source evidence references"]
}
```

**Stability notes:**
- `id` is permanent once created
- `status` transitions: `processing → complete | failed | verification_failed`
- `simplified_text` format is markdown, but structure may evolve
- `charts`, `chapters`, `claim_diff`, `evidence` are separate entities linked by document_id

---

### Chart
A re-visualized figure or table from the paper.

```json
{
  "id": "string — unique identifier (chart_abc123)",
  "document_id": "string — parent paper ID",
  "source_method": "data_extracted | image_fallback | omitted",
  "chart_type": "bar | line | pie | scatter | area | radar",
  "chart_data": {
    "labels": ["string[]"],
    "values": ["number[]"],
    "title": "string — chart title"
  } | null,
  "annotation": "string | null — plain-language explanation",
  "source_text": "string | null — original text backing this chart",
  "page_number": "integer | null — source page (0 if not applicable)",
  "display_order": "integer — ordering within document",
  "chapter_id": "string | null — linked chapter",
  "share_token": "string | null — lazy-generated share link token",
  "image_url": "string | null — URL to chart image (if image_blob exists)"
}
```

**Provenance fields:**
- `source_method`: How the chart was created (data extraction vs image capture)
- `source_text`: Original text that generated this chart
- `page_number`: Where in the original document
- `chapter_id`: Which section this chart belongs to

**Uncertainty:**
- `chart_data` is null when `source_method` is `image_fallback` (image-only)
- `annotation` may be null if explanation generation failed
- `image_url` is null when `source_method` is `data_extracted` (no image blob)

---

### Chapter
A detected section of the paper, used for per-chapter chart generation.

```json
{
  "id": "string — unique identifier (chapter_abc123)",
  "document_id": "string — parent paper ID",
  "title": "string — section heading",
  "summary": "string — section summary",
  "excerpt": "string — first paragraph or key text",
  "content": "string — full chapter text (from simplified text)",
  "display_order": "integer — ordering within document"
}
```

---

### ClaimDiff
Claim verification data comparing original vs simplified claims.

```json
{
  "original_claims": ["string[] — claims from original text"],
  "simplified_claims": ["string[] — claims from simplified text"],
  "mismatch_detected": "boolean — true if original and simplified claims differ",
  "mismatch_detail": "string | null — details of claim mismatch"
}
```

**Provenance:**
- `original_claims` and `simplified_claims` show the claim diff
- `mismatch_detected` flags potential simplification errors

**Uncertainty:**
- Mismatches indicate potential simplification errors, not necessarily incorrect output

---

### Evidence
A piece of evidence linking claims to source material.

```json
{
  "id": "string — unique identifier (evidence_abc123)",
  "paper_id": "string — parent paper ID",
  "page": "integer | null — source page",
  "figure_id": "string | null — linked figure",
  "table_id": "string | null — linked table",
  "section": "string | null — section name",
  "source_text": "string — verbatim source text",
  "source_reference": "string — formatted citation (e.g., 'Page 3, Figure 2')"
}
```

**Provenance:**
- `source_text` is always verbatim from the original
- `source_reference` provides human-readable location
- `figure_id` / `table_id` link to specific visual elements

---

### Table
An extracted table from the paper.

> **Note:** Table extraction is not yet implemented in PaperViz. This schema is defined for future implementation.

```json
{
  "id": "string — unique identifier (table_abc123)",
  "document_id": "string — parent paper ID",
  "page_number": "integer | null — source page",
  "caption": "string | null — table caption",
  "headers": ["string[] — column headers"],
  "rows": [
    ["string[] — cell values"]
  ],
  "source_text": "string | null — original text representation"
}
```

---

### Comparison
A structured comparison across multiple papers.

```json
{
  "papers": ["PaperSummary[] — individual paper summaries"],
  "dimensions": ["ComparisonDimension[] — side-by-side comparison"],
  "agreement": ["string[] — areas where papers agree"],
  "disagreement": ["string[] — areas where papers disagree"],
  "evidence_claims": ["EvidenceClaim[] — cross-paper claims"]
}
```

---

### PaperSummary
Structured fields extracted from a single paper for comparison.

```json
{
  "document_id": "string — paper ID",
  "title": "string — paper title",
  "research_question": "string — primary research question",
  "methodology": "string — research methodology",
  "dataset": "string — dataset description",
  "sample_size": "string — sample size",
  "findings": ["string[] — key findings"],
  "limitations": ["string[] — study limitations"],
  "figures": ["string[] — figure descriptions"],
  "evidence": ["string[] — supporting evidence"],
  "conclusions": "string — author conclusions"
}
```

---

### ComparisonDimension
A single dimension of comparison across papers.

```json
{
  "dimension": "string — dimension name (e.g., 'methodology')",
  "values": {
    "doc_abc123": "string — value for this paper"
  },
  "notes": "string | null — synthesis or observation"
}
```

---

### EvidenceClaim
A cross-paper claim with per-paper stance.

```json
{
  "claim": "string — the claim statement",
  "stances": {
    "doc_abc123": "supporting | contradicting | unclear"
  },
  "source_refs": {
    "doc_abc123": "string — evidence reference text"
  }
}
```

---

### DocumentListItem
Lightweight row for paper-history list. Carries preview + counts instead of full text.

```json
{
  "id": "string — document ID",
  "title": "string — paper title",
  "created_at": "integer — Unix timestamp (seconds)",
  "status": "processing | complete | failed | verification_failed",
  "summary_preview": "string — brief preview of simplified text",
  "chart_count": "integer — number of charts",
  "explanation_count": "integer — number of explanations"
}
```

---

### Collection
A user-created grouping of papers.

```json
{
  "id": "string — unique identifier (col_abc123)",
  "user_id": "string — owner ID",
  "name": "string — collection name",
  "created_at": "integer — Unix timestamp (seconds)"
}
```

---

### CollectionListItem
Lightweight row for collections list.

```json
{
  "id": "string — collection ID",
  "name": "string — collection name",
  "created_at": "integer — Unix timestamp (seconds)",
  "document_count": "integer — number of documents in collection"
}
```

---

### ExtractedChart
A chart/table candidate found during PDF extraction, before re-visualization.

```json
{
  "page_number": "integer — source page number (1-indexed)",
  "image_bytes": "bytes | null — raw chart image bytes if extracted"
}
```

> **Note:** This is an internal pipeline type, not exposed in API responses.

---

### PipelineOutput
Full result of one pipeline run, ready for persistence.

```json
{
  "status": "complete | failed | verification_failed",
  "simplified_text": "string — markdown-formatted simplified version",
  "error_message": "string — error details if failed",
  "verify": {
    "original_claims": ["string[]"],
    "simplified_claims": ["string[]"],
    "mismatch_detected": "boolean",
    "mismatch_detail": "string"
  },
  "charts": ["Chart[] — processed charts"],
  "chart_extraction_degraded": "boolean",
  "chapters": ["Chapter[] — detected sections"]
}
```

> **Note:** This is an internal pipeline type, not exposed in API responses.

---

## Identifier Conventions

| Entity | Prefix | Example |
|--------|--------|---------|
| Document | `doc_` | `doc_abc123` |
| Chart | `chart_` | `chart_abc123` |
| Chapter | `chapter_` | `chapter_abc123` |
| Claim | `claim_` | `claim_abc123` |
| Evidence | `evidence_` | `evidence_abc123` |
| Table | `table_` | `table_abc123` |
| Collection | `col_` | `col_abc123` |
| Share Token | (no prefix) | `abc123xyz` |

**Rules:**
- IDs are generated server-side using `repository.NewID()`
- IDs are immutable once created
- Share tokens are lazy-generated (not created until needed)

---

## Status Values

### Document Status
- `processing` — Pipeline is running
- `complete` — Analysis finished successfully
- `failed` — Pipeline encountered an error
- `verification_failed` — Claim verification detected issues

### Processing Stages
- `intake` — Text extraction
- `simplify` — LLM simplification
- `verify` — Claim verification
- `chapters` — Chapter detection
- `charts` — Chart generation

---

## Provenance Model

Every research output traces back to its source:

```
ClaimDiff
  ├── original_claims: ["verbatim claims from original"]
  ├── simplified_claims: ["verbatim claims from simplified"]
  └── mismatch_detected → mismatch_detail

Chart
  ├── source_method: "data_extracted" | "image_fallback"
  ├── source_text: "original text backing this chart"
  ├── page_number: 1
  └── chapter_id: "chapter_abc123"

Evidence
  ├── source_text: "verbatim source text"
  ├── source_reference: "Page 3, Figure 2"
  ├── figure_id: "chart_abc123" (optional)
  └── table_id: "table_abc123" (optional)
```

**Rule:** No output is presented without provenance. If provenance cannot be determined, the output must include `confidence: null` or a note explaining the gap.

---

## Uncertainty Model

| Field | Values | Meaning |
|-------|--------|---------|
| `confidence` | high, medium, low, null | Extraction reliability |
| `mismatch_detected` | boolean | Original vs simplified claim differs |
| `chart_extraction_degraded` | boolean | Image extraction had issues |
| `source_method` | data_extracted, image_fallback, omitted | How chart was created |

**Rule:** Never hide uncertainty. If the system is unsure, express it explicitly.

---

## Evolution Policy

This contract is pre-release (v0.x). Breaking changes may occur, but will be documented:

- **Minor version (0.x.1):** New optional fields, backward-compatible
- **Major version (0.x.0 → 1.0.0):** Stable, no breaking changes

**Deprecation notice:** Deprecated fields will be marked with `deprecated: true` and removed in the next major version.

---

## Usage by Interface

| Interface | Uses |
|-----------|------|
| Web UI | Full entity set, presentation layer |
| REST API | Document, Chart, ClaimDiff, Chapter, Evidence (via /api/documents) |
| MCP (future) | PaperSummary, Comparison, EvidenceClaim |
| SDK (future) | All entities |

**Rule:** Business logic lives in core layer. Web UI, API, MCP, and SDK are thin adapters over the same canonical data.
