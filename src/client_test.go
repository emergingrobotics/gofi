package gofi

import (
	"context"
	"testing"
	"time"

	"github.com/unifi-go/gofi/src/mock"
)

func TestNew_ValidConfig(t *testing.T) {
	config := &Config{
		Host:     "192.168.1.1",
		Username: "admin",
		Password: "password",
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client == nil {
		t.Fatal("New() returned nil client")
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{"nil config", nil},
		{"empty host", &Config{Username: "admin", Password: "pass"}},
		{"empty username", &Config{Host: "192.168.1.1", Password: "pass"}},
		{"empty password", &Config{Host: "192.168.1.1", Username: "admin"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)
			if err == nil {
				t.Error("New() should return error for invalid config")
			}
		})
	}
}

func TestNew_Defaults(t *testing.T) {
	config := &Config{
		Host:     "192.168.1.1",
		Username: "admin",
		Password: "password",
	}

	c, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	impl := c.(*client)

	if impl.config.Port != 443 {
		t.Errorf("Port = %d, want 443", impl.config.Port)
	}

	if impl.config.Site != "default" {
		t.Errorf("Site = %s, want default", impl.config.Site)
	}

	if impl.config.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns = %d, want 10", impl.config.MaxIdleConns)
	}
}

func TestNew_WithOptions(t *testing.T) {
	config := &Config{
		Host:     "192.168.1.1",
		Username: "admin",
		Password: "password",
	}

	c, err := New(config,
		WithTimeout(10*time.Second),
		WithSite("custom"),
		WithInsecureSkipVerify(),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	impl := c.(*client)

	if impl.config.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", impl.config.Timeout)
	}

	if impl.config.Site != "custom" {
		t.Errorf("Site = %s, want custom", impl.config.Site)
	}

	if !impl.config.SkipTLSVerify {
		t.Error("SkipTLSVerify should be true")
	}
}

func TestClient_Connect_Success(t *testing.T) {
	// Create mock server
	server := mock.NewServer()
	defer server.Close()

	config := &Config{
		Host:          server.Host(),
		Port:          server.Port(),
		Username:      "admin",
		Password:      "admin",
		SkipTLSVerify: true,
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Initially not connected
	if client.IsConnected() {
		t.Error("IsConnected() = true, want false before Connect()")
	}

	// Connect
	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	// Should be connected
	if !client.IsConnected() {
		t.Error("IsConnected() = false, want true after Connect()")
	}

	// Cleanup
	_ = client.Disconnect(context.Background())
}

func TestClient_Connect_InvalidCredentials(t *testing.T) {
	server := mock.NewServer()
	defer server.Close()

	config := &Config{
		Host:          server.Host(),
		Username:      "admin",
		Password:      "wrongpassword",
		SkipTLSVerify: true,
	}

	config.Port = server.Port()

	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Connect should fail
	err = client.Connect(context.Background())
	if err == nil {
		t.Fatal("Connect() should return error for invalid credentials")
	}

	// Should not be connected
	if client.IsConnected() {
		t.Error("IsConnected() = true, want false after failed Connect()")
	}
}

func TestClient_Connect_AlreadyConnected(t *testing.T) {
	server := mock.NewServer()
	defer server.Close()

	config := &Config{
		Host:          server.Host(),
		Username:      "admin",
		Password:      "admin",
		SkipTLSVerify: true,
	}

	config.Port = server.Port()

	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Connect once
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("First Connect() error = %v", err)
	}

	// Connect again should fail
	err = client.Connect(context.Background())
	if err != ErrAlreadyConnected {
		t.Errorf("Second Connect() error = %v, want ErrAlreadyConnected", err)
	}

	// Cleanup
	_ = client.Disconnect(context.Background())
}

