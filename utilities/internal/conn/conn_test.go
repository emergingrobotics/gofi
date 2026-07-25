package conn

import (
	"io"
	"strings"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{EnvUsername, EnvPassword, EnvControllerIP, EnvAPIKey, EnvConsoleID} {
		t.Setenv(k, "")
	}
}

func TestResolveConfig_Connector(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvAPIKey, "k123")
	t.Setenv(EnvConsoleID, "console-abc")
	cfg, err := ResolveConfig(io.Discard, "", 443, "default", false)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.APIKey != "k123" || cfg.ConsoleID != "console-abc" {
		t.Errorf("got APIKey=%q ConsoleID=%q", cfg.APIKey, cfg.ConsoleID)
	}
	if cfg.Host != "" {
		t.Errorf("Host = %q, want empty in connector mode", cfg.Host)
	}
	if cfg.SkipTLSVerify {
		t.Error("SkipTLSVerify = true, want false: connector mode (api.ui.com) must always verify TLS")
	}
}

// TestResolveConfig_ConnectorAlwaysVerifiesTLS confirms that connector mode
// ignores the secure flag entirely: api.ui.com has a valid CA-signed cert, so
// -k/--secure=false (the LAN self-signed accommodation) must not weaken TLS
// verification against the public connector endpoint.
func TestResolveConfig_ConnectorAlwaysVerifiesTLS(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvAPIKey, "k123")
	t.Setenv(EnvConsoleID, "console-abc")
	cfg, err := ResolveConfig(io.Discard, "", 443, "default", false)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.SkipTLSVerify {
		t.Error("SkipTLSVerify = true, want false: connector mode always verifies TLS regardless of secure flag")
	}
}

func TestResolveConfig_APIKeyRequiresConsole(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvAPIKey, "k123")
	if _, err := ResolveConfig(io.Discard, "", 443, "default", false); err == nil {
		t.Fatal("expected error: API key without console ID")
	}
}

func TestResolveConfig_Local(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvUsername, "admin")
	t.Setenv(EnvPassword, "pw")
	cfg, err := ResolveConfig(io.Discard, "10.0.0.1", 8443, "default", true)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Host != "10.0.0.1" || cfg.Port != 8443 || cfg.Username != "admin" {
		t.Errorf("got %+v", cfg)
	}
	if cfg.SkipTLSVerify {
		t.Error("SkipTLSVerify = true, want false when secure=true")
	}
}

func TestResolveConfig_LocalHostFromEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvUsername, "admin")
	t.Setenv(EnvPassword, "pw")
	t.Setenv(EnvControllerIP, "192.168.9.9")
	cfg, err := ResolveConfig(io.Discard, "", 443, "default", false)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Host != "192.168.9.9" {
		t.Errorf("Host = %q, want fallback from env", cfg.Host)
	}
}

func TestResolveConfig_NoCredentials(t *testing.T) {
	clearEnv(t)
	if _, err := ResolveConfig(io.Discard, "1.2.3.4", 443, "default", false); err == nil {
		t.Fatal("expected error when no credentials are set")
	}
}

func TestResolveConfig_KeyOverridesUserPassWithNote(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvAPIKey, "k123")
	t.Setenv(EnvConsoleID, "console-abc")
	t.Setenv(EnvUsername, "admin")
	t.Setenv(EnvPassword, "pw")
	var buf strings.Builder
	cfg, err := ResolveConfig(&buf, "", 443, "default", false)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.APIKey != "k123" {
		t.Error("expected API key to win over username/password")
	}
	if !strings.Contains(buf.String(), EnvAPIKey) {
		t.Errorf("expected override note mentioning %s, got %q", EnvAPIKey, buf.String())
	}
}
