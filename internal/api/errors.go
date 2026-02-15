// Package api provides the HTTP client layer for communicating with sable-api.
package api

import (
	"errors"
	"fmt"
)

// ResponseError represents an error response from the Sable API.
type ResponseError struct {
	StatusCode int
	Message    string
	Hint       string
}

func (e *ResponseError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("API error %d: %s (hint: %s)", e.StatusCode, e.Message, e.Hint)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// NotFoundError indicates a 404 response.
type NotFoundError struct{ ResponseError }

// UnauthorizedError indicates a 401 response.
type UnauthorizedError struct{ ResponseError }

// ForbiddenError indicates a 403 response.
type ForbiddenError struct{ ResponseError }

// ConflictError indicates a 409 response.
type ConflictError struct{ ResponseError }

// ValidationError indicates a 422 response.
type ValidationError struct{ ResponseError }

// As extracts a typed API error from an error chain using Go 1.26 errors.AsType.
//
//	if nf, ok := api.As[*NotFoundError](err); ok { ... }
func As[T error](err error) (T, bool) {
	return errors.AsType[T](err)
}

// NewFromStatus creates the appropriate typed error for a given HTTP status code.
func NewFromStatus(statusCode int, message, hint string) error {
	base := ResponseError{StatusCode: statusCode, Message: message, Hint: hint}
	switch statusCode {
	case 401:
		return &UnauthorizedError{base}
	case 403:
		return &ForbiddenError{base}
	case 404:
		return &NotFoundError{base}
	case 409:
		return &ConflictError{base}
	case 422:
		return &ValidationError{base}
	default:
		return &base
	}
}
