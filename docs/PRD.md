# PaperViz Product Requirements

> Current product requirements, source-verified on 2026-09-25.
>
> Status legend: **Implemented** means behavior exists in current source; **Partial** means useful code exists but end-to-end behavior is incomplete; **Gap** means required behavior is absent or mismatched; **Target** means desired outcome is not implemented.
>
> The lowercase `architecture.md` file is only a compatibility pointer. Architecture contracts live in [`ARCHITECTURE.md`](ARCHITECTURE.md).

---

## Product Summary

PaperViz transforms academic papers into simplified explanations, verified claims, structured research objects, and evidence-grounded figures.

The product is moving toward an agent-first model in which AI clients use PaperViz as a structured research-data tool. The web application remains necessary for authentication, API-key and billing management, and manual paper ingestion.

Current implementation supports both surfaces, but only web ingestion runs the complete processing pipeline. MCP provides deterministic text intake and shared-data retrieval; remote agent setup is not end-to-end complete.

## Product Goals

1. Reduce time needed to understand a paper's core claims and numeric findings.
2. Keep generated conclusions connected to source evidence.
3. Provide one data model usable through web and agent interfaces.
4. Preserve a simple, operable architecture for a solo maintainer.
5. Make uncertainty and failure visible.

## Non-Goals

- OCR for scanned-image PDFs.
- Generic chatbot interface.
- Model-generated chart values.
- ORM, job queue, message broker, or microservice architecture.
- Uploaded PDF persistence to disk.
- Speculative MCP tools.
- Duplicate canonical product routes.
- User-owned destructive or preference operations through MCP without explicit product approval.

---

## Target Users

### AI agent acting for a person

Agents need deterministic tools and structured research data:

- document metadata and sections;
- claims linked to source evidence;
- figures with provenance and grounding state;
- stable errors, rate limits, and bounded input.

### Person managing access and billing

People need email/password and Google authentication, an API key, subscription management, and usage visibility.

### Person analyzing a paper manually

People need PDF, text, DOI, or URL input; a chosen reading level; visible processing and failure states; claims, evidence, figures, and source material; and annotation, collection, export, and sharing controls.

## Problem Statement

Academic papers impose a comprehension tax through dense language, specialized terminology, and complex figures. Generic AI summaries can make prose easier to read but do not guarantee that claims remain faithful, numeric values came from the source, or uncertainty remains visible.

PaperViz treats research output as structured, inspectable data:

1. **Grounding over fluency.** Generated claims and figures must retain source relationships.
2. **Deterministic evidence.** Numeric chart values must come from extracted evidence, not model invention.
3. **Visible uncertainty.** Verification and grounding failures must remain visible.
4. **One shared data model.** REST and MCP must expose consistent research semantics.
5. **Simple operation.** A solo maintainer must be able to run and evolve the system without unnecessary infrastructure.

## Status Summary

| Capability | Status | Current behavior |
|---|---|---|
| PDF and pasted-text web ingestion | Implemented | Validates input, stores document, starts full pipeline |
| DOI and URL web ingestion | Implemented | Fetches source and starts full pipeline |
| Simplified and ELI5 output | Implemented | Gemini-backed transformation |
| Claim-diff verification | Implemented | Persists mismatch state and detail |
| Evidence-grounded charts | Implemented | Deterministic extraction, dataset construction, and grounding |
| Structured research objects | Implemented | Claims, evidence, tables, methods, results, citations, relationships |
| Email/password authentication | Implemented | Signup, login, session, logout |
| Google OAuth | Implemented in source | Provider configuration required for live use |
| API-key management | Implemented | Retrieve and regenerate user key |
| Stripe billing | Implemented in source | Checkout, portal, webhook; provider configuration required |
| Account and usage surfaces | Implemented | `/account`, account summary, usage |
| Annotations and collections | Implemented | Per-user ownership enforced |
| Research export | Implemented | Structured context; excludes full source and simplified text |
| Ephemeral sharing | Implemented | Document and figure share tokens |
| Local MCP server | Implemented | Five stdio tools sharing SQLite |
| Full MCP processing | Gap | `ingest_document` does not start the LLM pipeline |
| Remote MCP transport | Gap | `/agents` config targets `/api/mcp`, absent from current router |
| User-scoped MCP search | Gap | Search is global and stateless |
| Agent-scale concurrency | Target | Current one-connection SQLite and process-local controls are limiting |

