package mcp

import (
	"database/sql"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"paperviz/internal/external"
)

// NewMCPServer creates the MCP server and registers all research tools.
func NewMCPServer(db *sql.DB, gemini *external.GeminiClient) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "paperviz",
		Version: "0.1.0",
	}, nil)

	registerTools(server, db, gemini)

	return server
}
