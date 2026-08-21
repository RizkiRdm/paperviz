package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"paperviz/internal/external"
)

const paperExtractionPrompt = `Extract structured fields from this academic paper:
- "research_question": main research question or objective (1-2 sentences)
- "methodology": research methods used (1-2 sentences)
- "dataset": description of the dataset or data source (1 sentence)
- "sample_size": sample size or number of participants (1 sentence)
- "findings": array of key findings (each as separate string)
- "limitations": array of stated limitations (each as separate string)
- "figures": array of figure descriptions (brief description)
- "evidence": array of key evidence points (each as separate string)
- "conclusions": main conclusions drawn (1-2 sentences)

JSON shape: {"research_question":"...","methodology":"...","dataset":"...","sample_size":"...","findings":["..."],"limitations":["..."],"figures":["..."],"evidence":["..."],"conclusions":"..."}

Paper text:
%s`

type paperExtractionResult struct {
	ResearchQuestion string   `json:"research_question"`
	Methodology      string   `json:"methodology"`
	Dataset          string   `json:"dataset"`
	SampleSize       string   `json:"sample_size"`
	Findings         []string `json:"findings"`
	Limitations      []string `json:"limitations"`
	Figures          []string `json:"figures"`
	Evidence         []string `json:"evidence"`
	Conclusions      string   `json:"conclusions"`
}

const evidenceComparisonPrompt = `Compare evidence claims across these academic papers. For each distinct claim found in any paper's evidence, determine which papers support it, contradict it, or have unclear stance.

Papers:
%s

Return JSON array of claims. Each claim object:
{
  "claim": "specific factual claim about research findings",
  "stances": {"doc_id_1": "supporting", "doc_id_2": "contradicting", "doc_id_3": "unclear"},
  "source_refs": {"doc_id_1": "relevant evidence text from that paper"}
}

Claims should be specific (e.g., "Method X improves outcome Y by Z%%"), not vague themes.
Only include claims that appear in at least 2 papers' evidence.
JSON array shape: [{"claim":"...","stances":{...},"source_refs":{...}}]`

const comparisonSynthesisPrompt = `Compare these academic papers. For each dimension, provide synthesis notes.

Dimensions: %s

JSON array shape: [{"dimension":"...","notes":"synthesis observation"}, ...]

Paper summaries:
%s`

type dimensionSynthesis struct {
	Dimension string `json:"dimension"`
	Notes     string `json:"notes"`
}

type evidenceClaimResult struct {
	Claim      string            `json:"claim"`
	Stances    map[string]string `json:"stances"`
	SourceRefs map[string]string `json:"source_refs"`
}

// ExtractPaperSummary extracts structured fields from a single paper for multi-paper comparison.
func ExtractPaperSummary(ctx context.Context, client *external.GeminiClient, documentID, title, paperText string) (PaperSummary, error) {
	prompt := fmt.Sprintf(paperExtractionPrompt, paperText)
	parsed, err := external.ExtractJSON[paperExtractionResult](ctx, client, prompt, 0)
	if err != nil {
		return PaperSummary{}, fmt.Errorf("extract paper fields: %w", err)
	}

	return PaperSummary{
		DocumentID:       documentID,
		Title:            title,
		ResearchQuestion: parsed.ResearchQuestion,
		Methodology:      parsed.Methodology,
		Dataset:          parsed.Dataset,
		SampleSize:       parsed.SampleSize,
		Findings:         parsed.Findings,
		Limitations:      parsed.Limitations,
		Figures:          parsed.Figures,
		Evidence:         parsed.Evidence,
		Conclusions:      parsed.Conclusions,
	}, nil
}

// ComparePapers generates a structured comparison across multiple papers.
func ComparePapers(ctx context.Context, client *external.GeminiClient, papers []PaperSummary) (PaperComparison, error) {
	if len(papers) < 2 {
		return PaperComparison{}, fmt.Errorf("at least 2 papers required for comparison")
	}

	dimensions := buildComparisonDimensions(papers)

	dimensionNotes, err := synthesizeDimensions(ctx, client, dimensions, papers)
	if err != nil {
		slog.Warn("dimension synthesis failed, using raw values", "error", err)
	} else {
		for _, note := range dimensionNotes {
			for j, dim := range dimensions {
				if dim.Dimension == note.Dimension {
					dimensions[j].Notes = note.Notes
				}
			}
		}
	}

	agreement, disagreement := identifyAgreementsAndDisagreements(papers)

	evidenceClaims, err := CompareEvidence(ctx, client, papers)
	if err != nil {
		slog.Warn("evidence comparison failed, omitting evidence claims", "error", err)
	}

	return PaperComparison{
		Papers:         papers,
		Dimensions:     dimensions,
		Agreement:      agreement,
		Disagreement:   disagreement,
		EvidenceClaims: evidenceClaims,
	}, nil
}

