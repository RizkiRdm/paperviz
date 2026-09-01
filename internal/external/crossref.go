package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ErrDOINotFound means the DOI does not resolve in CrossRef's registry.
var ErrDOINotFound = errors.New("DOI not found in CrossRef")

// ErrNetworkError wraps transient HTTP failures (timeout, connection reset).
var ErrNetworkError = errors.New("network error during DOI lookup")

// doiRegexp validates DOI format per ISO 3297: 10.NNNN/suffix.
var doiRegexp = regexp.MustCompile(`^10.\d{4,9}/`)

// DOIInfo holds structured metadata returned by the CrossRef API.
type DOIInfo struct {
	Title    string `json:"title"`
	Authors  string `json:"authors"`
	PDFUrl   string `json:"pdf_url"`
	Abstract string `json:"abstract"`
}

// crossrefEnvelope is the top-level CrossRef API response wrapper.
type crossrefEnvelope struct {
	Status  string       `json:"status"`
	Message crossrefWork `json:"message"`
}

// crossrefWork represents a single work entry in the CrossRef response.
type crossrefWork struct {
	Title          []string         `json:"title"`
	Author         []crossrefAuthor `json:"author"`
	Link           []crossrefLink   `json:"link"`
	Abstract       string           `json:"abstract"`
	ContainerTitle []string         `json:"container-title"`
}

// crossrefAuthor represents a single author in the CrossRef response.
type crossrefAuthor struct {
	Given  string `json:"given"`
	Family string `json:"family"`
}

// crossrefLink represents a downloadable link entry in the CrossRef response.
type crossrefLink struct {
	URL  string `json:"URL"`
	Type string `json:"content-type"`
}

// crossrefHTTPClient is the shared HTTP client for CrossRef calls.
// Kept package-private so tests can swap it via tsServer.
var crossrefHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// FetchDOIInfo resolves a DOI via the CrossRef API and returns structured metadata.
func FetchDOIInfo(doi string) (*DOIInfo, error) {
	if err := validateDOI(doi); err != nil {
		return nil, err
	}

	url := "https://api.crossref.org/works/" + strings.TrimSpace(doi)
	resp, err := crossrefHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("crossref request %s: %w: %w", doi, ErrNetworkError, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read crossref response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("crossref DOI %s: %w", doi, ErrDOINotFound)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crossref returned status %d for DOI %s", resp.StatusCode, doi)
	}

	var env crossrefEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("unmarshal crossref response: %w", err)
	}

	return workToDOIInfo(&env.Message), nil
}

// workToDOIInfo converts a CrossRef work response into the simplified DOIInfo struct.
func workToDOIInfo(w *crossrefWork) *DOIInfo {
	info := &DOIInfo{
		Abstract: w.Abstract,
	}

	if len(w.Title) > 0 {
		info.Title = w.Title[0]
	}

	var authors []string
	for _, a := range w.Author {
		name := strings.TrimSpace(a.Given + " " + a.Family)
		if name != "" {
			authors = append(authors, name)
		}
	}
	info.Authors = strings.Join(authors, ", ")

	for _, link := range w.Link {
		if strings.Contains(link.Type, "pdf") {
			info.PDFUrl = link.URL
			break
		}
	}

	return info
}

// validateDOI checks that the DOI string matches the expected ISO 3297 pattern.
func validateDOI(doi string) error {
	doi = strings.TrimSpace(doi)
	if doi == "" {
		return fmt.Errorf("validate DOI: empty DOI: %w", ErrDOINotFound)
	}
	if !doiRegexp.MatchString(doi) {
		return fmt.Errorf("validate DOI: invalid format %q: %w", doi, ErrDOINotFound)
	}
	return nil
}
