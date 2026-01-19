package transport

import (
	"crypto/tls"
	"time"
)

// Config holds transport layer configuration.
type Config struct {
	// BaseURL is the base URL of the UniFi controller.
	BaseURL string

	// Timeout is the request timeout.
	Timeout time.Duration

	// TLSConfig is the TLS configuration.
	TLSConfig *tls.Config

	// MaxIdleConns is the maximum number of idle connections.
	MaxIdleConns int

	// MaxConnsPerHost is the maximum number of connections per host.
	MaxConnsPerHost int

	// IdleConnTimeout is the idle connection timeout.
	IdleConnTimeout time.Duration

	// UserAgent is the User-Agent header value.
	UserAgent string
}

// Option is a functional option for configuring the transport.
type Option func(*Config)

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithTLSConfig sets the TLS configuration.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(c *Config) {
		c.TLSConfig = tlsConfig
	}
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) Option {
	return func(c *Config) {
		c.MaxIdleConns = n
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Config) {
		c.UserAgent = ua
	}
}

// DefaultConfig returns a Config with default values.
func DefaultConfig(baseURL string) *Config {
	return &Config{
		BaseURL:          baseURL,
		Timeout:          30 * time.Second,
		MaxIdleConns:     10,
		MaxConnsPerHost:  10,
		IdleConnTimeout:  90 * time.Second,
		UserAgent:        "gofi/1.0",
	}
}
