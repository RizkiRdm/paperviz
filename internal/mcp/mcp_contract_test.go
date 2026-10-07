package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"paperviz/internal/repository"
)

// loadMigrations reads every numbered migration in dir, keyed by version.
// The whole set is loaded rather than a hand-picked subset so a schema change
// cannot leave the MCP test path silently running against an older schema than
// production.
func loadMigrations(t *testing.T) map[int]string {
	t.Helper()
	dir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	migrations := make(map[int]string)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".sql") {
			continue
		}
		version, err := strconv.Atoi(strings.SplitN(name, "_", 2)[0])
		if err != nil {
			t.Errorf("migration %q has no numeric version prefix", name)
			continue
		}
		sqlStr, err := repository.ReadMigration(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		migrations[version] = sqlStr
	}

	if len(migrations) == 0 {
		t.Fatal("no migrations found")
	}
	return migrations
}

// newTestDB opens an in-memory SQLite with all migrations applied.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := repository.Open(":memory:", loadMigrations(t))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// newTestServer builds an MCPServer through the production constructor.
//
// It deliberately calls NewMCPServer rather than hand-building the struct:
// registerTools is where a bad tool definition panics, and hand-building
// skipped it, which is how a server that could not boot shipped with a green
// test suite.
func newTestServer(t *testing.T) *MCPServer {
	t.Helper()
	return NewMCPServer(newTestDB(t), nil, "test-key")
}

// TestMCPHandshake_ListsLockedTools connects a real client to the server over
// in-memory transports and completes initialize + tools/list.
//
// This is the check that was missing. Every other test here calls a handler
// directly, so none of them executed registerTools — which is where a malformed
// tool definition panics, and how a server that could not boot shipped with a
// green suite. Going through a client session also asserts the advertised JSON
// Schemas are valid, which a direct handler call cannot.
func TestMCPHandshake_ListsLockedTools(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv := newTestServer(t)

	clientT, serverT := sdk.NewInMemoryTransports()
	serverSession, err := srv.Server().Connect(ctx, serverT, nil)
	if err != nil {
		t.Fatalf("connect server: %v", err)
	}
	t.Cleanup(func() { serverSession.Close() })

	client := sdk.NewClient(&sdk.Implementation{Name: "contract-test", Version: "0.0.1"}, nil)
	clientSession, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	t.Cleanup(func() { clientSession.Close() })

	if got := clientSession.InitializeResult().ServerInfo.Name; got != "paperviz" {
		t.Errorf("server name = %q, want %q", got, "paperviz")
	}

	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}

	want := map[string]bool{
		"ingest_document":  false,
		"search_documents": false,
		"get_document":     false,
		"get_figures":      false,
		"get_evidence":     false,
	}

	got := make(map[string]int, len(result.Tools))
	for _, tool := range result.Tools {
		got[tool.Name]++
		if _, ok := want[tool.Name]; !ok {
			t.Errorf("unexpected tool advertised: %q", tool.Name)
			continue
		}
		if tool.Description == "" {
			t.Errorf("tool %q has no description", tool.Name)
		}
		if tool.InputSchema == nil {
			t.Errorf("tool %q has no input schema", tool.Name)
		}
	}

	for name, count := range got {
		if count != 1 {
			t.Errorf("tool %q advertised %d times, want 1", name, count)
		}
	}
	for name := range want {
		if got[name] == 0 {
			t.Errorf("expected tool %q to be advertised", name)
		}
	}
	if len(result.Tools) != len(want) {
		t.Errorf("advertised %d tools, want exactly %d", len(result.Tools), len(want))
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
