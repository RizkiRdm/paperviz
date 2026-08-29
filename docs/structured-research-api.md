# Structured Research API — PaperViz

> Base URL: `https://paperviz.com` (or your deployment URL)
> Version: v0 (pre-release, subject to change)

## Authentication

PaperViz supports two authentication modes:

### Session-Based Auth (Authenticated Endpoints)

- **Signup/Login:** Creates a `session_token` httpOnly cookie (7-day expiry)
- **No API keys** — cookie-based authentication only
- **Protected endpoints** require valid session cookie

### Fingerprint-Based Auth (Unauthenticated Endpoints)

- **Fingerprint** = SHA256(IP + User-Agent + Accept-Language)
- Calculated server-side from request headers
- Rate limits and usage tracking applied per fingerprint
- **For API access:** Send standard HTTP headers (User-Agent, Accept-Language)

## Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `POST /api/documents` | 1 request | 30 seconds |
| All other endpoints | No limit | — |

## Usage Limits

| Tier | Monthly Paper Limit |
|------|---------------------|
| Free | 5 papers |
| Pro | Higher limit |
| Research | Highest limit |

Limits reset on the 1st of each month.

---

## Endpoints

### Authentication

#### POST /api/auth/signup
Create a new user account.

**Request:**
```json
{
  "email": "string (required) — valid email address",
  "password": "string (required) — minimum 8 characters"
}
```

**Response (201):**
```json
{
  "id": "usr_abc123",
  "email": "user@example.com"
}
```

**Status codes:**
- 201: Account created, session cookie set
- 400: Invalid request (invalid_email, password_too_short)
- 409: Email already taken

---

#### POST /api/auth/login
Authenticate an existing user.

**Request:**
```json
{
  "email": "string (required)",
  "password": "string (required)"
}
```

**Response (200):**
```json
{
  "id": "usr_abc123",
  "email": "user@example.com"
}
```

**Status codes:**
- 200: Login successful, session cookie set
- 400: Invalid request
- 401: Invalid credentials

---

#### POST /api/auth/logout
End current session.

**Response:** 200 OK (empty body, session cookie cleared)

---

#### GET /api/auth/me
Get current authenticated user.

**Response (200):**
```json
{
  "id": "usr_abc123",
  "email": "user@example.com"
}
```

**Status codes:**
- 200: User info
- 401: Not authenticated

---

### Documents

#### POST /api/documents
Create a new document for analysis. Accepts multipart form data.

**Request:** `multipart/form-data`
- `file` — PDF file (max 20 MiB) **OR**
- `text` — pasted text content (exactly one required)
- `reading_level` — `"simplified"` (default) or `"eli5"`

**Response (201):**
```json
{
  "document_id": "doc_abc123",
  "status": "processing"
}
```

**Status codes:**
- 201: Document created, processing started
- 400: Invalid input (file_too_large, missing_input, invalid_reading_level, invalid_file_type)
- 429: Rate limit exceeded (wait 30 seconds)
- 429: Monthly limit reached (upgrade required)

---

#### GET /api/documents/{id}
Get full document with all analysis results. Touches `last_accessed_at` to extend 7-day expiry.

**Response (200):**
```json
{
  "id": "doc_abc123",
  "title": "Paper Title",
  "status": "completed",
  "reading_level": "simplified",
  "original_text": "Full original text...",
  "simplified_text": "## Key Findings\n\nSimplified content...",
  "error_message": null,
  "chart_extraction_degraded": false,
  "processing_stage": null,
  "charts": [
    {
      "id": "chart_001",
      "source_method": "data_extracted",
      "chart_data": { "labels": [...], "datasets": [...] },
      "annotation": "This chart shows...",
      "page_number": 1,
      "chapter_id": "ch_001",
      "image_url": "/api/documents/doc_abc123/charts/chart_001/image"
    }
  ],
  "claim_diff": {
    "original_claims": ["claim 1", "claim 2"],
    "simplified_claims": ["simplified claim 1", "simplified claim 2"],
    "mismatch_detected": false,
    "mismatch_detail": null
  },
  "chapters": [
    {
      "id": "ch_001",
      "title": "Introduction",
      "summary": "Chapter summary...",
      "content": "Chapter content...",
      "display_order": 0
    }
  ],
  "evidence": [
    {
      "id": "ev_001",
      "page": 3,
      "figure_id": null,
      "table_id": null,
      "section": "Methods",
      "source_text": "Supporting text from original...",
      "source_reference": "Section 2.1, Page 3"
    }
  ]
}
```

**Status codes:**
- 200: Success
- 404: Document not found

