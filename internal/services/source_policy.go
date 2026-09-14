package services

// SourcePolicy defines per-source-type behavior for chart processing.
// guards in the pipeline use this to skip PDF-only operations for non-PDF sources.
type SourcePolicy struct {
	AllowImageCharts   bool // ReVisualizeCharts from extracted PDF images
	AllowImageFallback bool // image annotation fallback path
	RequirePDFBytes    bool // caller must supply PDFBytes for image ops
}

// defaultPolicy is the base config for text-evidence charts (all sources).
var defaultPolicy = SourcePolicy{
	AllowImageCharts:   false,
	AllowImageFallback: false,
	RequirePDFBytes:    false,
}

// GetSourcePolicy returns the behavior policy for a given source type.
func GetSourcePolicy(sourceType string) SourcePolicy {
	switch sourceType {
	case "pdf":
		return SourcePolicy{
			AllowImageCharts:   true,
			AllowImageFallback: true,
			RequirePDFBytes:    true,
		}
	case "doi", "url":
		// doi/url resolve to text upstream; if PDF bytes were obtained
		// during resolution, the caller sets SourceType="pdf" and
		// passes PDFBytes — otherwise policy falls through to pasted_text.
		return defaultPolicy
	default:
		// pasted_text, empty, or any future source type
		return defaultPolicy
	}
}

// ImageChartsAllowed reports whether this source type permits PDF image chart extraction.
func (p SourcePolicy) ImageChartsAllowed() bool {
	return p.AllowImageCharts
}

// ImageFallbackAllowed reports whether this source type permits image annotation fallback.
func (p SourcePolicy) ImageFallbackAllowed() bool {
	return p.AllowImageFallback
}
