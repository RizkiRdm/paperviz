package mcp

import "fmt"

// MCPError is a structured MCP-level error with a numeric code and message.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error returns the JSON representation of the MCP error.
func (e *MCPError) Error() string {
	return fmt.Sprintf("mcp error %d: %s", e.Code, e.Message)
}

// Sentinel errors for common MCP failure modes.
var (
	ErrAuthFailed         = &MCPError{Code: -32001, Message: "authentication failed"}
	ErrRateLimited        = &MCPError{Code: -32002, Message: "rate limit exceeded"}
	ErrSizeLimit          = &MCPError{Code: -32003, Message: "input size limit exceeded"}
	ErrJobLimit           = &MCPError{Code: -32004, Message: "concurrent job limit reached"}
	ErrTimeout            = &MCPError{Code: -32005, Message: "request timed out"}
	ErrInvalidInput       = &MCPError{Code: -32006, Message: "invalid input"}
	ErrDocumentNotFound   = &MCPError{Code: -32007, Message: "document not found"}
	ErrInternal           = &MCPError{Code: -32008, Message: "internal error"}
)

// NewAuthFailedError returns an auth failure error with a custom message.
func NewAuthFailedError(msg string) *MCPError {
	return &MCPError{Code: ErrAuthFailed.Code, Message: msg}
}

// NewRateLimitError returns a rate limit error with a custom message.
func NewRateLimitError(msg string) *MCPError {
	return &MCPError{Code: ErrRateLimited.Code, Message: msg}
}

// NewSizeLimitError returns a size limit error with a custom message.
func NewSizeLimitError(msg string) *MCPError {
	return &MCPError{Code: ErrSizeLimit.Code, Message: msg}
}

// NewJobLimitError returns a job limit error with a custom message.
func NewJobLimitError(msg string) *MCPError {
	return &MCPError{Code: ErrJobLimit.Code, Message: msg}
}
