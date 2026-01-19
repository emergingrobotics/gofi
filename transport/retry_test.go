package transport

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}

	if config.InitialBackoff != 100*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 100ms", config.InitialBackoff)
	}

	if config.MaxBackoff != 5*time.Second {
		t.Errorf("MaxBackoff = %v, want 5s", config.MaxBackoff)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("Multiplier = %f, want 2.0", config.Multiplier)
	}

	// Check retryable status codes
	expectedCodes := []int{429, 500, 502, 503, 504}
	if len(config.RetryableStatusCodes) != len(expectedCodes) {
		t.Errorf("len(RetryableStatusCodes) = %d, want %d", len(config.RetryableStatusCodes), len(expectedCodes))
	}
}

func TestRetryTransport_SuccessFirstAttempt(t *testing.T) {
	var attempts int32

	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create transport with retry
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	baseTransport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer baseTransport.Close()

	retryTransport := NewRetryTransport(baseTransport, DefaultRetryConfig())

	// Execute request
	req := NewRequest("GET", "/api/test")
	resp, err := retryTransport.Do(context.Background(), req)

	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	// Should only attempt once
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1", atomic.LoadInt32(&attempts))
	}
}

func TestRetryTransport_Retry500(t *testing.T) {
	var attempts int32

	// Create test server that fails twice then succeeds
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create transport with retry
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	baseTransport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer baseTransport.Close()

	retryConfig := DefaultRetryConfig()
	retryConfig.InitialBackoff = 10 * time.Millisecond // Speed up test
	retryTransport := NewRetryTransport(baseTransport, retryConfig)

	// Execute request
	req := NewRequest("GET", "/api/test")
	resp, err := retryTransport.Do(context.Background(), req)

	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	// Should attempt 3 times (2 failures + 1 success)
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", atomic.LoadInt32(&attempts))
	}
}

func TestRetryTransport_ExhaustedRetries(t *testing.T) {
	var attempts int32

	// Create test server that always fails
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create transport with retry
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	baseTransport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer baseTransport.Close()

	retryConfig := DefaultRetryConfig()
	retryConfig.MaxRetries = 2
	retryConfig.InitialBackoff = 10 * time.Millisecond
	retryTransport := NewRetryTransport(baseTransport, retryConfig)

	// Execute request
	req := NewRequest("GET", "/api/test")
	resp, err := retryTransport.Do(context.Background(), req)

	// Should not return error (just the failed response)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if resp.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", resp.StatusCode)
	}

	// Should attempt MaxRetries + 1 times
	expectedAttempts := int32(retryConfig.MaxRetries + 1)
	if atomic.LoadInt32(&attempts) != expectedAttempts {
		t.Errorf("attempts = %d, want %d", atomic.LoadInt32(&attempts), expectedAttempts)
	}
}

func TestRetryTransport_NoRetryOn4xx(t *testing.T) {
	var attempts int32

	// Create test server that returns 404
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create transport with retry
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	baseTransport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer baseTransport.Close()

	retryTransport := NewRetryTransport(baseTransport, DefaultRetryConfig())

	// Execute request
	req := NewRequest("GET", "/api/test")
	resp, err := retryTransport.Do(context.Background(), req)

	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", resp.StatusCode)
	}

	// Should only attempt once (no retry on 4xx)
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1 (4xx should not trigger retry)", atomic.LoadInt32(&attempts))
	}
}

func TestRetryTransport_ContextCancellation(t *testing.T) {
	var attempts int32

	// Create test server that always fails
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create transport with retry
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	baseTransport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer baseTransport.Close()

	retryConfig := DefaultRetryConfig()
	retryConfig.InitialBackoff = 100 * time.Millisecond
	retryTransport := NewRetryTransport(baseTransport, retryConfig)

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Execute request
	req := NewRequest("GET", "/api/test")
	_, err = retryTransport.Do(ctx, req)

	if err == nil {
		t.Fatal("Do() should return error when context is cancelled")
	}

	// Should only attempt once or twice (cancelled during backoff)
	if attempts := atomic.LoadInt32(&attempts); attempts > 2 {
		t.Errorf("attempts = %d, should be <=2 (cancelled during retry)", attempts)
	}
}

func TestRetryTransport_BackoffCalculation(t *testing.T) {
	retryConfig := &RetryConfig{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2.0,
	}

	rt := &RetryTransport{config: retryConfig}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1 * time.Second}, // Capped at MaxBackoff
		{5, 1 * time.Second}, // Still capped
	}

	for _, tt := range tests {
		got := rt.calculateBackoff(tt.attempt)
		if got != tt.want {
			t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestRetryTransport_CSRFTokenPassthrough(t *testing.T) {
	config := DefaultConfig("https://192.168.1.1")
	baseTransport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer baseTransport.Close()

	retryTransport := NewRetryTransport(baseTransport, DefaultRetryConfig())

	// Set CSRF token
	retryTransport.SetCSRFToken("test-token")

	// Verify it's passed through
	if token := retryTransport.GetCSRFToken(); token != "test-token" {
		t.Errorf("GetCSRFToken() = %s, want test-token", token)
	}
}
