package documents

import (
	"log/slog"

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

// GetReadModel loads read model aggregating all document relations.
func (s *Service) GetReadModel(id string) (*DocumentReadModel, error) {
	docRepo := repository.NewDocumentRepo(s.db)
	doc, err := docRepo.Get(id)
	if err != nil {
		return nil, err
	}
	chapterRepo := repository.NewChapterRepo(s.db)
	chapters, err := chapterRepo.ListByDocument(id)
	if err != nil {
		slog.Error("list chapters for readmodel failed", "error", err)
		return nil, err
	}
	chartRepo := repository.NewChartRepo(s.db)
	charts, err := chartRepo.ListByDocument(id)
	if err != nil {
		slog.Error("list charts for readmodel failed", "error", err)
		return nil, err
	}
	evidenceRepo := repository.NewEvidenceRepo(s.db)
	evidence, err := evidenceRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list evidence for readmodel failed", "error", err)
		return nil, err
	}
	claimRepo := repository.NewClaimRepo(s.db)
	claims, err := claimRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list claims for readmodel failed", "error", err)
		return nil, err
	}
	claimDiffRepo := repository.NewClaimDiffRepo(s.db)
	claimDiff, _ := claimDiffRepo.GetByDocument(id)
	tableRepo := repository.NewPaperTableRepo(s.db)
	tables, err := tableRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list tables for readmodel failed", "error", err)
		return nil, err
	}
	methodRepo := repository.NewMethodRepo(s.db)
	methods, err := methodRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list methods for readmodel failed", "error", err)
		return nil, err
	}
	resultRepo := repository.NewResultRepo(s.db)
	results, err := resultRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list results for readmodel failed", "error", err)
		return nil, err
	}
	citationRepo := repository.NewCitationRepo(s.db)
	citations, err := citationRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list citations for readmodel failed", "error", err)
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
