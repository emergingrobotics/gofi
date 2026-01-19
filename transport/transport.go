package transport

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync/atomic"
)

// Transport represents an HTTP transport for making requests.
type Transport interface {
	// Do executes an HTTP request.
	Do(ctx context.Context, req *Request) (*Response, error)

	// SetCSRFToken sets the CSRF token for subsequent requests.
	SetCSRFToken(token string)

	// GetCSRFToken returns the current CSRF token.
	GetCSRFToken() string

	// Close closes any idle connections.
	Close()
}

// httpTransport implements the Transport interface.
type httpTransport struct {
	client    *http.Client
	baseURL   *url.URL
	csrfToken atomic.Value // stores string
	userAgent string
}

// New creates a new HTTP transport.
func New(config *Config, opts ...Option) (Transport, error) {
	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Parse base URL
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Create cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Create HTTP transport
	transport := &http.Transport{
		TLSClientConfig:     config.TLSConfig,
		MaxIdleConns:        config.MaxIdleConns,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}

	// If TLS config not provided but we need to skip verification
	if transport.TLSClientConfig == nil && baseURL.Scheme == "https" {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: false, // Default to secure
		}
	}

	// Create HTTP client
	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects automatically
			return http.ErrUseLastResponse
		},
	}

	t := &httpTransport{
		client:    client,
		baseURL:   baseURL,
		userAgent: config.UserAgent,
	}

	// Initialize CSRF token as empty string
	t.csrfToken.Store("")

	return t, nil
}

// Do executes an HTTP request.
func (t *httpTransport) Do(ctx context.Context, req *Request) (*Response, error) {
	// Build full URL
	fullURL, err := t.baseURL.Parse(req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Serialize body if present
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	httpReq.Header.Set("Accept", "application/json")
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if t.userAgent != "" {
		httpReq.Header.Set("User-Agent", t.userAgent)
	}

	// Add CSRF token if available
	if token := t.GetCSRFToken(); token != "" {
		httpReq.Header.Set("X-CSRF-Token", token)
	}

	// Add custom headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Execute request
	httpResp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for CSRF token in response headers
	if csrfToken := httpResp.Header.Get("X-CSRF-Token"); csrfToken != "" {
		t.SetCSRFToken(csrfToken)
	}

	// Create response
	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Body:       body,
		Headers:    httpResp.Header,
	}

	return resp, nil
}

// SetCSRFToken sets the CSRF token.
func (t *httpTransport) SetCSRFToken(token string) {
	t.csrfToken.Store(token)
}

// GetCSRFToken returns the current CSRF token.
func (t *httpTransport) GetCSRFToken() string {
	if val := t.csrfToken.Load(); val != nil {
		if token, ok := val.(string); ok {
			return token
		}
	}
	return ""
}

// Close closes any idle connections.
func (t *httpTransport) Close() {
	t.client.CloseIdleConnections()
}
