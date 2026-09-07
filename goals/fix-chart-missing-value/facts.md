# Facts — Fix Chart Missing Value Silent Zero-Fill

1. Missing values (undefined/null) in chartData.values must be excluded from rows array, not defaulted to 0
2. Real zero values (0) must be preserved and rendered normally
3. Explanatory comment must be added above the fix explaining why missing values are dropped
4. The fix is isolated to frontend/src/components/data-chart.jsx — no backend changes
5. npm run build must succeed after the change
6. Chart type logic (bar/line/pie/scatter branches) must not be modified
7. Recharts dynamic import logic must not be modified
