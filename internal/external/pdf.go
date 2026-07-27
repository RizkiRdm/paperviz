package external

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// pdfParseTimeout is the MAX_LIMIT per PDF parse operation (ARCHITECTURE.md
// Section 4). PDF parsing is CPU-bound, not I/O-bound, so this is enforced
// via a goroutine + channel race rather than context cancellation (the
// underlying libraries do not accept a context).
const pdfParseTimeout = 10 * time.Second

// ErrNoTextLayer means the PDF has no extractable text (likely scanned).
// Callers MUST reject the upload before any LLM call (Failure Scenario 1).
var ErrNoTextLayer = errors.New("no extractable text layer")

// ExtractedImage is a single image pulled from a PDF page.
type ExtractedImage struct {
	PageNumber int
	Bytes      []byte
}

// ExtractText pulls plain text from PDF bytes held in memory. Returns
// ErrNoTextLayer if no text could be extracted (scanned/image-only PDF).
func ExtractText(pdfBytes []byte) (string, error) {
	type result struct {
		text string
		err  error
	}
	done := make(chan result, 1)

	go func() {
		reader, err := pdf.NewReader(bytes.NewReader(pdfBytes), int64(len(pdfBytes)))
		if err != nil {
			done <- result{err: fmt.Errorf("open pdf: %w", err)}
			return
		}
		textReader, err := reader.GetPlainText()
		if err != nil {
			done <- result{err: fmt.Errorf("extract text: %w", err)}
			return
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(textReader); err != nil {
			done <- result{err: fmt.Errorf("read extracted text: %w", err)}
			return
		}
		done <- result{text: buf.String()}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			return "", r.err
		}
		if strings.TrimSpace(r.text) == "" {
			return "", ErrNoTextLayer
		}
		return r.text, nil
	case <-time.After(pdfParseTimeout):
		slog.Warn("pdf text extraction timed out, background goroutine still running",
			"stage", "pdf_extract", "timeout_s", int(pdfParseTimeout.Seconds()))
		return "", fmt.Errorf("pdf text extraction timed out after %s", pdfParseTimeout)
	}
}

// ExtractTextByPage returns extracted text grouped by 1-indexed page
// number, so callers (the chart re-visualization service) can give Gemini
// the specific page's text as context for a specific chart image, instead
// of the whole document. This is a separate call from ExtractText because
// most callers (the simplification service) just want the flattened full
// text and don't need the extra per-page walk.
func ExtractTextByPage(pdfBytes []byte) (map[int]string, error) {
	type result struct {
		pages map[int]string
		err   error
	}
	done := make(chan result, 1)

	go func() {
		reader, err := pdf.NewReader(bytes.NewReader(pdfBytes), int64(len(pdfBytes)))
		if err != nil {
			done <- result{err: fmt.Errorf("open pdf: %w", err)}
			return
		}

		pages := make(map[int]string)
		total := reader.NumPage()
		for i := 1; i <= total; i++ {
			page := reader.Page(i)
			if page.V.IsNull() {
				continue // blank/non-content page, skip rather than error the whole doc
			}
			rows, err := page.GetTextByRow()
			if err != nil {
				continue // one unreadable page shouldn't fail the whole extraction
			}
			var buf bytes.Buffer
			for _, row := range rows {
				for _, word := range row.Content {
					buf.WriteString(word.S)
					buf.WriteString(" ")
				}
				buf.WriteString("\n")
			}
			pages[i] = buf.String()
		}
		done <- result{pages: pages}
	}()

	select {
	case r := <-done:
		return r.pages, r.err
	case <-time.After(pdfParseTimeout):
		slog.Warn("pdf per-page text extraction timed out, background goroutine still running",
			"stage", "pdf_extract_pages", "timeout_s", int(pdfParseTimeout.Seconds()))
		return nil, fmt.Errorf("pdf per-page text extraction timed out after %s", pdfParseTimeout)
	}
}

// ExtractImages pulls all embedded images from PDF bytes held in memory,
// per-page. Used by the Chart Re-visualization Service's image-fallback path.
func ExtractImages(pdfBytes []byte) ([]ExtractedImage, error) {
	type result struct {
		images []ExtractedImage
		err    error
	}
	done := make(chan result, 1)

	go func() {
		rs := bytes.NewReader(pdfBytes)
		// raw is []map[int]model.Image indexed by page position -> image ID -> image.
		// NOTE: model.Image's exact reader field name could not be verified against
		// pdfcpu's live API in this environment (no compiler available here to check).
		// If build fails here, check pdfcpu's model.Image struct for the correct
		// field (likely .Reader) and adjust.
		raw, err := api.ExtractImagesRaw(rs, []string{"all"}, nil)
		if err != nil {
			done <- result{err: fmt.Errorf("extract images: %w", err)}
			return
		}

		var images []ExtractedImage
		for pageIdx, imgsByID := range raw {
			for _, img := range imgsByID {
				b, err := io.ReadAll(img.Reader)
				if err != nil {
					continue // skip unreadable image, don't fail whole extraction
				}
				images = append(images, ExtractedImage{PageNumber: pageIdx + 1, Bytes: b})
			}
		}
		done <- result{images: images}
	}()

	select {
	case r := <-done:
		return r.images, r.err
	case <-time.After(pdfParseTimeout):
		slog.Warn("pdf image extraction timed out, background goroutine still running",
			"stage", "pdf_extract_images", "timeout_s", int(pdfParseTimeout.Seconds()))
		return nil, fmt.Errorf("pdf image extraction timed out after %s", pdfParseTimeout)
	}
}
