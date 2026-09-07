package services

import (
	"testing"

	"paperviz/internal/models"
)

func TestExtractNumericEvidence(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		page     int
		wantLen  int
		wantVals []struct {
			metric string
			entity string
			value  float64
			unit   string
		}
	}{
		{
			name:    "direct number",
			text:    "Model A achieved 72.4% accuracy on the benchmark.",
			page:    1,
			wantLen: 1,
			wantVals: []struct {
				metric string
				entity string
				value  float64
				unit   string
			}{
				{"accuracy", "Model A", 72.4, "%"},
			},
		},
		{
			name:    "two entities",
			text:    "Model A = 72.4% and Model B = 81.7% accuracy.",
			page:    2,
			wantLen: 2,
			wantVals: []struct {
				metric string
				entity string
				value  float64
				unit   string
			}{
				{"accuracy", "Model A", 72.4, "%"},
				{"accuracy", "Model B", 81.7, "%"},
			},
		},
		{
			name:    "before/after from-to",
			text:    "Accuracy improved from 72.4% to 81.7% after fine-tuning.",
			page:    3,
			wantLen: 2,
			wantVals: []struct {
				metric string
				entity string
				value  float64
				unit   string
			}{
				{"accuracy", "Accuracy improved", 72.4, "%"},
				{"accuracy", "Accuracy improved", 81.7, "%"},
			},
		},
		{
			name:    "ambiguous by-pct skipped",
			text:    "The model improved by 12% over baseline.",
			page:    4,
			wantLen: 0,
		},
		{
			name:    "no numbers",
			text:    "The model performed well on the benchmark task.",
			page:    5,
			wantLen: 0,
		},
		{
			name:    "multiple metrics",
			text:    "The system achieved 95% accuracy and F1 score of 0.89.",
			page:    6,
			wantLen: 2,
			wantVals: []struct {
				metric string
				entity string
				value  float64
				unit   string
			}{
				{"accuracy", "The system", 95, "%"},
				{"f1", "", 0.89, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractNumericEvidence(tt.text, tt.page)
			if len(got) != tt.wantLen {
				t.Fatalf("got %d items, want %d; items: %+v", len(got), tt.wantLen, got)
			}
			for i, w := range tt.wantVals {
				if got[i].Metric != w.metric {
					t.Errorf("item[%d].Metric = %q, want %q", i, got[i].Metric, w.metric)
				}
				if got[i].Entity != w.entity {
					t.Errorf("item[%d].Entity = %q, want %q", i, got[i].Entity, w.entity)
				}
				if got[i].Value != w.value {
					t.Errorf("item[%d].Value = %v, want %v", i, got[i].Value, w.value)
				}
				if got[i].Unit != w.unit {
					t.Errorf("item[%d].Unit = %q, want %q", i, got[i].Unit, w.unit)
				}
				if got[i].Source.Page != tt.page {
					t.Errorf("item[%d].Source.Page = %d, want %d", i, got[i].Source.Page, tt.page)
				}
				if got[i].Source.Text == "" {
					t.Errorf("item[%d].Source.Text is empty", i)
				}
			}
		})
	}
}

// TestExtractNumericEvidenceSourceText verifies every evidence item has source text.
func TestExtractNumericEvidenceSourceText(t *testing.T) {
	text := "We report beta = 0.42, 95% CI [0.31, 0.53] for the treatment effect."
	got := ExtractNumericEvidence(text, 10)
	if len(got) == 0 {
		t.Fatal("expected at least 1 evidence item")
	}
	for i, e := range got {
		if e.Source.Text == "" {
			t.Errorf("evidence[%d]: missing source text", i)
		}
		if e.Source.Page != 10 {
			t.Errorf("evidence[%d]: source page = %d, want 10", i, e.Source.Page)
		}
	}
}

// TestExtractNumericEvidenceValidate checks that extracted items have sane fields.
func TestExtractNumericEvidenceValidate(t *testing.T) {
	text := "Model X achieved 85.3% accuracy."
	got := ExtractNumericEvidence(text, 1)
	if len(got) == 0 {
		t.Fatal("expected at least 1 evidence item")
	}
	for i, e := range got {
		if e.Metric == "" {
			t.Errorf("evidence[%d]: empty metric", i)
		}
		if e.Source.Text == "" {
			t.Errorf("evidence[%d]: empty source text", i)
		}
		if e.Source.Page != 1 {
			t.Errorf("evidence[%d]: wrong page %d", i, e.Source.Page)
		}
	}
}

// TestSplitSentences verifies sentence splitting.
func TestSplitSentences(t *testing.T) {
	text := "First sentence. Second sentence. Third one!"
	got := splitSentences(text)
	if len(got) < 3 {
		t.Errorf("expected >= 3 sentences, got %d: %v", len(got), got)
	}
}

// TestNormalizeNumber verifies comma-separated number parsing.
func TestNormalizeNumber(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"1,234.5", 1234.5},
		{"72.4", 72.4},
		{"0.89", 0.89},
		{"1234", 1234},
	}
	for _, tt := range tests {
		got := normalizeNumber(tt.input)
		if got != tt.want {
			t.Errorf("normalizeNumber(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// Compile-time check that NumericEvidence is usable.
var _ models.NumericEvidence
