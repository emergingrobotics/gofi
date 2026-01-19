package mock

import (
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
