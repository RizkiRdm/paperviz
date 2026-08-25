package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"paperviz/internal/repository"
)

type SharedFigure struct {
	ChartID          string  `json:"chart_id"`
	OriginalImageURL string  `json:"original_image_url,omitempty"`
	Explanation      *string `json:"explanation,omitempty"`
	ChartData        *string `json:"chart_data,omitempty"`
	SourceMethod     string  `json:"source_method"`
	PageNumber       *int    `json:"page_number,omitempty"`
	SourcePaperTitle string  `json:"source_paper_title"`
	ReadingLevel     string  `json:"reading_level"`
	DocumentID       string  `json:"document_id"`
}

func GenerateShareToken(ctx context.Context, db *sql.DB, docID, chartID, userID string) (string, error) {
	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.Get(docID)
	if errors.Is(err, repository.ErrNotFound) {
		return "", fmt.Errorf("document not found")
	}
	if err != nil {
		return "", fmt.Errorf("get document: %w", err)
	}
	if doc.UserID == nil || *doc.UserID != userID {
		return "", fmt.Errorf("unauthorized")
	}

	chartRepo := repository.NewChartRepo(db)
	chart, err := chartRepo.GetByDocumentAndID(docID, chartID)
	if errors.Is(err, repository.ErrNotFound) {
		return "", fmt.Errorf("chart not found")
	}
	if err != nil {
		return "", fmt.Errorf("get chart: %w", err)
	}

	if chart.ShareToken != nil && *chart.ShareToken != "" {
		return *chart.ShareToken, nil
	}

	token, err := repository.NewID()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	if err := chartRepo.SetShareToken(chart.ID, token); err != nil {
		return "", fmt.Errorf("set share token: %w", err)
	}

	return token, nil
}

func RevokeShareToken(ctx context.Context, db *sql.DB, docID, chartID, userID string) error {
	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.Get(docID)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("document not found")
	}
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}
	if doc.UserID == nil || *doc.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	chartRepo := repository.NewChartRepo(db)
	return chartRepo.RevokeShareToken(chartID)
}

func GetSharedFigure(ctx context.Context, db *sql.DB, token string) (*SharedFigure, error) {
	chartRepo := repository.NewChartRepo(db)
	chart, err := chartRepo.GetByShareToken(token)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get chart: %w", err)
	}

	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.Get(chart.DocumentID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}

	if doc.Visibility == "private" {
		return nil, fmt.Errorf("not found")
	}

	if err := chartRepo.IncrementShareVisits(token); err != nil {
		slog.Warn("increment share visits failed", "error", err)
	}

	fig := &SharedFigure{
		ChartID:          chart.ID,
		Explanation:      chart.Annotation,
		ChartData:        chart.ChartData,
		SourceMethod:     chart.SourceMethod,
		PageNumber:       chart.PageNumber,
		SourcePaperTitle: doc.Title,
		ReadingLevel:     doc.ReadingLevel,
		DocumentID:       doc.ID,
	}

	if chart.ImageBlob != nil && len(chart.ImageBlob) > 0 {
		fig.OriginalImageURL = fmt.Sprintf("/api/documents/%s/charts/%s/image", doc.ID, chart.ID)
	}

	return fig, nil
}

// SharedPaper is the public payload for a full-paper share page — no original
// text, user ID, or tokens ever leave through it.
type SharedPaper struct {
	DocumentID     string            `json:"document_id"`
	Title          string            `json:"title"`
	ReadingLevel   string            `json:"reading_level"`
	CreatedAt      int64             `json:"created_at"`
	SimplifiedText *string           `json:"simplified_text,omitempty"`
	Charts         []SharedChartItem `json:"charts,omitempty"`
}

// SharedChartItem is one chart entry on a shared paper page.
type SharedChartItem struct {
	ChartID      string  `json:"chart_id"`
	SourceMethod string  `json:"source_method"`
	Annotation   *string `json:"annotation,omitempty"`
	PageNumber   *int    `json:"page_number,omitempty"`
	ImageURL     string  `json:"image_url,omitempty"`
}

