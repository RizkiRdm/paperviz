package mcp

import (
	"database/sql"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServer holds shared dependencies for all MCP tool handlers.
type MCPServer struct {
	db          *sql.DB
	apiKey      string
	userID      string
	scoped      bool
	rateLimiter *RateLimiter
	jobLimiter  *JobLimiter
	mcpServer   *mcp.Server
}

// scoped reports whether tool calls are owner-scoped (remote transport).
// The stdio transport is deliberately unscoped — its parity split is
// documented in docs/mcp-parity.md.
func (s *MCPServer) scopedMode() bool { return s.scoped }

// Server returns the underlying SDK server for transport wiring.
func (s *MCPServer) Server() *mcp.Server { return s.mcpServer }

// NewMCPServer creates the MCP server and registers all research tools.
// Stdio mode: unscoped, keyed by the single env API key.
func NewMCPServer(db *sql.DB, apiKey string) *MCPServer {
	return newServerWithDeps(db, apiKey, "", false, NewRateLimiter(), NewJobLimiter(3))
}

// newServerWithDeps builds a server over caller-provided shared limiters so
// the HTTP transport can rate-limit across per-request server instances.
func newServerWithDeps(db *sql.DB, limiterKey, userID string, scoped bool, rl *RateLimiter, jl *JobLimiter) *MCPServer {
	mcpSrv := &MCPServer{
		db:          db,
		apiKey:      limiterKey,
		userID:      userID,
		scoped:      scoped,
		rateLimiter: rl,
		jobLimiter:  jl,
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "paperviz",
		Version: "0.1.0",
	}, nil)

	mcpSrv.mcpServer = server

	registerTools(server, mcpSrv)

	return mcpSrv
}
