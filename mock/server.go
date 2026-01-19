package mock

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

// Server is a mock UniFi controller server.
type Server struct {
	server      *httptest.Server
	state       *State
	requireAuth bool
	requireCSRF bool
	scenario    Scenario
}

// NewServer creates a new mock server.
func NewServer(opts ...Option) *Server {
	s := &Server{
		state:       NewState(),
		requireAuth: true,
		requireCSRF: true,
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Create HTTP server with TLS
	s.server = httptest.NewUnstartedServer(s)
	s.server.TLS = &tls.Config{
		InsecureSkipVerify: true,
	}
	s.server.StartTLS()

	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply scenario if set
	if s.scenario != nil {
		if s.scenario.Apply(w, r) {
			return
		}
	}

	// Route requests
	path := r.URL.Path

	// Auth endpoints (no auth required)
	if path == "/api/auth/login" {
		s.handleLogin(w, r)
		return
	}

	if path == "/api/logout" {
		s.handleLogout(w, r)
		return
	}

	if path == "/api/status" {
		s.handleStatus(w, r)
		return
	}

	// All other endpoints require authentication
	if s.requireAuth {
		if !s.isAuthenticated(r) {
			writeUnauthorized(w)
			return
		}
	}

	// Check CSRF token for non-GET requests
	if s.requireCSRF && r.Method != "GET" && r.Method != "HEAD" {
		if !s.validateCSRF(r) {
			writeForbidden(w, "Invalid CSRF token")
			return
		}
	}

	if path == "/api/self" {
		s.handleSelf(w, r)
		return
	}

	// Extract site from path for API calls
	site := ""
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "s" && i+1 < len(parts) {
			site = parts[i+1]
			break
		}
	}

	// Device endpoints
	if strings.Contains(path, "/stat/device") ||
	   strings.Contains(path, "/basicstat/device") ||
	   (strings.Contains(path, "/rest/device/") && r.Method == "PUT") ||
	   strings.Contains(path, "/cmd/devmgr") {
		s.handleDevices(w, r, site)
		return
	}

	// Site endpoints
	if strings.HasPrefix(path, "/api/self/sites") ||
	   strings.Contains(path, "/api/s/") ||
	   strings.Contains(path, "/stat/health") ||
	   strings.Contains(path, "/stat/sysinfo") {
		s.handleSites(w, r, "")
		return
	}

	// Default: 404
	writeNotFound(w)
}

// Close shuts down the server.
func (s *Server) Close() {
	if s.server != nil {
		s.server.Close()
	}
}

// URL returns the server's base URL.
func (s *Server) URL() string {
	if s.server != nil {
		return s.server.URL
	}
	return ""
}

// Host returns the server's host (without scheme).
func (s *Server) Host() string {
	if s.server != nil {
		host, _, _ := net.SplitHostPort(s.server.Listener.Addr().String())
		return host
	}
	return ""
}

// Port returns the server's port.
func (s *Server) Port() int {
	if s.server != nil {
		_, port, _ := net.SplitHostPort(s.server.Listener.Addr().String())
		// Port is a string, parse it
		if p, err := strconv.Atoi(port); err == nil {
			return p
		}
	}
	return 0
}

// State returns the server's state (for test manipulation).
func (s *Server) State() *State {
	return s.state
}

// isAuthenticated checks if the request has a valid session.
func (s *Server) isAuthenticated(r *http.Request) bool {
	// Check for session cookie
	cookie, err := r.Cookie("unifises")
	if err != nil {
		return false
	}

	_, exists := s.state.GetSession(cookie.Value)
	return exists
}

// validateCSRF validates the CSRF token.
func (s *Server) validateCSRF(r *http.Request) bool {
	// Get CSRF token from header
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		return false
	}

	// Get session
	cookie, err := r.Cookie("unifises")
	if err != nil {
		return false
	}

	session, exists := s.state.GetSession(cookie.Value)
	if !exists {
		return false
	}

	return session.CSRFToken == token
}

// generateToken generates a random session token.
func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateCSRFToken generates a random CSRF token.
func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
