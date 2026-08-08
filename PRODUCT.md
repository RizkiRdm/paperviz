# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Undergraduate students in three primary situations:
1. **Reading**: skimming or studying a dense academic paper before class or assignment
2. **Writing**: extracting and understanding charts/data from source papers for a literature review
3. **Presenting**: preparing simplified explanations and re-rendered visuals for class presentations

All three share one job: make sense of complex academic papers quickly, without drowning in jargon or squinting at tiny chart images.

## Product Purpose

PaperViz converts academic papers (PDF upload or pasted text) into simplified-language versions with re-visualized charts. Students get a plain-language summary tuned to their reading level (Simplified or ELI5) plus server-side re-rendered charts with natural-language annotations. The goal: make papers accessible without losing substance. Success means a student can understand a paper's core ideas and data in minutes instead of hours.

## Positioning

Paper-aware simplification — not generic AI summarization. A competitor could ask ChatGPT to summarize text, but PaperViz does two things generic tools don't: (1) extracts and re-visualizes charts from PDFs with plain-language annotations, and (2) tunes simplification to academic reading levels (Simplified vs ELI5) rather than generic output. The ephemeral share link is a delivery mechanism, not the product itself.

## Operating Context

- Student uploads a PDF or pastes text via a web interface (no login required)
- Backend extracts text + chart images from PDF (in-memory, no disk writes)
- Gemini API simplifies content to chosen reading level; charts are re-rendered server-side with overlaid annotations
- Output served at an ephemeral share link (expires after 7 days of inactivity)
- Single Go binary + SQLite — no Docker required, container-friendly
- Solo-developer project with ~1hr/day human oversight

## Capabilities and Constraints

**Confirmed functionality:**
- PDF text extraction (text-layer only, no OCR)
- Chart image extraction from PDFs
- Gemini-powered simplification (Simplified / ELI5 reading levels)
- Server-side chart re-visualization with annotation overlay
- 7-day ephemeral share links (no authentication)
- Chapter-based chart extraction (one Gemini call per chapter)
- Claim-diff verification (detects simplification drift)
- Background expiry sweep (hourly, deletes expired documents)

**Technical constraints:**
- Go 1.24+ backend (chi router, modernc.org/sqlite — pure Go, no CGO)
- React 19 + Vite 8 + Tailwind CSS v4 frontend
- Google Gemini API (direct HTTP client, no SDK)
- SQLite database (ephemeral data, 7-day expiry)
- No job queue, message broker, or microservice split
- No Docker/orchestration requirement for MVP

**Terminology:**
- "Reading level" = Simplified or ELI5
- "Chart source method" = data_extracted, image_fallback, or omitted
- "Document status" = processing, complete, failed, verification_failed

**Undecided:**
- Deployment target (VPS / Railway / Fly.io / split deploy — not chosen yet)
- OCR support for image-only PDFs (currently text-layer only)
- User accounts (currently no-auth; sessions table exists in schema but not wired)

## Brand Commitments

- **Name**: PaperViz
- **Voice**: Direct, student-friendly, no condescension
- **License**: MIT
- **Design system**: Dub-style (defined in DESIGN.md) — light theme, monochrome palette with one electric blue accent, Satoshi display headlines, Inter body text, Geist Mono for code
- **Visual identity**: Quiet editorial SaaS aesthetic, hairline borders, dense monochrome typography, compact components

## Evidence on Hand

- `DESIGN.md`: Full token system (colors, typography, spacing, shadows, components)
- `README.md`: Product description, tech stack, deployment options
- `AGENTS.md`: Project rules, known issues, DB reset protocol, security hardening
- `frontend/src/`: React components, pages (result-page, login-page, signup-page)
- `internal/`: Go backend (handlers, services, repository, external)
- `migrations/`: SQLite schema (001_init, 002_users, 003_chapters)
- `graphify-out/`: Knowledge graph of codebase architecture

**Absences that future work must not fabricate:**
- No PLAN.md (referenced in README but file missing)
- No test fixtures or sample PDFs in repository

## Product Principles

1. **Paper-aware, not generic** — simplification and chart handling must understand academic paper structure, not just text
2. **Ephemeral by default** — no accounts, no data hoarding, links expire; privacy is a feature, not a liability
3. **Student-first speed** — the tool must be faster than reading the paper, or it fails its job
4. **Chart fidelity** — re-visualized charts must preserve data truth while adding plain-language annotation
5. **Agent-autonomous development** — solo-dev with ~1hr/day oversight; phases execute autonomously, cross-phase changes require human sign-off

## Accessibility & Inclusion

No specific accessibility requirements established. Future work should consider:
- Screen reader compatibility for simplified text output
- Color contrast compliance (current design uses near-black text on white — high contrast)
- Mobile responsiveness (students access on phones)
