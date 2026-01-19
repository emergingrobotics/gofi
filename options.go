package gofi

import (
	"crypto/tls"
	"time"
)

// Option configures a Client.
type Option func(*Config)

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithRetry configures retry behavior.
func WithRetry(maxRetries int, initialBackoff time.Duration) Option {
	return func(c *Config) {
		if c.RetryConfig == nil {
			c.RetryConfig = &RetryConfig{}
		}
		c.RetryConfig.MaxRetries = maxRetries
		c.RetryConfig.InitialBackoff = initialBackoff
		if c.RetryConfig.MaxBackoff == 0 {
			c.RetryConfig.MaxBackoff = 5 * time.Second
		}
	}
}

// WithLogger sets a custom logger.
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithTLSConfig sets custom TLS configuration.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(c *Config) {
		c.TLSConfig = tlsConfig
	}
}

// WithInsecureSkipVerify disables TLS certificate verification.
// WARNING: This is insecure and should only be used for testing.
func WithInsecureSkipVerify() Option {
	return func(c *Config) {
		c.SkipTLSVerify = true
	}
}

// WithSite sets the default site.
func WithSite(site string) Option {
	return func(c *Config) {
		c.Site = site
	}
}
