package tracore

import (
	"encoding/json"
	"errors"
	"fmt"
)

// TracoreError is returned by every SDK call that fails. For HTTP failures Status
// is the response status code and Code/Message/Details come from the API's
// `{ "error": { code, message, details } }` envelope. For client-side failures
// (transport, encoding) Status is 0 and Code is "transport_error".
type TracoreError struct {
	// Status is the HTTP status code, or 0 for a client-side failure.
	Status int
	// Code is the machine-readable error code from the API, when provided.
	Code string
	// Message is the human-readable error message.
	Message string
	// Details carries additional structured error context, when provided.
	Details map[string]interface{}
}

func (e *TracoreError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("tracore: %s (status %d, code %q)", e.Message, e.Status, e.Code)
	}
	return fmt.Sprintf("tracore: %s (status %d)", e.Message, e.Status)
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

// transportError wraps a client-side (non-HTTP) failure.
func transportError(err error) *TracoreError {
	return &TracoreError{Code: "transport_error", Message: err.Error()}
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