// GenerateDocumentShareToken lazily creates (or returns the existing) share
// token for a document, bumping private docs to unlisted so the link works.
func GenerateDocumentShareToken(ctx context.Context, db *sql.DB, docID, userID string) (string, error) {
	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.Get(docID)
	if errors.Is(err, repository.ErrNotFound) {
		return "", fmt.Errorf("document not found")
	}
	if err != nil {
		return "", fmt.Errorf("get document: %w", err)
	}
	if doc.UserID == nil || *doc.UserID != userID {
		return "", fmt.Errorf("unauthorized")
	}

	if doc.ShareToken != nil && *doc.ShareToken != "" {
		return *doc.ShareToken, nil
	}

	token, err := repository.NewID()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	if err := docRepo.SetShareToken(doc.ID, token); err != nil {
		return "", fmt.Errorf("set share token: %w", err)
	}

	if doc.Visibility == "private" {
		if err := docRepo.UpdateVisibility(doc.ID, "unlisted"); err != nil {
			return "", fmt.Errorf("update visibility: %w", err)
		}
	}

	return token, nil
}

// RevokeDocumentShareToken clears a document's share token, returning an
// unlisted document to private — public visibility is never downgraded.
func RevokeDocumentShareToken(ctx context.Context, db *sql.DB, docID, userID string) error {
	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.Get(docID)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("document not found")
	}
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}
	if doc.UserID == nil || *doc.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	if err := docRepo.RevokeShareToken(doc.ID); err != nil {
		return fmt.Errorf("revoke share token: %w", err)
	}

	if doc.Visibility == "unlisted" {
		if err := docRepo.UpdateVisibility(doc.ID, "private"); err != nil {
			return fmt.Errorf("update visibility: %w", err)
		}
	}

	return nil
}

// SetDocumentVisibility updates a document's visibility after validation;
// switching to private also revokes the share token so old links die.
func SetDocumentVisibility(ctx context.Context, db *sql.DB, docID, userID, visibility string) error {
	switch visibility {
	case "public", "unlisted", "private":
	default:
		return fmt.Errorf("invalid visibility")
	}

	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.Get(docID)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("document not found")
	}
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}
	if doc.UserID == nil || *doc.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	if err := docRepo.UpdateVisibility(doc.ID, visibility); err != nil {
		return fmt.Errorf("update visibility: %w", err)
	}

	if visibility == "private" {
		if err := docRepo.RevokeShareToken(doc.ID); err != nil {
			return fmt.Errorf("revoke share token: %w", err)
		}
	}

	return nil
}

// GetSharedPaper resolves a paper share token into the public payload,
// refusing private documents and unknown tokens with a uniform "not found".
func GetSharedPaper(ctx context.Context, db *sql.DB, token string) (*SharedPaper, error) {
	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.GetByShareToken(token)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get document by share token: %w", err)
	}

	if doc.Visibility == "private" {
		return nil, fmt.Errorf("not found")
	}

	if err := docRepo.IncrementShareVisits(token); err != nil {
		slog.Warn("increment share visits failed", "error", err)
	}

	paper := &SharedPaper{
		DocumentID:     doc.ID,
		Title:          doc.Title,
		ReadingLevel:   doc.ReadingLevel,
		CreatedAt:      doc.CreatedAt,
		SimplifiedText: doc.SimplifiedText,
	}

	chartRepo := repository.NewChartRepo(db)
	charts, err := chartRepo.ListByDocument(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("list charts: %w", err)
	}
	for _, c := range charts {
		item := SharedChartItem{
			ChartID:      c.ID,
			SourceMethod: c.SourceMethod,
			Annotation:   c.Annotation,
			PageNumber:   c.PageNumber,
		}
		if c.ImageBlob != nil && len(c.ImageBlob) > 0 {
			item.ImageURL = fmt.Sprintf("/api/documents/%s/charts/%s/image", doc.ID, c.ID)
		}
		paper.Charts = append(paper.Charts, item)
	}

	return paper, nil
}

// TrackReferralConversion records a share→visit conversion against whichever
// entity (document or chart) owns the referral token.
func TrackReferralConversion(ctx context.Context, db *sql.DB, ref string) error {
	if ref == "" {
		return fmt.Errorf("invalid ref")
	}

	docRepo := repository.NewDocumentRepo(db)
	err := docRepo.IncrementShareConversions(ref)
	if errors.Is(err, repository.ErrNotFound) {
		chartRepo := repository.NewChartRepo(db)
		cerr := chartRepo.IncrementShareConversions(ref)
		if errors.Is(cerr, repository.ErrNotFound) {
			return fmt.Errorf("not found")
		}
		if cerr != nil {
			return fmt.Errorf("increment chart conversions: %w", cerr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("increment document conversions: %w", err)
	}
	return nil
}
