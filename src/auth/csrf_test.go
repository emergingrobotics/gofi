package auth

import (
	"net/http"
	"sync"
	"testing"
)

func TestNewCSRFHandler(t *testing.T) {
	h := NewCSRFHandler()
	if h == nil {
		t.Fatal("NewCSRFHandler() returned nil")
	}

	// Initial token should be empty
	if token := h.Get(); token != "" {
		t.Errorf("Get() = %s, want empty string", token)
	}
}

func TestCSRFHandler_SetGet(t *testing.T) {
	h := NewCSRFHandler()

	// Set token
	h.Set("test-csrf-token")

	// Get token
	got := h.Get()
	if got != "test-csrf-token" {
		t.Errorf("Get() = %s, want test-csrf-token", got)
	}
}

func TestCSRFHandler_ConcurrentAccess(t *testing.T) {
	h := NewCSRFHandler()

	var wg sync.WaitGroup
	concurrency := 100

	// Concurrent writes
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			token := "token-" + string(rune(n))
			h.Set(token)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = h.Get()
		}()
	}

	wg.Wait()

	// Should have some token set (last write wins)
	if token := h.Get(); token == "" {
		t.Error("Expected some token to be set after concurrent operations")
	}
}

func TestCSRFHandler_UpdateFromResponse_Header(t *testing.T) {
	h := NewCSRFHandler()

	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("X-CSRF-Token", "header-csrf-token")

	h.UpdateFromResponse(resp)

	got := h.Get()
	if got != "header-csrf-token" {
		t.Errorf("Get() = %s, want header-csrf-token", got)
	}
}

func TestCSRFHandler_UpdateFromResponse_Cookie(t *testing.T) {
	h := NewCSRFHandler()

	// Create a response with a cookie
	cookie := &http.Cookie{
		Name:  "csrf_token",
		Value: "cookie-csrf-token",
	}

	// Create a response with the cookie
	resp := &http.Response{
		Header: make(http.Header),
	}
	http.SetCookie(&httpWriter{header: resp.Header}, cookie)

	// Create a request to associate with the response
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp.Request = req

	h.UpdateFromResponse(resp)

	// Should extract CSRF token from cookie
	if got := h.Get(); got != "cookie-csrf-token" {
		t.Errorf("Get() = %s, want cookie-csrf-token", got)
	}
}

func TestCSRFHandler_UpdateFromResponse_NilResponse(t *testing.T) {
	h := NewCSRFHandler()
	h.Set("initial-token")

	// Should not panic with nil response
	h.UpdateFromResponse(nil)

	// Token should remain unchanged
	if got := h.Get(); got != "initial-token" {
		t.Errorf("Get() = %s, want initial-token (should not change on nil response)", got)
	}
}

func TestCSRFHandler_UpdateFromResponse_NoToken(t *testing.T) {
	h := NewCSRFHandler()
	h.Set("initial-token")

	resp := &http.Response{
		Header: http.Header{},
	}

	h.UpdateFromResponse(resp)

	// Token should remain unchanged
	if got := h.Get(); got != "initial-token" {
		t.Errorf("Get() = %s, want initial-token (should not change when no token in response)", got)
	}
}

// httpWriter is a minimal ResponseWriter for testing
type httpWriter struct {
	header http.Header
}

func (w *httpWriter) Header() http.Header {
	return w.header
}

func (w *httpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *httpWriter) WriteHeader(statusCode int) {}
