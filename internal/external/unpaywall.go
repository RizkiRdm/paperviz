package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrDOINotFoundReexport is re-declared here to keep the sentinel accessible
// via this file's imports; callers should use ErrDOINotFound from crossref.go.
// Declared only so this file compiles independently if crossref.go is absent.
// In practice both files share the package and ErrDOINotFound is defined once.

// ErrNoOpenAccess means the DOI exists but no open-access PDF URL was found.
var ErrNoOpenAccess = errors.New("no open-access PDF available")

// unpaywallHTTPClient is the shared HTTP client for Unpaywall calls.
var unpaywallHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// unpaywallEnvelope is the top-level Unpaywall API response.
type unpaywallEnvelope struct {
	Doi            string               `json:"doi"`
	IsOa           bool                 `json:"is_oa"`
	BestOaLocation *unpaywallOaLocation `json:"best_oa_location"`
}

// unpaywallOaLocation represents the best open-access location for a work.
type unpaywallOaLocation struct {
	UrlForPdf string `json:"url_for_pdf"`
	Url       string `json:"url"`
}

// FetchOpenAccessPDF returns the open-access PDF URL for a DOI, or nil if none exists.
func FetchOpenAccessPDF(doi string) (*string, error) {
	if err := validateDOI(doi); err != nil {
		return nil, err
	}

	url := "https://api.unpaywall.org/v2/" + strings.TrimSpace(doi) + "?email=paperviz@example.com"
	resp, err := unpaywallHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unpaywall request %s: %w: %w", doi, ErrNetworkError, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read unpaywall response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return nil, fmt.Errorf("unpaywall DOI %s: %w", doi, ErrDOINotFound)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unpaywall returned status %d for DOI %s", resp.StatusCode, doi)
	}

	var env unpaywallEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("unmarshal unpaywall response: %w", err)
	}

	if !env.IsOa || env.BestOaLocation == nil {
		return nil, nil
	}

	pdfUrl := env.BestOaLocation.UrlForPdf
	if pdfUrl == "" {
		pdfUrl = env.BestOaLocation.Url
	}
	if pdfUrl == "" {
		return nil, nil
	}

	return &pdfUrl, nil
}
