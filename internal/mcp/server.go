package mcp

import (
	"database/sql"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServer holds shared dependencies for all MCP tool handlers.
type MCPServer struct {
	db          *sql.DB
	apiKey      string
	rateLimiter *RateLimiter
	jobLimiter  *JobLimiter
	mcpServer   *mcp.Server
}

func (s *MCPServer) Server() *mcp.Server { return s.mcpServer }

// NewMCPServer creates the MCP server and registers all research tools.
func NewMCPServer(db *sql.DB, apiKey string) *MCPServer {
	mcpSrv := &MCPServer{
		db:          db,
		apiKey:      apiKey,
		rateLimiter: NewRateLimiter(),
		jobLimiter:  NewJobLimiter(3),
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "paperviz",
		Version: "0.1.0",
	}, nil)

	mcpSrv.mcpServer = server

	registerTools(server, mcpSrv)

	return mcpSrv
}
