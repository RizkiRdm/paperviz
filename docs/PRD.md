# PRD.md — PaperViz

## Project Summary

**Overview**
PaperViz converts academic research papers (PDF or pasted text) into two outputs: (1) a language-simplified version at a user-selected reading level, and (2) re-visualized charts/graphs that are more legible than the paper's originals. Output is served at a unique, no-auth, shareable link that expires after 7 days.

**Objective**
Reduce the reading barrier of academic papers for undergraduate students, without altering the research findings or claims.

**Value Proposition**
A student pastes or uploads a paper and gets back the same research, in language they can actually parse, with charts that explain themselves instead of requiring caption archaeology.

---

## Target Users

**Primary Persona: Undergraduate Student (S1)**
- Motivation: Needs to read a paper for coursework, thesis literature review, or assignment — under time pressure.
- Frustration: Academic register (passive voice, field-specific jargon, dense sentence structure) slows comprehension. Charts often assume prior familiarity with the field's visual conventions (e.g., unlabeled axes referencing earlier terminology, dense multi-panel figures).
- Technical sophistication: Low-to-moderate. Comfortable with web apps, not with LaTeX, not with statistical notation beyond intro-level coursework.

**Secondary personas (not designed for in MVP, may benefit incidentally):** high school students, self-directed learners. Not validated — do not build for them explicitly.

---

## Problem Statement

Academic papers are written for peer researchers, not novice readers. Two specific failures:

1. **Language barrier**: Field-specific terminology and academic sentence structure impose a comprehension tax that is unrelated to the difficulty of the underlying idea. A student can understand the concept but not the sentence.
2. **Visual barrier**: Papers typically contain the minimum number of static charts needed to support a claim to an expert reviewer — not to teach a newcomer. There is no interpretive layer explaining what a chart shows or why it matters.

**Why it matters**: Both failures push non-expert readers toward abandoning the primary source and relying on secondary summaries (blog posts, AI chat answers) that may distort or omit findings. This is a literacy and academic-integrity concern, not just convenience.

**Existing alternatives and why they fall short**:
- Asking an AI chatbot to "explain this paper" — no persistent artifact, no visual re-rendering, inconsistent depth, not shareable.
- Reading secondary sources (blogs, Wikipedia) — several removes from the primary source, may misrepresent findings.
- Office hours / asking a TA — not scalable, not available on demand.

---

## Success Metrics

- Simplification job (average paper, ~6000 words) completes in **<60s** end-to-end.
- Chart re-visualization produces at least **1 improved chart per paper** for papers containing at least 1 extractable data table or figure, in **≥70%** of test papers (measured against a fixed test set of 20 papers during Phase 2 validation).
- Generated share link is accessible and renders correctly within **<2s** load time.
- Meaning-preservation: automated claim-diff check runs on **100%** of jobs. Jobs with detected claim mismatches MUST NOT be published silently — flagged jobs are either blocked or shown with an explicit warning (decision in ARCHITECTURE.md).
- Claim-diff false-positive rate (flagging a correct simplification as mismatched) kept low enough to not block normal usage — validated manually against the 20-paper test set before release; no papers in that set should be blocked incorrectly.

This is a correctness requirement, not a nice-to-have. A job that fails claim-diff verification is not a shippable result.

---

## Core Capabilities (MVP)

1. **Document ingestion** — accept PDF upload or pasted raw text.
2. **Text extraction** — extract body text and tabular/numeric data from PDF.
3. **Simplification engine** — rewrite extracted text at a user-selected reading level, preserving all factual claims, findings, and figures. Levels: Academic (original), Simplified (general audience), ELI5 (child-level vocabulary, same facts).
4. **Chart re-visualization engine** — for each detected chart/table:
   - Primary path: extract underlying data points from text/table → re-render as a new, annotated chart.
   - Fallback path: if data extraction fails, extract the original chart image from the PDF and overlay a plain-language annotation/caption alongside it.
5. **Meaning-preservation verification** — automated claim-extraction diff: extract factual claims from source text and from simplified text separately (2 LLM calls), compare. Mismatches flag the job as failed verification; result is not published as-is.
6. **Result publishing** — generate a unique, non-guessable share link. No account required to view or create.
7. **Link expiry** — stored result and link expire 7 days after last access.

---

## Non-MVP Features

Explicitly excluded — see ARCHITECTURE.md Non-Goals for enforcement:

- Interactive charts (sliders, "what if I change X" simulations).
- arXiv / URL-based paper fetch.
- Non-PDF document formats (Word, LaTeX source, HTML papers).
- User accounts / authentication / saved history tied to identity.
- Mobile app.
- Real-time collaboration / multi-user editing.
- Multi-language translation (simplification changes register, not language).
- Citation-graph or related-paper discovery features.

---

## User Flows

### Primary Flow: Upload → Simplify → View
1. User lands on homepage, uploads a PDF or pastes text.
2. User selects target reading level (Simplified / ELI5). Academic (original) is always available as a baseline comparison.
3. System processes: extracts text + data → runs simplification → runs chart re-visualization.
4. System returns a unique share link. Result page shows: simplified text (with toggle to compare against original), re-visualized charts inline with annotations.
5. User can copy/share the link. Link is valid for 7 days from last access.

### Edge Case: PDF with no extractable charts
- System completes text simplification normally.
- Chart section of result page shows "No charts detected in this document" — does not block delivery of the text output.

### Edge Case: PDF is scanned/image-only (no text layer)
- System detects no extractable text layer.
- User is shown an explicit error: "This PDF appears to be a scanned image without selectable text. PaperViz requires a text-based PDF." No silent partial failure.

### Failure Scenario: Chart data extraction fails, image fallback also fails
- Chart is omitted from output with an inline note: "Original chart could not be reprocessed — refer to source PDF page N."
- Does not block the rest of the document from processing.

### Failure Scenario: LLM simplification call fails or times out
- Retry once. If still failing, surface a clear error to the user and do not publish a partial/corrupted result link.

---

## High-Level Tech Stack

- **Frontend**: React + Vite + Tailwind CSS + shadcn/ui.
- **Backend**: Go (single binary).
- **Database**: SQLite via `modernc.org/sqlite` (no CGO).
- **LLM Provider**: Google Gemini (AI Studio, direct integration — no internal AI gateway dependency for MVP).
- **PDF processing**: Go-native or CLI-invoked PDF text/image extraction (library selection deferred to ARCHITECTURE.md Context Lock).
- **Charting (frontend render)**: Recharts.

---

## Technical Assumptions

- **Scale**: Solo-dev, low-traffic MVP. No assumption of concurrent high load. Single-instance deployment is sufficient.
- **Storage**: SQLite is sufficient given 7-day retention and expected low volume; no justification exists yet for a networked database.
- **PDF text layer required**: OCR for scanned PDFs is explicitly out of scope — see Non-MVP Features.
- **LLM cost/rate limits**: Gemini AI Studio free tier is assumed sufficient for MVP validation traffic. If usage exceeds free-tier limits, this is a scaling decision to revisit post-validation, not a Day 1 concern.
- **No PII handling**: Papers are assumed to be public/non-sensitive academic documents. No special data-protection engineering required for MVP. This assumption breaks if users upload non-academic or sensitive documents — not handled in MVP.
- **Single-language assumption**: MVP assumes English-language source papers. Non-English input handling is undefined behavior in MVP, not silently supported.
