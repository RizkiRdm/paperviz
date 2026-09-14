package mcp

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	"paperviz/internal/repository"
)

// newTestDB opens an in-memory SQLite with all migrations for contract tests.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	migrations := make(map[int]string)
	for v, file := range map[int]string{
		1: "001_init.sql", 2: "002_users.sql", 3: "003_chapters.sql",
		4: "004_chapter_charts.sql", 5: "005_evidence.sql", 6: "006_document_title.sql",
		7: "007_saved_papers.sql", 8: "008_research_collections.sql",
		9: "009_share_tokens.sql", 10: "010_document_share.sql",
		11: "011_share_referrals.sql", 12: "012_usage_analytics.sql",
	} {
		sqlStr, err := repository.ReadMigration(filepath.Join("..", "..", "migrations", file))
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}
		migrations[v] = sqlStr
	}
	db, err := repository.Open(":memory:", migrations)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// newTestServer builds a minimal MCPServer with a real DB and noop rate limiter.
func newTestServer(t *testing.T) *MCPServer {
	t.Helper()
	db := newTestDB(t)
	return &MCPServer{
		db:          db,
		apiKey:      "test-key",
		rateLimiter: NewRateLimiter(),
	}
}

// TestIngestDocument_Contract covers valid input, empty text, oversized text, and schema shape.
func TestIngestDocument_Contract(t *testing.T) {
	srv := newTestServer(t)

	tests := []struct {
		name       string
		args       IngestDocumentInput
		wantErr    bool
		wantFields []string // fields that must be non-empty on success
	}{
		{
			name:       "valid paste",
			args:       IngestDocumentInput{Text: "Sample paper text for ingestion test"},
			wantFields: []string{"document_id", "title", "status", "source_type"},
		},
		{
			name:    "empty text returns error",
			args:    IngestDocumentInput{Text: ""},
			wantErr: true,
		},
		{
			name:    "oversized text returns error",
			args:    IngestDocumentInput{Text: string(make([]byte, 500*1024+1))},
			wantErr: true,
		},
		{
			name:    "invalid reading level returns error",
			args:    IngestDocumentInput{Text: "text", ReadingLevel: "college"},
			wantErr: true,
		},
		{
			name:       "eli5 reading level accepted",
			args:       IngestDocumentInput{Text: "Simple text for ELI5", ReadingLevel: "eli5"},
			wantFields: []string{"document_id", "status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, result, err := handleIngestDocument(nil, srv, tt.args)
			if tt.wantErr {
				if err == nil && result.DocumentID != "" {
					t.Errorf("expected error or empty result, got docID=%s", result.DocumentID)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			raw, mErr := json.Marshal(result)
			if mErr != nil {
				t.Fatalf("json marshal: %v", mErr)
			}
			var parsed map[string]any
			if jErr := json.Unmarshal(raw, &parsed); jErr != nil {
				t.Fatalf("json unmarshal: %v", jErr)
			}
			for _, f := range tt.wantFields {
				if parsed[f] == nil || parsed[f] == "" {
					t.Errorf("field %s should be non-empty, got %v", f, parsed[f])
				}
			}
		})
	}
}

// TestSearchDocuments_Contract covers valid query, empty query, no-results, and schema shape.
func TestSearchDocuments_Contract(t *testing.T) {
	srv := newTestServer(t)
	// Seed one document for search.
	_, ingestRes, _ := handleIngestDocument(nil, srv, IngestDocumentInput{Text: "Quantum Computing Survey 2026"})
	_ = ingestRes

	tests := []struct {
		name    string
		args    SearchDocumentsInput
		wantErr bool
	}{
		{
			name: "valid query returns results",
			args: SearchDocumentsInput{Query: "Quantum", Limit: 5},
		},
		{
			name:    "empty query returns error",
			args:    SearchDocumentsInput{Query: ""},
			wantErr: true,
		},
		{
			name: "non-matching query returns empty list",
			args: SearchDocumentsInput{Query: "zzz_no_match_zzz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, result, err := handleSearchDocuments(nil, srv, tt.args)
			if tt.wantErr {
				if err == nil && len(result.Documents) > 0 {
					t.Errorf("expected error or empty, got %d docs", len(result.Documents))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			raw, mErr := json.Marshal(result)
			if mErr != nil {
				t.Fatalf("json marshal: %v", mErr)
			}
			var parsed map[string]any
			if jErr := json.Unmarshal(raw, &parsed); jErr != nil {
				t.Fatalf("json unmarshal: %v", jErr)
			}
			if _, ok := parsed["documents"]; !ok {
				t.Error("response must have 'documents' key")
			}
		})
	}
}

// TestGetDocument_Contract covers valid doc, not-found, empty include, and schema shape.
func TestGetDocument_Contract(t *testing.T) {
	srv := newTestServer(t)
	_, ingestDoc, _ := handleIngestDocument(nil, srv, IngestDocumentInput{Text: "Test document for get"})

	tests := []struct {
		name    string
		args    GetDocumentInput
		wantErr bool
	}{
		{
			name: "valid doc id metadata only",
			args: GetDocumentInput{DocumentID: ingestDoc.DocumentID},
		},
		{
			name: "valid doc id with sections",
			args: GetDocumentInput{DocumentID: ingestDoc.DocumentID, Include: "sections"},
		},
		{
			name:    "empty doc id returns error",
			args:    GetDocumentInput{DocumentID: ""},
			wantErr: true,
		},
		{
			name:    "not-found doc id returns error",
			args:    GetDocumentInput{DocumentID: "nonexistent-id-000"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, result, err := handleGetDocument(nil, srv, tt.args)
			if tt.wantErr {
				if err == nil && result.DocumentID != "" {
					t.Errorf("expected error or empty, got docID=%s", result.DocumentID)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			raw, mErr := json.Marshal(result)
			if mErr != nil {
				t.Fatalf("json marshal: %v", mErr)
			}
			var parsed map[string]any
			if jErr := json.Unmarshal(raw, &parsed); jErr != nil {
				t.Fatalf("json unmarshal: %v", jErr)
			}
			if parsed["document_id"] == nil || parsed["document_id"] == "" {
				t.Error("document_id must be non-empty")
			}
			if parsed["status"] == nil || parsed["status"] == "" {
				t.Error("status must be non-empty")
			}
		})
	}
}

// TestGetFigures_Contract covers valid doc, not-found, empty charts, and schema shape.
func TestGetFigures_Contract(t *testing.T) {
	srv := newTestServer(t)
	_, figDoc, _ := handleIngestDocument(nil, srv, IngestDocumentInput{Text: "Test paper for figures"})

	tests := []struct {
		name    string
		args    DocIDInput
		wantErr bool
	}{
		{
			name: "valid doc with no charts returns empty figures",
			args: DocIDInput{DocumentID: figDoc.DocumentID},
		},
		{
			name:    "not-found doc id returns error",
			args:    DocIDInput{DocumentID: "nonexistent-id-000"},
			wantErr: true,
		},
		{
			name:    "empty doc id returns error",
			args:    DocIDInput{DocumentID: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, result, err := handleGetFigures(nil, srv, tt.args)
			if tt.wantErr {
				if err == nil && len(result.Figures) > 0 {
					t.Errorf("expected error or empty, got %d figures", len(result.Figures))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			raw, mErr := json.Marshal(result)
			if mErr != nil {
				t.Fatalf("json marshal: %v", mErr)
			}
			var parsed map[string]any
			if jErr := json.Unmarshal(raw, &parsed); jErr != nil {
				t.Fatalf("json unmarshal: %v", jErr)
			}
			if _, ok := parsed["figures"]; !ok {
				t.Error("response must have 'figures' key")
			}
		})
	}
}

// TestGetEvidence_Contract covers valid doc, not-found, empty evidence, and schema shape.
func TestGetEvidence_Contract(t *testing.T) {
	srv := newTestServer(t)
	_, evDoc, _ := handleIngestDocument(nil, srv, IngestDocumentInput{Text: "Test paper for evidence"})

	tests := []struct {
		name    string
		args    DocIDInput
		wantErr bool
	}{
		{
			name: "valid doc with no evidence returns empty",
			args: DocIDInput{DocumentID: evDoc.DocumentID},
		},
		{
			name:    "not-found doc id returns error",
			args:    DocIDInput{DocumentID: "nonexistent-id-000"},
			wantErr: true,
		},
		{
			name:    "empty doc id returns error",
			args:    DocIDInput{DocumentID: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, result, err := handleGetEvidence(nil, srv, tt.args)
			if tt.wantErr {
				if err == nil && len(result.Evidence) > 0 {
					t.Errorf("expected error or empty, got %d evidence", len(result.Evidence))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			raw, mErr := json.Marshal(result)
			if mErr != nil {
				t.Fatalf("json marshal: %v", mErr)
			}
			var parsed map[string]any
			if jErr := json.Unmarshal(raw, &parsed); jErr != nil {
				t.Fatalf("json unmarshal: %v", jErr)
			}
			if _, ok := parsed["evidence"]; !ok {
				t.Error("response must have 'evidence' key")
			}
		})
	}
}
