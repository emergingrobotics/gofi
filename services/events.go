package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/types"
	"github.com/unifi-go/gofi/websocket"
)

// eventService implements EventService.
type eventService struct {
	baseURL   string
	wsClient  *websocket.Client
	eventCh   chan types.Event
	errorCh   chan error
	closeCh   chan struct{}
	tlsConfig *tls.Config
}

// NewEventService creates a new event service.
func NewEventService(baseURL string, tlsConfig *tls.Config) EventService {
	return &eventService{
		baseURL:   baseURL,
		tlsConfig: tlsConfig,
		eventCh:   make(chan types.Event, 100),
		errorCh:   make(chan error, 10),
		closeCh:   make(chan struct{}),
	}
}

// Subscribe subscribes to events for a site.
func (e *eventService) Subscribe(ctx context.Context, site string) (<-chan types.Event, <-chan error, error) {
	// Build WebSocket URL
	wsPath := internal.BuildWebSocketPath(site)

	// Convert https:// to wss://
	wsURL := "wss" + e.baseURL[5:] + wsPath // Strip "https" and add "wss"

	// Create WebSocket client
	var opts []websocket.Option
	if e.tlsConfig != nil {
		opts = append(opts, websocket.WithTLSConfig(e.tlsConfig))
	}

	client, err := websocket.New(wsURL, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create WebSocket client: %w", err)
	}

	e.wsClient = client

	// Connect
	if err := client.Connect(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Start reading events
	go e.readLoop()

	return e.eventCh, e.errorCh, nil
}

// readLoop reads events from the WebSocket.
func (e *eventService) readLoop() {
	defer func() {
		close(e.eventCh)
		close(e.errorCh)
	}()

	for {
		select {
		case <-e.closeCh:
			return
		default:
			if e.wsClient == nil {
				return
			}

			message, err := e.wsClient.ReadMessage()
			if err != nil {
				select {
				case e.errorCh <- fmt.Errorf("read error: %w", err):
				case <-e.closeCh:
				default:
				}
				return
			}

			// Parse event
			var event types.Event
			if err := json.Unmarshal(message, &event); err != nil {
				select {
				case e.errorCh <- fmt.Errorf("parse error: %w", err):
				case <-e.closeCh:
				default:
				}
				continue
			}

			// Send event
			select {
			case e.eventCh <- event:
			case <-e.closeCh:
				return
			case <-time.After(1 * time.Second):
				// Drop event if channel is full
			}
		}
	}
}

// Close closes the event stream.
func (e *eventService) Close() error {
	close(e.closeCh)

	if e.wsClient != nil {
		return e.wsClient.Close()
	}

	return nil
}
