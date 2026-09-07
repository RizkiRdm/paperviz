package models

import (
	"testing"
)

// TestChartSpecValidate runs table-driven validation tests for all chart types.
func TestChartSpecValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    ChartSpec
		wantErr bool
	}{
		{
			name: "valid bar spec",
			spec: ChartSpec{
				Type:  "bar",
				Title: "Model Accuracy",
				Data: &BarData{
					Categories: []string{"Model A", "Model B", "Model C"},
					Series: []Series{
						{Name: "Accuracy", Values: []float64{0.92, 0.87, 0.95}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid line spec",
			spec: ChartSpec{
				Type:  "line",
				Title: "Training Loss Over Time",
				Data: &LineData{
					X: []string{"Epoch 1", "Epoch 2", "Epoch 3"},
					Series: []Series{
						{Name: "Loss", Values: []float64{1.5, 0.8, 0.3}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid scatter spec",
			spec: ChartSpec{
				Type:  "scatter",
				Title: "Latency vs Accuracy",
				Data: &ScatterData{
					X: []float64{10.5, 20.3, 30.1},
					Y: []float64{0.92, 0.87, 0.95},
				},
			},
			wantErr: false,
		},
		{
			name: "valid pie spec",
			spec: ChartSpec{
				Type:  "pie",
				Title: "Error Distribution",
				Data: &PieData{
					Labels: []string{"Timeout", "Auth", "Network"},
					Values: []float64{45, 30, 25},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid chart type",
			spec: ChartSpec{
				Type:  "radar",
				Title: "Radar Chart",
				Data:  &BarData{},
			},
			wantErr: true,
		},
		{
			name: "bar series length mismatch",
			spec: ChartSpec{
				Type:  "bar",
				Title: "Mismatched Bar",
				Data: &BarData{
					Categories: []string{"A", "B"},
					Series: []Series{
						{Name: "X", Values: []float64{1, 2, 3}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "line series length mismatch",
			spec: ChartSpec{
				Type:  "line",
				Title: "Mismatched Line",
				Data: &LineData{
					X: []string{"1", "2"},
					Series: []Series{
						{Name: "Y", Values: []float64{1, 2, 3}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "scatter x y length mismatch",
			spec: ChartSpec{
				Type:  "scatter",
				Title: "Mismatched Scatter",
				Data: &ScatterData{
					X: []float64{1, 2, 3},
					Y: []float64{1, 2},
				},
			},
			wantErr: true,
		},
		{
			name: "scatter with wrong data type",
			spec: ChartSpec{
				Type:  "scatter",
				Title: "Wrong Type Scatter",
				Data: &BarData{
					Categories: []string{"A"},
					Series:     []Series{{Name: "X", Values: []float64{1}}},
				},
			},
			wantErr: true,
		},
		{
			name: "pie labels values length mismatch",
			spec: ChartSpec{
				Type:  "pie",
				Title: "Mismatched Pie",
				Data: &PieData{
					Labels: []string{"A", "B"},
					Values: []float64{1},
				},
			},
			wantErr: true,
		},
		{
			name: "pie negative value",
			spec: ChartSpec{
				Type:  "pie",
				Title: "Negative Pie",
				Data: &PieData{
					Labels: []string{"A"},
					Values: []float64{-5},
				},
			},
			wantErr: true,
		},
		{
			name: "empty title",
			spec: ChartSpec{
				Type: "bar",
				Data: &BarData{},
			},
			wantErr: true,
		},
		{
			name: "empty bar categories",
			spec: ChartSpec{
				Type:  "bar",
				Title: "Empty Bar",
				Data:  &BarData{},
			},
			wantErr: true,
		},
		{
			name: "empty scatter",
			spec: ChartSpec{
				Type:  "scatter",
				Title: "Empty Scatter",
				Data:  &ScatterData{},
			},
			wantErr: true,
		},
		{
			name: "empty pie labels",
			spec: ChartSpec{
				Type:  "pie",
				Title: "Empty Pie",
				Data:  &PieData{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
