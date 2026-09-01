package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// CitationRepo provides CRUD access to the citations table.
type CitationRepo struct {
	db dbExecutor
}

// NewCitationRepo returns a CitationRepo backed by the given executor.
func NewCitationRepo(db dbExecutor) *CitationRepo {
	return &CitationRepo{db: db}
}

// Insert writes one citation row, linked to its parent paper.
func (r *CitationRepo) Insert(c Citation) error {
	_, err := r.db.Exec(
		`INSERT INTO citations (id, paper_id, cited_paper_id, authors, title, year, venue, doi, url, source_page)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.PaperID, c.CitedPaperID, c.Authors, c.Title, c.Year, c.Venue, c.DOI, c.URL, c.SourcePage,
	)
	if err != nil {
		return fmt.Errorf("insert citation: %w", err)
	}
	return nil
}

// ListByPaper returns all citations for a paper, ordered chronologically.
func (r *CitationRepo) ListByPaper(paperID string) ([]Citation, error) {
	rows, err := r.db.Query(
		`SELECT id, paper_id, cited_paper_id, authors, title, year, venue, doi, url, source_page
		FROM citations WHERE paper_id = ? ORDER BY year ASC`, paperID,
	)
	if err != nil {
		return nil, fmt.Errorf("list citations: %w", err)
	}
	defer rows.Close()

	var list []Citation
	for rows.Next() {
		var c Citation
		if err := rows.Scan(&c.ID, &c.PaperID, &c.CitedPaperID, &c.Authors, &c.Title, &c.Year, &c.Venue, &c.DOI, &c.URL, &c.SourcePage); err != nil {
			return nil, fmt.Errorf("scan citation: %w", err)
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate citations: %w", err)
	}
	return list, nil
}

// GetByID returns a single citation, scoped to its parent paper.
// Returns ErrNotFound if no row matches either the id or the paper scope.
func (r *CitationRepo) GetByID(paperID, citationID string) (*Citation, error) {
	row := r.db.QueryRow(
		`SELECT id, paper_id, cited_paper_id, authors, title, year, venue, doi, url, source_page
		FROM citations WHERE id = ? AND paper_id = ?`, citationID, paperID,
	)
	var c Citation
	err := row.Scan(&c.ID, &c.PaperID, &c.CitedPaperID, &c.Authors, &c.Title, &c.Year, &c.Venue, &c.DOI, &c.URL, &c.SourcePage)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get citation: %w", err)
	}
	return &c, nil
}
