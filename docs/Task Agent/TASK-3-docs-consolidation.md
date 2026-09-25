# TASK-3 — Archive Stale Docs, Reconcile Duplicate PRD Files

> Historical task record. The archive files and duplicate PRD targets referenced below were removed during documentation cleanup; do not rerun this task.

## Metadata
- id: TASK-3
- priority: P2
- depends_on: none
- estimated_scope: file moves + small edits, no code change, no test impact
- session_type: single-session, isolated

## Context (read this, do not skip)

`docs/PROJECT_STATE.md` is the current source of truth (last updated
2026-09-04, mentions Phase 12 work). Two other files are stale and
contradict it:

- `docs/progress.md` — last entry 2026-08-19, describes Chunk 3.2 as
  "next" work. This is roughly 15 shipped chunks behind reality.
- `docs/current_task.md` — same staleness, also stops at Chunk 3.2/3.3.

There are also two files that may be duplicates:
- `docs/PRD.md`
- `docs/prd.md`

This task does NOT delete anything. It archives stale files with a
clear pointer to the current source of truth, and resolves the PRD
duplicate question with evidence, not a guess.

## Mandatory Investigation Steps (do these BEFORE moving/editing any file)

```
diff docs/PRD.md docs/prd.md
```

```
head -5 docs/progress.md docs/current_task.md docs/PROJECT_STATE.md
```

```
git log --follow --oneline docs/PRD.md
git log --follow --oneline docs/prd.md
```

Record the `diff` result and the `git log` result — you will need them to
decide which PRD file (if either) is safe to archive. Do NOT modify or
move any file during this step.

## Scope — What To Do

### Step 1: Create archive directory

```
mkdir -p docs/archive
```

### Step 2: Archive `progress.md` and `current_task.md`

Move both files into `docs/archive/`, keeping the same filename:

```
git mv docs/progress.md docs/archive/progress.md
git mv docs/current_task.md docs/archive/current_task.md
```

(If `git mv` fails because the file isn't tracked yet in this session's
state, use plain `mv` instead — do not fail the task over this.)

At the very top of BOTH moved files, insert this exact line as a new
first line, above the existing `# Project Progress` / `# Current Task`
heading:

```
> ARCHIVED 2026-09-04 — superseded by docs/PROJECT_STATE.md. Do not use
> this file to determine current project state.
```

### Step 3: Resolve the PRD duplicate

Use the `diff` output from the investigation step:

- **If `diff` shows no differences (identical files):** archive the one
  with the OLDER last-commit date per `git log --follow`, keep the
  newer one in place at `docs/`. Add the same archived-notice line to
  the top of the archived copy, pointing to the file kept in place
  (e.g. `> ARCHIVED — superseded by docs/PRD.md`).
- **If `diff` shows real differences:** do NOT archive either file.
  Instead, stop and report the difference in your completion report —
  a human needs to decide which content is correct. Do not guess or
  merge content yourself.

### Step 4: Update any references

```
grep -rln "docs/progress.md\|docs/current_task.md" --include="*.md" .
```

If any other `.md` file links to the old paths, update the link to
point at `docs/archive/...` instead. Do NOT edit `docs/PROJECT_STATE.md`
content beyond fixing a broken link path if one exists — do not rewrite
its prose.

## Out of Scope — Do NOT Do These

- Do NOT delete any file. Archive only.
- Do NOT edit the content/body of `docs/PROJECT_STATE.md`.
- Do NOT touch any file outside `docs/` (no code changes in this task).
- Do NOT attempt to merge or rewrite PRD content if the two PRD files
  differ — flag it and stop, per Step 3.
- Do NOT archive `docs/PLAN.md`, `docs/decisions.md`, or any other file
  not explicitly named above — only `progress.md`, `current_task.md`,
  and conditionally one PRD file are in scope.

## Validation

```
ls docs/archive/
```
Expected: contains the moved files.

```
head -3 docs/archive/progress.md
```
Expected: shows the archived-notice line first.

```
grep -rln "docs/progress.md\|docs/current_task.md" --include="*.md" .
```
Expected: no results outside `docs/archive/` itself (i.e. no dangling
references left in active docs).

## Acceptance Criteria

- [ ] `docs/archive/` exists and contains `progress.md` and
      `current_task.md`, each with the archived-notice line at the top.
- [ ] `docs/PROJECT_STATE.md` is unchanged in content.
- [ ] PRD duplicate resolved per the diff-based rule in Step 3, OR
      explicitly flagged as unresolved with the diff output pasted in
      the completion report if the files differ.
- [ ] No other active `.md` file has a dangling link to the old
      `docs/progress.md` / `docs/current_task.md` paths.

## Completion Report Format

```
TASK-3 STATUS: <done | blocked | partial>
diff docs/PRD.md docs/prd.md: <identical | differs — paste diff>
PRD resolution: <archived <file> | left both in place, needs human decision>
Files archived: <list>
Dangling references fixed: <list or none found>
```
