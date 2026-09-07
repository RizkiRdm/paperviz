package services

import (
	"regexp"
	"strconv"
	"strings"

	"paperviz/internal/models"
)

// Regex patterns for numeric extraction.
var (
	reFromTo    = regexp.MustCompile(`(?i)\b([\w][\w\s]*?)\s+from\s+([\d.,]+)\s*(%|ms|seconds?|minutes?|hours?)\s+to\s+([\d.,]+)\s*(%|ms|seconds?|minutes?|hours?)`)
	reStat      = regexp.MustCompile(`(?i)\b(\w+)\s*=\s*([\d.,]+)\s*(?:,\s*[\d.]+%\s*CI\s*\[.*?\])?`)
	reEntityVal = regexp.MustCompile(`(?i)\b([\w][\w\s]*?)\s+(?:achieved|obtained|reported|had|reached|scored)\s+([\d.,]+)\s*(%|ms|seconds?|minutes?|hours?)`)
)

// knownMetrics is the set of recognized metric keywords.
var knownMetrics = map[string]bool{
	"accuracy": true, "f1": true, "precision": true, "recall": true,
	"latency": true, "throughput": true, "speed": true, "error": true,
	"loss": true, "time": true, "rate": true,
	"bleu": true, "rouge": true, "auc": true, "rmse": true, "mae": true,
	"beta": true, "alpha": true, "gamma": true, "p-value": true,
	"sensitivity": true, "specificity": true,
}

// conjunctions are words that cannot start an entity in "X = value" patterns.
var conjunctions = map[string]bool{
	"and": true, "or": true, "but": true, "the": true, "a": true, "an": true,
	"with": true, "for": true, "in": true, "on": true, "at": true, "to": true,
	"of": true, "by": true, "from": true, "as": true, "is": true, "was": true,
	"are": true, "were": true, "has": true, "have": true, "had": true,
}

// metricOrder is a deterministic list of known metrics for context detection.
var metricOrder = []string{
	"accuracy", "f1", "precision", "recall", "latency", "throughput",
	"speed", "error", "loss", "time", "rate", "bleu", "rouge", "auc",
	"rmse", "mae", "beta", "alpha", "gamma", "p-value", "sensitivity", "specificity",
}

// splitSentences splits text into sentences by common delimiters.
func splitSentences(text string) []string {
	raw := regexp.MustCompile(`(?m)[.!]\s+|\n\s*\n`).Split(text, -1)
	var out []string
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if len(s) > 5 {
			out = append(out, s)
		}
	}
	return out
}

// normalizeNumber strips commas from numbers like "1,234.5".
func normalizeNumber(s string) float64 {
	s = strings.TrimRight(s, ".,")
	s = strings.ReplaceAll(s, ",", "")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// isStopWord returns true for words that aren't metric names.
func isStopWord(w string) bool {
	w = strings.ToLower(w)
	if knownMetrics[w] {
		return false
	}
	stops := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "as": true,
		"is": true, "was": true, "are": true, "were": true, "been": true,
		"has": true, "have": true, "had": true, "it": true, "we": true,
		"they": true, "this": true, "that": true, "from": true, "than": true,
		"between": true, "model": true, "baseline": true, "system": true,
		"algorithm": true, "method": true, "approach": true, "technique": true,
		"result": true, "results": true, "performance": true,
		"improved": true, "increased": true, "decreased": true,
		"compared": true, "using": true, "over": true,
	}
	return stops[w]
}

// detectMetricFromContext finds the metric name that appears closest to a number in the sentence.
func detectMetricFromContext(sentence string) string {
	lower := strings.ToLower(sentence)
	bestMetric := "value"
	bestDist := len(lower)
	for _, m := range metricOrder {
		idx := strings.Index(lower, m)
		if idx < 0 {
			continue
		}
		if idx < bestDist {
			bestDist = idx
			bestMetric = m
		}
	}
	return bestMetric
}

// extractFromTo handles "X improved from 72.4% to 81.7%" → two evidence items.
func extractFromTo(sentence string, page int) []models.NumericEvidence {
	matches := reFromTo.FindStringSubmatch(sentence)
	if matches == nil {
		return nil
	}
	entity := strings.TrimSpace(matches[1])
	metric := detectMetricFromContext(sentence)
	return []models.NumericEvidence{
		{
			Metric: metric,
			Entity: entity,
			Value:  normalizeNumber(matches[2]),
			Unit:   matches[3],
			Source: models.EvidenceSource{Page: page, Text: sentence},
		},
		{
			Metric: metric,
			Entity: entity,
			Value:  normalizeNumber(matches[4]),
			Unit:   matches[5],
			Source: models.EvidenceSource{Page: page, Text: sentence},
		},
	}
}

