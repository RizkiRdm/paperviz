package documents

import (
	"log/slog"
	"time"

	"paperviz/internal/repository"
)

// DocumentReadModel aggregates doc+chapters+charts+evidence+claims for single read.
type DocumentReadModel struct {
	Document  repository.Document
	Chapters  []repository.Chapter
	Charts    []repository.Chart
	Evidence  []repository.Evidence
	Claims    []repository.Claim
	ClaimDiff *repository.ClaimDiff
	Tables    []repository.PaperTable
	Methods   []repository.Method
	Results   []repository.Result
	Citations []repository.Citation
}

// GetReadModel loads read model aggregating all document relations in single transaction.
func (s *Service) GetReadModel(id string) (*DocumentReadModel, error) {
	tx, err := s.db.Begin()
	if err != nil {
		slog.Error("begin readmodel tx failed", "error", err)
		return nil, err
	}
	defer tx.Rollback()
	docRepo := repository.NewDocumentRepo(tx)
	doc, err := docRepo.Get(id)
	if err != nil {
		return nil, err
	}
	if err := docRepo.TouchLastAccessed(id, time.Now().Unix()); err != nil {
		slog.Error("touch last_accessed_at failed", "document_id", id, "error", err)
	}
	chapterRepo := repository.NewChapterRepo(tx)
	chapters, err := chapterRepo.ListByDocument(id)
	if err != nil {
		slog.Error("list chapters for readmodel failed", "error", err)
		return nil, err
	}
	chartRepo := repository.NewChartRepo(tx)
	charts, err := chartRepo.ListByDocument(id)
	if err != nil {
		slog.Error("list charts for readmodel failed", "error", err)
		return nil, err
	}
	evidenceRepo := repository.NewEvidenceRepo(tx)
	evidence, err := evidenceRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list evidence for readmodel failed", "error", err)
		return nil, err
	}
	claimRepo := repository.NewClaimRepo(tx)
	claims, err := claimRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list claims for readmodel failed", "error", err)
		return nil, err
	}
	claimDiffRepo := repository.NewClaimDiffRepo(tx)
	claimDiff, _ := claimDiffRepo.GetByDocument(id)
	tableRepo := repository.NewPaperTableRepo(tx)
	tables, err := tableRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list tables for readmodel failed", "error", err)
		return nil, err
	}
	methodRepo := repository.NewMethodRepo(tx)
	methods, err := methodRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list methods for readmodel failed", "error", err)
		return nil, err
	}
	resultRepo := repository.NewResultRepo(tx)
	results, err := resultRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list results for readmodel failed", "error", err)
		return nil, err
	}
	citationRepo := repository.NewCitationRepo(tx)
	citations, err := citationRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list citations for readmodel failed", "error", err)
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		slog.Error("commit readmodel tx failed", "error", err)
		return nil, err
	}
	return &DocumentReadModel{
		Document:  *doc,
		Chapters:  chapters,
		Charts:    charts,
		Evidence:  evidence,
		Claims:    claims,
		ClaimDiff: claimDiff,
		Tables:    tables,
		Methods:   methods,
		Results:   results,
		Citations: citations,
	}, nil
}