---

#### GET /api/documents/{id}/charts/{chartId}/image
Get chart as binary image.

**Response:** Binary image data (Content-Type detected: image/png, image/jpeg, etc.)

**Status codes:**
- 200: Image data
- 404: Chart not found or no image

---

#### GET /api/documents
List all documents for authenticated user. **Requires auth.**

**Query Parameters:**
- `limit` — max results (default 20, max 100)
- `offset` — pagination offset (default 0)

**Response (200):**
```json
{
  "documents": [
    {
      "id": "doc_abc123",
      "title": "Paper Title",
      "status": "completed",
      "created_at": 1724800000,
      "summary_preview": "Brief preview...",
      "chart_count": 3,
      "explanation_count": 5
    }
  ],
  "total": 10,
  "limit": 20,
  "offset": 0
}
```

**Status codes:**
- 200: Success
- 401: Not authenticated

---

#### GET /api/documents/stats
Get document statistics for authenticated user. **Requires auth.**

**Response (200):**
```json
{
  "total": 10,
  "saved": 3,
  "collections": 2
}
```

---

#### PUT /api/documents/{id}/save
Toggle saved status. **Requires auth.**

**Request:**
```json
{
  "saved": true
}
```

**Response (200):**
```json
{
  "saved": true
}
```

---

#### PATCH /api/documents/{id}
Update document title. **Requires auth.**

**Request:**
```json
{
  "title": "New Title"
}
```

**Response (200):**
```json
{
  "title": "New Title"
}
```

---

#### DELETE /api/documents/{id}
Delete a document. **Requires auth.**

**Response (200):**
```json
{
  "status": "deleted"
}
```

---

### Comparison

#### POST /api/documents/compare
Compare multiple papers (2-10 documents).

**Request:**
```json
{
  "document_ids": ["doc_abc123", "doc_def456"]
}
```

**Response (200):**
```json
{
  "papers": [
    {
      "document_id": "doc_abc123",
      "title": "Paper A",
      "research_question": "What is the effect of X on Y?",
      "methodology": "RCT with 200 participants",
      "dataset": "Stanford NLP Corpus",
      "sample_size": "200",
      "findings": ["Finding 1", "Finding 2"],
      "limitations": ["Limitation 1"],
      "figures": ["Figure 1 description"],
      "evidence": ["Evidence text 1"],
      "conclusions": "X has significant effect on Y"
    }
  ],
  "dimensions": [
    {
      "dimension": "methodology",
      "values": {
        "doc_abc123": "Paper A used RCT",
        "doc_def456": "Paper B used survey"
      },
      "notes": "Papers use different methodologies..."
    }
  ],
  "agreement": ["Papers share common research themes: learning, education"],
  "disagreement": ["Papers use different research methodologies"],
  "evidence_claims": [
    {
      "claim": "Method X improves outcome Y by 15%",
      "stances": {
        "doc_abc123": "supporting",
        "doc_def456": "contradicting"
      },
      "source_refs": {
        "doc_abc123": "Section 3.2: We found 15% improvement...",
        "doc_def456": "Table 4: No significant improvement observed..."
      }
    }
  ]
}
```

**Status codes:**
- 200: Comparison complete
- 400: Invalid document count (must be 2-10)
- 404: One or more documents not found
- 500: Extraction or comparison failed

---

### Sharing (Chart-Level)

#### POST /api/documents/{id}/charts/{chartId}/share
Generate share link for a specific chart. **Requires auth.**

**Response (200):**
```json
{
  "share_token": "abc123xyz",
  "share_url": "/share/fig/abc123xyz"
}
```

**Status codes:**
- 200: Token generated
- 401: Not authenticated
- 403: Not document owner
- 404: Document or chart not found

---

#### DELETE /api/documents/{id}/charts/{chartId}/share
Revoke chart share link. **Requires auth.**

**Response:** 204 No Content

---

### Sharing (Document-Level)

#### POST /api/documents/{id}/share
Generate share link for entire document. **Requires auth.**

**Response (200):**
```json
{
  "share_url": "/share/doc/abc123xyz"
}
```

---

#### DELETE /api/documents/{id}/share
Revoke document share link. **Requires auth.**

**Response:** 204 No Content

---

#### PATCH /api/documents/{id}/visibility
Update document visibility. **Requires auth.**

**Request:**
```json
{
  "visibility": "unlisted"
}
```

**Valid values:** `"unlisted"`, `"public"`

**Response (200):**
```json
{
  "visibility": "unlisted"
}
```

---

### Public Share Access

