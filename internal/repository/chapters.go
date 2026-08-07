package repository

import "fmt"

type ChapterRepo struct {
	db dbExecutor
}

func NewChapterRepo(db dbExecutor) *ChapterRepo {
	return &ChapterRepo{db: db}
}

func (r *ChapterRepo) Insert(c Chapter) error {
	_, err := r.db.Exec(
		`INSERT INTO chapters (id, document_id, title, summary, excerpt, display_order)
		VALUES (?, ?, ?, ?, ?, ?)`,
		c.ID, c.DocumentID, c.Title, c.Summary, c.Excerpt, c.DisplayOrder,
	)
	if err != nil {
		return fmt.Errorf("insert chapter: %w", err)
	}
	return nil
}

func (r *ChapterRepo) ListByDocument(documentID string) ([]Chapter, error) {
	rows, err := r.db.Query(
		`SELECT id, document_id, title, summary, excerpt, display_order
		FROM chapters WHERE document_id = ? ORDER BY display_order ASC`, documentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()

	var chapters []Chapter
	for rows.Next() {
		var c Chapter
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.Title, &c.Summary, &c.Excerpt, &c.DisplayOrder); err != nil {
			return nil, fmt.Errorf("scan chapter: %w", err)
		}
		chapters = append(chapters, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chapters: %w", err)
	}
	return chapters, nil
}
