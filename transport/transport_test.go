package transport

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	config := DefaultConfig("https://192.168.1.1")
	transport, err := New(config)

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if transport == nil {
		t.Fatal("New() returned nil transport")
	}

	// Cleanup
	transport.Close()
}

func TestNew_InvalidURL(t *testing.T) {
	config := DefaultConfig("://invalid-url")
	_, err := New(config)

	if err == nil {
		t.Fatal("New() should return error for invalid URL")
	}
}

func TestNew_WithOptions(t *testing.T) {
	config := DefaultConfig("https://192.168.1.1")
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	transport, err := New(config,
		WithTimeout(10*time.Second),
		WithTLSConfig(tlsConfig),
		WithUserAgent("test/1.0"),
	)

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Verify options were applied
	httpT := transport.(*httpTransport)
	if httpT.userAgent != "test/1.0" {
		t.Errorf("UserAgent = %s, want test/1.0", httpT.userAgent)
	}

	transport.Close()
}

func TestTransport_Do_GET(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %s, want GET", r.Method)
		}

		if r.URL.Path != "/api/test" {
			t.Errorf("Path = %s, want /api/test", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	// Create transport
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	transport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer transport.Close()

	// Execute request
	req := NewRequest("GET", "/api/test")
	resp, err := transport.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	if !resp.IsSuccess() {
		t.Error("IsSuccess() should return true")
	}

	var result map[string]string
	if err := resp.Parse(&result); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("result[status] = %s, want ok", result["status"])
	}
}

func TestTransport_Do_POST(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}

		// Check Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
		}

		// Read and verify body
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode body: %v", err)
		}

		if body["key"] != "value" {
			t.Errorf("body[key] = %s, want value", body["key"])
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer server.Close()

	// Create transport
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	transport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer transport.Close()

	// Execute request
	req := NewRequest("POST", "/api/create").
		WithBody(map[string]string{"key": "value"})

	resp, err := transport.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
	}
}

func TestTransport_CSRFToken(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return CSRF token in first request
		if r.URL.Path == "/api/login" {
			w.Header().Set("X-CSRF-Token", "test-csrf-token")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Verify CSRF token in subsequent requests
		if r.URL.Path == "/api/test" {
			token := r.Header.Get("X-CSRF-Token")
			if token != "test-csrf-token" {
				t.Errorf("X-CSRF-Token = %s, want test-csrf-token", token)
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create transport
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	transport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer transport.Close()

	// First request to get CSRF token
	req1 := NewRequest("POST", "/api/login")
	resp1, err := transport.Do(context.Background(), req1)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if resp1.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp1.StatusCode)
	}

	// Verify CSRF token was stored
	token := transport.GetCSRFToken()
	if token != "test-csrf-token" {
		t.Errorf("GetCSRFToken() = %s, want test-csrf-token", token)
	}

	// Second request should include CSRF token
	req2 := NewRequest("GET", "/api/test")
	resp2, err := transport.Do(context.Background(), req2)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if resp2.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200 (CSRF token should be valid)", resp2.StatusCode)
	}
}

func TestTransport_SetGetCSRFToken(t *testing.T) {
	config := DefaultConfig("https://192.168.1.1")
	transport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer transport.Close()

	// Initially empty
	if token := transport.GetCSRFToken(); token != "" {
		t.Errorf("GetCSRFToken() = %s, want empty string", token)
	}

	// Set token
	transport.SetCSRFToken("my-token")

	// Verify it was set
	if token := transport.GetCSRFToken(); token != "my-token" {
		t.Errorf("GetCSRFToken() = %s, want my-token", token)
	}
}

func TestTransport_CustomHeaders(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "custom-value" {
			t.Errorf("X-Custom header = %s, want custom-value", r.Header.Get("X-Custom"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create transport
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	transport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer transport.Close()

	// Execute request with custom header
	req := NewRequest("GET", "/api/test").
		WithHeader("X-Custom", "custom-value")

	_, err = transport.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
}

func TestTransport_ContextCancellation(t *testing.T) {
	// Create test server with delay
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create transport
	config := DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	config.Timeout = 5 * time.Second
	transport, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer transport.Close()

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Execute request
	req := NewRequest("GET", "/api/test")
	_, err = transport.Do(ctx, req)

	if err == nil {
		t.Fatal("Do() should return error when context is cancelled")
	}
}
