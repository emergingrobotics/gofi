package mock

import (
	"encoding/json"
	"net/http"

	"github.com/unifi-go/gofi/types"
)

// handleLogin handles login requests.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	// Parse credentials
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Validate credentials
	if !s.state.ValidateCredentials(creds.Username, creds.Password) {
		writeAPIError(w, http.StatusUnauthorized, "error", "Invalid credentials")
		return
	}

	// Create session
	token := generateToken()
	csrfToken := generateCSRFToken()

	session := &Session{
		Username:  creds.Username,
		CSRFToken: csrfToken,
	}

	s.state.CreateSession(token, session)

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "unifises",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Set CSRF token in header
	w.Header().Set("X-CSRF-Token", csrfToken)

	// Return success
	resp := types.APIResponse[interface{}]{
		Meta: types.ResponseMeta{RC: "ok"},
		Data: []interface{}{},
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleLogout handles logout requests.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	// Get session cookie
	cookie, err := r.Cookie("unifises")
	if err == nil {
		s.state.DeleteSession(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "unifises",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Return success
	resp := types.APIResponse[interface{}]{
		Meta: types.ResponseMeta{RC: "ok"},
		Data: []interface{}{},
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleStatus handles status requests (no auth required).
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := &types.Status{
		Up:      true,
		Version: "7.5.174",
	}

	// Status endpoint returns direct JSON, not API response wrapper
	writeJSON(w, http.StatusOK, status)
}

// handleSelf handles self requests.
func (s *Server) handleSelf(w http.ResponseWriter, r *http.Request) {
	// Get session
	cookie, err := r.Cookie("unifises")
	if err != nil {
		writeUnauthorized(w)
		return
	}

	session, exists := s.state.GetSession(cookie.Value)
	if !exists {
		writeUnauthorized(w)
		return
	}

	admin := &types.AdminUser{
		Name:  session.Username,
		Email: session.Username + "@example.com",
	}

	// Self endpoint returns data in API response format (array)
	writeAPIResponse(w, []interface{}{*admin})
}
