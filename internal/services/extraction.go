package services

import (
	"errors"
	"fmt"

	"paperviz/internal/external"
)

// ExtractFromPDF extracts body text and candidate chart images from PDF
// bytes. Pure function of (input) -> (output, error), per ARCHITECTURE.md
// Section 5 Module Rules — no side effects, no persistence.
func ExtractFromPDF(pdfBytes []byte) (ExtractResult, error) {
	text, err := external.ExtractText(pdfBytes)
	if errors.Is(err, external.ErrNoTextLayer) {
		return ExtractResult{}, ErrNoTextLayer
	}
	if err != nil {
		return ExtractResult{}, fmt.Errorf("extract text: %w", err)
	}

	images, err := external.ExtractImages(pdfBytes)
	if err != nil {
		// Image extraction failure is not fatal to the pipeline — the
		// document still has valid text. Charts will simply be unavailable.
		images = nil
	}

	// Per-page text lets the chart service give Gemini the right local
	// context for each chart image (see external.ExtractTextByPage). If
	// this fails, we still have the flattened body text above, so the
	// document as a whole is not blocked — only chart annotation quality
	// degrades (falls back to whole-document context in the pipeline).
	pages, err := external.ExtractTextByPage(pdfBytes)
	if err != nil {
		pages = nil
	}

	charts := make([]ExtractedChart, 0, len(images))
	for _, img := range images {
		charts = append(charts, ExtractedChart{
			PageNumber: img.PageNumber,
			ImageBytes: img.Bytes,
		})
	}

	return ExtractResult{Text: text, Pages: pages, Charts: charts}, nil
}

// PDFDocument is a deep domain object encapsulating extracted PDF text,
// per-page text mapping, and bounded chart images.
type PDFDocument struct {
	Text   string
	Pages  map[int]string
	Charts []ExtractedChart
}

// ParsePDF parses PDF bytes, extracts text, pages, and bounded chart images.
func ParsePDF(pdfBytes []byte, maxCharts int) (*PDFDocument, error) {
	res, err := ExtractFromPDF(pdfBytes)
	if err != nil {
		return nil, err
	}
	charts := res.Charts
	if len(charts) > maxCharts {
		charts = charts[:maxCharts]
	}
	return &PDFDocument{
		Text:   res.Text,
		Pages:  res.Pages,
		Charts: charts,
	}, nil
}
