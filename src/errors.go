package gofi

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
var (
	// ErrNotConnected is returned when an operation requires a connection but the client is not connected.
	ErrNotConnected = errors.New("not connected to UniFi controller")

	// ErrAlreadyConnected is returned when Connect() is called but the client is already connected.
	ErrAlreadyConnected = errors.New("already connected to UniFi controller")

	// ErrAuthenticationFailed is returned when login credentials are invalid.
	ErrAuthenticationFailed = errors.New("authentication failed: invalid credentials")

	// ErrSessionExpired is returned when the session token has expired.
	ErrSessionExpired = errors.New("session expired")

	// ErrInvalidCSRFToken is returned when the CSRF token is invalid or missing.
	ErrInvalidCSRFToken = errors.New("invalid or missing CSRF token")

	// ErrNotFound is returned when a requested resource is not found.
	ErrNotFound = errors.New("resource not found")

	// ErrPermissionDenied is returned when the user lacks permission for an operation.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrAlreadyExists is returned when attempting to create a resource that already exists.
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidRequest is returned when the request is malformed or invalid.
	ErrInvalidRequest = errors.New("invalid request")

	// ErrRateLimited is returned when too many requests have been made.
	ErrRateLimited = errors.New("rate limited: too many requests")

	// ErrServerError is returned when the server encounters an internal error.
	ErrServerError = errors.New("server error")

	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("operation timed out")

	// ErrInvalidConfig is returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")
)

// APIError represents an error returned by the UniFi API.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int

	// RC is the UniFi response code (e.g., "error", "ok").
	RC string

	// Message is the error message from the API.
	Message string

	// Endpoint is the API endpoint that returned the error.
	Endpoint string

	// Err is the underlying sentinel error, if any.
	Err error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error [%d]: %s (rc=%s, endpoint=%s)", e.StatusCode, e.Message, e.RC, e.Endpoint)
	}
	return fmt.Sprintf("API error [%d]: rc=%s, endpoint=%s", e.StatusCode, e.RC, e.Endpoint)
}

// Is implements error comparison for errors.Is().
func (e *APIError) Is(target error) bool {
	if e.Err != nil && errors.Is(e.Err, target) {
		return true
	}

	// Check if target is an APIError with matching properties
	if apiErr, ok := target.(*APIError); ok {
		if apiErr.StatusCode != 0 && apiErr.StatusCode != e.StatusCode {
			return false
		}
		if apiErr.RC != "" && apiErr.RC != e.RC {
			return false
		}
		return true
	}

	return false
}

// Unwrap returns the underlying error.
func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new APIError with the given parameters.
func NewAPIError(statusCode int, rc, message, endpoint string) *APIError {
	err := &APIError{
		StatusCode: statusCode,
		RC:         rc,
		Message:    message,
		Endpoint:   endpoint,
	}

	// Map to sentinel errors based on status code and RC
	switch statusCode {
	case 401:
		err.Err = ErrAuthenticationFailed
	case 403:
		if rc == "error_invalid_csrf_token" {
			err.Err = ErrInvalidCSRFToken
		} else {
			err.Err = ErrPermissionDenied
		}
	case 404:
		err.Err = ErrNotFound
	case 409:
		err.Err = ErrAlreadyExists
	case 429:
		err.Err = ErrRateLimited
	case 500, 502, 503, 504:
		err.Err = ErrServerError
	default:
		if rc == "error" || rc == "error_invalid" {
			err.Err = ErrInvalidRequest
		}
	}

	return err
}

// ValidationError represents a validation error for input data.
type ValidationError struct {
	// Field is the name of the field that failed validation.
	Field string

	// Message is the validation error message.
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
