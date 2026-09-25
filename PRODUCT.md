# Product

<!-- impeccable:product-schema 1 -->

## Product Statement

PaperViz converts academic papers into simplified explanations, verified claims, structured research objects, and evidence-grounded figures. It is being shaped into an agent-first research tool, while the web application remains the complete manual-ingestion surface and the home for authentication, API keys, and billing.

## Users

### Primary direction: AI agents acting for people

Agents need paper data they can inspect and cite:

- metadata and sections;
- claims linked to source evidence;
- figures with provenance and grounding state;
- deterministic retrieval without hidden model reasoning.

### Human web user

People use the web application to:

- upload a text-layer PDF or paste text;
- import a paper through DOI or URL;
- choose Simplified or ELI5 reading level;
- inspect processing, verification, and figure status;
- annotate, organize, export, or share research context;
- authenticate, manage an API key, and manage billing.

### End user

The person directing an agent or using the web app should understand a paper's core claims and data faster while retaining enough source evidence to check generated interpretations.

## Product Problem

Generic AI summarization improves readability but does not guarantee research integrity. It can:

- omit or distort claims;
- generate plausible numbers that were not present in the source;
- hide uncertainty behind fluent prose;
- make charts difficult to trace back to paper evidence.

PaperViz's product opportunity is not generic summarization. It is paper-aware, inspectable research transformation.

## Product Surfaces

### Web and REST

The web path supports four input modes:

- PDF;
- pasted text;
- DOI;
- URL.

Web ingestion runs the complete asynchronous pipeline: simplification, claim verification, chapter detection, evidence extraction, dataset construction, chart planning, grounding validation, and persistence.

### MCP

The intended agent surface is MCP. The current repository ships five stdio tools:

- `ingest_document`;
- `search_documents`;
- `get_document`;
- `get_figures`;
- `get_evidence`.

MCP is additive to REST. It shares data and domain semantics but does not duplicate the web UI or contain hidden LLM reasoning.

## Core Capabilities

### Paper understanding

- text-layer PDF extraction in memory;
- pasted-text ingestion;
- DOI and guarded URL import;
- Simplified and ELI5 output;
- chapter detection;
- claim-diff verification with visible mismatch detail.

### Evidence-grounded figures

- deterministic numeric and table evidence extraction;
- candidate dataset construction;
- Gemini-assisted chart selection and explanation;
- deterministic grounding validation;
- chart provenance and frontend grounding status;
- bounded image-chart fallback.

### Research context

- claims and evidence;
- tables, methods, results, and citations;
- claim-evidence links and paper relationships;
- per-user annotations and collections;
- structured research export;
- ephemeral document and figure sharing.

### Human and agent support

- email/password and Google authentication;
- session cookies and API keys;
- Stripe checkout, portal, and webhook handling;
- usage and account summaries;
- local stdio MCP server.

## Trust and Integrity Principles

1. **AI may transform evidence, not manufacture it.** Numeric chart values come from extracted datasets.
2. **Verification failure is evidence.** A detected mismatch must remain visible.
3. **Unsupported figures do not render as trustworthy charts.** Deterministic grounding controls this outcome.
4. **Source location matters.** Claims, evidence, and figures retain page, section, figure, table, or source references where available.
5. **Shared and exported data must respect copyright and ownership.** Public payloads and exports must not leak full source text or another user's private data.

## Current Implementation Reality

The agent-first direction is not complete:

- `ingest_document` stores pasted text but does not start the full LLM pipeline;
- MCP-created documents can remain in `processing` state;
- `/agents` generates remote configuration for `/api/mcp`, but the current HTTP router does not register that transport;
- MCP search is global and stateless, not scoped to an authenticated user's library;
- current deployment architecture is one Go process plus one SQLite database, not a horizontally scaled system.

These gaps must be documented honestly and verified in code before being described as shipped.

## Technical Constraints

- Go 1.25 backend with `chi` and raw `database/sql`;
- SQLite through `modernc.org/sqlite`, WAL enabled, no CGO;
- React 19, Vite 8, Tailwind CSS 4, shadcn/ui, and Recharts;
- direct Gemini HTTP integration;
- no ORM;
- no job queue or message broker;
- no microservice split;
- no uploaded PDF bytes persisted to disk;
- MCP remains additive and stateless unless a deliberate architecture decision changes that rule.

## Non-Goals

- OCR for scanned-image PDFs;
- generic chatbot experience;
- invented chart data;
- silently publishing verification failures as clean results;
- speculative MCP tools without demonstrated demand;
- duplicate canonical product routes;
- exposing destructive or user-owned operations through MCP by convenience.

## Product Success

PaperViz succeeds when:

1. a person or agent can move from source paper to structured research context with low friction;
2. generated explanations preserve inspectable source relationships;
3. chart values remain traceable to extracted evidence;
4. failures and uncertainty are visible rather than hidden;
5. the system remains simple enough for a solo maintainer to operate.

## Brand Commitments

- **Name:** PaperViz
- **Voice:** direct, student-friendly, and technically precise;
- **Design:** quiet editorial interface, monochrome foundation, one electric-blue accent;
- **UI source of truth:** `DESIGN.md`;
- **Product direction:** agent-first without hiding current implementation gaps.

## Current Product Sources

| Concern | Source |
|---|---|
| Product direction | `PRODUCT.md` |
| Product requirements and status | `docs/PRD.md` |
| Canonical user flow | `docs/product/current-user-flow.md` |
| Architecture | `docs/ARCHITECTURE.md` |
| Processing behavior | `docs/DATA_PIPELINE.md` |
| Current engineering state | `docs/PROJECT_STATE.md` |
| UI system | `DESIGN.md` |

For current behavior, source code remains authoritative when any document lags.
