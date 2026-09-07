# Audit Task Index — 2026-09-04

## Purpose

This index exists so an agent starting a fresh session picks EXACTLY ONE
task file below and does ONLY that scope. Do not read multiple task
files in one session unless explicitly told to. Do not start work not
listed in one of these files.

## Execution Order

1. `TASK-1-auth-rate-limit.md` — P0, no dependency, do this first.
2. `TASK-2-chart-missing-value-fix.md` — P0, no dependency, independent
   of TASK-1, can be done in any order relative to it.
3. `TASK-3-docs-consolidation.md` — P2, no dependency, docs-only, zero
   risk to running code, safe to do anytime.

## Explicitly Out Of Scope For All Three Tasks Above

The chart engine rework (evidence extraction, grounding validation,
multimodal image support) is a SEPARATE, ALREADY-PLANNED workstream
that lives at `docs/paperviz-chart-agent-task-chunks.md` (chunks C1
through C17). Do NOT start that work from this index. Do NOT create a
new plan for it. If asked to work on charts, open that file directly
and start at chunk C2 — chunk C1 (audit) is already satisfied by the
2026-09-04 review referenced in this index.

## Rule For Every Task File In This Index

Each task file is self-contained: it has its own Context, mandatory
investigation steps, exact scope, out-of-scope list, validation
commands, and acceptance criteria. Read only the ONE task file assigned
for the current session. Do not scan the rest of the repository beyond
what that file's investigation steps tell you to grep/read.
