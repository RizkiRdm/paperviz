# Goal — Fix Chart Missing Value Silent Zero-Fill

Fix silent zero-fill in `frontend/src/components/data-chart.jsx` where missing data points (`undefined`/`null`) are rendered as zero instead of being excluded from the chart.

## Done Condition
- Missing values excluded from rows array
- Real zeros preserved
- Comment explaining the fix added
- `npm run build` succeeds
- Changes committed and pushed

## References
- Facts: [facts.md](facts.md)
- Plan: [plan.md](plan.md)
