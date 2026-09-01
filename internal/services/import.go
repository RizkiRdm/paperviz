package services

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"paperviz/internal/external"
)

// Sentinel errors for the Import Service, surfaced to handlers as
// snake_case error codes per ARCHITECTURE.md Internal Contracts.
var (
	ErrInvalidDOI     = errors.New("invalid DOI format")
	ErrDOINotFound    = errors.New("DOI not found in external registry")
	ErrPaywallBlocked = errors.New("paper is paywalled with no open-access PDF")
	ErrInvalidURL     = errors.New("invalid URL provided")
	ErrSSRFBlocked    = errors.New("URL resolves to a private/internal address")
	ErrNotPDF         = errors.New("URL does not serve a PDF file")
)

// importHTTPClient is the shared HTTP client for all Import Service requests.
// Kept package-private so tests can swap it via tsServer round-tripper.
var importHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// fetchDOIInfo is the function used to resolve DOI metadata via CrossRef.
// Swappable for testing without modifying the external package.
var fetchDOIInfo = external.FetchDOIInfo

// fetchOpenAccessPDF is the function used to find open-access PDF URLs.
// Swappable for testing without modifying the external package.
var fetchOpenAccessPDF = external.FetchOpenAccessPDF

// isPrivateIP reports whether hostname resolves to a private/loopback address.
// Swappable for testing to avoid blocking test server addresses.
var isPrivateIP = blockPrivateIP

// doiPattern validates DOI format per ISO 3297: 10.NNNN/suffix.
var doiPattern = regexp.MustCompile(`^10.\d{4,9}/`)

// privateIPRanges holds net.IPNet ranges for RFC 1918 + loopback addresses.
// Resolved once at init to avoid repeated parsing.
var privateIPRanges []*net.IPNet

func init() {
	privateCIDRs := []string{
		"127.0.0.0/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"::1/128",
		"fc00::/7",
	}
	for _, cidr := range privateCIDRs {
		_, ipNet, _ := net.ParseCIDR(cidr)
		if ipNet != nil {
			privateIPRanges = append(privateIPRanges, ipNet)
		}
	}
}

// FetchByDOI resolves a DOI to PDF bytes via CrossRef metadata and Unpaywall
// open-access lookup, then downloads the PDF from the best available URL.
func FetchByDOI(doi string) ([]byte, error) {
	doi = strings.TrimSpace(doi)
	if !doiPattern.MatchString(doi) {
		return nil, fmt.Errorf("fetch doi %q: %w", doi, ErrInvalidDOI)
	}

	info, err := fetchDOIInfo(doi)
	if err != nil {
		if errors.Is(err, external.ErrDOINotFound) {
			return nil, fmt.Errorf("fetch doi %s: %w", doi, ErrDOINotFound)
		}
		return nil, fmt.Errorf("fetch doi %s: %w", doi, err)
	}

	pdfURL, err := fetchOpenAccessPDF(doi)
	if err != nil {
		if errors.Is(err, external.ErrDOINotFound) {
			return nil, fmt.Errorf("fetch doi %s: %w", doi, ErrDOINotFound)
		}
		return nil, fmt.Errorf("fetch doi %s unpaywall: %w", doi, err)
	}

	downloadURL := ""
	if pdfURL != nil && *pdfURL != "" {
		downloadURL = *pdfURL
	} else if info.PDFUrl != "" {
		downloadURL = info.PDFUrl
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("fetch doi %s: %w", doi, ErrPaywallBlocked)
	}

	pdfBytes, err := downloadPDF(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("fetch doi %s download: %w", doi, err)
	}

	return pdfBytes, nil
}

// FetchByURL downloads a PDF from a direct URL with SSRF protection and
// source-specific URL resolution for arxiv, PubMed, and generic hosts.
func FetchByURL(rawURL string) ([]byte, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("fetch url: %w", ErrInvalidURL)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetch url parse %q: %w", rawURL, err)
	}

	// SSRF protection: only allow HTTPS scheme.
	if parsed.Scheme != "https" {
		return nil, fmt.Errorf("fetch url %s (scheme %s): %w", rawURL, parsed.Scheme, ErrInvalidURL)
	}

	if err := isPrivateIP(parsed.Hostname()); err != nil {
		return nil, fmt.Errorf("fetch url %s: %w", rawURL, err)
	}

	downloadURL, err := resolveSourceURL(parsed)
	if err != nil {
		return nil, err
	}

	// Download the PDF with content-type validation.
	pdfBytes, err := downloadPDF(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("fetch url %s: %w", rawURL, err)
	}

	return pdfBytes, nil
}

// resolveSourceURL maps known academic hosts to their direct PDF download URLs.
func resolveSourceURL(parsed *url.URL) (string, error) {
	host := strings.ToLower(parsed.Hostname())

	switch {
	case strings.HasSuffix(host, "arxiv.org"):
		return resolveArxivURL(parsed)
	case strings.HasSuffix(host, "pubmed.ncbi.nlm.nih.gov"):
		return resolvePubMedURL(parsed)
	default:
		// Generic: use the URL as-is for direct PDF download.
		return parsed.String(), nil
	}
}

