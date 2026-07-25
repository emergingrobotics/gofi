package mock

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer()
	defer server.Close()

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.URL() == "" {
		t.Error("URL() returned empty string")
	}

	if server.State() == nil {
		t.Error("State() returned nil")
	}
}

func TestNewServer_WithOptions(t *testing.T) {
	server := NewServer(
		WithoutAuth(),
		WithoutCSRF(),
	)
	defer server.Close()

	if server.requireAuth {
		t.Error("requireAuth should be false with WithoutAuth()")
	}

	if server.requireCSRF {
		t.Error("requireCSRF should be false with WithoutCSRF()")
	}
}

func TestNewServer_WithFixtures(t *testing.T) {
	fixtures := DefaultFixtures()
	server := NewServer(WithFixtures(fixtures))
	defer server.Close()

	// Default fixtures should have loaded
	sites := server.State().ListSites()
	if len(sites) == 0 {
		t.Error("Fixtures not loaded")
	}
}

func TestNewServer_WithScenario(t *testing.T) {
	scenario := ScenarioServerError
	server := NewServer(WithScenario(scenario))
	defer server.Close()

	if server.scenario == nil {
		t.Error("Scenario not set")
	}
}

func TestServer_URL(t *testing.T) {
	server := NewServer()
	defer server.Close()

	url := server.URL()
	if url == "" {
		t.Error("URL() returned empty string")
	}

	// Should be https
	if len(url) < 8 || url[:8] != "https://" {
		t.Errorf("URL should start with https://, got %s", url)
	}
}

func TestServer_State(t *testing.T) {
	server := NewServer()
	defer server.Close()

	state := server.State()
	if state == nil {
		t.Fatal("State() returned nil")
	}

	// Should have default admin user
	if !state.ValidateCredentials("admin", "admin") {
		t.Error("Default admin user not present")
	}
}

func TestServer_APIKeyAuth(t *testing.T) {
	srv := NewServer(WithAPIKey("good-key"))
	defer srv.Close()
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	// Correct key, no cookie -> authenticated (200).
	req, _ := http.NewRequest("GET", srv.URL()+"/proxy/network/api/s/default/rest/networkconf", nil)
	req.Header.Set("X-API-KEY", "good-key")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("correct key status = %d, want 200", resp.StatusCode)
	}

	// Wrong key -> 403 with exact body.
	req2, _ := http.NewRequest("GET", srv.URL()+"/proxy/network/api/s/default/rest/networkconf", nil)
	req2.Header.Set("X-API-KEY", "bad-key")
	resp2, _ := client.Do(req2)
	if resp2.StatusCode != http.StatusForbidden {
		t.Errorf("wrong key status = %d, want 403", resp2.StatusCode)
	}
	body, _ := io.ReadAll(resp2.Body)
	var parsed map[string]interface{}
	_ = json.Unmarshal(body, &parsed)
	if parsed["code"] != "forbidden" || parsed["message"] != "insufficient permissions" {
		t.Errorf("wrong-key body = %s, want forbidden/insufficient permissions", body)
	}
}

func TestServer_ConnectorPrefixStrip(t *testing.T) {
	srv := NewServer(WithAPIKey("good-key"))
	defer srv.Close()
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	req, _ := http.NewRequest("GET",
		srv.URL()+"/v1/connector/consoles/abc123/proxy/network/api/s/default/rest/networkconf", nil)
	req.Header.Set("X-API-KEY", "good-key")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("connector-prefixed status = %d, want 200 (prefix should be stripped)", resp.StatusCode)
	}
}
