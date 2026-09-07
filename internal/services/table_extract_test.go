package services

import (
	"testing"

	"paperviz/internal/models"
)

func TestExtractTableData(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		page    int
		wantLen int
		check   func(t *testing.T, items []models.NumericEvidence)
	}{
		{
			name: "pipe table 3x3",
			text: "| Model   | Accuracy | F1    |\n" +
				"|---------|----------|-------|\n" +
				"| Model A | 85.3     | 0.82  |\n" +
				"| Model B | 91.2     | 0.89  |\n" +
				"| Model C | 78.5     | 0.76  |",
			page:    1,
			wantLen: 6,
			check: func(t *testing.T, items []models.NumericEvidence) {
				// 3 rows × 2 numeric columns = 6 items
				if len(items) != 6 {
					t.Fatalf("got %d items, want 6", len(items))
				}
				// Verify first item
				if items[0].Metric != "accuracy" {
					t.Errorf("items[0].Metric = %q, want %q", items[0].Metric, "accuracy")
				}
				if items[0].Entity != "Model A" {
					t.Errorf("items[0].Entity = %q, want %q", items[0].Entity, "Model A")
				}
				if items[0].Value != 85.3 {
					t.Errorf("items[0].Value = %v, want 85.3", items[0].Value)
				}
				// Verify F1 item
				if items[1].Metric != "f1" {
					t.Errorf("items[1].Metric = %q, want %q", items[1].Metric, "f1")
				}
				if items[1].Value != 0.82 {
					t.Errorf("items[1].Value = %v, want 0.82", items[1].Value)
				}
			},
		},
		{
			name: "tab-separated table",
			text: "Model\tAccuracy\tF1\n" +
				"Model A\t85.3\t0.82\n" +
				"Model B\t91.2\t0.89",
			page:    2,
			wantLen: 4,
			check: func(t *testing.T, items []models.NumericEvidence) {
				if items[0].Entity != "Model A" {
					t.Errorf("items[0].Entity = %q, want %q", items[0].Entity, "Model A")
				}
				if items[0].Value != 85.3 {
					t.Errorf("items[0].Value = %v, want 85.3", items[0].Value)
				}
			},
		},
		{
			name:    "empty input",
			text:    "",
			page:    3,
			wantLen: 0,
		},
		{
			name: "table with missing cells",
			text: "| Model   | Accuracy | F1    |\n" +
				"|---------|----------|-------|\n" +
				"| Model A | 85.3     |       |\n" +
				"| Model B |          | 0.89  |",
			page:    4,
			wantLen: 2,
			check: func(t *testing.T, items []models.NumericEvidence) {
				// Missing cells skipped: Model A has only accuracy, Model B has only F1
				if items[0].Entity != "Model A" || items[0].Metric != "accuracy" {
					t.Errorf("items[0] = %+v, want Model A accuracy", items[0])
				}
				if items[1].Entity != "Model B" || items[1].Metric != "f1" {
					t.Errorf("items[1] = %+v, want Model B f1", items[1])
				}
			},
		},
		{
			name: "units in header",
			text: "| Model   | Accuracy (%) | Latency (ms) |\n" +
				"|---------|--------------|--------------|\n" +
				"| Model A | 85.3         | 120          |",
			page:    5,
			wantLen: 2,
			check: func(t *testing.T, items []models.NumericEvidence) {
				if items[0].Unit != "%" {
					t.Errorf("items[0].Unit = %q, want %%", items[0].Unit)
				}
				if items[1].Unit != "ms" {
					t.Errorf("items[1].Unit = %q, want ms", items[1].Unit)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := ExtractTableData(tt.text, tt.page)
			if len(items) != tt.wantLen {
				t.Fatalf("got %d items, want %d; items: %+v", len(items), tt.wantLen, items)
			}
			if tt.check != nil {
				tt.check(t, items)
			}
			// Verify source fields on all items
			for i, item := range items {
				if item.Source.Page != tt.page {
					t.Errorf("items[%d].Source.Page = %d, want %d", i, item.Source.Page, tt.page)
				}
				if item.Source.Text == "" {
					t.Errorf("items[%d].Source.Text is empty", i)
				}
			}
		})
	}
}

// TestIsTableRow verifies table row detection.
func TestIsTableRow(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"| a | b | c |", true},
		{"a\tb\tc", true},
		{"plain text", false},
		{"", false},
	}
	for _, tt := range tests {
		got := isTableRow(tt.input)
		if got != tt.want {
			t.Errorf("isTableRow(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// TestParseHeader verifies header parsing with units.
func TestParseHeader(t *testing.T) {
	tests := []struct {
		input      string
		wantMetric string
		wantUnit   string
	}{
		{"Accuracy", "accuracy", ""},
		{"Accuracy (%)", "accuracy", "%"},
		{"Latency (ms)", "latency", "ms"},
		{"F1 Score", "f1 score", ""},
	}
	for _, tt := range tests {
		metric, unit := parseHeader(tt.input)
		if metric != tt.wantMetric {
			t.Errorf("parseHeader(%q).metric = %q, want %q", tt.input, metric, tt.wantMetric)
		}
		if unit != tt.wantUnit {
			t.Errorf("parseHeader(%q).unit = %q, want %q", tt.input, unit, tt.wantUnit)
		}
	}
}