## Functional Requirements

### Input

1. The system must accept one text-layer PDF or pasted-text document through REST.
2. PDF input must be limited to 20 MiB and validated as PDF content.
3. Image-only PDFs must return a clear `no_text_layer` error.
4. DOI import must accept valid DOI syntax and resolve source material.
5. URL import must use HTTP or HTTPS and block private-network destinations.
6. MCP intake must accept text only and enforce a 500 KiB input limit.

### Processing

1. Web ingestion must return a document ID without waiting for full processing.
2. Processing must expose a user-facing stage while running.
3. The frontend must poll no faster than every two seconds.
4. Processing must have explicit long-running, timeout, and retry states.
5. The pipeline must simplify, verify, detect chapters, build evidence and datasets, plan figures, validate grounding, and persist results.
6. Background processing must have an explicit timeout.
7. One figure failure must not abort unrelated figure processing.

### Explanation and claims

1. Reading levels must be `simplified` and `eli5`.
2. Simplified output must be checked against original claims.
3. Verification mismatch must remain visible in API and UI.
4. A verification failure must not be represented as a clean result.
5. Claim-diff and claims must be persisted atomically with successful pipeline output.

### Figures and grounding

1. Numeric values must originate from deterministic evidence extraction.
2. Candidate datasets must be built from extracted evidence.
3. Gemini may select chart type, title, and explanation but not values.
4. Grounding validation must run deterministically.
5. Unsupported data must not render as verified chart data.
6. Missing values must remain missing rather than become zero.
7. Provenance must identify supporting evidence where available.
8. Embedded-image fallback must be bounded per document.

### Research context

1. The system must store or expose claims, evidence, tables, methods, results, citations, and relationships.
2. Evidence must retain source text and available source references.
3. Annotations must belong to authenticated users.
4. Collections must enforce owner access.
5. Export must exclude original and simplified full text.
6. Shared public payloads must not expose private source text or user identifiers.

### Authentication and billing

1. Email/password authentication must use secure session cookies.
2. Passwords must meet complexity requirements.
3. Google OAuth must validate state to prevent CSRF.
4. Authentication endpoints must be rate limited.
5. Users must be able to retrieve and regenerate API keys.
6. Stripe checkout, portal, and webhook routes must exist.
7. Required integration configuration must fail loudly at startup.

### MCP

1. MCP must run over stdio in the current repository.
2. MCP must share the migrated SQLite data model with REST.
3. MCP must expose exactly five documented tools: `ingest_document`, `search_documents`, `get_document`, `get_figures`, and `get_evidence`.
4. MCP tool calls must not invoke Gemini.
5. MCP must not expose user list, save, rename, delete, share, visibility, or referral operations.
6. MCP must enforce process API-key authentication, per-key rate limits, and job limits.
7. Remote MCP must not be advertised as working until an HTTP transport route exists and passes end-to-end tests.

## Primary User Flows

### Web analysis

```text
open /
  → choose PDF, paste, DOI, or URL
  → select reading level where applicable
  → receive document ID
  → follow processing
  → inspect understanding, evidence, figures, source, and management sections
```

### Local MCP data access

```text
configure AI client for local stdio binary
  → provide process environment
  → call one of five tools
  → read/write shared SQLite research data within MCP boundaries
```

Current limitation: MCP-only intake does not advance through the full pipeline.

### Intended remote agent flow

```text
sign in
  → obtain API key
  → open /agents
  → connect AI client
  → ingest and retrieve grounded research data
```

This flow remains partial until remote MCP transport and MCP processing behavior are implemented and tested.

## Quality Requirements

### Correctness

