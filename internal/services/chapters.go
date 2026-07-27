package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"paperviz/internal/external"
)

const maxChapters = 10

const chapterDetectionPrompt = `You are analyzing a simplified academic paper's text to identify its
logical chapters or sections (e.g. Introduction, Methodology, Results,
Discussion, Conclusion — but use whatever structure actually fits this
specific paper, not a fixed template).

For each chapter/section you identify, extract:
- "title": a short section title (a few words)
- "summary": one sentence describing what this section covers
- "excerpt": the actual text belonging to this section (can be shortened
  if very long, but MUST preserve any numbers, statistics, or comparisons
  that appear in it — do not summarize away numeric data)

Identify at most %d chapters. If the paper naturally has fewer distinct
sections than that, return fewer — do not invent artificial splits.

Respond with ONLY a JSON array in this exact shape, nothing else:
[
  {"title": "...", "summary": "...", "excerpt": "..."},
  ...
]

Simplified paper text:
%s`

type chapterJSON struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Excerpt string `json:"excerpt"`
}

func DetectChapters(ctx context.Context, client *external.GeminiClient, simplifiedText string) ([]Chapter, error) {
	prompt := fmt.Sprintf(chapterDetectionPrompt, maxChapters, simplifiedText)
	raw, err := client.Generate(ctx, prompt, true, 0)
	if err != nil {
		return nil, fmt.Errorf("detect chapters: %w", err)
	}

	trimmed := strings.TrimSpace(raw)
	trimmed = stripJSONFences(trimmed)
	if trimmed == "" || trimmed == "[]" || trimmed == "null" {
		slog.Info("chapter detection: no chapters found", "stage", "chapters")
		return nil, nil
	}

	var parsed []chapterJSON
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil, fmt.Errorf("parse chapters JSON: %w", err)
	}

	if len(parsed) > maxChapters {
		parsed = parsed[:maxChapters]
	}

	chapters := make([]Chapter, 0, len(parsed))
	for _, c := range parsed {
		if strings.TrimSpace(c.Excerpt) == "" {
			continue
		}
		chapters = append(chapters, Chapter{
			Title:   c.Title,
			Summary: c.Summary,
			Excerpt: c.Excerpt,
		})
	}

	slog.Info("chapter detection complete", "stage", "chapters", "chapters_found", len(chapters))
	return chapters, nil
}
