# Task 1 — Inspect the Existing Project

Inspect `[file]` and the surrounding project structure before making changes.

Identify:

1. How AI responses are received.
2. How Markdown responses are currently rendered.
3. How the `<article>` content is generated.
4. How the `Original` section is displayed.
5. How charts and visualizations are rendered.
6. How images from papers are handled.
7. How mathematical formulas are handled.
8. How the current `ELI5` category works.
9. Where the category logic is implemented.
10. How chapters are currently generated.

Do not make major changes yet.

First understand the existing implementation and identify which files/components need to be changed.

---

# Task 2 — Fix Markdown Rendering

Improve the Markdown rendering used for AI responses.

The AI response is Markdown, so use a proper Markdown rendering library instead of manually parsing Markdown.

The renderer should correctly support:

- Headings
- Paragraphs
- Bold/italic
- Lists
- Tables
- Blockquotes
- Code blocks
- Inline code
- Links
- Images
- LaTeX/math expressions

The rendered result must produce clean HTML suitable for an `<article>`.

Do not display raw Markdown syntax to the user.

Do not change unrelated functionality.

---

# Task 3 — Fix the `<article>` Presentation

Improve the visual structure of the rendered AI response.

The `<article>` should look like a properly formatted reading experience.

Improve:

- Typography
- Heading hierarchy
- Paragraph spacing
- Lists
- Tables
- Code blocks
- Blockquotes
- Links
- Images
- Mathematical formulas

The result should be readable and visually consistent with the existing application.

Do not change the actual AI-generated content.

Only improve how it is rendered.

---

# Task 4 — Fix the Original Section

The current `Original` section is messy and difficult to read.

Improve its rendering and structure.

Requirements:

- Preserve the original information.
- Do not summarize the content.
- Do not remove important information.
- Preserve images and figures when available.
- Preserve mathematical formulas.
- Preserve tables when possible.
- Make the content readable as a proper article.

If the original content is already Markdown, render it using the same reliable Markdown pipeline.

---

# Task 5 — Add Proper Mathematical Rendering

Add proper support for mathematical formulas.

Use a reliable math-rendering library such as:

- KaTeX
- MathJax

Support both:

- Inline mathematics
- Block mathematics

For example:

`E = mc^2`

should render as mathematics instead of displaying raw LaTeX syntax.

Make sure math rendering works together with the Markdown renderer.

Do not break normal Markdown rendering.

---

# Task 6 — Fix Paper Images and Figures

Improve how images from papers are displayed.

When an AI response references or contains a paper figure, architecture diagram, flowchart, or other visual, display it properly.

Support:

- Paper figures
- Architecture diagrams
- System diagrams
- Flowcharts
- Experimental figures
- Other relevant images

Images should not appear as broken Markdown, raw URLs, or broken HTML.

Keep images inside the appropriate content when they belong to the article.

---

# Task 7 — Change the Layout to Two Columns

Change the current chart/visual layout.

Do not place charts underneath the AI response.

Create a responsive two-column layout:

### Left column
Contains:

- AI response
- Rendered Markdown
- Article content
- Original content

### Right column
Contains:

- Charts
- Visualizations
- Paper figures
- Other visual elements

The layout should be responsive.

On smaller screens, allow the columns to stack vertically.

Do not redesign the entire application. Only change the relevant layout.

---

# Task 8 — Remove ELI5

Remove the existing `ELI5` category.

Remove it from:

- UI
- Category selection
- Relevant state/logic
- Prompts or AI-generation logic
- Any conditional rendering related to ELI5

Make sure there are no broken references after removing it.

Do not replace it yet.

The new categories will be implemented in the next tasks.

---

# Task 9 — Implement the `Simplified` Category

Add a new category called:

`Simplified`

Its purpose is to make academic papers easier to understand while still preserving meaningful technical content.

Rules:

- Make difficult concepts easier to understand.
- Do not make the content excessively simple.
- Preserve important technical concepts.
- Do not remove important information.
- Keep the explanation useful for students.
- Use a chapter-based structure.
- Chapters should be based on the paper's content.
- This category can have a limit on chart/visual creation.

The goal is:

> Make the paper easier to understand without turning it into an oversimplified explanation.

Make sure this behavior is clearly separated from the old ELI5 behavior.

---

# Task 10 — Implement the `FB Words` Category

Add a second category:

`FB Words`

FB Words means **Focus on Buzzwords**.

This category must behave differently from `Simplified`.

Its purpose is to simplify overly academic terminology, not to simplify the entire paper.

For example:

- `Interdisciplinary` → `involving multiple fields`
- `Multidisciplinary` → `involving several areas of study`
- `Paradigm Shift` → `a major change in the way something is understood`

Rules:

- Do not summarize.
- Do not shorten the content.
- Do not remove information.
- Do not remove technical details.
- Do not simplify entire paragraphs unnecessarily.
- Only replace unnecessarily complicated academic/buzzword terminology.
- Preserve the original meaning.
- Preserve the original amount of information.
- Chapters should come from the paper itself.
- Do not create an artificial chapter system.
- There is no chart/visual creation limit for this category.

The goal is:

> Keep the full content of the paper while replacing unnecessarily complicated academic buzzwords with clearer language.

---

# Task 11 — Verify the Category Differences

Verify that the two categories behave differently.

| Feature | Simplified | FB Words |
|---|---|---|
| Simplifies content | Yes | No |
| Removes information | No | No |
| Simplifies buzzwords | Yes | Main purpose |
| Generated chapter system | Yes | No |
| Uses paper chapters | Yes | Yes |
| Chart creation limit | Yes | No |
| Preserves full information | Yes | Yes |

Do not allow `FB Words` to accidentally behave like `Simplified`.

Do not allow `Simplified` to behave like the old ELI5 mode.

---

# Task 12 — Final Integration and Testing

Now verify the entire implementation.

Test:

1. AI Markdown rendering.
2. `<article>` formatting.
3. Original content rendering.
4. Mathematical formulas.
5. Paper images.
6. Charts.
7. Two-column layout.
8. Responsive layout.
9. `Simplified` category.
10. `FB Words` category.
11. ELI5 removal.
12. Chapter behavior.
13. Chart limits.

Check for:

- TypeScript errors
- Build errors
- Runtime errors
- Broken imports
- Missing dependencies
- Broken components
- Incorrect state handling

Fix any problems you find.

Do not rewrite unrelated parts of the project.

Keep the implementation clean, maintainable, and consistent with the existing codebase.