func buildComparisonDimensions(papers []PaperSummary) []ComparisonDimension {
	dims := []ComparisonDimension{
		{Dimension: "research_question", Values: make(map[string]string)},
		{Dimension: "methodology", Values: make(map[string]string)},
		{Dimension: "dataset", Values: make(map[string]string)},
		{Dimension: "sample_size", Values: make(map[string]string)},
		{Dimension: "findings", Values: make(map[string]string)},
		{Dimension: "limitations", Values: make(map[string]string)},
		{Dimension: "evidence", Values: make(map[string]string)},
		{Dimension: "conclusions", Values: make(map[string]string)},
	}

	for _, paper := range papers {
		dims[0].Values[paper.DocumentID] = paper.ResearchQuestion
		dims[1].Values[paper.DocumentID] = paper.Methodology
		dims[2].Values[paper.DocumentID] = paper.Dataset
		dims[3].Values[paper.DocumentID] = paper.SampleSize
		dims[4].Values[paper.DocumentID] = strings.Join(paper.Findings, "; ")
		dims[5].Values[paper.DocumentID] = strings.Join(paper.Limitations, "; ")
		dims[6].Values[paper.DocumentID] = strings.Join(paper.Evidence, "; ")
		dims[7].Values[paper.DocumentID] = paper.Conclusions
	}

	return dims
}

func synthesizeDimensions(ctx context.Context, client *external.GeminiClient, dims []ComparisonDimension, papers []PaperSummary) ([]dimensionSynthesis, error) {
	var dimensionNames []string
	for _, d := range dims {
		dimensionNames = append(dimensionNames, d.Dimension)
	}

	var paperSummaries []string
	for _, p := range papers {
		summary := fmt.Sprintf("Paper '%s': RQ: %s, Methods: %s, Findings: %s, Conclusions: %s",
			p.Title, p.ResearchQuestion, p.Methodology, strings.Join(p.Findings, "; "), p.Conclusions)
		paperSummaries = append(paperSummaries, summary)
	}

	prompt := fmt.Sprintf(comparisonSynthesisPrompt,
		strings.Join(dimensionNames, "\n"),
		strings.Join(paperSummaries, "\n\n"))

	parsed, err := external.ExtractJSON[[]dimensionSynthesis](ctx, client, prompt, 0)
	if err != nil {
		return nil, fmt.Errorf("synthesize dimensions: %w", err)
	}

	return parsed, nil
}

func identifyAgreementsAndDisagreements(papers []PaperSummary) (agreement []string, disagreement []string) {
	if len(papers) < 2 {
		return agreement, disagreement
	}

	rq1 := strings.ToLower(papers[0].ResearchQuestion)
	rq2 := strings.ToLower(papers[1].ResearchQuestion)

	commonKeywords := findCommonKeywords(rq1, rq2)
	if len(commonKeywords) > 0 {
		agreement = append(agreement, fmt.Sprintf("Papers share common research themes: %s", strings.Join(commonKeywords, ", ")))
	}

	m1 := strings.ToLower(papers[0].Methodology)
	m2 := strings.ToLower(papers[1].Methodology)

	bothExperimental := strings.Contains(m1, "experiment") && strings.Contains(m2, "experiment")
	bothSurvey := strings.Contains(m1, "survey") && strings.Contains(m2, "survey")
	bothQualitative := strings.Contains(m1, "qualitative") && strings.Contains(m2, "qualitative")
	bothQuantitative := strings.Contains(m1, "quantitative") && strings.Contains(m2, "quantitative")

	if bothExperimental || bothSurvey || bothQualitative || bothQuantitative {
		agreement = append(agreement, "Papers use similar research methodology")
	} else {
		disagreement = append(disagreement, "Papers use different research methodologies")
	}

	return agreement, disagreement
}

// CompareEvidence identifies cross-paper claims and per-paper stance.
func CompareEvidence(ctx context.Context, client *external.GeminiClient, papers []PaperSummary) ([]EvidenceClaim, error) {
	if len(papers) < 2 {
		return nil, nil
	}

	var paperDescriptions []string
	for _, p := range papers {
		desc := fmt.Sprintf("Document '%s' (id: %s):\n  Evidence: %s",
			p.Title, p.DocumentID, strings.Join(p.Evidence, "; "))
		paperDescriptions = append(paperDescriptions, desc)
	}

	prompt := fmt.Sprintf(evidenceComparisonPrompt, strings.Join(paperDescriptions, "\n\n"))

	parsed, err := external.ExtractJSON[[]evidenceClaimResult](ctx, client, prompt, 0)
	if err != nil {
		return nil, fmt.Errorf("compare evidence: %w", err)
	}

	var claims []EvidenceClaim
	for _, r := range parsed {
		claims = append(claims, EvidenceClaim{
			Claim:      r.Claim,
			Stances:    r.Stances,
			SourceRefs: r.SourceRefs,
		})
	}
	return claims, nil
}

func findCommonKeywords(text1, text2 string) []string {
	words1 := strings.Fields(text1)
	words2 := strings.Fields(text2)

	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "is": true, "are": true,
		"was": true, "were": true, "in": true, "of": true, "for": true,
		"to": true, "and": true, "or": true, "but": true, "not": true,
		"with": true, "on": true, "at": true, "by": true, "from": true,
	}

	wordSet := make(map[string]bool)
	for _, w := range words1 {
		w = strings.Trim(w, ".,;:!?\"'()-")
		if len(w) > 3 && !stopWords[w] {
			wordSet[w] = true
		}
	}

	var common []string
	seen := make(map[string]bool)
	for _, w := range words2 {
		w = strings.Trim(w, ".,;:!?\"'()-")
		if len(w) > 3 && !stopWords[w] && wordSet[w] && !seen[w] {
			common = append(common, w)
			seen[w] = true
		}
	}

	return common
}
