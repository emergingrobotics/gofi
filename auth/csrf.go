package auth

import (
	"net/http"
	"sync/atomic"
)

// CSRFHandler handles CSRF token storage and updates.
type CSRFHandler struct {
	token atomic.Value // stores string
}

// NewCSRFHandler creates a new CSRFHandler.
func NewCSRFHandler() *CSRFHandler {
	h := &CSRFHandler{}
	h.token.Store("")
	return h
}

// Get returns the current CSRF token.
func (c *CSRFHandler) Get() string {
	if val := c.token.Load(); val != nil {
		if token, ok := val.(string); ok {
			return token
		}
	}
	return ""
}

// Set sets the CSRF token.
func (c *CSRFHandler) Set(token string) {
	c.token.Store(token)
}

// UpdateFromResponse updates the CSRF token from an HTTP response.
func (c *CSRFHandler) UpdateFromResponse(resp *http.Response) {
	if resp == nil {
		return
	}

	// Check for CSRF token in headers
	if token := resp.Header.Get("X-CSRF-Token"); token != "" {
		c.Set(token)
		return
	}

	// Also check cookies for CSRF token
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrf_token" || cookie.Name == "X-CSRF-Token" {
			c.Set(cookie.Value)
			return
		}
	}
}
