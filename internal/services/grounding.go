package services

import (
	"math"

	"paperviz/internal/models"
)

// GroundingStatus represents the validation result.
type GroundingStatus string

const (
	GroundingVerified    GroundingStatus = "verified"
	GroundingPartial     GroundingStatus = "partial"
	GroundingUnsupported GroundingStatus = "unsupported"
)

// GroundingResult holds the validation outcome.
type GroundingResult struct {
	Status GroundingStatus
	Errors []string
}

// ValidateGrounding checks that chart spec data is grounded in evidence.
func ValidateGrounding(spec models.ChartSpec, datasets []models.CandidateDataset) GroundingResult {
	// Build evidence index from datasets for fast lookup.
	evidenceIndex := buildEvidenceIndex(datasets)
	result := GroundingResult{Status: GroundingVerified}

	// Rule 1: chart type must be supported.
	if !isValidChartType(spec.Type) {
		result.Errors = append(result.Errors, "unsupported chart type: "+spec.Type)
		result.Status = GroundingUnsupported
		return result
	}

	// Rule 10 + 2: validate structure per chart type and collect numeric values.
	var allValues []float64
	switch spec.Type {
	case "bar":
		allValues = validateBarGrounding(spec, datasets, evidenceIndex, &result)
	case "line":
		allValues = validateLineGrounding(spec, datasets, evidenceIndex, &result)
	case "scatter":
		allValues = validateScatterGrounding(spec, datasets, evidenceIndex, &result)
	case "pie":
		allValues = validatePieGrounding(spec, datasets, evidenceIndex, &result)
	}

	// Rule 3: all numeric values must be valid floats.
	for i, v := range allValues {
		if math.IsNaN(v) {
			result.Errors = append(result.Errors, "NaN value at index "+itoa(i))
			result.Status = GroundingUnsupported
		}
		if math.IsInf(v, 0) {
			result.Errors = append(result.Errors, "Inf value at index "+itoa(i))
			result.Status = GroundingUnsupported
		}
	}

	// Downgrade to unsupported if any errors accumulated.
	if len(result.Errors) > 0 && result.Status != GroundingUnsupported {
		result.Status = GroundingUnsupported
	}
	return result
}

// isValidChartType checks if the chart type is in the allowed set.
func isValidChartType(t string) bool {
	switch t {
	case "bar", "line", "scatter", "pie":
		return true
	}
	return false
}

// buildEvidenceIndex creates a map from evidence ID to DatasetPoint for quick lookup.
func buildEvidenceIndex(datasets []models.CandidateDataset) map[string]models.DatasetPoint {
	idx := make(map[string]models.DatasetPoint)
	for _, ds := range datasets {
		for _, pt := range ds.Points {
			if pt.EvidenceID != "" {
				idx[pt.EvidenceID] = pt
			}
		}
	}
	return idx
}

// validateBarGrounding validates bar chart structure and evidence, returning all numeric values.
func validateBarGrounding(spec models.ChartSpec, datasets []models.CandidateDataset, evidenceIndex map[string]models.DatasetPoint, result *GroundingResult) []float64 {
	data, ok := spec.Data.(*models.BarData)
	if !ok {
		result.Errors = append(result.Errors, "bar chart data must be *BarData")
		result.Status = GroundingUnsupported
		return nil
	}
	// Rule 2: required fields.
	if len(data.Categories) == 0 {
		result.Errors = append(result.Errors, "bar chart requires categories")
		result.Status = GroundingUnsupported
		return nil
	}
	if len(data.Series) == 0 {
		result.Errors = append(result.Errors, "bar chart requires series")
		result.Status = GroundingUnsupported
		return nil
	}
	// Rule 4: dimensions must match.
	var allValues []float64
	for _, s := range data.Series {
		if len(s.Values) != len(data.Categories) {
			result.Errors = append(result.Errors, "series "+s.Name+" length mismatch")
			result.Status = GroundingUnsupported
			return allValues
		}
		allValues = append(allValues, s.Values...)
	}
	validateEvidenceTraceability(datasets, evidenceIndex, data.Categories, data.Series, result)
	return allValues
}

