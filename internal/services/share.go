package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
