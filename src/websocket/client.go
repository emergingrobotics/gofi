package websocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client.
type Client struct {
	url       string
	conn      *websocket.Conn
	mu        sync.RWMutex
	tlsConfig *tls.Config
	headers   http.Header
	dialer    *websocket.Dialer
}

// Config holds WebSocket client configuration.
type Config struct {
	TLSConfig *tls.Config
	Headers   http.Header
}

// Option configures a WebSocket client.
type Option func(*Config)

// WithTLSConfig sets the TLS configuration.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(c *Config) {
		c.TLSConfig = tlsConfig
	}
}

// WithHeaders sets custom headers.
func WithHeaders(headers http.Header) Option {
	return func(c *Config) {
		c.Headers = headers
	}
}

// New creates a new WebSocket client.
func New(wsURL string, opts ...Option) (*Client, error) {
	// Parse URL to validate
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	if u.Scheme != "ws" && u.Scheme != "wss" {
		return nil, fmt.Errorf("invalid WebSocket scheme: %s (must be ws or wss)", u.Scheme)
	}

	config := &Config{
		Headers: make(http.Header),
	}

	for _, opt := range opts {
		opt(config)
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  config.TLSConfig,
	}

	c := &Client{
		url:       wsURL,
		headers:   config.Headers,
		tlsConfig: config.TLSConfig,
		dialer:    dialer,
	}

	return c, nil
}

// Connect establishes the WebSocket connection.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return fmt.Errorf("already connected")
	}

	conn, _, err := c.dialer.DialContext(ctx, c.url, c.headers)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	return nil
}

// Close closes the WebSocket connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}

// ReadMessage reads a message from the WebSocket.
func (c *Client) ReadMessage() ([]byte, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return message, nil
}

// WriteMessage writes a message to the WebSocket.
func (c *Client) WriteMessage(data []byte) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// IsConnected returns true if the WebSocket is connected.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}