- `go test ./...` must pass.
- Service changes require success and error tests.
- Pipeline changes require acceptance and failure regression coverage.
- Claim verification must catch the corrupted-passage regression.
- Chart validation must reject unsupported values.
- Ownership tests must cover annotations and collections.

### Security

- No secrets in source control.
- No full document text in logs.
- No uploaded PDF bytes on disk.
- No SSRF through URL import.
- No cross-user annotation or collection access.
- No unverified OAuth callback.
- No unscoped destructive MCP operation.

### Reliability

- Explicit errors for expected failures.
- No panic for validation, network, or API errors.
- Bounded request and processing timeouts.
- Degraded figure output must not corrupt the whole document.
- Verification and grounding failure states must be preserved.

### Frontend

- Every `catch` must provide user-facing feedback.
- Retry must preserve input.
- Development errors must reach `console.error`.
- No silent zero-fill for missing chart data.
- No stale or duplicate canonical route.
- UI must use `DESIGN.md` tokens.

## Success Metrics

These are targets, not current verified measurements:

1. A typical supported paper completes end to end in under 60 seconds.
2. Claim verification runs for every completed web pipeline.
3. No rendered chart contains values absent from extracted evidence.
4. A new maintainer can complete local setup using documentation.
5. An agent can complete setup and first successful tool call without manual code changes.
6. Agent tool latency remains acceptable under measured concurrent load.
7. Verification, grounding, and processing failures are understandable without reading server logs.

## Current Product Gaps

### P0: complete agent path

- Implement a real remote MCP transport or change `/agents` to match the supported local transport.
- Define whether MCP ingestion starts full processing, triggers a separate processing operation, or remains intake-only.
- Add end-to-end tests from agent configuration to grounded output.

### P1: ownership and scale

- Decide whether MCP search should be user-scoped.
- Benchmark and define supported concurrency.
- Complete handler-to-application/repository decoupling.
- Remove or clearly quarantine stale static SEO surfaces.

### P2: product hardening

- Measure real processing latency and agent-tool latency.
- Validate OAuth, Stripe, and remote MCP with live provider configuration.
- Decide whether `/account` should retain document-management UI or remain account-only.

## Acceptance Scenarios

### Acceptance 1: text-layer PDF

Given a valid PDF under 20 MiB with selectable text, when a person submits it, then the system returns a processing document ID, runs the full pipeline, and presents terminal result state with explanation, claims, figures where supported, and source.

### Acceptance 2: pasted text

Given at least 50 characters of pasted text, when a person submits it, then the system creates a simplified document without requiring PDF image extraction.

### Acceptance 3: unsupported scan

Given an image-only PDF, when a person submits it, then the system returns a clear no-text-layer error and does not start processing.

### Acceptance 4: grounded chart

Given extractable numeric evidence, when chart planning succeeds, then every rendered value traces to a candidate dataset and evidence provenance.

### Acceptance 5: unsupported chart

Given missing, incompatible, or untraceable chart data, when grounding runs, then the figure is marked unsupported or omitted rather than rendered as verified data.

### Acceptance 6: claim mismatch

Given a simplified claim that conflicts with the source, when verification completes, then the document exposes verification failure and mismatch detail.

### Acceptance 7: user-owned context

Given one authenticated user owns an annotation or collection, when another user attempts access, then the system returns forbidden or not found without leaking private content.

### Acceptance 8: MCP boundary

When MCP is called, then only documented deterministic tools run, no Gemini call occurs, and user-owned management operations remain unavailable.

## Source Documents

- Product direction: [`../PRODUCT.md`](../PRODUCT.md)
- Product flow: [`product/current-user-flow.md`](product/current-user-flow.md)
- Architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md)
- Pipeline: [`DATA_PIPELINE.md`](DATA_PIPELINE.md)
- MCP parity: [`mcp-parity.md`](mcp-parity.md)
- API contract: [`openapi.yaml`](openapi.yaml)
- Current engineering state: [`PROJECT_STATE.md`](PROJECT_STATE.md)
