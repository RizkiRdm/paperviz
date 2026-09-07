package models

import (
	"errors"
	"fmt"
)

// ChartProvenance traces chart values back to paper source.
type ChartProvenance struct {
	PaperID         string   `json:"paper_id,omitempty"`
	PageNumbers     []int    `json:"page_numbers,omitempty"`
	EvidenceIDs     []string `json:"evidence_ids,omitempty"`
	SourceMethod    string   `json:"source_method"` // text_evidence | table_extraction | image_extraction
	DatasetID       string   `json:"dataset_id,omitempty"`
	GroundingStatus string   `json:"grounding_status"` // verified | partial | unsupported
}

// ChartSpec is the validated chart specification ready for rendering.
type ChartSpec struct {
	Type       string          `json:"type"` // bar, line, scatter, pie
	Title      string          `json:"title"`
	Data       interface{}     `json:"data"` // one of BarData, LineData, ScatterData, PieData
	XLabel     string          `json:"x_label,omitempty"`
	YLabel     string          `json:"y_label,omitempty"`
	Takeaway   string          `json:"takeaway,omitempty"`
	Provenance ChartProvenance `json:"provenance"`
}

// BarData is for categorical comparisons.
type BarData struct {
	Categories []string `json:"categories"`
	Series     []Series `json:"series"`
}

// LineData is for time-series or sequential data.
type LineData struct {
	X      []string `json:"x"`
	Series []Series `json:"series"`
}

// ScatterData requires numeric X and Y.
type ScatterData struct {
	X []float64 `json:"x"`
	Y []float64 `json:"y"`
}

// PieData represents parts of a whole.
type PieData struct {
	Labels []string  `json:"labels"`
	Values []float64 `json:"values"`
}

// Series is a named set of values for bar/line charts.
type Series struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
}

// validTypes lists allowed chart types.
var validTypes = map[string]bool{
	"bar":     true,
	"line":    true,
	"scatter": true,
	"pie":     true,
}

// Validate checks the ChartSpec for correctness based on its type.
func (s ChartSpec) Validate() error {
	if s.Title == "" {
		return errors.New("title is required")
	}
	if !validTypes[s.Type] {
		return fmt.Errorf("invalid chart type: %s", s.Type)
	}
	switch s.Type {
	case "bar":
		return s.validateBar()
	case "line":
		return s.validateLine()
	case "scatter":
		return s.validateScatter()
	case "pie":
		return s.validatePie()
	}
	return nil
}

// validateBar checks bar chart data consistency.
func (s ChartSpec) validateBar() error {
	data, ok := s.Data.(*BarData)
	if !ok {
		return errors.New("data must be *BarData for bar chart")
	}
	if len(data.Categories) == 0 {
		return errors.New("bar chart requires at least one category")
	}
	for i, series := range data.Series {
		if len(series.Values) != len(data.Categories) {
			return fmt.Errorf("bar series %d length %d != categories length %d", i, len(series.Values), len(data.Categories))
		}
	}
	return nil
}

// validateLine checks line chart data consistency.
func (s ChartSpec) validateLine() error {
	data, ok := s.Data.(*LineData)
	if !ok {
		return errors.New("data must be *LineData for line chart")
	}
	if len(data.X) == 0 {
		return errors.New("line chart requires at least one x value")
	}
	for i, series := range data.Series {
		if len(series.Values) != len(data.X) {
			return fmt.Errorf("line series %d length %d != x length %d", i, len(series.Values), len(data.X))
		}
	}
	return nil
}

// validateScatter checks scatter chart data consistency.
func (s ChartSpec) validateScatter() error {
	data, ok := s.Data.(*ScatterData)
	if !ok {
		return errors.New("data must be *ScatterData for scatter chart")
	}
	if len(data.X) == 0 {
		return errors.New("scatter chart requires at least one data point")
	}
	if len(data.X) != len(data.Y) {
		return fmt.Errorf("scatter x length %d != y length %d", len(data.X), len(data.Y))
	}
	return nil
}

// validatePie checks pie chart data consistency.
func (s ChartSpec) validatePie() error {
	data, ok := s.Data.(*PieData)
	if !ok {
		return errors.New("data must be *PieData for pie chart")
	}
	if len(data.Labels) == 0 {
		return errors.New("pie chart requires at least one label")
	}
	if len(data.Labels) != len(data.Values) {
		return fmt.Errorf("pie labels length %d != values length %d", len(data.Labels), len(data.Values))
	}
	for i, v := range data.Values {
		if v < 0 {
			return fmt.Errorf("pie value at index %d is negative: %f", i, v)
		}
	}
	return nil
}
