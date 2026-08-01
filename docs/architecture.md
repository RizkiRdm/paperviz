# Architecture Memory

Canonical specification: [`ARCHITECTURE.md`](./ARCHITECTURE.md). Do not duplicate its full contracts here.

## Current Addendum

- Frontend chart rendering is component-split: `ChartCard` lazy-loads `DataChart`, which dynamically imports Recharts only when a `data_extracted` chart mounts.
- Backend architecture remains unchanged: single Go binary, React static frontend, SQLite, and strict `handlers → services → repository/external` flow.
