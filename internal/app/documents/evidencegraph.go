package documents

import (
	"log/slog"

	"paperviz/internal/repository"
)

// ClaimWithEvidence pairs a claim with the evidence rows that reference it.
type ClaimWithEvidence struct {
	Claim    repository.Claim
	Evidence []repository.Evidence
}

// EvidenceGraph is the typed DTO behind GET /api/documents/:id/evidence-graph.
type EvidenceGraph struct {
	Claims    []ClaimWithEvidence
	Tables    []repository.PaperTable
	Methods   []repository.Method
	Results   []repository.Result
	Citations []repository.Citation
}

// EvidenceGraph aggregates a document's claims (each with its evidence),
// tables, methods, results, and citations in a single read transaction.
func (s *Service) EvidenceGraph(id string) (*EvidenceGraph, error) {
	if _, err := s.requireDoc(id); err != nil {
		return nil, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		slog.Error("begin evidence graph tx failed", "error", err)
		return nil, err
	}
	defer tx.Rollback()

	claimRepo := repository.NewClaimRepo(tx)
	claims, err := claimRepo.ListByPaper(id)
	if err != nil {
		slog.Error("list claims for evidence graph failed", "error", err)
		return nil, err
	}
	claimResults := make([]ClaimWithEvidence, 0, len(claims))
	for _, c := range claims {
		evidence, err := claimRepo.GetEvidence(c.ID)
		if err != nil {
			slog.Error("get evidence for claim failed", "claim_id", c.ID, "error", err)
			return nil, err
		}
		claimResults = append(claimResults, ClaimWithEvidence{Claim: c, Evidence: evidence})
	}
	tables, err := repository.NewPaperTableRepo(tx).ListByPaper(id)
	if err != nil {
		slog.Error("list tables for evidence graph failed", "error", err)
		return nil, err
	}
	methods, err := repository.NewMethodRepo(tx).ListByPaper(id)
	if err != nil {
		slog.Error("list methods for evidence graph failed", "error", err)
		return nil, err
	}
	results, err := repository.NewResultRepo(tx).ListByPaper(id)
	if err != nil {
		slog.Error("list results for evidence graph failed", "error", err)
		return nil, err
	}
	citations, err := repository.NewCitationRepo(tx).ListByPaper(id)
	if err != nil {
		slog.Error("list citations for evidence graph failed", "error", err)
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		slog.Error("commit evidence graph tx failed", "error", err)
		return nil, err
	}
	return &EvidenceGraph{
		Claims:    claimResults,
		Tables:    tables,
		Methods:   methods,
		Results:   results,
		Citations: citations,
	}, nil
}
