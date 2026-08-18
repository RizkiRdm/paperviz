# Chunk Y — Fix result page text rendering (frontend, React)

## Scope
THREE concerns in `frontend/src/pages/result-page.jsx` and its immediate
dependencies. Each is independent — do them in order, commit-worthy after
each, and re-verify the previous one still holds before moving to the next.
Do NOT touch: chapter-detection backend, chart generation backend, chart
type/color logic in `data-chart.jsx`, auth pages, dashboard page.

Do NOT redesign anything not listed below. If you notice other visual
issues while in this file, report them at the end — do not fix them in this
chunk.

## Mandatory investigation (before writing any code)

1. `grep -n "className" frontend/src/pages/result-page.jsx | grep -i "prose\|whitespace"`
   — find the exact article/content container className string. Paste it.
2. `cat frontend/package.json` — confirm `@tailwindcss/typography` and
   `react-markdown` are NOT currently in `dependencies`. If either already
   exists, skip the install step for that one and report it.
3. `grep -n "displayedText\|chapterContent\|showOriginal" frontend/src/pages/result-page.jsx`
   — read enough surrounding context (view the file, don't just grep) to
   understand exactly what value gets rendered in the toggle-to-Original
   state vs the default state, so your fix touches the right variable.

Report your findings from steps 1-3 before proceeding to fixes.

## Fix 1 — Long/unbroken text overflowing its container

**Symptom:** toggling to "Original" text pushes content outside its
container horizontally.

**Root cause:** the content container has no CSS rule forcing long unbroken
tokens (PDF-extraction artifacts: merged column text, long DOIs, broken
hyphenation) to wrap.

**Fix:** add a Tailwind utility that forces word-breaking on the container
that renders `displayedText` / `chapterContent`. Use `break-words` (Tailwind
v4 utility for `overflow-wrap: break-word`). If `break-words` alone doesn't
fully solve it for pathological single-token cases, use the arbitrary-value
form `[overflow-wrap:anywhere]` instead — test both, keep whichever actually
fixes it, don't add both redundantly.

**Verification:** after the fix, manually construct a test string with a
150+ character unbroken token (e.g. a long fake URL) and confirm via
`npm run build` + visual check (or a quick temporary console log of computed
styles) that it wraps instead of overflowing. Remove any temporary test code
before finishing.

## Fix 2 — Raw markdown syntax showing as literal text

**Symptom:** when Gemini's output contains markdown-style formatting
(`**bold**`, `## Header`, `- list item`), it renders as literal characters
instead of formatted text.

**Root cause:** no markdown parser is used; content is rendered as a raw
string inside a `whitespace-pre-wrap` div. The `prose` Tailwind class present
in the className is currently inert — `@tailwindcss/typography` (the plugin
that gives `prose` its styling) is not installed.

**Fix:**
1. `npm install react-markdown --save` (check current major version resolves
   cleanly against React 19 in this project — if npm reports a peer dep
   conflict, report it and stop rather than force-installing).
2. `npm install -D @tailwindcss/typography` and register the plugin in
   `frontend/src/index.css` (or wherever Tailwind's `@plugin`/config
   directives live in this v4 project — check how Tailwind v4 plugins are
   registered in this codebase's existing CSS entry file first, don't assume
   the v3-style `tailwind.config.js` pattern applies).
3. Replace the raw string render (`{displayedText}` / `{chapterContent}`)
   with `<ReactMarkdown>{...}</ReactMarkdown>` wrapped in a container that
   keeps the `prose prose-neutral max-w-none` classes (now functional) plus
   the `break-words` fix from Fix 1.
4. Do NOT change what text is passed in — only change how it's rendered.
   `original_text` and `simplified_text` come from the backend as-is.

**Verification:** confirm a chapter/document whose content includes markdown
syntax (check a few existing documents via the dashboard, or manually insert
`**test**` into a local dev document's simplified_text via sqlite3 for a
quick check) now renders as actual bold/headers, not literal asterisks/hashes.

## Fix 3 — Chapter navigation: tabs → sidebar

**Current state:** chapter navigation is a horizontal tab strip
(`role="tablist"`) at the top of the content area.

**Target:** vertical sidebar list on desktop (left of content), collapsing
to the existing tab/dropdown pattern on mobile breakpoints (check existing
Tailwind breakpoint usage in this file for the convention already in use —
`sm:`/`md:` prefixes, match the project's existing pattern rather than
inventing a new one).

**Requirements — do not regress accessibility that already exists:**
- Keep `role="tablist"` / `role="tab"` / `aria-selected` semantics — a
  sidebar list of tabs is still a tablist, just oriented differently. Add
  `aria-orientation="vertical"` to the tablist container.
- Keep existing arrow-key navigation behavior. Vertical layout means
  Up/Down should move focus between chapters (confirm current implementation
  uses Left/Right — if so, you'll need to either support both arrow pairs or
  switch to Up/Down; check `aria-orientation="vertical"` conventions for
  which arrow keys are expected in that case before deciding).
- Chapter titles that are long must truncate with `title` attribute (native
  tooltip) showing the full text on hover — do not let long titles break the
  sidebar's fixed width.

**Verification:** manually tab through chapters via keyboard only (no mouse)
before/after and confirm equivalent navigability. Note any accessibility
regression in your report even if you weren't able to fully fix it.

## Proof of completion (required, not optional)

```bash
npm run build
npm run lint   # or oxlint directly if that's the configured script — check package.json scripts first
```
Paste raw output. Then paste a short description (not a screenshot — you
can't take one) of what you manually verified for each of the 3 fixes above,
per the Verification steps listed under each.

## Out of scope / do not touch
- Chart-type variety (bar-only issue) — likely a model-prompting behavior,
  not a code bug; not part of this chunk
- The `ChapterIndex: -1` backend bug — that's Chunk X, a different file,
  running independently. Do not attempt to fix `charts.go` from this chunk.
- `WarningBanner` missing `mismatch_detail` prop — separate, smaller
  follow-up, not included here on purpose (flag it in your report if you
  notice it while in this file, but don't fix it without a dedicated chunk)
