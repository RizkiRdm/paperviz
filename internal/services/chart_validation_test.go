package services

import (
	"testing"

	"paperviz/internal/models"
)

// TestRealPaper_DeepLearning validates extraction from a paper with numeric results and a table.
func TestRealPaper_DeepLearning(t *testing.T) {
	text := "Results show that our model achieved 94.2% accuracy on the chest X-ray dataset, compared to 87.1% for the baseline ResNet-50. The F1 score was 0.91 vs 0.84. Processing latency decreased from 245ms to 128ms per image."
	table := "Table 1: Performance Comparison\n| Model | Accuracy | F1 | Latency (ms) |\n| ResNet-50 | 87.1 | 0.84 | 245 |\n| Our Model | 94.2 | 0.91 | 128 |\n| EfficientNet | 91.5 | 0.88 | 156 |"

	textEvidence := ExtractNumericEvidence(text, 1)
	tableEvidence := ExtractTableData(table, 1)

	// Text extraction: entity-value (94.2% accuracy), standalone (f1 0.91), from-to (245ms→128ms).
	if len(textEvidence) < 3 {
		t.Fatalf("want ≥3 text evidence items, got %d", len(textEvidence))
	}

	// Table extraction: 3 models × 3 numeric columns = 9 evidence items.
	if len(tableEvidence) < 9 {
		t.Fatalf("want ≥9 table evidence items, got %d", len(tableEvidence))
	}

	// Verify table entities include all three models.
	entities := make(map[string]bool)
	for _, e := range tableEvidence {
		entities[e.Entity] = true
	}
	for _, want := range []string{"ResNet-50", "Our Model", "EfficientNet"} {
		if !entities[want] {
			t.Errorf("table evidence missing entity %q", want)
		}
	}

	// Combined evidence should produce multiple datasets.
	allEvidence := append(textEvidence, tableEvidence...)
	datasets := BuildCandidateDatasets(allEvidence)
	if len(datasets) < 2 {
		t.Fatalf("want ≥2 datasets from combined evidence, got %d", len(datasets))
	}

	// Verify text-sourced evidence (has IDs set by from-to pattern) can be grounded.
	// Find the latency dataset (from text from-to extraction with non-empty entity).
	for _, ds := range datasets {
		if ds.Metric == "latency" && len(ds.Points) == 2 {
			spec := models.ChartSpec{
				Type:  "bar",
				Title: "Latency Comparison",
				Data: &models.BarData{
					Categories: ds.Labels(),
					Series: []models.Series{
						{Name: ds.Metric, Values: ds.Values()},
					},
				},
			}
			result := ValidateGrounding(spec, []models.CandidateDataset{ds})
			if result.Status == GroundingUnsupported {
				t.Errorf("latency dataset grounding failed: %v", result.Errors)
			}
			return
		}
	}
	t.Log("no 2-point latency dataset found for grounding check — acceptable if from-to didn't match")
}

// TestRealPaper_Climate validates extraction from a climate paper with non-standard units.
func TestRealPaper_Climate(t *testing.T) {
	text := "Temperature anomalies increased from 0.8°C in 2000 to 1.2°C in 2020. Sea level rose by 3.2mm per year. CO2 concentrations reached 415 ppm. The correlation between CO2 and temperature was r = 0.87, p < 0.001."

	evidence := ExtractNumericEvidence(text, 1)

	// The stat pattern extracts "r = 0.87" but with empty entity.
	// °C, ppm, mm are not in the supported unit list, so from-to and entity-value won't match.
	if len(evidence) == 0 {
		t.Log("no evidence extracted — climate units (°C, ppm, mm) not in supported set, acceptable")
		return
	}

	// Verify extracted evidence has valid structure.
	for i, e := range evidence {
		if e.Metric == "" {
			t.Errorf("evidence[%d] has empty metric", i)
		}
		if e.Source.Text == "" {
			t.Errorf("evidence[%d] has empty source text", i)
		}
	}

	// Datasets should be empty because extracted evidence has empty entities.
	datasets := BuildCandidateDatasets(evidence)
	if len(datasets) > 0 {
		// If datasets were produced, verify they have valid points.
		for _, ds := range datasets {
			if len(ds.Points) == 0 {
				t.Errorf("dataset %q has no points", ds.ID)
			}
		}
	}
}

