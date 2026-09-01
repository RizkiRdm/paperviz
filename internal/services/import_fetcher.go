package services

import (
	"fmt"
	"strings"

	"paperviz/internal/external"
)

// paperFetcher implements the ImportService interface using existing service functions.
type paperFetcher struct{}

// NewPaperFetcher returns a concrete ImportService backed by FetchByDOI/FetchByURL.
func NewPaperFetcher() *paperFetcher { return &paperFetcher{} }

// FetchByDOI resolves a DOI to paper text and title via CrossRef + Unpaywall.
func (f *paperFetcher) FetchByDOI(doi string) (string, string, error) {
	pdfBytes, err := FetchByDOI(doi)
	if err != nil {
		return "", "", fmt.Errorf("fetch by DOI %s: %w", doi, err)
	}
	text, err := external.ExtractText(pdfBytes)
	if err != nil {
		return "", "", fmt.Errorf("extract text for DOI %s: %w", doi, err)
	}
	title := extractTitle(text, doi)
	return text, title, nil
}

// FetchByURL downloads a PDF from a URL and extracts its text.
func (f *paperFetcher) FetchByURL(url string) (string, string, error) {
	pdfBytes, err := FetchByURL(url)
	if err != nil {
		return "", "", fmt.Errorf("fetch by URL %s: %w", url, err)
	}
	text, err := external.ExtractText(pdfBytes)
	if err != nil {
		return "", "", fmt.Errorf("extract text for URL %s: %w", url, err)
	}
	title := extractTitle(text, url)
	return text, title, nil
}

// extractTitle returns the first meaningful line of extracted text, or a fallback.
func extractTitle(text, fallback string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 10 {
			return line
		}
	}
	return fallback
}
