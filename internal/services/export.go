package services

import (
	"database/sql"
	"fmt"

	"paperviz/internal/repository"
)

// ResearchExport assembles all research data for a document into a single exportable struct.
type ResearchExport struct {
	Document       ExportDocument                `json:"document"`
	Claims         []repository.Claim            `json:"claims"`
	Evidence       []repository.Evidence         `json:"evidence"`
	Tables         []repository.PaperTable       `json:"tables"`
	Methods        []repository.Method           `json:"methods"`
	Results        []repository.Result           `json:"results"`
	Citations      []repository.Citation         `json:"citations"`
	Relationships  []repository.PaperRelationship `json:"relationships"`
	Annotations    []repository.Annotation       `json:"annotations"`
	Collections    []ExportCollection            `json:"collections"`
}

// ExportDocument holds document metadata without text content for copyright compliance.
type ExportDocument struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	SourceType   string `json:"source_type"`
	ReadingLevel string `json:"reading_level"`
	CreatedAt    int64  `json:"created_at"`
	Status       string `json:"status"`
}

// ExportCollection represents a collection that contains this document.
type ExportCollection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ExportResearchContext assembles all research data for a document into a single JSON-serializable struct.
func ExportResearchContext(db *sql.DB, documentID string) (*ResearchExport, error) {
	// Fetch the document; return error if not found.
	docRepo := repository.NewDocumentRepo(db)
	doc, err := docRepo.Get(documentID)
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}

	// Fetch all related data using existing repos.
	claims, err := repository.NewClaimRepo(db).ListByPaper(documentID)
	if err != nil {
		return nil, fmt.Errorf("list claims: %w", err)
	}

	evidence, err := repository.NewEvidenceRepo(db).ListByPaper(documentID)
	if err != nil {
		return nil, fmt.Errorf("list evidence: %w", err)
	}

	tables, err := repository.NewPaperTableRepo(db).ListByPaper(documentID)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}

	methods, err := repository.NewMethodRepo(db).ListByPaper(documentID)
	if err != nil {
		return nil, fmt.Errorf("list methods: %w", err)
	}

	results, err := repository.NewResultRepo(db).ListByPaper(documentID)
	if err != nil {
		return nil, fmt.Errorf("list results: %w", err)
	}

	citations, err := repository.NewCitationRepo(db).ListByPaper(documentID)
	if err != nil {
		return nil, fmt.Errorf("list citations: %w", err)
	}

	relationships, err := repository.NewPaperRelationshipRepo(db).ListByPaper(documentID)
	if err != nil {
		return nil, fmt.Errorf("list relationships: %w", err)
	}

	// Fetch all annotations for the document (all users).
	annotations, err := repository.NewAnnotationRepo(db).ListByDocument(documentID)
	if err != nil {
		return nil, fmt.Errorf("list annotations: %w", err)
	}

	// Fetch collections the document belongs to via the join table.
	collections, err := listCollectionsByDocument(db, documentID)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}

	// Assemble into ResearchExport struct.
	export := &ResearchExport{
		Document: ExportDocument{
			ID:           doc.ID,
			Title:        doc.Title,
			SourceType:   doc.SourceType,
			ReadingLevel: doc.ReadingLevel,
			CreatedAt:    doc.CreatedAt,
			Status:       doc.Status,
		},
		Claims:        claims,
		Evidence:      evidence,
		Tables:        tables,
		Methods:       methods,
		Results:       results,
		Citations:     citations,
		Relationships: relationships,
		Annotations:   annotations,
		Collections:   collections,
	}

	return export, nil
}

// listCollectionsByDocument queries document_collections joined with collections.
func listCollectionsByDocument(db *sql.DB, documentID string) ([]ExportCollection, error) {
	rows, err := db.Query(
		`SELECT c.id, c.name
		FROM collections c
		INNER JOIN document_collections dc ON dc.collection_id = c.id
		WHERE dc.document_id = ?
		ORDER BY c.created_at ASC`, documentID,
	)
	if err != nil {
		return nil, fmt.Errorf("query collections by document: %w", err)
	}
	defer rows.Close()

	var collections []ExportCollection
	for rows.Next() {
		var ec ExportCollection
		if err := rows.Scan(&ec.ID, &ec.Name); err != nil {
			return nil, fmt.Errorf("scan collection: %w", err)
		}
		collections = append(collections, ec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate collections: %w", err)
	}
	return collections, nil
}
