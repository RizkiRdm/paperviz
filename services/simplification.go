// Package services holds PaperViz's business logic: pipeline orchestration,
// PDF extraction, LLM-based simplification, claim-diff verification, and
// chart re-visualization. Per ARCHITECTURE.md, this layer MUST NOT import
// net/http types and MUST NOT contain SQL — it only calls external/ (Gemini,
// PDF libs) and returns plain Go values to the handlers layer.
package services

import (
	"context"
	"fmt"

	"paperviz/external"
)

// simplifiedPrompt and eli5Prompt are the two reading-level prompt templates
// named in PRD.md ("Simplified" = general audience, "ELI5" = child-level
// vocabulary). Both explicitly instruct the model to preserve every fact,
// number, and finding — this is the PRD's core correctness requirement
// (Success Metrics: "Meaning-preservation ... is a correctness requirement,
// not a nice-to-have"). The claim-diff verification service (verification.go)
// is what actually checks this instruction was followed; the prompt is the
// first line of defense, not the only one.
const simplifiedPrompt = `You are simplifying an academic paper's text for an undergraduate
student who is not a specialist in this field. Rewrite the passage below in
plain, general-audience English:

- Replace jargon with everyday words, or briefly define jargon you must keep.
- Break long, passive-voice sentences into shorter, active-voice sentences.
- Do NOT remove, change, round, or invent any number, statistic, name, date,
  or factual claim. Every fact in the original must still be present and
  accurate in your rewrite.
- Do NOT add commentary, opinions, or information not present in the original.
- Keep the same overall structure and order of ideas.

Original passage:
%s

Rewrite:`

const eli5Prompt = `You are explaining an academic paper's text to a curious 8-year-old.
Rewrite the passage below using simple, child-friendly vocabulary and short
sentences:

- Use concrete, everyday words and short sentences a child would understand.
- You may use simple analogies to explain a concept, but do NOT change what
  the original is actually claiming.
- Do NOT remove, change, round, or invent any number, statistic, name, date,
  or factual claim. Every fact in the original must still be present and
  accurate in your rewrite, even if described in simpler terms.
- Do NOT add commentary, opinions, or information not present in the original.

Original passage:
%s

Rewrite:`

// Simplify rewrites text at the given reading level using the Gemini client.
// It is a pure function of its inputs plus one external call — no hidden
// config reads, no persistence, per ARCHITECTURE.md Section 5 Module Rules.
//
// level must be the string "simplified" or "eli5" (PRD.md reading levels).
// Handlers validate the incoming request field before the pipeline ever
// reaches this call, so an invalid value here indicates a bug upstream,
// not bad user input — see handlers/documents.go for the request-time check.
func Simplify(ctx context.Context, client *external.GeminiClient, text, level string) (string, error) {
	var template string
	switch level {
	case "simplified":
		template = simplifiedPrompt
	case "eli5":
		template = eli5Prompt
	default:
		return "", ErrInvalidReadingLevel
	}

	prompt := fmt.Sprintf(template, text)
	result, err := client.Generate(ctx, prompt, false)
	if err != nil {
		return "", fmt.Errorf("simplify text: %w", err)
	}
	return result, nil
}
