package services

import (
	"fmt"
	"sort"
	"strings"

	"paperviz/internal/models"
)

// BuildCandidateDatasets groups NumericEvidence items into chart-ready datasets.
func BuildCandidateDatasets(evidence []models.NumericEvidence) []models.CandidateDataset {
	groups := groupEvidence(evidence)
	return buildDatasets(groups)
}

// groupEvidence clusters evidence items by metric+unit, skipping empty entities.
func groupEvidence(evidence []models.NumericEvidence) map[string][]models.NumericEvidence {
	groups := make(map[string][]models.NumericEvidence)
	for _, e := range evidence {
		if strings.TrimSpace(e.Entity) == "" {
			continue
		}
		key := datasetKey(e.Metric, e.Unit)
		groups[key] = append(groups[key], e)
	}
	return groups
}

// datasetKey returns a composite key for grouping by metric and unit.
func datasetKey(metric, unit string) string {
	return fmt.Sprintf("%s||%s", metric, unit)
}

// buildDatasets converts grouped evidence into sorted CandidateDataset slices.
func buildDatasets(groups map[string][]models.NumericEvidence) []models.CandidateDataset {
	if len(groups) == 0 {
		return nil
	}
	datasets := make([]models.CandidateDataset, 0, len(groups))
	for key, items := range groups {
		metric, unit := splitKey(key)
		points := buildPoints(items)
		datasets = append(datasets, models.CandidateDataset{
			ID:     fmt.Sprintf("ds_%s_%s", slug(metric), slug(unit)),
			Title:  formatTitle(metric, unit),
			Metric: metric,
			Unit:   unit,
			Points: points,
		})
	}
	sort.Slice(datasets, func(i, j int) bool {
		return datasets[i].Metric < datasets[j].Metric
	})
	return datasets
}

// buildPoints sorts evidence items by entity name and maps them to DatasetPoints.
func buildPoints(items []models.NumericEvidence) []models.DatasetPoint {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Entity < items[j].Entity
	})
	points := make([]models.DatasetPoint, len(items))
	for i, item := range items {
		points[i] = models.DatasetPoint{
			Label:      item.Entity,
			Value:      item.Value,
			EvidenceID: item.ID,
		}
	}
	return points
}

// splitKey extracts metric and unit from a composite key string.
func splitKey(key string) (string, string) {
	parts := strings.SplitN(key, "||", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return key, ""
}

// slug converts a string to a lowercase, hyphenated identifier.
func slug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

// formatTitle produces a human-readable title from metric and unit.
func formatTitle(metric, unit string) string {
	if unit != "" {
		return fmt.Sprintf("%s (%s)", metric, unit)
	}
	return metric
}
