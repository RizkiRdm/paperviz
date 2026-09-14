package documents

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"paperviz/internal/external"
	"paperviz/internal/repository"
	"paperviz/internal/services"
)

// Service owns document domain logic. Handlers are transport only: parse → call Service → render.
type Service struct {
	db     *sql.DB
	gemini *external.GeminiClient
}

// New creates Service with explicit config (no global state).
func New(db *sql.DB, gemini *external.GeminiClient) *Service {
	return &Service{db: db, gemini: gemini}
}

// Create validates, inserts, and starts pipeline async.
func (s *Service) Create(readingLevel string, hasFile bool, pdfBytes []byte, pastedText string, userID *string) (string, string, error) {
	res, code, err := services.ValidateAndInsert(s.db, readingLevel, hasFile, pdfBytes, pastedText, userID)
	if err != nil {
		return "", code, err
	}
	go services.RunPipelineAndPersist(s.db, s.gemini, res.DocumentID, services.PipelineInput{
		OriginalText: res.OriginalText,
		SourceType:   res.SourceType,
		ReadingLevel: readingLevel,
		PDFBytes:     res.PDFBytes,
	})
	return res.DocumentID, "", nil
}

// Get returns document and refreshes expiry window.
func (s *Service) Get(id string) (*repository.Document, error) {
	repo := repository.NewDocumentRepo(s.db)
	doc, err := repo.Get(id)
	if err != nil {
		return nil, err
	}
	if err := repo.TouchLastAccessed(id, time.Now().Unix()); err != nil {
		slog.Error("touch last_accessed_at failed", "document_id", id, "error", err)
	}
	return doc, nil
}

// List returns paginated document summaries for user.
func (s *Service) List(userID string, limit, offset int) ([]repository.DocumentListItem, error) {
	if userID == "" {
		return nil, errors.New("unauthenticated")
	}
	repo := repository.NewDocumentRepo(s.db)
	items, err := repo.ListSummariesByUser(userID, limit, offset)
	if err != nil {
		slog.Error("list documents failed", "error", err)
		return nil, err
	}
	return items, nil
}

// GetMetadata returns document metadata row.
func (s *Service) GetMetadata(id string) (*repository.Document, error) {
	repo := repository.NewDocumentRepo(s.db)
	doc, err := repo.Get(id)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// GetSections returns chapters for document.
func (s *Service) GetSections(id string) ([]repository.Chapter, error) {
	repo := repository.NewChapterRepo(s.db)
	chapters, err := repo.ListByDocument(id)
	if err != nil {
		slog.Error("list chapters failed", "error", err)
		return nil, err
	}
	return chapters, nil
}

// GetEvidence returns evidence rows for paper.
func (s *Service) GetEvidence(paperID string) ([]repository.Evidence, error) {
	repo := repository.NewEvidenceRepo(s.db)
	rows, err := repo.ListByPaper(paperID)
	if err != nil {
		slog.Error("list evidence failed", "error", err)
		return nil, err
	}
	return rows, nil
}

// GetFigures returns charts for document.
func (s *Service) GetFigures(documentID string) ([]repository.Chart, error) {
	repo := repository.NewChartRepo(s.db)
	charts, err := repo.ListByDocument(documentID)
	if err != nil {
		slog.Error("list charts failed", "error", err)
		return nil, err
	}
	return charts, nil
}

// GetClaims returns claims for paper.
func (s *Service) GetClaims(paperID string) ([]repository.Claim, error) {
	repo := repository.NewClaimRepo(s.db)
	claims, err := repo.ListByPaper(paperID)
	if err != nil {
		slog.Error("list claims failed", "error", err)
		return nil, err
	}
	return claims, nil
}