// validateLineGrounding validates line chart structure and evidence, returning all numeric values.
func validateLineGrounding(spec models.ChartSpec, datasets []models.CandidateDataset, evidenceIndex map[string]models.DatasetPoint, result *GroundingResult) []float64 {
	data, ok := spec.Data.(*models.LineData)
	if !ok {
		result.Errors = append(result.Errors, "line chart data must be *LineData")
		result.Status = GroundingUnsupported
		return nil
	}
	if len(data.X) == 0 {
		result.Errors = append(result.Errors, "line chart requires x values")
		result.Status = GroundingUnsupported
		return nil
	}
	if len(data.Series) == 0 {
		result.Errors = append(result.Errors, "line chart requires series")
		result.Status = GroundingUnsupported
		return nil
	}
	var allValues []float64
	for _, s := range data.Series {
		if len(s.Values) != len(data.X) {
			result.Errors = append(result.Errors, "series "+s.Name+" length mismatch")
			result.Status = GroundingUnsupported
			return allValues
		}
		allValues = append(allValues, s.Values...)
	}
	validateEvidenceTraceability(datasets, evidenceIndex, data.X, data.Series, result)
	return allValues
}

// validateScatterGrounding validates scatter chart structure and evidence, returning all numeric values.
func validateScatterGrounding(spec models.ChartSpec, datasets []models.CandidateDataset, evidenceIndex map[string]models.DatasetPoint, result *GroundingResult) []float64 {
	data, ok := spec.Data.(*models.ScatterData)
	if !ok {
		result.Errors = append(result.Errors, "scatter chart data must be *ScatterData")
		result.Status = GroundingUnsupported
		return nil
	}
	if len(data.X) == 0 {
		result.Errors = append(result.Errors, "scatter chart requires data points")
		result.Status = GroundingUnsupported
		return nil
	}
	if len(data.X) != len(data.Y) {
		result.Errors = append(result.Errors, "scatter x/y length mismatch")
		result.Status = GroundingUnsupported
		return allFloats(data.X, data.Y)
	}
	return allFloats(data.X, data.Y)
}

// validatePieGrounding validates pie chart structure and evidence, returning all numeric values.
func validatePieGrounding(spec models.ChartSpec, datasets []models.CandidateDataset, evidenceIndex map[string]models.DatasetPoint, result *GroundingResult) []float64 {
	data, ok := spec.Data.(*models.PieData)
	if !ok {
		result.Errors = append(result.Errors, "pie chart data must be *PieData")
		result.Status = GroundingUnsupported
		return nil
	}
	if len(data.Labels) == 0 {
		result.Errors = append(result.Errors, "pie chart requires labels")
		result.Status = GroundingUnsupported
		return nil
	}
	if len(data.Labels) != len(data.Values) {
		result.Errors = append(result.Errors, "pie labels/values length mismatch")
		result.Status = GroundingUnsupported
		return data.Values
	}
	// Rule 8: pie values must be non-negative.
	for i, v := range data.Values {
		if v < 0 {
			result.Errors = append(result.Errors, "pie negative value at index "+itoa(i))
			result.Status = GroundingUnsupported
		}
	}
	return data.Values
}

// allFloats merges two float64 slices into one.
func allFloats(a, b []float64) []float64 {
	out := make([]float64, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

// validateEvidenceTraceability checks that every numeric value is traceable to dataset evidence.
func validateEvidenceTraceability(datasets []models.CandidateDataset, evidenceIndex map[string]models.DatasetPoint, categories []string, series []models.Series, result *GroundingResult) {
	// Rule 5: units must be consistent across datasets used.
	units := make(map[string]bool)
	for _, ds := range datasets {
		if ds.Unit != "" {
			units[ds.Unit] = true
		}
	}
	if len(units) > 1 {
		result.Errors = append(result.Errors, "inconsistent units across datasets")
		result.Status = GroundingUnsupported
		return
	}

	// Rule 6+7+8: every category must appear in a dataset point with evidence.
	catSet := make(map[string]bool, len(categories))
	for _, c := range categories {
		catSet[c] = true
	}
	foundEvidence := false
	for _, ds := range datasets {
		for _, pt := range ds.Points {
			if catSet[pt.Label] {
				if pt.EvidenceID == "" {
					result.Errors = append(result.Errors, "missing evidence for "+pt.Label)
					result.Status = GroundingUnsupported
					return
				}
				if _, ok := evidenceIndex[pt.EvidenceID]; !ok {
					result.Errors = append(result.Errors, "evidence ID "+pt.EvidenceID+" not found in datasets")
					result.Status = GroundingUnsupported
					return
				}
				foundEvidence = true
			}
		}
	}
	if !foundEvidence && len(categories) > 0 {
		result.Errors = append(result.Errors, "no evidence traceable to chart categories")
		result.Status = GroundingPartial
	}
}

// itoa converts an int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
