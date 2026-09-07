package models

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		evidence NumericEvidence
		wantErr  bool
	}{
		{
			name: "valid evidence with all fields",
			evidence: NumericEvidence{
				ID:      "ev1",
				Metric:  "accuracy",
				Entity:  "Model A",
				Value:   95.5,
				Unit:    "%",
				Source:  EvidenceSource{Page: 1, Text: "Model A achieves 95.5% accuracy."},
				Context: "Results section",
			},
			wantErr: false,
		},
		{
			name: "missing metric",
			evidence: NumericEvidence{
				ID:     "ev2",
				Entity: "Model A",
				Value:  95.5,
				Source: EvidenceSource{Page: 1, Text: "Model A achieves 95.5% accuracy."},
			},
			wantErr: true,
		},
		{
			name: "missing entity",
			evidence: NumericEvidence{
				ID:     "ev3",
				Metric: "accuracy",
				Value:  95.5,
				Source: EvidenceSource{Page: 1, Text: "Model A achieves 95.5% accuracy."},
			},
			wantErr: true,
		},
		{
			name: "missing source text",
			evidence: NumericEvidence{
				ID:     "ev4",
				Metric: "accuracy",
				Entity: "Model A",
				Value:  95.5,
				Source: EvidenceSource{Page: 1},
			},
			wantErr: true,
		},
		{
			name: "zero page number",
			evidence: NumericEvidence{
				ID:     "ev5",
				Metric: "accuracy",
				Entity: "Model A",
				Value:  95.5,
				Source: EvidenceSource{Page: 0, Text: "Model A achieves 95.5% accuracy."},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.evidence.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
