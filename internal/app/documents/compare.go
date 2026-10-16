package documents

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"paperviz/internal/external"
	"paperviz/internal/repository"
	"paperviz/internal/services"
)

// Domain errors for POST /api/documents/compare, mapped to the same wire
// codes the handler produced when the orchestration lived inline.
var (
	// ErrDocumentMissing means one of the compared documents does not exist.
	ErrDocumentMissing = errors.New("document not found")

	// ErrExtractionFailed means the per-paper summary extraction failed.
	ErrExtractionFailed = errors.New("extraction failed")

	// ErrComparisonFailed means the cross-paper comparison call failed.
	ErrComparisonFailed = errors.New("comparison failed")
)

// Compare extracts a summary per document, compares the papers with the
// requesting user's model client, and records an analytics event. The client
// is resolved per request upstream — there is no server-side fallback key.
func (s *Service) Compare(ctx context.Context, client *external.LLM, docIDs []string) (services.PaperComparison, error) {
	var papers []services.PaperSummary
	docRepo := repository.NewDocumentRepo(s.db)
	for _, docID := range docIDs {
		doc, err := docRepo.Get(docID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return services.PaperComparison{}, fmt.Errorf("%w: %s", ErrDocumentMissing, docID)
			}
			slog.Error("get document for comparison failed", "error", err)
			return services.PaperComparison{}, err
		}
		summary, err := services.ExtractPaperSummary(ctx, client, doc.ID, doc.Title, doc.OriginalText)
		if err != nil {
			slog.Error("extract paper summary failed", "document_id", docID, "error", err)
			return services.PaperComparison{}, ErrExtractionFailed
		}
		papers = append(papers, summary)
	}
	comparison, err := services.ComparePapers(ctx, client, papers)
	if err != nil {
		slog.Error("compare papers failed", "error", err)
		return services.PaperComparison{}, ErrComparisonFailed
	}
	s.recordComparisonEvent(docIDs)
	return comparison, nil
}

// recordComparisonEvent tracks a completed comparison for the analytics
// dashboard; a tracking failure must never fail the user's request.
func (s *Service) recordComparisonEvent(docIDs []string) {
	eventID, _ := repository.NewID()
	if eventID == "" {
		return
	}
	_, _ = s.db.Exec(
		`INSERT INTO analytics_events (id, event_type, entity_id, metadata, created_at) VALUES (?, 'comparison', ?, ?, ?)`,
		eventID, strings.Join(docIDs, ","), "", time.Now().Unix(),
	)
}
