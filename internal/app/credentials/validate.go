// Package credentials — validation lives beside the resolver so a key check
// reuses the same decrypt-and-bind path as a real analysis call.
package credentials

import (
	"context"

	"paperviz/internal/external"
)

// ValidationResult is the safe wire outcome of a key check. It never carries
// the key, the provider's raw body, or any credential material.
type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// PingFunc attempts one tiny live call with a bound client. Production uses
// DefaultPing; tests substitute a stub so no provider is ever contacted.
type PingFunc func(ctx context.Context, client *external.LLM) error

// DefaultPing sends a one-token prompt through the shared transport, so a
// validate request inherits its retry budget and concurrency slot.
func DefaultPing(ctx context.Context, client *external.LLM) error {
	_, err := client.Generate(ctx, "ping", false, 1)
	return err
}

// Validator proves one stored credential works against its provider.
type Validator struct {
	resolver *Resolver
	ping     PingFunc
}

// NewValidator binds a Validator to a resolver. A nil ping selects the live
// DefaultPing; tests pass a stub.
func NewValidator(resolver *Resolver, ping PingFunc) *Validator {
	if ping == nil {
		ping = DefaultPing
	}
	return &Validator{resolver: resolver, ping: ping}
}

// Validate resolves the credential scoped to userID and pings the provider.
// ErrNotFound and ErrCredentialUnreadable are returned as errors for the
// handler to map to 404/409; provider failures return a result instead.
func (v *Validator) Validate(ctx context.Context, userID, id string) (ValidationResult, error) {
	if v == nil || v.resolver == nil {
		return ValidationResult{Valid: false, Reason: "unavailable", Message: "Validation is unavailable."}, nil
	}
	client, err := v.resolver.ForCredential(ctx, userID, id)
	if err != nil {
		return ValidationResult{}, err
	}
	if err := v.ping(ctx, client); err != nil {
		return classifyPingError(err), nil
	}
	return ValidationResult{Valid: true}, nil
}

// classifyPingError maps a provider failure onto a UI-safe reason without
// echoing the provider's raw message, which can contain key fragments.
func classifyPingError(err error) ValidationResult {
	switch {
	case external.IsAuthError(err):
		return ValidationResult{Valid: false, Reason: "invalid_key", Message: "That key was rejected by the provider. Check it and try again."}
	case external.IsQuotaError(err):
		return ValidationResult{Valid: false, Reason: "quota_exceeded", Message: "The provider reports exhausted quota for that key."}
	default:
		return ValidationResult{Valid: false, Reason: "unavailable", Message: "The provider could not be reached. Retry shortly."}
	}
}
