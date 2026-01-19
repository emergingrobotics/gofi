package transport

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	baseURL := "https://192.168.1.1"
	cfg := DefaultConfig(baseURL)

	if cfg.BaseURL != baseURL {
		t.Errorf("BaseURL = %s, want %s", cfg.BaseURL, baseURL)
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}

	if cfg.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns = %d, want 10", cfg.MaxIdleConns)
	}

	if cfg.UserAgent != "gofi/1.0" {
		t.Errorf("UserAgent = %s, want gofi/1.0", cfg.UserAgent)
	}
}

func TestWithTimeout(t *testing.T) {
	cfg := &Config{}
	opt := WithTimeout(10 * time.Second)
	opt(cfg)

	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 10*time.Second)
	}
}

func TestWithTLSConfig(t *testing.T) {
	cfg := &Config{}
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	opt := WithTLSConfig(tlsConfig)
	opt(cfg)

	if cfg.TLSConfig != tlsConfig {
		t.Error("TLSConfig not set correctly")
	}
}

func TestWithMaxIdleConns(t *testing.T) {
	cfg := &Config{}
	opt := WithMaxIdleConns(20)
	opt(cfg)

	if cfg.MaxIdleConns != 20 {
		t.Errorf("MaxIdleConns = %d, want 20", cfg.MaxIdleConns)
	}
}

func TestWithUserAgent(t *testing.T) {
	cfg := &Config{}
	opt := WithUserAgent("test-agent/1.0")
	opt(cfg)

	if cfg.UserAgent != "test-agent/1.0" {
		t.Errorf("UserAgent = %s, want test-agent/1.0", cfg.UserAgent)
	}
}

func TestOptionsChaining(t *testing.T) {
	cfg := DefaultConfig("https://test.local")

	opts := []Option{
		WithTimeout(5 * time.Second),
		WithMaxIdleConns(5),
		WithUserAgent("custom/2.0"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 5*time.Second)
	}

	if cfg.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", cfg.MaxIdleConns)
	}

	if cfg.UserAgent != "custom/2.0" {
		t.Errorf("UserAgent = %s, want custom/2.0", cfg.UserAgent)
	}
}
