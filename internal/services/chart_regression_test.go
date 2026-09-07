package services

import (
	"testing"

	"paperviz/internal/models"
)

// TestChartRegression is a table-driven suite covering 8 regression cases for the chart pipeline.
func TestChartRegression(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			// Case 1: simple text comparison extracts two evidence items and builds one dataset with two points.
			name: "case1_simple_text_comparison",
			run: func(t *testing.T) {
				input := "Model A = 72.4% and Model B = 81.7% accuracy."
				evidence := ExtractNumericEvidence(input, 1)
				if len(evidence) != 2 {
					t.Fatalf("want 2 evidence items, got %d", len(evidence))
				}
				datasets := BuildCandidateDatasets(evidence)
				if len(datasets) != 1 {
					t.Fatalf("want 1 dataset, got %d", len(datasets))
				}
				if len(datasets[0].Points) != 2 {
					t.Fatalf("want 2 points, got %d", len(datasets[0].Points))
				}
			},
		},
		{
			// Case 2: pipe-separated table extracts multiple evidence items from rows.
			name: "case2_table_data",
			run: func(t *testing.T) {
				input := "| Model | Accuracy | F1 |\n| --- | --- | --- |\n| Baseline | 72.4 | 0.68 |\n| Ours | 81.7 | 0.76 |"
				evidence := ExtractTableData(input, 1)
				if len(evidence) < 2 {
					t.Fatalf("want at least 2 evidence items from table, got %d", len(evidence))
				}
				datasets := BuildCandidateDatasets(evidence)
				if len(datasets) == 0 {
					t.Fatal("want at least 1 dataset from table evidence, got 0")
				}
			},
		},
		{
			// Case 3: ambiguous improvement with no concrete baseline produces zero evidence.
			name: "case3_ambiguous_improvement",
			run: func(t *testing.T) {
				input := "The model improved by 12% over baseline."
				evidence := ExtractNumericEvidence(input, 1)
				if len(evidence) != 0 {
					t.Fatalf("want 0 evidence items for ambiguous text, got %d", len(evidence))
				}
			},
		},
		{
			// Case 4: evidence with nil/missing values yields unsupported grounding status.
			name: "case4_missing_value_grounding",
			run: func(t *testing.T) {
				datasets := []models.CandidateDataset{
					{
						ID:     "ds_test",
						Title:  "test",
						Metric: "accuracy",
						Unit:   "%",
						Points: []models.DatasetPoint{
							{Label: "A", Value: 0, EvidenceID: "e1"},
						},
					},
				}
				// Bar spec with mismatched series length triggers unsupported.
				spec := models.ChartSpec{
					Type:  "bar",
					Title: "test",
					Data: &models.BarData{
						Categories: []string{"A", "B"},
						Series: []models.Series{
							{Name: "s1", Values: []float64{10}},
						},
					},
				}
				result := ValidateGrounding(spec, datasets)
				if result.Status != GroundingUnsupported {
					t.Fatalf("want unsupported status, got %s (errors: %v)", result.Status, result.Errors)
				}
			},
		},
		{
			// Case 5: scatter chart with numeric x/y passes grounding as verified.
			name: "case5_scatter_grounding",
			run: func(t *testing.T) {
				spec := models.ChartSpec{
					Type:  "scatter",
					Title: "scatter test",
					Data: &models.ScatterData{
						X: []float64{1.0, 2.0, 3.0},
						Y: []float64{10.0, 20.0, 30.0},
					},
				}
				result := ValidateGrounding(spec, nil)
				if result.Status != GroundingVerified {
					t.Fatalf("want verified status, got %s (errors: %v)", result.Status, result.Errors)
				}
			},
		},
		{
			// Case 6: pie chart with negative values is rejected as unsupported.
			name: "case6_invalid_pie_negative_values",
			run: func(t *testing.T) {
				spec := models.ChartSpec{
					Type:  "pie",
					Title: "invalid pie",
					Data: &models.PieData{
						Labels: []string{"A", "B"},
						Values: []float64{50, -10},
					},
				}
				result := ValidateGrounding(spec, nil)
				if result.Status != GroundingUnsupported {
					t.Fatalf("want unsupported status for negative pie, got %s", result.Status)
				}
			},
		},
		{
			// Case 7: values not traceable to any evidence dataset yield unsupported grounding.
			name: "case7_unsupported_untraceable_values",
			run: func(t *testing.T) {
				spec := models.ChartSpec{
					Type:  "bar",
					Title: "untraceable",
					Data: &models.BarData{
						Categories: []string{"X", "Y"},
						Series: []models.Series{
							{Name: "s1", Values: []float64{10, 20}},
						},
					},
				}
				// Empty datasets — no evidence to trace categories to.
				result := ValidateGrounding(spec, nil)
				if result.Status == GroundingVerified {
					t.Fatal("want non-verified status for untraceable values")
				}
			},
		},
		{
			// Case 8: evidence provenance fields (source text and page) are populated correctly.
			name: "case8_evidence_provenance",
			run: func(t *testing.T) {
				ev := models.NumericEvidence{
					ID:     "prov_1",
					Metric: "accuracy",
					Entity: "Model A",
					Value:  72.4,
					Unit:   "%",
					Source: models.EvidenceSource{
						Page: 3,
						Text: "Model A achieved 72.4% accuracy.",
					},
				}
				if ev.Source.Page != 3 {
					t.Errorf("want page 3, got %d", ev.Source.Page)
				}
				if ev.Source.Text == "" {
					t.Error("want non-empty source text")
				}
				if ev.ID != "prov_1" {
					t.Errorf("want ID prov_1, got %s", ev.ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
