package mock

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

// newTestClient returns an HTTP client that accepts self-signed certs.
func newTestClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func TestHandleLogin_Success(t *testing.T) {
	server := NewServer()
	defer server.Close()

	// Create login request
	body := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", server.URL()+"/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Check for session cookie
	var foundCookie bool
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "unifises" {
			foundCookie = true
			if cookie.Value == "" {
				t.Error("Session cookie value is empty")
			}
		}
	}

	if !foundCookie {
		t.Error("Session cookie not set")
	}

	// Check for CSRF token
	csrfToken := resp.Header.Get("X-CSRF-Token")
	if csrfToken == "" {
		t.Error("CSRF token not set in response header")
	}

	// Parse response
	var apiResp types.APIResponse[interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if apiResp.Meta.RC != "ok" {
		t.Errorf("RC = %s, want ok", apiResp.Meta.RC)
	}
}

func TestHandleLogin_InvalidCredentials(t *testing.T) {
	server := NewServer()
	defer server.Close()

	body := map[string]string{
		"username": "admin",
		"password": "wrongpassword",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", server.URL()+"/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestHandleLogin_InvalidMethod(t *testing.T) {
	server := NewServer()
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL()+"/api/auth/login", nil)

	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleLogout(t *testing.T) {
	server := NewServer()
	defer server.Close()

	// Login first
	loginBody := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	loginBodyBytes, _ := json.Marshal(loginBody)

	loginReq, _ := http.NewRequest("POST", server.URL()+"/api/auth/login", bytes.NewReader(loginBodyBytes))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := newTestClient().Do(loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	defer loginResp.Body.Close()

	// Get session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range loginResp.Cookies() {
		if cookie.Name == "unifises" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("No session cookie returned from login")
	}

	// Logout
	logoutReq, _ := http.NewRequest("POST", server.URL()+"/api/logout", nil)
	logoutReq.AddCookie(sessionCookie)

	logoutResp, err := newTestClient().Do(logoutReq)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	defer logoutResp.Body.Close()

	if logoutResp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", logoutResp.StatusCode, http.StatusOK)
	}
}

func TestHandleStatus(t *testing.T) {
	server := NewServer()
	defer server.Close()

	// Status endpoint doesn't require auth
	req, _ := http.NewRequest("GET", server.URL()+"/api/status", nil)

	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var status types.Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !status.Up {
		t.Error("Status.Up = false, want true")
	}
}

func TestHandleSelf(t *testing.T) {
	server := NewServer()
	defer server.Close()

	// Login first
	loginBody := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	loginBodyBytes, _ := json.Marshal(loginBody)

	loginReq, _ := http.NewRequest("POST", server.URL()+"/api/auth/login", bytes.NewReader(loginBodyBytes))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := newTestClient().Do(loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	defer loginResp.Body.Close()

	// Get session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range loginResp.Cookies() {
		if cookie.Name == "unifises" {
			sessionCookie = cookie
			break
		}
	}

	// Request self
	selfReq, _ := http.NewRequest("GET", server.URL()+"/api/self", nil)
	selfReq.AddCookie(sessionCookie)

	selfResp, err := newTestClient().Do(selfReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer selfResp.Body.Close()

	if selfResp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", selfResp.StatusCode, http.StatusOK)
	}

	var apiResp types.APIResponse[types.AdminUser]
	if err := json.NewDecoder(selfResp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(apiResp.Data) == 0 {
		t.Fatal("No data in response")
	}

	if apiResp.Data[0].Name != "admin" {
		t.Errorf("Name = %s, want admin", apiResp.Data[0].Name)
	}
}

func TestServer_AuthRequired(t *testing.T) {
	server := NewServer()
	defer server.Close()

	// Try to access /api/self without authentication
	req, _ := http.NewRequest("GET", server.URL()+"/api/self", nil)

	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestServer_WithoutAuth(t *testing.T) {
	server := NewServer(WithoutAuth())
	defer server.Close()

	// When auth is disabled, the server should skip auth checks
	// The /api/self endpoint still needs a session to return user info,
	// but it shouldn't return 401 Unauthorized - it should proceed to the handler
	// which will then fail for lack of session (different error)

	// Actually, let's test that the auth check is bypassed by checking
	// that we get past the auth layer (even if the handler itself fails)
	req, _ := http.NewRequest("GET", server.URL()+"/some/endpoint", nil)

	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should get 404 Not Found (route doesn't exist) instead of 401
	if resp.StatusCode == http.StatusUnauthorized {
		t.Error("Should not return 401 when auth is disabled")
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Status = %d, want %d (should get 404 for unknown route)", resp.StatusCode, http.StatusNotFound)
	}
}
