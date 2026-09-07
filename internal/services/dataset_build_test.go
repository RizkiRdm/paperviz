package services

import (
	"testing"

	"paperviz/internal/models"
)

// TestBuildCandidateDatasetsSameMetric verifies same-metric items group into one dataset.
func TestBuildCandidateDatasetsSameMetric(t *testing.T) {
	evidence := []models.NumericEvidence{
		{ID: "e1", Metric: "accuracy", Entity: "Model A", Value: 0.95, Unit: "%", Source: models.EvidenceSource{Page: 1, Text: "A"}},
		{ID: "e2", Metric: "accuracy", Entity: "Model B", Value: 0.88, Unit: "%", Source: models.EvidenceSource{Page: 1, Text: "B"}},
	}
	datasets := BuildCandidateDatasets(evidence)
	if len(datasets) != 1 {
		t.Fatalf("want 1 dataset, got %d", len(datasets))
	}
	if len(datasets[0].Points) != 2 {
		t.Fatalf("want 2 points, got %d", len(datasets[0].Points))
	}
	// Points should be sorted by entity name.
	if datasets[0].Points[0].Label != "Model A" {
		t.Errorf("want first point label Model A, got %s", datasets[0].Points[0].Label)
	}
	if datasets[0].Points[1].Label != "Model B" {
		t.Errorf("want second point label Model B, got %s", datasets[0].Points[1].Label)
	}
}

// TestBuildCandidateDatasetsDifferentMetrics verifies different metrics produce separate datasets.
func TestBuildCandidateDatasetsDifferentMetrics(t *testing.T) {
	evidence := []models.NumericEvidence{
		{ID: "e1", Metric: "accuracy", Entity: "Model A", Value: 0.95, Unit: "%", Source: models.EvidenceSource{Page: 1, Text: "A"}},
		{ID: "e2", Metric: "latency", Entity: "Model A", Value: 120, Unit: "ms", Source: models.EvidenceSource{Page: 1, Text: "B"}},
	}
	datasets := BuildCandidateDatasets(evidence)
	if len(datasets) != 2 {
		t.Fatalf("want 2 datasets, got %d", len(datasets))
	}
}

// TestBuildCandidateDatasetsEmpty verifies empty input returns nil.
func TestBuildCandidateDatasetsEmpty(t *testing.T) {
	datasets := BuildCandidateDatasets(nil)
	if datasets != nil {
		t.Errorf("want nil, got %v", datasets)
	}
}

// TestBuildCandidateDatasetsSkipsEmptyEntity verifies items with blank entity are excluded.
func TestBuildCandidateDatasetsSkipsEmptyEntity(t *testing.T) {
	evidence := []models.NumericEvidence{
		{ID: "e1", Metric: "accuracy", Entity: "", Value: 0.95, Unit: "%", Source: models.EvidenceSource{Page: 1, Text: "A"}},
		{ID: "e2", Metric: "accuracy", Entity: "Model A", Value: 0.88, Unit: "%", Source: models.EvidenceSource{Page: 1, Text: "B"}},
	}
	datasets := BuildCandidateDatasets(evidence)
	if len(datasets) != 1 {
		t.Fatalf("want 1 dataset, got %d", len(datasets))
	}
	if len(datasets[0].Points) != 1 {
		t.Fatalf("want 1 point (empty entity skipped), got %d", len(datasets[0].Points))
	}
}

// TestBuildCandidateDatasetsDifferentUnits verifies same metric with different units splits.
func TestBuildCandidateDatasetsDifferentUnits(t *testing.T) {
	evidence := []models.NumericEvidence{
		{ID: "e1", Metric: "latency", Entity: "Model A", Value: 120, Unit: "ms", Source: models.EvidenceSource{Page: 1, Text: "A"}},
		{ID: "e2", Metric: "latency", Entity: "Model B", Value: 2, Unit: "seconds", Source: models.EvidenceSource{Page: 1, Text: "B"}},
	}
	datasets := BuildCandidateDatasets(evidence)
	if len(datasets) != 2 {
		t.Fatalf("want 2 datasets for different units, got %d", len(datasets))
	}
}
