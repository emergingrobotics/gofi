package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/unifi-go/gofi/transport"
)

// Manager manages authentication state.
type Manager interface {
	// Login authenticates with the UniFi controller.
	Login(ctx context.Context) error

	// Logout ends the current session.
	Logout(ctx context.Context) error

	// EnsureAuthenticated ensures there is a valid session, refreshing if needed.
	EnsureAuthenticated(ctx context.Context) error

	// Session returns the current session, or nil if not authenticated.
	Session() *Session

	// IsAuthenticated returns true if there is a valid session.
	IsAuthenticated() bool
}

// manager implements the Manager interface.
type manager struct {
	transport transport.Transport
	username  string
	password  string

	session *Session
	csrf    *CSRFHandler

	mu         sync.RWMutex
	refreshing bool
	refreshCh  chan struct{}
}

// New creates a new authentication manager.
func New(transport transport.Transport, username, password string) Manager {
	return &manager{
		transport: transport,
		username:  username,
		password:  password,
		csrf:      NewCSRFHandler(),
	}
}

// Login authenticates with the UniFi controller.
func (m *manager) Login(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Prepare login request
	loginReq := map[string]string{
		"username": m.username,
		"password": m.password,
	}

	// Create HTTP request
	req := transport.NewRequest("POST", "/api/auth/login").
		WithBody(loginReq)

	// Execute login
	resp, err := m.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	// Check response status - UDM Pro v10+ returns 403 with JSON error on auth failure
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		// Try to parse the error response (newer UDM format)
		var errResp struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(resp.Body, &errResp); err == nil && errResp.Message != "" {
			return fmt.Errorf("login failed: %s", errResp.Message)
		}
		return fmt.Errorf("login failed: status %d (check credentials)", resp.StatusCode)
	}

	if !resp.IsSuccess() {
		// Try to parse error response
		var errResp struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(resp.Body, &errResp); err == nil && errResp.Message != "" {
			return fmt.Errorf("login failed: %s", errResp.Message)
		}
		return fmt.Errorf("login failed: status %d, body: %s", resp.StatusCode, truncateBody(resp.Body))
	}

	// Try parsing as standard UniFi response with meta.rc
	var loginResp struct {
		Meta struct {
			RC      string `json:"rc"`
			Message string `json:"msg"`
		} `json:"meta"`
		Errors []string `json:"errors"`
	}

	if err := json.Unmarshal(resp.Body, &loginResp); err != nil {
		// If we can't parse but got 200, might be a different response format
		// Check if we got a CSRF token which indicates success
		csrfToken := resp.Headers.Get("X-CSRF-Token")
		if csrfToken == "" {
			return fmt.Errorf("failed to parse login response: %w (body: %s)", err, truncateBody(resp.Body))
		}
		// Got CSRF token, assume success
		m.csrf.Set(csrfToken)
		m.transport.SetCSRFToken(csrfToken)
		m.session = &Session{
			Token:     "authenticated",
			CSRFToken: csrfToken,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Username:  m.username,
			CreatedAt: time.Now(),
		}
		return nil
	}

	// Check for errors array (some UDM versions use this)
	if len(loginResp.Errors) > 0 {
		return fmt.Errorf("login failed: %s", loginResp.Errors[0])
	}

	// Check meta.rc
	if loginResp.Meta.RC != "" && loginResp.Meta.RC != "ok" {
		msg := loginResp.Meta.Message
		if msg == "" {
			msg = loginResp.Meta.RC
		}
		return fmt.Errorf("login failed: %s", msg)
	}

	// Extract CSRF token from response headers
	csrfToken := resp.Headers.Get("X-CSRF-Token")
	if csrfToken != "" {
		m.csrf.Set(csrfToken)
		m.transport.SetCSRFToken(csrfToken)
	}

	// Create session
	m.session = &Session{
		Token:      "authenticated", // Cookie-based, actual token is in transport
		CSRFToken:  csrfToken,
		ExpiresAt:  time.Now().Add(24 * time.Hour), // Default 24h expiration
		Username:   m.username,
		CreatedAt:  time.Now(),
	}

	return nil
}

// truncateBody returns a truncated version of the response body for error messages.
func truncateBody(body []byte) string {
	const maxLen = 200
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "..."
}

// Logout ends the current session.
func (m *manager) Logout(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.session == nil {
		return nil // Already logged out
	}

	// Create logout request
	req := transport.NewRequest("POST", "/api/logout")

	// Execute logout (ignore errors, we're clearing session anyway)
	_, _ = m.transport.Do(ctx, req)

	// Clear session
	m.session = nil
	m.csrf.Set("")
	m.transport.SetCSRFToken("")

	return nil
}

// EnsureAuthenticated ensures there is a valid session, refreshing if needed.
func (m *manager) EnsureAuthenticated(ctx context.Context) error {
	// Quick check without lock
	m.mu.RLock()
	session := m.session
	refreshing := m.refreshing
	refreshCh := m.refreshCh
	m.mu.RUnlock()

	// If valid session and not needing refresh, we're done
	if session != nil && session.IsValid() && !session.NeedsRefresh() {
		return nil
	}

	// If another goroutine is already refreshing, wait for it
	if refreshing && refreshCh != nil {
		select {
		case <-refreshCh:
			// Refresh complete, check if we have a valid session now
			m.mu.RLock()
			session := m.session
			m.mu.RUnlock()

			if session == nil || !session.IsValid() {
				return fmt.Errorf("session refresh failed")
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// We need to authenticate/refresh
	m.mu.Lock()

	// Double-check after acquiring lock
	if m.session != nil && m.session.IsValid() && !m.session.NeedsRefresh() {
		m.mu.Unlock()
		return nil
	}

	// Set up refresh coordination
	if m.refreshing {
		// Another goroutine started refreshing between our checks
		refreshCh := m.refreshCh
		m.mu.Unlock()

		if refreshCh != nil {
			select {
			case <-refreshCh:
				m.mu.RLock()
				session := m.session
				m.mu.RUnlock()

				if session == nil || !session.IsValid() {
					return fmt.Errorf("session refresh failed")
				}
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Fall through to retry
		return m.EnsureAuthenticated(ctx)
	}

	// We're the refreshing goroutine
	m.refreshing = true
	m.refreshCh = make(chan struct{})
	refreshCh = m.refreshCh

	m.mu.Unlock()

	// Perform authentication
	err := m.Login(ctx)

	// Clean up refresh state and notify waiters
	m.mu.Lock()
	m.refreshing = false
	close(refreshCh)
	m.refreshCh = nil
	m.mu.Unlock()

	return err
}

// Session returns the current session.
func (m *manager) Session() *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.session
}

// IsAuthenticated returns true if there is a valid session.
func (m *manager) IsAuthenticated() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.session != nil && m.session.IsValid()
}
