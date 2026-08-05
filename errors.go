package tracore

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Synthetic client-side error codes. Set on TracoreError.Code for failures that
// never produced an HTTP response, so retry/breaker layers can classify without
// restating string literals.
const (
	// ErrCodeTransport marks a network, timeout, or encoding failure (retryable).
	ErrCodeTransport = "transport_error"
	// ErrCodeConfig marks invalid client-side input (permanent; a caller bug).
	ErrCodeConfig = "config_error"
	// ErrCodePollingTimeout marks a poll loop that exhausted MaxAttempts without
	// the run reaching a terminal status. The run is still live server-side.
	ErrCodePollingTimeout = "polling_timeout"
)

// Server-side `error.code` vocabulary, mirroring the contract's ApiErrorCode
// schema. The wire field stays a free-form string — new codes are not a breaking
// change — so match on these constants rather than an exhaustive switch.
// Note: key-validation failures arrive as ErrCodeValidation with
// `Details["code"] == "invalid_key"`, not as a top-level code. The last five
// constants surface only on internal billing surfaces and the Stripe webhook,
// never on public API endpoints.
const (
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeValidation          = "VALIDATION_ERROR"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeInternal            = "INTERNAL_ERROR"
	ErrCodePlanLimitExceeded   = "plan_limit_exceeded"
	ErrCodeRateLimited         = "rate_limited"
	ErrCodeForbiddenKey        = "forbidden_key"
	ErrCodeInvalidProvider     = "invalid_provider"
	ErrCodeInvalidModel        = "invalid_model"
	ErrCodeNoKey               = "no_key"
	ErrCodeNotImplemented      = "not_implemented"
	ErrCodeEmailNotVerified    = "email_not_verified"
	ErrCodeStripeNotConfigured = "stripe_not_configured"
	ErrCodeNoStripeCustomer    = "no_stripe_customer"
	ErrCodeCheckoutMissingUser = "checkout_missing_user"
	ErrCodeInvalidSignature    = "invalid_signature"
)

// TracoreError is returned by every SDK call that fails. For HTTP failures Status
// is the response status code and Code/Message/Details come from the API's
// `{ "error": { code, message, details } }` envelope. For client-side failures
// (transport, encoding) Status is 0 and Code is ErrCodeTransport.
type TracoreError struct {
	// Status is the HTTP status code, or 0 for a client-side failure.
	Status int
	// Code is the machine-readable error code from the API, when provided.
	Code string
	// Message is the human-readable error message.
	Message string
	// Details carries additional structured error context, when provided.
	Details map[string]interface{}

	// cause is the underlying error for client-side failures, reachable via
	// Unwrap so errors.Is/As see through the SDK wrapper (context.Canceled,
	// net.Error timeouts, caller sentinels round-tripping the transport).
	cause error
}

func (e *TracoreError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("tracore: %s (status %d, code %q)", e.Message, e.Status, e.Code)
	}
	return fmt.Sprintf("tracore: %s (status %d)", e.Message, e.Status)
}

// Unwrap returns the underlying cause (nil for pure HTTP envelope errors), making
// the wrapper transparent to errors.Is and errors.As.
func (e *TracoreError) Unwrap() error {
	return e.cause
}

// AsTracoreError unwraps err into a *TracoreError if it is one. It is a thin
// convenience over errors.As for the common call-site check.
//
//	if te, ok := tracore.AsTracoreError(err); ok && te.Status == 404 { ... }
func AsTracoreError(err error) (*TracoreError, bool) {
	var te *TracoreError
	if errors.As(err, &te) {
		return te, true
	}
	return nil, false
}

// normalizeError builds a *TracoreError from a non-2xx response. It parses the
// standard API error envelope and falls back to a status-only message when the
// body is empty or not an envelope.
func normalizeError(status int, body []byte) *TracoreError {
	var env struct {
		Error struct {
			Code    string                 `json:"code"`
			Message string                 `json:"message"`
			Details map[string]interface{} `json:"details"`
		} `json:"error"`
	}
	if len(body) > 0 && json.Unmarshal(body, &env) == nil && env.Error.Message != "" {
		return &TracoreError{
			Status:  status,
			Code:    env.Error.Code,
			Message: env.Error.Message,
			Details: env.Error.Details,
		}
	}
	return &TracoreError{Status: status, Message: fmt.Sprintf("request failed with status %d", status)}
}

// transportError wraps a client-side (non-HTTP) failure, keeping the original
// error reachable via Unwrap.
func transportError(err error) *TracoreError {
	return &TracoreError{Code: ErrCodeTransport, Message: err.Error(), cause: err}
}

// decode resolves an oapi-codegen response into either the typed success payload
// or a normalized error. payload is the response's JSON2xx field (nil unless the
// status matched a documented success), status/body come from the same response.
func decode[T any](payload *T, status int, body []byte) (*T, error) {
	if payload == nil {
		return nil, normalizeError(status, body)
	}
	return payload, nil
}
