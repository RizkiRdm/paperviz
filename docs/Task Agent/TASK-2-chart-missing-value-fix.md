# TASK-2 — Fix Silent Zero-Fill In Chart Rendering

## Metadata
- id: TASK-2
- priority: P0
- depends_on: none
- estimated_scope: 1 file edit, no new dependency
- session_type: single-session, isolated

## Context (read this, do not skip)

File: `frontend/src/components/data-chart.jsx`

Current code (approx line 20-23):

```js
const rows = (chartData.labels || []).map((label, i) => ({
  name: label,
  value: chartData.values?.[i] ?? 0,
}))
```

Problem: if `chartData.values[i]` is `undefined` or `null` (missing data
point), the `?? 0` silently renders it as the number zero on the chart.
This is misleading — a missing value and a real value of 0 look
identical to the reader. This project has an explicit rule elsewhere
(`docs/paperviz-chart-agent-task-chunks.md`, Rule 3) that says fabricating
or filling missing numeric values is not allowed. This same principle
applies here even though this specific file predates that doc.

## Mandatory Investigation Steps (do these BEFORE editing any file)

```
grep -n "chartData.values" frontend/src/components/data-chart.jsx
```

```
sed -n '1,60p' frontend/src/components/data-chart.jsx
```

Confirm:
1. The exact line number of the `rows` mapping.
2. That `rows` is consumed later by `LineChart`/`BarChart`/`PieChart`/
   `ScatterChart` via `data={rows}` — do not assume, verify by reading
   the `renderChart()` function in the same file.

Do NOT modify any code during this step.

## Scope — What To Change

In `frontend/src/components/data-chart.jsx`, change the `rows`
construction so that missing values are EXCLUDED from the array instead
of defaulted to 0.

Target behavior:
- If `chartData.values[i]` is `undefined` or `null`, that `(label, value)`
  pair must NOT appear in `rows` at all.
- If `chartData.values[i]` is a real number (including a real `0`), keep
  it as-is. Do not filter out legitimate zeros — only filter out missing
  entries.

Example of the fix shape (adapt to actual surrounding code, do not
copy-paste blindly without checking variable names match):

```js
const rows = (chartData.labels || [])
  .map((label, i) => ({ name: label, value: chartData.values?.[i] }))
  .filter((row) => row.value !== undefined && row.value !== null)
```

Add a one-line comment above this block explaining why:

```js
// Missing values are dropped, not defaulted to 0 — a missing data point
// and a real zero must not look the same on the chart.
```

## Out of Scope — Do NOT Do These

- Do NOT change chart type logic (`bar`/`line`/`pie`/`scatter` branches in
  `renderChart()`).
- Do NOT change the `recharts` dynamic import logic at the top of the
  file.
- Do NOT touch any backend Go file. This is frontend-only.
- Do NOT add a loading/empty-state UI for "some values missing" unless
  explicitly asked in a future task — just drop the missing points
  silently from the render, no new UI element.

## Validation — Run These Commands And Paste Raw Output As Proof

```
cd frontend && npm run build
```
Expected: build succeeds, no new warnings referencing `data-chart.jsx`.

If an oxlint config is present:
```
cd frontend && npx oxlint src/components/data-chart.jsx
```
Expected: no new errors on this file.

## Manual Test (describe result even if you cannot run a browser)

Construct a test case mentally or via a temporary console log:
input `chartData = { chart_type: "bar", labels: ["A","B","C"], values: [10, null, 30] }`
must produce `rows` with exactly 2 entries (A=10, C=30), not 3.

If you have a way to verify this at runtime (e.g. a temporary
`console.log(rows)` during local dev, removed before finishing), do so
and paste the output. If not possible in this session, state that
explicitly.

## Acceptance Criteria

- [ ] `chartData.values?.[i] ?? 0` no longer appears in the file.
- [ ] Missing values (`undefined`/`null`) are excluded from `rows`
      entirely, not converted to `0`.
- [ ] Real zero values (`0`) are still preserved and rendered normally.
- [ ] Explanatory comment added above the fix.
- [ ] `npm run build` succeeds — paste raw output.

## Completion Report Format

```
TASK-2 STATUS: <done | blocked | partial>
Files changed: frontend/src/components/data-chart.jsx
Build result: <paste raw `npm run build` tail>
Manual/logic verification: <describe or paste>
```