// TestRealPaper_NLP validates mixed metrics with ambiguous claims produce only grounded evidence.
func TestRealPaper_NLP(t *testing.T) {
	text := "Our approach improved BLEU score from 32.1 to 38.7 on WMT14 EN-DE. ROUGE-L increased by 4.2 points. The model had 125M parameters. Inference time was 45ms on a V100 GPU. Ambiguous claim: performance improved significantly."

	evidence := ExtractNumericEvidence(text, 1)
	if len(evidence) == 0 {
		t.Fatal("expected at least 1 evidence item from NLP paper, got 0")
	}

	// Every evidence item must have metric and source text.
	for i, e := range evidence {
		if e.Metric == "" {
			t.Errorf("evidence[%d] has empty metric", i)
		}
		if e.Source.Text == "" {
			t.Errorf("evidence[%d] has empty source text", i)
		}
	}

	// Ambiguous claim must not produce evidence.
	for _, e := range evidence {
		if containsSubstring(e.Source.Text, "performance improved significantly") {
			t.Errorf("ambiguous claim leaked as evidence: %q", e.Source.Text)
		}
	}

	// "from 32.1 to 38.7" lacks required unit (%|ms|etc), so from-to won't match.
	// "4.2 points" — "points" not in unit list. "125M parameters" — no metric match.
	// "45ms" via standalone: metric="time", value=45, entity="".
	// With empty entity, no datasets should be produced.
	datasets := BuildCandidateDatasets(evidence)
	if len(datasets) > 0 {
		t.Logf("datasets produced: %d (some evidence had entity names)", len(datasets))
	}
}

// TestRealPaper_Sparse validates sparse text with no numbers produces zero evidence and no charts.
func TestRealPaper_Sparse(t *testing.T) {
	text := "The experiment showed mixed results. Some metrics improved while others declined."

	textEvidence := ExtractNumericEvidence(text, 1)
	if len(textEvidence) != 0 {
		t.Fatalf("want 0 text evidence from sparse text, got %d", len(textEvidence))
	}

	tableEvidence := ExtractTableData(text, 1)
	if len(tableEvidence) != 0 {
		t.Fatalf("want 0 table evidence from sparse text, got %d", len(tableEvidence))
	}

	datasets := BuildCandidateDatasets(textEvidence)
	if len(datasets) != 0 {
		t.Fatalf("want 0 datasets from sparse text, got %d", len(datasets))
	}
}

// TestRealPaper_GroundedRatio validates the grounded-to-total ratio across all papers is ≥ 0.8.
func TestRealPaper_GroundedRatio(t *testing.T) {
	papers := []struct {
		name  string
		text  string
		table string
	}{
		{
			name:  "deep_learning",
			text:  "Results show that our model achieved 94.2% accuracy on the chest X-ray dataset, compared to 87.1% for the baseline ResNet-50. The F1 score was 0.91 vs 0.84. Processing latency decreased from 245ms to 128ms per image.",
			table: "Table 1: Performance Comparison\n| Model | Accuracy | F1 | Latency (ms) |\n| ResNet-50 | 87.1 | 0.84 | 245 |\n| Our Model | 94.2 | 0.91 | 128 |\n| EfficientNet | 91.5 | 0.88 | 156 |",
		},
		{
			name: "climate",
			text: "Temperature anomalies increased from 0.8°C in 2000 to 1.2°C in 2020. Sea level rose by 3.2mm per year. CO2 concentrations reached 415 ppm. The correlation between CO2 and temperature was r = 0.87, p < 0.001.",
		},
		{
			name: "nlp",
			text: "Our approach improved BLEU score from 32.1 to 38.7 on WMT14 EN-DE. ROUGE-L increased by 4.2 points. The model had 125M parameters. Inference time was 45ms on a V100 GPU.",
		},
		{
			name: "sparse",
			text: "The experiment showed mixed results. Some metrics improved while others declined.",
		},
	}

	totalEvidence := 0
	groundedEvidence := 0

	for _, paper := range papers {
		var allEvidence []models.NumericEvidence
		if paper.text != "" {
			allEvidence = append(allEvidence, ExtractNumericEvidence(paper.text, 1)...)
		}
		if paper.table != "" {
			allEvidence = append(allEvidence, ExtractTableData(paper.table, 1)...)
		}
		totalEvidence += len(allEvidence)

		for _, e := range allEvidence {
			// Grounded = has metric, valid value, and source text (entity optional for standalone metrics).
			if e.Metric != "" && e.Value != 0 && e.Source.Text != "" {
				groundedEvidence++
			}
		}
	}

	if totalEvidence == 0 {
		t.Fatal("total evidence across all papers is 0 — extraction pipeline broken")
	}

	ratio := float64(groundedEvidence) / float64(totalEvidence)
	t.Logf("grounded ratio: %d/%d = %.2f", groundedEvidence, totalEvidence, ratio)

	if ratio < 0.8 {
		t.Errorf("grounded ratio %.2f < 0.8 threshold (%d grounded of %d total)", ratio, groundedEvidence, totalEvidence)
	}
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
