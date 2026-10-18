package documents

import (
	"errors"
	"log/slog"

	"paperviz/internal/repository"
)

// Stats holds the per-user dashboard counters served by GET /api/documents/stats.
type Stats struct {
	Total       int
	Saved       int
	Collections int
}

// requireDoc loads a document or returns ErrNotFound, so every sub-resource
// endpoint answers 404 for a missing document before listing anything.
func (s *Service) requireDoc(id string) (*repository.Document, error) {
	doc, err := repository.NewDocumentRepo(s.db).Get(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return doc, nil
}

// Stats returns the caller's document, saved, and collection counts.
func (s *Service) Stats(userID string) (Stats, error) {
	var out Stats
	docRepo := repository.NewDocumentRepo(s.db)
	total, err := docRepo.CountByUser(userID)
	if err != nil {
		slog.Error("count documents failed", "error", err)
		return out, err
	}
	saved, err := docRepo.CountSavedByUser(userID)
	if err != nil {
		slog.Error("count saved failed", "error", err)
		return out, err
	}
	collections, err := repository.NewCollectionRepo(s.db).ListByUser(userID)
	if err != nil {
		slog.Error("count collections failed", "error", err)
		return out, err
	}
	out = Stats{Total: total, Saved: saved, Collections: len(collections)}
	return out, nil
}

// ListClaims returns all claims for an existing document.
func (s *Service) ListClaims(id string) ([]repository.Claim, error) {
	if _, err := s.requireDoc(id); err != nil {
		return nil, err
	}
	return repository.NewClaimRepo(s.db).ListByPaper(id)
}

// ListTables returns all paper tables for an existing document.
func (s *Service) ListTables(id string) ([]repository.PaperTable, error) {
	if _, err := s.requireDoc(id); err != nil {
		return nil, err
	}
	return repository.NewPaperTableRepo(s.db).ListByPaper(id)
}

// ListMethods returns all methods for an existing document.
func (s *Service) ListMethods(id string) ([]repository.Method, error) {
	if _, err := s.requireDoc(id); err != nil {
		return nil, err
	}
	return repository.NewMethodRepo(s.db).ListByPaper(id)
}

// ListResults returns all results for an existing document.
func (s *Service) ListResults(id string) ([]repository.Result, error) {
	if _, err := s.requireDoc(id); err != nil {
		return nil, err
	}
	return repository.NewResultRepo(s.db).ListByPaper(id)
}

// ListCitations returns all citations for an existing document.
func (s *Service) ListCitations(id string) ([]repository.Citation, error) {
	if _, err := s.requireDoc(id); err != nil {
		return nil, err
	}
	return repository.NewCitationRepo(s.db).ListByPaper(id)
}

// Chart returns one chart scoped to its parent document, so a bare chart ID
// can never address another document's figure.
func (s *Service) Chart(docID, chartID string) (*repository.Chart, error) {
	chart, err := repository.NewChartRepo(s.db).GetByDocumentAndID(docID, chartID)
	if err != nil {
		return nil, err
	}
	return chart, nil
}
