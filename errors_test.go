package gofi

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		contains string
	}{
		{
			"with message",
			&APIError{StatusCode: 404, RC: "error", Message: "Not found", Endpoint: "/api/test"},
			"API error [404]: Not found (rc=error, endpoint=/api/test)",
		},
		{
			"without message",
			&APIError{StatusCode: 500, RC: "error", Endpoint: "/api/test"},
			"API error [500]: rc=error, endpoint=/api/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.contains {
				t.Errorf("Error() = %q, want %q", got, tt.contains)
			}
		})
	}
}

func TestAPIError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    *APIError
		target error
		want   bool
	}{
		{
			"matches sentinel via Unwrap",
			&APIError{StatusCode: 404, Err: ErrNotFound},
			ErrNotFound,
			true,
		},
		{
			"matches sentinel via status code",
			&APIError{StatusCode: 401, Err: ErrAuthenticationFailed},
			ErrAuthenticationFailed,
			true,
		},
		{
			"does not match different sentinel",
			&APIError{StatusCode: 404, Err: ErrNotFound},
			ErrPermissionDenied,
			false,
		},
		{
			"matches APIError with same status code",
			&APIError{StatusCode: 404, RC: "error"},
			&APIError{StatusCode: 404},
			true,
		},
		{
			"does not match APIError with different status code",
			&APIError{StatusCode: 404},
			&APIError{StatusCode: 500},
			false,
		},
		{
			"matches APIError with same RC",
			&APIError{RC: "error_invalid"},
			&APIError{RC: "error_invalid"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Is(tt.err, tt.target)
			if got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	underlying := ErrNotFound
	err := &APIError{
		StatusCode: 404,
		Err:        underlying,
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != underlying {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
	}

	// Test with errors.Is through wrapping
	if !errors.Is(err, ErrNotFound) {
		t.Error("errors.Is() should return true for wrapped ErrNotFound")
	}
}

func TestNewAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		rc         string
		message    string
		endpoint   string
		wantErr    error
	}{
		{"401 maps to ErrAuthenticationFailed", 401, "error", "Invalid credentials", "/api/login", ErrAuthenticationFailed},
		{"403 with CSRF maps to ErrInvalidCSRFToken", 403, "error_invalid_csrf_token", "Bad token", "/api/test", ErrInvalidCSRFToken},
		{"403 maps to ErrPermissionDenied", 403, "error", "Forbidden", "/api/test", ErrPermissionDenied},
		{"404 maps to ErrNotFound", 404, "error", "Not found", "/api/test", ErrNotFound},
		{"409 maps to ErrAlreadyExists", 409, "error", "Conflict", "/api/test", ErrAlreadyExists},
		{"429 maps to ErrRateLimited", 429, "error", "Too many requests", "/api/test", ErrRateLimited},
		{"500 maps to ErrServerError", 500, "error", "Internal error", "/api/test", ErrServerError},
		{"502 maps to ErrServerError", 502, "error", "Bad gateway", "/api/test", ErrServerError},
		{"503 maps to ErrServerError", 503, "error", "Unavailable", "/api/test", ErrServerError},
		{"504 maps to ErrServerError", 504, "error", "Timeout", "/api/test", ErrServerError},
		{"error_invalid maps to ErrInvalidRequest", 400, "error_invalid", "Bad request", "/api/test", ErrInvalidRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.statusCode, tt.rc, tt.message, tt.endpoint)

			if err.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", err.StatusCode, tt.statusCode)
			}

			if err.RC != tt.rc {
				t.Errorf("RC = %s, want %s", err.RC, tt.rc)
			}

			if err.Message != tt.message {
				t.Errorf("Message = %s, want %s", err.Message, tt.message)
			}

			if err.Endpoint != tt.endpoint {
				t.Errorf("Endpoint = %s, want %s", err.Endpoint, tt.endpoint)
			}

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error to wrap %v, but it doesn't", tt.wantErr)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name  string
		err   *ValidationError
		want  string
	}{
		{
			"with field",
			&ValidationError{Field: "username", Message: "required"},
			"validation error: username: required",
		},
		{
			"without field",
			&ValidationError{Message: "invalid input"},
			"validation error: invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("email", "invalid format")

	if err.Field != "email" {
		t.Errorf("Field = %s, want email", err.Field)
	}

	if err.Message != "invalid format" {
		t.Errorf("Message = %s, want 'invalid format'", err.Message)
	}

	expected := "validation error: email: invalid format"
	if err.Error() != expected {
		t.Errorf("Error() = %s, want %s", err.Error(), expected)
	}
}

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors are defined and unique
	sentinelErrors := []error{
		ErrNotConnected,
		ErrAlreadyConnected,
		ErrAuthenticationFailed,
		ErrSessionExpired,
		ErrInvalidCSRFToken,
		ErrNotFound,
		ErrPermissionDenied,
		ErrAlreadyExists,
		ErrInvalidRequest,
		ErrRateLimited,
		ErrServerError,
		ErrTimeout,
		ErrInvalidConfig,
	}

	// Check all are non-nil
	for i, err := range sentinelErrors {
		if err == nil {
			t.Errorf("Sentinel error at index %d is nil", i)
		}
	}

	// Check all have unique error messages
	seen := make(map[string]bool)
	for _, err := range sentinelErrors {
		msg := err.Error()
		if seen[msg] {
			t.Errorf("Duplicate error message: %s", msg)
		}
		seen[msg] = true
	}
}
