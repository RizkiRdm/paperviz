package services

import (
	"math"
	"testing"

	"paperviz/internal/models"
)

// validBarSpec returns a valid bar ChartSpec for testing.
func validBarSpec() models.ChartSpec {
	return models.ChartSpec{
		Type:  "bar",
		Title: "Test Bar",
		Data: &models.BarData{
			Categories: []string{"A", "B", "C"},
			Series: []models.Series{
				{Name: "X", Values: []float64{1, 2, 3}},
			},
		},
	}
}

// validBarDatasets returns datasets matching validBarSpec.
func validBarDatasets() []models.CandidateDataset {
	return []models.CandidateDataset{
		{
			ID:     "ds1",
			Title:  "Dataset 1",
			Metric: "score",
			Unit:   "%",
			Points: []models.DatasetPoint{
				{Label: "A", Value: 1, EvidenceID: "e1"},
				{Label: "B", Value: 2, EvidenceID: "e2"},
				{Label: "C", Value: 3, EvidenceID: "e3"},
			},
		},
	}
}

// validScatterSpec returns a valid scatter ChartSpec for testing.
func validScatterSpec() models.ChartSpec {
	return models.ChartSpec{
		Type:  "scatter",
		Title: "Test Scatter",
		Data: &models.ScatterData{
			X: []float64{1, 2, 3},
			Y: []float64{4, 5, 6},
		},
	}
}

func TestValidateGrounding_ValidBar(t *testing.T) {
	got := ValidateGrounding(validBarSpec(), validBarDatasets())
	if got.Status != GroundingVerified {
		t.Errorf("expected verified, got %s; errors: %v", got.Status, got.Errors)
	}
}

func TestValidateGrounding_ValidScatter(t *testing.T) {
	got := ValidateGrounding(validScatterSpec(), nil)
	if got.Status != GroundingVerified {
		t.Errorf("expected verified, got %s; errors: %v", got.Status, got.Errors)
	}
}

func TestValidateGrounding_NaNValue(t *testing.T) {
	spec := models.ChartSpec{
		Type:  "bar",
		Title: "NaN Chart",
		Data: &models.BarData{
			Categories: []string{"A"},
			Series: []models.Series{
				{Name: "X", Values: []float64{math.NaN()}},
			},
		},
	}
	got := ValidateGrounding(spec, nil)
	if got.Status != GroundingUnsupported {
		t.Errorf("expected unsupported for NaN, got %s", got.Status)
	}
}

func TestValidateGrounding_MismatchedLengths(t *testing.T) {
	spec := models.ChartSpec{
		Type:  "bar",
		Title: "Mismatch",
		Data: &models.BarData{
			Categories: []string{"A", "B"},
			Series: []models.Series{
				{Name: "X", Values: []float64{1}},
			},
		},
	}
	got := ValidateGrounding(spec, nil)
	if got.Status != GroundingUnsupported {
		t.Errorf("expected unsupported for length mismatch, got %s", got.Status)
	}
}

func TestValidateGrounding_NoEvidence(t *testing.T) {
	spec := models.ChartSpec{
		Type:  "bar",
		Title: "No Evidence",
		Data: &models.BarData{
			Categories: []string{"A", "B"},
			Series: []models.Series{
				{Name: "X", Values: []float64{1, 2}},
			},
		},
	}
	got := ValidateGrounding(spec, nil)
	if got.Status == GroundingVerified {
		t.Error("expected non-verified for missing evidence")
	}
}

func TestValidateGrounding_InvalidType(t *testing.T) {
	spec := models.ChartSpec{
		Type:  "radar",
		Title: "Bad Type",
		Data:  &models.BarData{},
	}
	got := ValidateGrounding(spec, nil)
	if got.Status != GroundingUnsupported {
		t.Errorf("expected unsupported for invalid type, got %s", got.Status)
	}
}

func TestValidateGrounding_EmptySpec(t *testing.T) {
	got := ValidateGrounding(models.ChartSpec{}, nil)
	if got.Status != GroundingUnsupported {
		t.Errorf("expected unsupported for empty spec, got %s", got.Status)
	}
}

func TestValidateGrounding_PieNegative(t *testing.T) {
	spec := models.ChartSpec{
		Type:  "pie",
		Title: "Negative Pie",
		Data: &models.PieData{
			Labels: []string{"A", "B"},
			Values: []float64{10, -5},
		},
	}
	got := ValidateGrounding(spec, nil)
	if got.Status != GroundingUnsupported {
		t.Errorf("expected unsupported for negative pie value, got %s", got.Status)
	}
}

func TestValidateGrounding_PieValid(t *testing.T) {
	spec := models.ChartSpec{
		Type:  "pie",
		Title: "Valid Pie",
		Data: &models.PieData{
			Labels: []string{"A", "B"},
			Values: []float64{30, 70},
		},
	}
	got := ValidateGrounding(spec, nil)
	if got.Status != GroundingVerified {
		t.Errorf("expected verified for valid pie, got %s; errors: %v", got.Status, got.Errors)
	}
}

func TestValidateGrounding_LineValid(t *testing.T) {
	spec := models.ChartSpec{
		Type:  "line",
		Title: "Valid Line",
		Data: &models.LineData{
			X: []string{"t1", "t2"},
			Series: []models.Series{
				{Name: "S", Values: []float64{1, 2}},
			},
		},
	}
	datasets := []models.CandidateDataset{
		{
			ID:     "ds1",
			Title:  "Dataset 1",
			Metric: "score",
			Unit:   "%",
			Points: []models.DatasetPoint{
				{Label: "t1", Value: 1, EvidenceID: "e1"},
				{Label: "t2", Value: 2, EvidenceID: "e2"},
			},
		},
	}
	got := ValidateGrounding(spec, datasets)
	if got.Status != GroundingVerified {
		t.Errorf("expected verified for valid line, got %s; errors: %v", got.Status, got.Errors)
	}
}

func TestValidateGrounding_ScatterMismatchedXY(t *testing.T) {
	spec := models.ChartSpec{
		Type:  "scatter",
		Title: "Bad Scatter",
		Data: &models.ScatterData{
			X: []float64{1, 2},
			Y: []float64{1},
		},
	}
	got := ValidateGrounding(spec, nil)
	if got.Status != GroundingUnsupported {
		t.Errorf("expected unsupported for scatter x/y mismatch, got %s", got.Status)
	}
}
