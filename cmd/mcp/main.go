// Command mcp is PaperViz's MCP server entrypoint. It exposes 5 tools
// (ingest_document, search_documents, get_document, get_figures, get_evidence)
// over stdio transport (default) or stateless StreamableHTTP (MCP_HTTP_ADDR).
// No LLM calls in any tool path, so no model key is required — configuration
// comes from environment variables and the same DB as the main server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	papervizMCP "paperviz/internal/mcp"
	"paperviz/internal/repository"
)

func main() {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "paperviz.db"
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	db, err := repository.Open(dbPath, migrationsDir)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// MCP_HTTP_ADDR selects the remote transport: every request then
	// authenticates by user API key (PAPERVIZ_API_KEY is stdio-only).
	if addr := os.Getenv("MCP_HTTP_ADDR"); addr != "" {
		srv := &http.Server{
			Addr:              addr,
			Handler:           papervizMCP.NewHTTPServer(db).Handler(),
			ReadHeaderTimeout: 5 * time.Second,
		}
		log.Fatalf("MCP HTTP server failed: %v", srv.ListenAndServe())
	}

	papervizAPIKey := os.Getenv("PAPERVIZ_API_KEY")
	if papervizAPIKey == "" {
		log.Fatal("PAPERVIZ_API_KEY environment variable is required")
	}

	srv := papervizMCP.NewMCPServer(db, papervizAPIKey)

	if err := srv.Server().Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}