// extractStatistical handles "beta = 0.42, 95% CI [0.31, 0.53]".
func extractStatistical(sentence string, page int) []models.NumericEvidence {
	matches := reStat.FindStringSubmatch(sentence)
	if matches == nil {
		return nil
	}
	metric := strings.ToLower(matches[1])
	if isStopWord(metric) {
		return nil
	}
	return []models.NumericEvidence{
		{
			Metric: metric,
			Entity: "",
			Value:  normalizeNumber(matches[2]),
			Unit:   "",
			Source: models.EvidenceSource{Page: page, Text: sentence},
		},
	}
}

// extractEntityValue handles "Model A achieved 72.4% accuracy".
func extractEntityValue(sentence string, page int) []models.NumericEvidence {
	matches := reEntityVal.FindStringSubmatch(sentence)
	if matches == nil {
		return nil
	}
	entity := strings.TrimSpace(matches[1])
	metric := detectMetricFromContext(sentence)
	return []models.NumericEvidence{
		{
			Metric: metric,
			Entity: entity,
			Value:  normalizeNumber(matches[2]),
			Unit:   matches[3],
			Source: models.EvidenceSource{Page: page, Text: sentence},
		},
	}
}

// extractEq handles "Model A = 72.4% and Model B = 81.7%" pattern.
// Splits on " and " to avoid capturing conjunctions as part of entities.
func extractEq(sentence string, page int) []models.NumericEvidence {
	reSingle := regexp.MustCompile(`(?i)\b([A-Za-z][\w]*(?:\s+[A-Za-z][\w]*)*)\s+=\s+([\d.,]+)\s*(%|ms|seconds?|minutes?|hours?)`)
	// Split sentence on " and " to avoid cross-entity capture
	parts := strings.Split(sentence, " and ")
	var out []models.NumericEvidence
	for _, part := range parts {
		matches := reSingle.FindAllStringSubmatch(part, -1)
		for _, m := range matches {
			entity := strings.TrimSpace(m[1])
			firstWord := strings.ToLower(strings.Split(entity, " ")[0])
			if conjunctions[firstWord] {
				continue
			}
			metric := detectMetricFromContext(sentence)
			out = append(out, models.NumericEvidence{
				Metric: metric,
				Entity: entity,
				Value:  normalizeNumber(m[2]),
				Unit:   m[3],
				Source: models.EvidenceSource{Page: page, Text: sentence},
			})
		}
	}
	return out
}

// extractStandalone handles metric + number without entity.
// Uses a dynamic regex built from knownMetrics to avoid matching non-metric words.
func extractStandalone(sentence string, page int) []models.NumericEvidence {
	metricNames := make([]string, 0, len(knownMetrics))
	for m := range knownMetrics {
		metricNames = append(metricNames, regexp.QuoteMeta(m))
	}
	re := regexp.MustCompile(`(?i)\b(` + strings.Join(metricNames, "|") + `)\b\s+(?:\w+\s+){0,3}([\d][\d.,]*)\s*(%|ms|seconds?|minutes?|hours?)?`)
	matches := re.FindAllStringSubmatch(sentence, -1)
	var out []models.NumericEvidence
	for _, m := range matches {
		word := strings.ToLower(strings.TrimSpace(m[1]))
		out = append(out, models.NumericEvidence{
			Metric: word,
			Entity: "",
			Value:  normalizeNumber(m[2]),
			Unit:   m[3],
			Source: models.EvidenceSource{Page: page, Text: sentence},
		})
	}
	return out
}

// ExtractNumericEvidence parses text and returns numeric claims with source provenance.
func ExtractNumericEvidence(text string, pageNumber int) []models.NumericEvidence {
	sentences := splitSentences(text)
	var all []models.NumericEvidence

	for _, sent := range sentences {
		// Pattern 1: from X to Y (before/after) — highest priority
		fromTo := extractFromTo(sent, pageNumber)
		if len(fromTo) > 0 {
			all = append(all, fromTo...)
			continue
		}

		// Pattern 2: statistical (beta = 0.42, ...)
		stat := extractStatistical(sent, pageNumber)
		if len(stat) > 0 {
			all = append(all, stat...)
			continue
		}

		// Pattern 3: entity + value (Model A achieved 72.4%)
		ev := extractEntityValue(sent, pageNumber)
		if len(ev) > 0 {
			all = append(all, ev...)
		}

		// Pattern 4: entity = value (Model A = 72.4%)
		eq := extractEq(sent, pageNumber)
		if len(eq) > 0 {
			all = append(all, eq...)
		}

		// Pattern 5: standalone metric + number — always run to catch
		// additional metrics in the same sentence (e.g. "95% accuracy and F1 0.89")
		standalone := extractStandalone(sent, pageNumber)
		all = append(all, standalone...)
	}

	return all
}