#### GET /share/fig/{shareToken}
Access a shared chart figure (no auth required).

**Response:** Chart data JSON (same shape as chart in document response)

---

#### GET /share/doc/{shareToken}
Access a shared paper (no auth required).

**Response:** Document JSON (same shape as GET /api/documents/{id})

---

#### POST /share-referrals
Track share-to-analysis conversion.

**Request:**
```json
{
  "ref": "share_token_value"
}
```

**Response:** 204 No Content

---

### Usage

#### GET /api/usage
Get current usage stats for this fingerprint.

**Response (200):**
```json
{
  "tier": "free",
  "papers_used": 3,
  "limit": 5,
  "reset_date": "2026-09-01"
}
```

---

### Collections

#### POST /api/collections
Create a collection. **Requires auth.**

**Request:**
```json
{
  "name": "My Research"
}
```

**Response (201):**
```json
{
  "id": "col_abc123",
  "name": "My Research",
  "created_at": 1724800000,
  "document_count": 0
}
```

---

#### GET /api/collections
List all collections. **Requires auth.**

**Response (200):**
```json
{
  "collections": [
    {
      "id": "col_abc123",
      "name": "My Research",
      "created_at": 1724800000,
      "document_count": 5
    }
  ]
}
```

---

#### GET /api/collections/{id}
Get collection with its documents. **Requires auth.**

**Response (200):**
```json
{
  "id": "col_abc123",
  "name": "My Research",
  "created_at": 1724800000,
  "documents": [
    {
      "id": "doc_abc123",
      "title": "Paper Title",
      "status": "completed",
      "created_at": 1724800000,
      "summary_preview": "Brief preview...",
      "chart_count": 3,
      "explanation_count": 5
    }
  ]
}
```

---

#### PATCH /api/collections/{id}
Rename a collection. **Requires auth.**

**Request:**
```json
{
  "name": "New Name"
}
```

**Response (200):**
```json
{
  "name": "New Name"
}
```

---

#### DELETE /api/collections/{id}
Delete a collection. **Requires auth.**

**Response (200):**
```json
{
  "status": "deleted"
}
```

---

#### POST /api/collections/{id}/documents
Add document to collection. **Requires auth.**

**Request:**
```json
{
  "document_id": "doc_abc123"
}
```

**Response (200):**
```json
{
  "status": "added"
}
```

---

#### DELETE /api/collections/{id}/documents/{docId}
Remove document from collection. **Requires auth.**

**Response (200):**
```json
{
  "status": "removed"
}
```

---

### Analytics (Internal)

#### GET /analytics
Get aggregate analytics summary. **Requires auth.**

---

#### POST /api/analytics/pricing-view
Track pricing page view event. **No auth required.**

**Response:** 204 No Content

---

#### POST /api/analytics/upgrade-intent
Track upgrade CTA click. **No auth required.**

**Response:** 204 No Content

---

## Error Format

All errors follow consistent format:

```json
{
  "error": "error_code"
}
```

**Common error codes:**

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `not_found` | 404 | Resource doesn't exist |
| `invalid_json` | 400 | Request body is malformed JSON |
| `rate_limit_exceeded` | 429 | Too many requests (wait and retry) |
| `unauthenticated` | 401 | Authentication required |
| `unauthorized` | 401 | Authentication required |
| `forbidden` | 403 | Insufficient permissions |
| `internal_error` | 500 | Server error |
| `email_taken` | 409 | Email already registered |
| `invalid_email` | 400 | Malformed email address |
| `password_too_short` | 400 | Password < 8 characters |
| `file_too_large` | 400 | Upload exceeds 20 MiB |
| `missing_input` | 400 | Neither file nor text provided |
| `invalid_file_type` | 400 | Uploaded file is not a valid PDF |
| `invalid_reading_level` | 400 | Must be "simplified" or "eli5" |
| `no_text_layer` | 422 | PDF has no extractable text |
| `name_required` | 400 | Collection name is empty |
| `document_id_required` | 400 | Document ID is empty |
| `title_required` | 400 | Title is empty |
| `monthly limit reached` | 429 | Paper limit reached for tier |
| `document_ids_must_be_2_to_10` | 400 | Comparison requires 2-10 docs |
| `document_not_found` | 404 | Document for comparison not found |
| `extraction_failed` | 500 | Paper summary extraction failed |
| `comparison_failed` | 500 | Paper comparison failed |
| `invalid_visibility` | 400 | Visibility must be unlisted/public |
| `invalid_ref` | 400 | Referral token is empty |

---

## Data Models

