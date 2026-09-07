# Plan — Fix Chart Missing Value Silent Zero-Fill

## Problem
`data-chart.jsx` line 24 uses `?? 0` which silently converts missing data points to zero. A missing value and a real zero look identical on the chart, violating the project's rule against fabricating missing numeric values.

## Solution
Filter out missing values from the rows array instead of defaulting to 0.

## Steps

### Step 1: Edit data-chart.jsx
**File:** `frontend/src/components/data-chart.jsx`

**Change:** Replace lines 22-25:
```js
const rows = (chartData.labels || []).map((label, i) => ({
  name: label,
  value: chartData.values?.[i] ?? 0,
}))
```

With:
```js
// Missing values are dropped, not defaulted to 0 — a missing data point
// and a real zero must not look the same on the chart.
const rows = (chartData.labels || [])
  .map((label, i) => ({ name: label, value: chartData.values?.[i] }))
  .filter((row) => row.value !== undefined && row.value !== null)
```

**Verification:** grep for `?? 0` in the file — should return zero matches.

### Step 2: Build verification
**Command:** `cd frontend && npm run build`

**Expected:** Build succeeds with no new warnings referencing data-chart.jsx.

### Step 3: Commit
**Command:** `git add frontend/src/components/data-chart.jsx && git commit -m "fix(chart): exclude missing values instead of silent zero-fill"`

## Risks
- None identified. Single file, isolated change, no backend impact.

## Open Questions
- None.
