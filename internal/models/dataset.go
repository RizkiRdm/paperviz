package models

// CandidateDataset is a chart-ready group of evidence points sharing the same metric and unit.
type CandidateDataset struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Metric string         `json:"metric"`
	Unit   string         `json:"unit"`
	Points []DatasetPoint `json:"points"`
}

// DatasetPoint is one labeled value with evidence provenance.
type DatasetPoint struct {
	Label      string  `json:"label"`
	Value      float64 `json:"value"`
	EvidenceID string  `json:"evidence_id"`
}

// Labels returns the label strings from all points in dataset order.
func (d CandidateDataset) Labels() []string {
	labels := make([]string, len(d.Points))
	for i, p := range d.Points {
		labels[i] = p.Label
	}
	return labels
}

// Values returns the numeric values from all points in dataset order.
func (d CandidateDataset) Values() []float64 {
	vals := make([]float64, len(d.Points))
	for i, p := range d.Points {
		vals[i] = p.Value
	}
	return vals
}
