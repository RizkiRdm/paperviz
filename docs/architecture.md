# Architecture Memory

Canonical specification: [`ARCHITECTURE.md`](./ARCHITECTURE.md). Do not duplicate its full contracts here.

## Current Addendum

- Frontend chart rendering is component-split: `ChartCard` lazy-loads `DataChart`, which dynamically imports Recharts only when a `data_extracted` chart mounts. `DataChart` supports bar/line/pie/scatter via `chartData.type`.
- Processing stage tracking: `documents.processing_stage` column stores current pipeline stage during processing. Pipeline receives `OnStage` callback, calls it at simplifying/verifying/generating_charts transitions. Cleared to NULL on pipeline completion.
- Share dialog: modal overlay in `result-page.jsx` with URL input, copy button, and 7-day expiry note. Triggered from header "Share" button.
- Backend architecture remains unchanged: single Go binary, React static frontend, SQLite, and strict `handlers → services → repository/external` flow.
