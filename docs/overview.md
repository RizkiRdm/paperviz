# PaperViz Overview

This document explains what PaperViz is, why it exists, how its two product surfaces fit together, and which guarantees the system is designed to provide. It is not a setup guide; use [Getting Started](getting-started.md) to run the application.

## Problem

Academic papers are difficult to use quickly because they combine dense language, specialized terminology, claims, tables, and figures. Generic language-model summaries can make prose easier to read, but they do not guarantee that:

- important claims remain faithful to the source;
- numeric values came from the paper rather than the model;
- chart data can be traced back to source evidence;
- uncertainty or verification failures remain visible.

PaperViz addresses this by treating research output as structured, inspectable data rather than as an unqualified generated summary.

## Intended Product Model

PaperViz is moving toward an agent-first product:

1. A person signs in, manages billing, and obtains an API key.
2. An AI agent uses PaperViz tools to ingest and retrieve research data.
3. The web application remains available as a manual escape hatch for people who want to upload or paste a paper directly.

The intended agent experience is not fully realized in the current repository. The code currently ships a local stdio MCP server with five deterministic tools, while the `/agents` page generates configuration for a remote HTTP MCP endpoint that is not registered in the HTTP router. MCP ingestion also does not start the full LLM pipeline. These gaps are documented in [Current implementation reality](#current-implementation-reality).

## Users and Outcomes

### AI agent acting for a person

An agent needs structured paper data it can cite and inspect:

- document metadata;
- extracted sections;
- claims and supporting evidence;
- figures with source locations and provenance;
- grounded chart values.

### Person using the web application

A person may need to:

- upload a PDF or paste text;
- import a paper through DOI or URL;
- choose Simplified or ELI5 output;
- inspect verification and figure-grounding status;
- annotate, organize, export, or share research context;
- manage authentication, API keys, and billing.

### Student or researcher

The intended end outcome is faster understanding: a person should be able to identify a paper's main claims, inspect their evidence, and understand important figures without treating generated text as infallible.

## Two Product Surfaces

### Web and REST

The React application calls the Go HTTP API. Web and DOI/URL ingestion start the complete asynchronous processing pipeline:

```text
intake → simplify → verify claims → build evidence → plan figures → persist → poll result
```

The HTTP server also serves the built React application, authentication, account management, billing, collections, annotations, sharing, exports, usage, and structured research endpoints.

### MCP

The MCP entrypoint runs as a separate process over stdio and shares the SQLite database with the web application. It currently registers:

| Tool | Behavior |
|---|---|
| `ingest_document` | Validates and stores pasted text, then returns a document ID |
| `search_documents` | Performs stateless global title search |
| `get_document` | Returns metadata and optional sections, evidence, or figures |
| `get_figures` | Returns chart data, source locations, and provenance |
| `get_evidence` | Returns evidence and linked claims |

MCP is designed as an additive interface over shared data and domain logic. It is not a second business-logic implementation and does not contain hidden reasoning loops.

## Web Processing Pipeline

### 1. Intake

The web path accepts one of:

- a text-layer PDF up to 20 MiB;
- pasted text;
- a DOI resolved through external providers;
- an HTTPS URL resolved through guarded external fetching.

The service validates input, extracts PDF text, creates a document ID, and inserts a `processing` row. Uploaded PDF bytes remain in memory and are not written to disk.

### 2. Simplification

Gemini rewrites source text at the selected reading level: `simplified` or `eli5`. This generated explanation is useful content, but it is not treated as automatically correct.

### 3. Claim verification

PaperViz extracts original and simplified claims, compares them, and stores a claim-diff record. If a meaningful mismatch is detected, the document is marked `verification_failed` and the mismatch detail is surfaced instead of being hidden behind a success state.

### 4. Chapter detection

Simplified text is divided into a bounded set of chapters. Chapters give the result structure and provide context for figure planning.

### 5. Evidence extraction

Deterministic parsers extract numeric evidence and table data from chapter excerpts. Each value can retain source text and location metadata.

### 6. Dataset construction

Extracted evidence is grouped by metric and unit into candidate datasets. This step is deterministic; the language model does not create the numeric values.

### 7. Figure planning and grounding

Gemini may choose whether a dataset merits a chart, which supported chart type fits, and how to title or explain it. Values come from the candidate dataset. A deterministic grounding validator then checks structure, dimensions, units, traceability, and chart-specific constraints. Unsupported figures are not rendered as trustworthy charts.

### 8. Persistence

The final document state is written with its claim diff, chapters, claims, charts, and provenance. The frontend polls the document while processing and then presents the result sections.

## Trust Model

PaperViz uses two related but distinct controls.

### Claim verification

Claim verification asks: **Does the simplified explanation preserve the source's claims?**

- Source and simplified claims are compared.
- Mismatch status and detail are persisted.
- Failed verification remains visible.

### Figure grounding

Figure grounding asks: **Do chart values and structure come from extracted evidence?**

- Numeric evidence is extracted deterministically.
- Candidate datasets are built deterministically.
- Gemini selects presentation, not values.
- Invalid or untraceable data is rejected.
- Frontend rendering exposes grounding and provenance state.

This separation prevents a fluent explanation or chart title from being mistaken for evidence.

## Data Model

PaperViz stores research context as linked records rather than one opaque response:

- documents and processing state;
- users, sessions, API keys, and billing fields;
- chapters and sections;
- charts, chart data, provenance, and limited image blobs;
- evidence, claims, and claim-evidence links;
- tables, methods, results, and citations;
- annotations and research collections;
- sharing, referrals, and usage records.

Exact schemas live in [`../migrations/`](../migrations/) and repository code. The machine-readable research object contract is documented in [`canonical-research-output-contract.md`](canonical-research-output-contract.md).

## Data Lifecycle

- PDF uploads are processed in memory; raw uploaded PDF bytes are not persisted.
- Extracted source text and derived research objects are stored in SQLite.
- At most five embedded-image charts per document are sent through the image fallback path.
- Document reads refresh the access timestamp used by expiry.
- Expired sessions are removed at server startup; expired documents are swept at startup and hourly.
- Sharing uses expiring public tokens and excludes source text from public document payloads.
- Export includes structured research context and user-owned material, not original or simplified full text.

SQLite runs in WAL mode with foreign keys enabled and one open connection. This favors correctness and simplicity for a single low-concurrency instance, not horizontal scale.

## Architecture Style

PaperViz is a modular monolith:

```text
React UI
   ↓
REST handlers
   ↓
application/domain services
   ↓
repositories and external providers
```

MCP is a peer transport adapter:

```text
MCP tools
   ↓
shared services and repositories
```

The design deliberately avoids:

- ORM abstractions;
- background job infrastructure or message brokers;
- microservice decomposition;
- LLM gateways;
- hidden reasoning inside MCP;
- user-owned destructive operations exposed through MCP.

See [`ARCHITECTURE.md`](ARCHITECTURE.md) for contracts and non-goals.

## Current Implementation Reality

The following distinctions matter when reading plans or older documents:

| Area | Current behavior |
|---|---|
| Web ingestion | Starts full asynchronous processing pipeline |
| DOI/URL ingestion | Fetches source and starts full asynchronous processing pipeline |
| MCP ingestion | Stores pasted text only; does not start processing pipeline |
| MCP transport | Local stdio binary in this repository |
| `/agents` generated config | Points to remote `https://paperviz.com/api/mcp`; no matching route exists in current HTTP router |
| MCP user scope | Stateless and global; not tied to REST session ownership |
| Figure grounding | Deterministic validation before chart persistence/rendering |
| Raw PDF persistence | Prohibited and not implemented |
| Horizontal scaling | Not supported by current single-instance SQLite design |

These are current-state observations, not aspirational architecture.

## Key Vocabulary

| Term | Meaning |
|---|---|
| Document | One ingested paper and its processing state |
| Chapter | A bounded section of simplified output |
| Claim | A statement extracted from the original paper |
| Evidence | Source text or structured context supporting research claims and figures |
| Candidate dataset | Deterministically grouped numeric evidence with shared metric and unit |
| Chart provenance | Links from a chart to dataset and evidence identifiers |
| Grounding | Deterministic validation that a figure is supported by traceable data |
| Verification | Comparison between original and simplified claims |
| Source method | How a chart was produced: extracted data, image fallback, or omitted |
| MCP | Tool protocol used by AI clients to call PaperViz data operations |

## Where to Go Next

- Run the application: [Getting Started](getting-started.md)
- Change the system safely: [Maintainer Guide](maintainer-guide.md)
- Look up routes, files, and settings: [Codebase Reference](codebase-reference.md)
- Study pipeline internals: [Data Pipeline](DATA_PIPELINE.md)
- Review current engineering status: [Project State](PROJECT_STATE.md)