### Document
| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique document identifier |
| title | string | Paper title (extracted or user-provided) |
| status | string | `processing`, `completed`, or `error` |
| reading_level | string | `simplified` or `eli5` |
| original_text | string | Full original text |
| simplified_text | string? | Simplified version with markdown (null if processing) |
| error_message | string? | Error details if status is "error" |
| chart_extraction_degraded | bool | Whether chart extraction had issues |
| processing_stage | string? | Current processing step |
| charts | array | Re-visualized charts |
| claim_diff | object? | Claim verification data |
| chapters | array | Content chapters |
| evidence | array | Source evidence references |

### Chart
| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique chart identifier |
| source_method | string | `data_extracted` or `image_capture` |
| chart_data | object? | Recharts-compatible data (null for image-only) |
| annotation | string? | Plain-language explanation |
| page_number | int? | Source page (null if not applicable) |
| chapter_id | string? | Associated chapter ID |
| image_url | string? | URL to chart image (if image_blob exists) |

### ClaimDiff
| Field | Type | Description |
|-------|------|-------------|
| original_claims | array | Claims from original text |
| simplified_claims | array | Claims from simplified text |
| mismatch_detected | bool | Whether verification found discrepancies |
| mismatch_detail | string? | Description of mismatch if detected |

### Chapter
| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique chapter identifier |
| title | string | Chapter heading |
| summary | string | Brief summary |
| content | string | Full chapter text |
| display_order | int | Sort order |

### Evidence
| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique evidence identifier |
| page | int? | Source page number |
| figure_id | string? | Referenced figure |
| table_id | string? | Referenced table |
| section | string? | Source section name |
| source_text | string | Original text excerpt |
| source_reference | string | Human-readable reference |

### PaperSummary (Comparison)
| Field | Type | Description |
|-------|------|-------------|
| document_id | string | Source document ID |
| title | string | Paper title |
| research_question | string | Primary research question |
| methodology | string | Research methodology |
| dataset | string | Dataset description |
| sample_size | string | Sample size |
| findings | array | Key findings |
| limitations | array | Study limitations |
| figures | array | Referenced figures |
| evidence | array | Supporting evidence |
| conclusions | string | Main conclusions |

### ComparisonDimension
| Field | Type | Description |
|-------|------|-------------|
| dimension | string | Comparison dimension name |
| values | object | Map of document_id → value |
| notes | string? | AI synthesis notes |

### EvidenceClaim
| Field | Type | Description |
|-------|------|-------------|
| claim | string | Specific factual claim |
| stances | object | Map of document_id → stance (supporting/contradicting/unclear) |
| source_refs | object | Map of document_id → source text |

---

## Examples

### Create account
```bash
curl -X POST https://paperviz.com/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "you@example.com", "password": "securepass123"}'
```

### Upload a paper (with session cookie)
```bash
curl -X POST https://paperviz.com/api/documents \
  -b "session_token=abc123" \
  -F "file=@paper.pdf" \
  -F "reading_level=simplified"
```

### Paste text
```bash
curl -X POST https://paperviz.com/api/documents \
  -b "session_token=abc123" \
  -F "text=Your paper text here..." \
  -F "reading_level=eli5"
```

### Get results
```bash
curl https://paperviz.com/api/documents/doc_abc123
```

### List your documents
```bash
curl https://paperviz.com/api/documents \
  -b "session_token=abc123"
```

### Compare papers
```bash
curl -X POST https://paperviz.com/api/documents/compare \
  -H "Content-Type: application/json" \
  -d '{"document_ids": ["doc_abc123", "doc_def456"]}'
```

### Create collection
```bash
curl -X POST https://paperviz.com/api/collections \
  -b "session_token=abc123" \
  -H "Content-Type: application/json" \
  -d '{"name": "ML Research"}'
```

### Add paper to collection
```bash
curl -X POST https://paperviz.com/api/collections/col_abc123/documents \
  -b "session_token=abc123" \
  -H "Content-Type: application/json" \
  -d '{"document_id": "doc_xyz789"}'
```

### Share a paper
```bash
curl -X POST https://paperviz.com/api/documents/doc_abc123/share \
  -b "session_token=abc123"
```

### Access shared paper
```bash
curl https://paperviz.com/share/doc/abc123xyz
```

---

## Limitations (MVP)

- **No API keys** — session cookie auth only
- **No webhooks** — polling required for async operations
- **No batch operations** — one document at a time
- **No streaming** — synchronous responses only
- **No table extraction** — tables captured as images
- **No OCR** — PDF must have text layer

---

## Changelog

- **v0 (2026-08-29):** Initial documentation of existing endpoints
