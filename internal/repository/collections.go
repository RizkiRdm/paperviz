package repository

import (
	"fmt"
)

// CollectionRepo provides CRUD access to the collections table.
type CollectionRepo struct {
	db dbExecutor
}

func NewCollectionRepo(db dbExecutor) *CollectionRepo {
	return &CollectionRepo{db: db}
}

// Insert creates a new collection.
func (r *CollectionRepo) Insert(c Collection) error {
	_, err := r.db.Exec(
		`INSERT INTO collections (id, user_id, name, created_at) VALUES (?, ?, ?, ?)`,
		c.ID, c.UserID, c.Name, c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert collection: %w", err)
	}
	return nil
}

// Get retrieves a collection by ID. Returns ErrNotFound if no row matches.
func (r *CollectionRepo) Get(id string) (*Collection, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, name, created_at FROM collections WHERE id = ?`, id,
	)
	var c Collection
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt)
	if err != nil {
		return nil, ErrNotFound
	}
	return &c, nil
}

// ListByUser returns collections belonging to a user, ordered by most recent.
func (r *CollectionRepo) ListByUser(userID string) ([]CollectionListItem, error) {
	rows, err := r.db.Query(
		`SELECT c.id, c.name, c.created_at,
		        (SELECT COUNT(*) FROM document_collections dc WHERE dc.collection_id = c.id) AS document_count
		FROM collections c WHERE c.user_id = ? ORDER BY c.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list collections by user: %w", err)
	}
	defer rows.Close()

	var items []CollectionListItem
	for rows.Next() {
		var it CollectionListItem
		if err := rows.Scan(&it.ID, &it.Name, &it.CreatedAt, &it.DocumentCount); err != nil {
			return nil, fmt.Errorf("scan collection: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate collections: %w", err)
	}
	return items, nil
}

// UpdateName sets a new name on a collection.
func (r *CollectionRepo) UpdateName(id string, name string) error {
	_, err := r.db.Exec(`UPDATE collections SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return fmt.Errorf("update collection name: %w", err)
	}
	return nil
}

// Delete removes a collection. document_collections cascade via FK.
func (r *CollectionRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM collections WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete collection: %w", err)
	}
	return nil
}

// AddDocument adds a document to a collection.
func (r *CollectionRepo) AddDocument(collectionID, documentID string) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO document_collections (collection_id, document_id) VALUES (?, ?)`,
		collectionID, documentID,
	)
	if err != nil {
		return fmt.Errorf("add document to collection: %w", err)
	}
	return nil
}

// RemoveDocument removes a document from a collection.
func (r *CollectionRepo) RemoveDocument(collectionID, documentID string) error {
	_, err := r.db.Exec(
		`DELETE FROM document_collections WHERE collection_id = ? AND document_id = ?`,
		collectionID, documentID,
	)
	if err != nil {
		return fmt.Errorf("remove document from collection: %w", err)
	}
	return nil
}

// ListDocuments returns documents in a collection.
func (r *CollectionRepo) ListDocuments(collectionID string) ([]DocumentListItem, error) {
	rows, err := r.db.Query(
		`SELECT d.id, d.title, d.created_at, d.status, d.saved,
		        substr(coalesce(d.simplified_text, ''), 1, 240) AS summary_preview,
		        (SELECT COUNT(*) FROM charts c WHERE c.document_id = d.id) AS chart_count,
		        (SELECT COUNT(*) FROM charts c WHERE c.document_id = d.id AND c.annotation IS NOT NULL AND c.annotation != '') AS explanation_count
		FROM documents d
		INNER JOIN document_collections dc ON dc.document_id = d.id
		WHERE dc.collection_id = ?
		ORDER BY d.created_at DESC`,
		collectionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list documents in collection: %w", err)
	}
	defer rows.Close()

	var items []DocumentListItem
	for rows.Next() {
		var it DocumentListItem
		if err := rows.Scan(&it.ID, &it.Title, &it.CreatedAt, &it.Status, &it.Saved, &it.SummaryPreview, &it.ChartCount, &it.ExplanationCount); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate documents: %w", err)
	}
	return items, nil
}
