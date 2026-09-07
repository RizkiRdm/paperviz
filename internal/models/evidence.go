package models

import (
	"errors"
	"fmt"
)

// NumericEvidence represents a numeric claim extracted from a paper with full provenance.
type NumericEvidence struct {
	ID         string         `json:"id"`
	Metric     string         `json:"metric"` // e.g. "accuracy", "latency", "F1"
	Entity     string         `json:"entity"` // e.g. "Model A", "Baseline"
	Value      float64        `json:"value"`
	Unit       string         `json:"unit"` // e.g. "%", "ms", "seconds"
	Source     EvidenceSource `json:"source"`
	Context    string         `json:"context,omitempty"`
	Group      string         `json:"group,omitempty"`
	Experiment string         `json:"experiment,omitempty"`
	Time       string         `json:"time,omitempty"`
	Confidence string         `json:"confidence,omitempty"`
}

// EvidenceSource points back to where the evidence came from in the paper.
type EvidenceSource struct {
	Page int    `json:"page"`
	Text string `json:"text"` // verbatim source text
}

// Validate checks that all required fields are present and valid.
func (e NumericEvidence) Validate() error {
	// Check required fields
	if e.Metric == "" {
		return errors.New("metric is required")
	}
	if e.Entity == "" {
		return errors.New("entity is required")
	}
	if e.Source.Text == "" {
		return errors.New("source text is required")
	}
	if e.Source.Page <= 0 {
		return fmt.Errorf("source page must be > 0, got %d", e.Source.Page)
	}
	return nil
}