func TestClient_Disconnect(t *testing.T) {
	server := mock.NewServer()
	defer server.Close()

	config := &Config{
		Host:          server.Host(),
		Username:      "admin",
		Password:      "admin",
		SkipTLSVerify: true,
	}

	config.Port = server.Port()

	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Connect
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	// Disconnect
	if err := client.Disconnect(context.Background()); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}

	// Should not be connected
	if client.IsConnected() {
		t.Error("IsConnected() = true, want false after Disconnect()")
	}

	// Disconnecting again should be safe
	if err := client.Disconnect(context.Background()); err != nil {
		t.Errorf("Second Disconnect() error = %v, want nil", err)
	}
}

func TestClient_IsConnected(t *testing.T) {
	server := mock.NewServer()
	defer server.Close()

	config := &Config{
		Host:          server.Host(),
		Username:      "admin",
		Password:      "admin",
		SkipTLSVerify: true,
	}

	config.Port = server.Port()

	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Initially false
	if client.IsConnected() {
		t.Error("IsConnected() should be false initially")
	}

	// After connect: true
	_ = client.Connect(context.Background())
	if !client.IsConnected() {
		t.Error("IsConnected() should be true after Connect()")
	}

	// After disconnect: false
	_ = client.Disconnect(context.Background())
	if client.IsConnected() {
		t.Error("IsConnected() should be false after Disconnect()")
	}
}

func TestOptions(t *testing.T) {
	config := &Config{
		Host:     "192.168.1.1",
		Username: "admin",
		Password: "password",
	}

	// Test individual options
	WithTimeout(5 * time.Second)(config)
	if config.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", config.Timeout)
	}

	WithSite("mysite")(config)
	if config.Site != "mysite" {
		t.Errorf("Site = %s, want mysite", config.Site)
	}

	WithInsecureSkipVerify()(config)
	if !config.SkipTLSVerify {
		t.Error("SkipTLSVerify should be true")
	}

	WithRetry(5, 50*time.Millisecond)(config)
	if config.RetryConfig == nil {
		t.Fatal("RetryConfig should not be nil")
	}
	if config.RetryConfig.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", config.RetryConfig.MaxRetries)
	}
}

func TestOptions_APIKeyAndConnector(t *testing.T) {
	cfg := &Config{}
	WithAPIKey("k123")(cfg)
	WithConnector("console-abc")(cfg)
	if cfg.APIKey != "k123" {
		t.Errorf("APIKey = %q, want k123", cfg.APIKey)
	}
	if cfg.ConsoleID != "console-abc" {
		t.Errorf("ConsoleID = %q, want console-abc", cfg.ConsoleID)
	}
}

func TestNew_APIKeyConnector(t *testing.T) {
	c, err := New(&Config{}, WithAPIKey("k"), WithConnector("abc"))
	if err != nil {
		t.Fatalf("New with API key+connector: %v", err)
	}
	if c == nil {
		t.Fatal("client is nil")
	}
}

func TestNew_NoCredentials(t *testing.T) {
	_, err := New(&Config{Host: "1.2.3.4"})
	if err == nil {
		t.Fatal("expected validation error with no credentials")
	}
}

func TestNew_BothCredentials(t *testing.T) {
	_, err := New(&Config{ConsoleID: "abc", APIKey: "k", Username: "u", Password: "p"})
	if err == nil {
		t.Fatal("expected validation error when both credential sets present")
	}
}

func TestNew_HostAndConsoleMutuallyExclusive(t *testing.T) {
	_, err := New(&Config{Host: "1.2.3.4", APIKey: "k", ConsoleID: "abc"})
	if err == nil {
		t.Fatal("expected error when Host and ConsoleID both set")
	}
}

func TestNew_APIKeyRequiresConnector(t *testing.T) {
	_, err := New(&Config{APIKey: "k"})
	if err == nil {
		t.Fatal("expected error: API key without ConsoleID")
	}
}

func TestNew_DefaultTimeoutIs30s(t *testing.T) {
	cfg := &Config{Host: "1.2.3.4", Username: "u", Password: "p"}
	if _, err := New(cfg); err != nil {
		t.Fatalf("New: %v", err)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("resolved Timeout = %v, want 30s", cfg.Timeout)
	}
}
