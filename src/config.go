package gofi

import (
	"crypto/tls"
	"time"
)

// Config holds the configuration for connecting to a UDM Pro.
type Config struct {
	// Host is the IP address or hostname of the UDM Pro.
	Host string

	// Port is the HTTPS port (default: 443).
	Port int

	// Username for local admin authentication.
	Username string

	// Password for local admin authentication.
	Password string

	// APIKey authenticates via the X-API-KEY header. When set, Username and
	// Password are ignored and no login request is made. Requires ConsoleID
	// (connector mode).
	APIKey string

	// ConsoleID is the Site Manager console ID. When set, requests are routed
	// through the connector proxy at https://api.ui.com and Host/Port are unused.
	ConsoleID string

	// BaseURL overrides the derived base URL (host:port, or the connector's
	// https://api.ui.com). Mainly for testing against a mock server.
	BaseURL string

	// Site is the default site ID (default: "default").
	Site string

	// TLS configuration
	TLSConfig *tls.Config

	// SkipTLSVerify disables TLS certificate verification.
	// WARNING: Only use for development/testing.
	SkipTLSVerify bool

	// Timeout for HTTP requests (default: 30s).
	Timeout time.Duration

	// MaxIdleConns is the maximum number of idle connections (default: 10).
	MaxIdleConns int

	// RetryConfig configures automatic retries.
	RetryConfig *RetryConfig

	// Logger for debug output (optional).
	Logger Logger
}

// RetryConfig configures retry behavior.
type RetryConfig struct {
	// MaxRetries is the maximum number of retries (default: 3).
	MaxRetries int

	// InitialBackoff is the initial backoff duration (default: 100ms).
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration (default: 5s).
	MaxBackoff time.Duration

	// RetryableErrors are error types that trigger retries.
	RetryableErrors []error
}

// Logger is a simple logging interface.
type Logger interface {
	// Debug logs a debug message.
	Debug(msg string, keysAndValues ...interface{})

	// Info logs an informational message.
	Info(msg string, keysAndValues ...interface{})

	// Warn logs a warning message.
	Warn(msg string, keysAndValues ...interface{})

	// Error logs an error message.
	Error(msg string, keysAndValues ...interface{})
}
