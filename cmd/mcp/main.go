// Command mcp is PaperViz's MCP server entrypoint. It exposes 5 tools
// (ingest_document, search_documents, get_document, get_figures, get_evidence)
// over stdio transport. No LLM calls in any tool path, so no model key is
// required — configuration comes from environment variables and the same DB
// as the main server.
package main

import (
	"context"
	"log"
	"os"

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

	papervizAPIKey := os.Getenv("PAPERVIZ_API_KEY")
	if papervizAPIKey == "" {
		log.Fatal("PAPERVIZ_API_KEY environment variable is required")
	}

	db, err := repository.Open(dbPath, migrationsDir)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	srv := papervizMCP.NewMCPServer(db, papervizAPIKey)

	if err := srv.Server().Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}
