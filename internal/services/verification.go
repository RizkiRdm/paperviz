package services

import (
	"context"
	"encoding/json"
	"fmt"

	"paperviz/internal/external"
)

const dualClaimExtractionPrompt = `Extract every factual claim from EACH of the two passages below, as two
separate JSON arrays of strings. A "factual claim" is any specific number,
statistic, finding, comparison, name, date, or result stated in the text —
the kind of detail that would be wrong if changed. Do NOT include
opinions, transitions, or restatements of the same claim twice.

Respond with ONLY a JSON object in this exact shape, nothing else:
{"original_claims": ["claim1", "claim2", ...], "simplified_claims": ["claim1", "claim2", ...]}

Original passage:
%s

Simplified passage:
%s`

type dualClaimExtractionResult struct {
	OriginalClaims   []string `json:"original_claims"`
	SimplifiedClaims []string `json:"simplified_claims"`
}

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

// DiffClaims runs the full claim-diff verification: extract claims from
// both the original and simplified text together (1 dual-extraction LLM
// call), then ask Gemini to judge whether they're factually consistent.
// This is what stands between "the model wrote something plausible" and
// "the model preserved the paper's actual findings" — treat this as a
// correctness gate, not a formality.
//
// A mismatch here means the pipeline MUST NOT publish the simplified text
// as verified (see ARCHITECTURE.md Acceptance Scenario 4) — the caller
// (pipeline.go) is responsible for setting status=verification_failed
// rather than complete when MismatchDetected is true.
//
// Call count: 2 total (1 dual-extraction + 1 comparison), down from 3
// before the merge (2 single-extraction + 1 comparison).
func DiffClaims(ctx context.Context, client *external.GeminiClient, originalText, simplifiedText string) (VerifyResult, error) {
	prompt := fmt.Sprintf(dualClaimExtractionPrompt, originalText, simplifiedText)
	raw, err := client.Generate(ctx, prompt, true, 0)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("extract claims: %w", err)
	}

	var dual dualClaimExtractionResult
	if err := json.Unmarshal([]byte(raw), &dual); err != nil {
		return VerifyResult{}, fmt.Errorf("parse dual claims JSON: %w", err)
	}

	originalJSON, err := json.Marshal(dual.OriginalClaims)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("marshal original claims: %w", err)
	}
	simplifiedJSON, err := json.Marshal(dual.SimplifiedClaims)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("marshal simplified claims: %w", err)
	}

	prompt = fmt.Sprintf(claimComparisonPrompt, originalJSON, simplifiedJSON)
	raw, err = client.Generate(ctx, prompt, true, 0)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("compare claims: %w", err)
	}

	var comparison claimComparisonResult
	if err := json.Unmarshal([]byte(raw), &comparison); err != nil {
		return VerifyResult{}, fmt.Errorf("parse comparison JSON: %w", err)
	}

	return VerifyResult{
		OriginalClaims:   dual.OriginalClaims,
		SimplifiedClaims: dual.SimplifiedClaims,
		MismatchDetected: comparison.MismatchDetected,
		MismatchDetail:   comparison.Detail,
	}, nil
}
