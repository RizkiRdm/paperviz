# Shareable Figure Explanation

## Goal
Build public share pages for individual figures with unique URLs, visibility controls, and clear distinction between original figure and PaperViz explanation.

## Facts
See `facts.md` — 13 accepted facts covering URL structure, share page content, visibility inheritance, token generation, revocation, and frontend components.

## Plan
See `plan.md` — 7 steps: migration, repository, service, handler, frontend share page, chart card share button, visibility integration.

## Done Condition
- [ ] `/share/fig/{token}` serves a public page with original image + explanation + attribution + CTA
- [ ] Share token generated lazily on first share, revocable by owner
- [ ] Document visibility (public/unlisted/private) controls share page access
- [ ] Per-chart share button in result page ChartCard
- [ ] All existing tests pass, new tests for share functionality
- [ ] Frontend builds clean, `gofmt` clean
