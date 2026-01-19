package transport

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// InitialBackoff is the initial backoff duration.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	MaxBackoff time.Duration

	// Multiplier is the backoff multiplier.
	Multiplier float64

	// RetryableStatusCodes are HTTP status codes that should trigger a retry.
	RetryableStatusCodes []int
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2.0,
		RetryableStatusCodes: []int{
			429, // Too Many Requests
			500, // Internal Server Error
			502, // Bad Gateway
			503, // Service Unavailable
			504, // Gateway Timeout
		},
	}
}

// RetryTransport wraps a Transport with retry logic.
type RetryTransport struct {
	transport Transport
	config    *RetryConfig
}

// NewRetryTransport creates a new RetryTransport.
func NewRetryTransport(transport Transport, config *RetryConfig) *RetryTransport {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &RetryTransport{
		transport: transport,
		config:    config,
	}
}

// Do executes a request with retry logic.
func (r *RetryTransport) Do(ctx context.Context, req *Request) (*Response, error) {
	var lastErr error
	var resp *Response

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Execute request
		resp, lastErr = r.transport.Do(ctx, req)

		// If no error and successful response, return immediately
		if lastErr == nil && !r.shouldRetry(resp) {
			return resp, nil
		}

		// Don't retry on last attempt
		if attempt == r.config.MaxRetries {
			break
		}

		// Calculate backoff
		backoff := r.calculateBackoff(attempt)

		// Wait before retry
		select {
		case <-time.After(backoff):
			// Continue to next attempt
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", r.config.MaxRetries+1, lastErr)
	}

	return resp, nil
}

// shouldRetry determines if a response should trigger a retry.
func (r *RetryTransport) shouldRetry(resp *Response) bool {
	if resp == nil {
		return true
	}

	for _, code := range r.config.RetryableStatusCodes {
		if resp.StatusCode == code {
			return true
		}
	}

	return false
}

// calculateBackoff calculates the backoff duration for a given attempt.
func (r *RetryTransport) calculateBackoff(attempt int) time.Duration {
	backoff := float64(r.config.InitialBackoff) * math.Pow(r.config.Multiplier, float64(attempt))

	if backoff > float64(r.config.MaxBackoff) {
		backoff = float64(r.config.MaxBackoff)
	}

	return time.Duration(backoff)
}

// SetCSRFToken sets the CSRF token on the underlying transport.
func (r *RetryTransport) SetCSRFToken(token string) {
	r.transport.SetCSRFToken(token)
}

// GetCSRFToken returns the CSRF token from the underlying transport.
func (r *RetryTransport) GetCSRFToken() string {
	return r.transport.GetCSRFToken()
}

// Close closes the underlying transport.
func (r *RetryTransport) Close() {
	r.transport.Close()
}
