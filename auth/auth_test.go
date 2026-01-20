package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/unifi-go/gofi/transport"
)

func TestNew(t *testing.T) {
	config := transport.DefaultConfig("https://192.168.1.1")
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	mgr := New(trans, "admin", "password")
	if mgr == nil {
		t.Fatal("New() returned nil")
	}

	// Should not be authenticated initially - this is expected
	_ = mgr.IsAuthenticated()
}

func TestManager_Login_Success(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" {
			t.Errorf("Path = %s, want /api/auth/login", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}

		// Set CSRF token in response
		w.Header().Set("X-CSRF-Token", "test-csrf-token")

		// Return success
		resp := map[string]interface{}{
			"meta": map[string]string{
				"rc": "ok",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create manager
	config := transport.DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	mgr := New(trans, "admin", "password")

	// Login
	err = mgr.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	// Should be authenticated
	if !mgr.IsAuthenticated() {
		t.Error("IsAuthenticated() = false, want true after successful login")
	}

	// Should have session
	session := mgr.Session()
	if session == nil {
		t.Fatal("Session() = nil, want non-nil after login")
	}

	if session.Username != "admin" {
		t.Errorf("Session.Username = %s, want admin", session.Username)
	}

	if session.CSRFToken != "test-csrf-token" {
		t.Errorf("Session.CSRFToken = %s, want test-csrf-token", session.CSRFToken)
	}
}

func TestManager_Login_Failure(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]interface{}{
			"meta": map[string]string{
				"rc": "error",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create manager
	config := transport.DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	mgr := New(trans, "admin", "wrongpassword")

	// Login should fail
	err = mgr.Login(context.Background())
	if err == nil {
		t.Fatal("Login() should return error for failed authentication")
	}

	// Should not be authenticated
	if mgr.IsAuthenticated() {
		t.Error("IsAuthenticated() = true, want false after failed login")
	}
}

func TestManager_Logout(t *testing.T) {
	// Create test server
	loginCalled := false
	logoutCalled := false

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" {
			loginCalled = true
			w.Header().Set("X-CSRF-Token", "test-csrf-token")
			resp := map[string]interface{}{
				"meta": map[string]string{"rc": "ok"},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/api/logout" {
			logoutCalled = true
			w.WriteHeader(http.StatusOK)
			return
		}
	}))
	defer server.Close()

	// Create manager
	config := transport.DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	mgr := New(trans, "admin", "password")

	// Login first
	if err := mgr.Login(context.Background()); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if !loginCalled {
		t.Error("Login endpoint was not called")
	}

	if !mgr.IsAuthenticated() {
		t.Fatal("Should be authenticated after login")
	}

	// Logout
	if err := mgr.Logout(context.Background()); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	if !logoutCalled {
		t.Error("Logout endpoint was not called")
	}

	// Should not be authenticated
	if mgr.IsAuthenticated() {
		t.Error("IsAuthenticated() = true, want false after logout")
	}

	// Session should be nil
	if session := mgr.Session(); session != nil {
		t.Error("Session() should be nil after logout")
	}
}

func TestManager_EnsureAuthenticated(t *testing.T) {
	// Create test server
	loginCount := 0

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" {
			loginCount++
			w.Header().Set("X-CSRF-Token", "test-csrf-token")
			resp := map[string]interface{}{
				"meta": map[string]string{"rc": "ok"},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}))
	defer server.Close()

	// Create manager
	config := transport.DefaultConfig(server.URL)
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	mgr := New(trans, "admin", "password")

	// EnsureAuthenticated should log in
	if err := mgr.EnsureAuthenticated(context.Background()); err != nil {
		t.Fatalf("EnsureAuthenticated() error = %v", err)
	}

	if loginCount != 1 {
		t.Errorf("Login count = %d, want 1", loginCount)
	}

	if !mgr.IsAuthenticated() {
		t.Error("Should be authenticated after EnsureAuthenticated")
	}

	// Calling again should not trigger another login
	if err := mgr.EnsureAuthenticated(context.Background()); err != nil {
		t.Fatalf("EnsureAuthenticated() error = %v", err)
	}

	if loginCount != 1 {
		t.Errorf("Login count = %d, want 1 (should not login again)", loginCount)
	}
}

func TestManager_IsAuthenticated(t *testing.T) {
	config := transport.DefaultConfig("https://192.168.1.1")
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	mgr := New(trans, "admin", "password")

	// Initially not authenticated
	if mgr.IsAuthenticated() {
		t.Error("IsAuthenticated() = true, want false initially")
	}
}
