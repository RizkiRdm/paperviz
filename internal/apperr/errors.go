// Package apperr defines typed error sentinels for the application layer.
// Handlers map these to HTTP status codes; services return them to signal
// domain failures without leaking infrastructure details.
package apperr

import (
	"errors"
	"fmt"
)

// Error categories — each maps to an HTTP status in handler layer.
var (
	// Validation signals bad input from the client (400).
	Validation = errors.New("validation failed")
	// NotFound signals the requested resource does not exist (404).
	NotFound = errors.New("not found")
	// Permission signals the caller lacks ownership or access (403).
	Permission = errors.New("permission denied")
	// Processing signals a pipeline or background job failure (422/500).
	Processing = errors.New("processing failed")
	// Provider signals an external service error (LLM, storage) (502).
	Provider = errors.New("provider error")
	// Internal signals an unexpected system failure (500).
	Internal = errors.New("internal error")
)

// HTTPStatus maps an apperr sentinel to the appropriate HTTP status code.
// Returns 500 for unrecognized errors.
func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, Validation):
		return 400
	case errors.Is(err, NotFound):
		return 404
	case errors.Is(err, Permission):
		return 403
	case errors.Is(err, Processing):
		return 422
	case errors.Is(err, Provider):
		return 502
	default:
		return 500
	}
}

// Code maps an apperr sentinel to a snake_case error code for JSON responses.
// Returns "internal_error" for unrecognized errors.
func Code(err error) string {
	switch {
	case errors.Is(err, Validation):
		return "validation_error"
	case errors.Is(err, NotFound):
		return "not_found"
	case errors.Is(err, Permission):
		return "forbidden"
	case errors.Is(err, Processing):
		return "processing_error"
	case errors.Is(err, Provider):
		return "provider_error"
	default:
		return "internal_error"
	}
}

// Wrap wraps a sentinel error with context while preserving the sentinel chain.
// Usage: apperr.Wrap(apperr.NotFound, "document %s", id)
func Wrap(sentinel error, msg string, args ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(msg, args...), sentinel)
}
