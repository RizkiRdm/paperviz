package services

import (
	"context"
	"encoding/json"
	"fmt"

	"paperviz/internal/external"
)

// claimExtractionPrompt asks Gemini to list factual claims (numbers, named
// findings, statistics) from a passage as a JSON string array. Used on both
// the original and simplified text so they can be compared afterward.
//
// responseMimeType=application/json (passed via the `asJSON` flag in
// GeminiClient.Generate) is what makes Gemini return parseable JSON instead
// of a prose list — do not remove that flag from either call site below.
const claimExtractionPrompt = `Extract every factual claim from the passage below as a JSON array of
strings. A "factual claim" is any specific number, statistic, finding,
comparison, name, date, or result stated in the text — the kind of detail
that would be wrong if changed. Do NOT include opinions, transitions, or
restatements of the same claim twice. Output ONLY a JSON array of strings,
nothing else.

Passage:
%s`

// claimComparisonPrompt asks Gemini to compare two claim lists and report
// whether they disagree on any fact. This is a second, independent LLM call
// from claim extraction — comparing lists programmatically (e.g. string
// equality) would produce false positives for claims that are reworded but
// factually identical ("40%%" vs "two out of five"), so the comparison
// itself needs judgment, not just diffing.
const claimComparisonPrompt = `Below are two lists of factual claims extracted from an original academic
passage and a simplified rewrite of that passage. Determine whether the
simplified version's claims are factually consistent with the original's —
wording may differ, but every number, finding, and result must match.

Respond with ONLY a JSON object in this exact shape:
{"mismatch_detected": true or false, "detail": "short explanation, empty string if no mismatch"}

Original claims:
%s

Simplified claims:
%s`

// claimComparisonResult mirrors the JSON shape claimComparisonPrompt asks
// Gemini to return. Unexported: only used to unmarshal within this file.
type claimComparisonResult struct {
	MismatchDetected bool   `json:"mismatch_detected"`
	Detail           string `json:"detail"`
}

// extractClaims calls Gemini to pull a JSON array of factual claims out of
// a passage. Shared by both sides of DiffClaims below.
func extractClaims(ctx context.Context, client *external.GeminiClient, text string) ([]string, error) {
	prompt := fmt.Sprintf(claimExtractionPrompt, text)
	raw, err := client.Generate(ctx, prompt, true, 0)
	if err != nil {
		return nil, fmt.Errorf("extract claims: %w", err)
	}

	var claims []string
	if err := json.Unmarshal([]byte(raw), &claims); err != nil {
		return nil, fmt.Errorf("parse claims JSON: %w", err)
	}
	return claims, nil
}

// DiffClaims runs the full claim-diff verification: extract claims from
// both the original and simplified text (2 separate LLM calls, per PRD.md
// Core Capability 5), then ask Gemini to judge whether they're factually
// consistent. This is what stands between "the model wrote something
// plausible" and "the model preserved the paper's actual findings" —
// treat this as a correctness gate, not a formality.
//
// A mismatch here means the pipeline MUST NOT publish the simplified text
// as verified (see ARCHITECTURE.md Acceptance Scenario 4) — the caller
// (pipeline.go) is responsible for setting status=verification_failed
// rather than complete when MismatchDetected is true.
func DiffClaims(ctx context.Context, client *external.GeminiClient, originalText, simplifiedText string) (VerifyResult, error) {
	originalClaims, err := extractClaims(ctx, client, originalText)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("extract original claims: %w", err)
	}

	simplifiedClaims, err := extractClaims(ctx, client, simplifiedText)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("extract simplified claims: %w", err)
	}

	originalJSON, err := json.Marshal(originalClaims)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("marshal original claims: %w", err)
	}
	simplifiedJSON, err := json.Marshal(simplifiedClaims)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("marshal simplified claims: %w", err)
	}

	prompt := fmt.Sprintf(claimComparisonPrompt, originalJSON, simplifiedJSON)
	raw, err := client.Generate(ctx, prompt, true, 0)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("compare claims: %w", err)
	}

	var comparison claimComparisonResult
	if err := json.Unmarshal([]byte(raw), &comparison); err != nil {
		return VerifyResult{}, fmt.Errorf("parse comparison JSON: %w", err)
	}

	return VerifyResult{
		OriginalClaims:   originalClaims,
		SimplifiedClaims: simplifiedClaims,
		MismatchDetected: comparison.MismatchDetected,
		MismatchDetail:   comparison.Detail,
	}, nil
}