// resolveArxivURL converts an arxiv abstract/HTML URL to a direct PDF URL.
func resolveArxivURL(parsed *url.URL) (string, error) {
	path := parsed.Path
	// Strip trailing /abs/ or /pdf/ prefixes to get the bare ID.
	path = strings.TrimPrefix(path, "/abs/")
	path = strings.TrimPrefix(path, "/pdf/")
	path = strings.TrimPrefix(path, "/html/")
	path = strings.TrimSuffix(path, ".pdf")
	path = strings.TrimLeft(path, "/")

	if path == "" {
		return "", fmt.Errorf("resolve arxiv url %s: %w", parsed.String(), ErrInvalidURL)
	}

	return "https://arxiv.org/pdf/" + path + ".pdf", nil
}

// resolvePubMedURL resolves a PubMed article page to its PMC PDF URL via the NCBI API.
func resolvePubMedURL(parsed *url.URL) (string, error) {
	// Extract PMID from path: /12345678/ or /12345678
	path := strings.Trim(parsed.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", fmt.Errorf("resolve pubmed url %s: %w", parsed.String(), ErrInvalidURL)
	}
	pmid := parts[0]

	// Query NCBI for PMC ID to get PDF link.
	apiURL := fmt.Sprintf("https://eutils.ncbi.nlm.nih.gov/entrez/eutils/elink.fcgi?dbfrom=pubmed&db=pmc&id=%s&retmode=json", pmid)
	resp, err := importHTTPClient.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("resolve pubmed %s: %w", pmid, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read pubmed elink response: %w", err)
	}

	// Extract PMC ID from response using simple string search.
	// The JSON structure is deeply nested; string search is sufficient here.
	bodyStr := string(body)
	pmcPrefix := `"pmc"`
	idx := strings.Index(bodyStr, pmcPrefix)
	if idx < 0 {
		// No PMC ID found — fall back to PubMed abstract page (no direct PDF).
		return "", fmt.Errorf("resolve pubmed %s: no PMC PDF available: %w", pmid, ErrPaywallBlocked)
	}

	// Find the PMC ID value following the prefix.
	rest := bodyStr[idx+len(pmcPrefix):]
	// Skip past colon, quotes, and whitespace to the numeric ID.
	rest = strings.TrimLeft(rest, `": 	`)
	idEnd := 0
	for idEnd < len(rest) && (rest[idEnd] >= '0' && rest[idEnd] <= '9') {
		idEnd++
	}
	if idEnd == 0 {
		return "", fmt.Errorf("resolve pubmed %s: no PMC PDF available: %w", pmid, ErrPaywallBlocked)
	}
	pmcID := rest[:idEnd]

	return fmt.Sprintf("https://www.ncbi.nlm.nih.gov/pmc/articles/PMC%s/pdf/", pmcID), nil
}

// blockPrivateIP resolves hostname and rejects loopback/RFC1918 addresses.
func blockPrivateIP(hostname string) error {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("resolve hostname %s: %w: %w", hostname, ErrSSRFBlocked, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("resolve hostname %s: no records: %w", hostname, ErrSSRFBlocked)
	}
	for _, ip := range ips {
		for _, ipNet := range privateIPRanges {
			if ipNet.Contains(ip) {
				return fmt.Errorf("hostname %s resolves to private IP %s: %w", hostname, ip, ErrSSRFBlocked)
			}
		}
	}
	return nil
}

// downloadPDF fetches a URL and validates the response is a PDF before
// reading the full body. Uses HEAD-then-GET to check Content-Type early.
func downloadPDF(rawURL string) ([]byte, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse url %s: %w", rawURL, err)
	}

	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, fmt.Errorf("download %s: unsupported scheme %s: %w", rawURL, parsed.Scheme, ErrNotPDF)
	}

	// HEAD request to check Content-Type before downloading full body.
	headResp, err := importHTTPClient.Head(rawURL)
	if err == nil {
		headResp.Body.Close()
		ct := headResp.Header.Get("Content-Type")
		if ct != "" && !isPDFContentType(ct) {
			return nil, fmt.Errorf("download %s: content-type %q: %w", rawURL, ct, ErrNotPDF)
		}
	}

	// GET request to download the full PDF body.
	resp, err := importHTTPClient.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %d: %w", rawURL, resp.StatusCode, ErrNotPDF)
	}

	// Re-check Content-Type on GET response.
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !isPDFContentType(ct) {
		return nil, fmt.Errorf("download %s: content-type %q: %w", rawURL, ct, ErrNotPDF)
	}

	// Read up to 100MB to prevent abuse. Real PDFs are well under this.
	const maxPDFSize = 100 * 1024 * 1024
	limitedReader := io.LimitReader(resp.Body, maxPDFSize+1)
	pdfBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("read download %s: %w", rawURL, err)
	}
	if len(pdfBytes) > maxPDFSize {
		return nil, fmt.Errorf("download %s: exceeds max size: %w", rawURL, ErrNotPDF)
	}

	// Validate PDF magic bytes as final check.
	if len(pdfBytes) < 5 || string(pdfBytes[:5]) != "%PDF-" {
		slog.Warn("downloaded file is not a PDF despite content-type", "url", rawURL, "ct", ct)
		return nil, fmt.Errorf("download %s: not a valid PDF file: %w", rawURL, ErrNotPDF)
	}

	return pdfBytes, nil
}

// isPDFContentType reports whether the Content-Type header indicates a PDF.
func isPDFContentType(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "application/pdf") || strings.Contains(ct, "pdf")
}
