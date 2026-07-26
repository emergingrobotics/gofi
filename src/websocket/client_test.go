package websocket

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestNew(t *testing.T) {
	client, err := New("wss://192.168.1.1/events")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNew_InvalidURL(t *testing.T) {
	_, err := New("://invalid")
	if err == nil {
		t.Error("New() should return error for invalid URL")
	}
}

func TestNew_InvalidScheme(t *testing.T) {
	_, err := New("http://example.com/ws")
	if err == nil {
		t.Error("New() should return error for non-ws scheme")
	}

	if err != nil && !strings.Contains(err.Error(), "scheme") {
		t.Errorf("Error should mention scheme, got: %v", err)
	}
}

func TestClient_Connect(t *testing.T) {
	// Create test WebSocket server
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("Upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// Echo messages
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if err := conn.WriteMessage(mt, message); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "wss" + strings.TrimPrefix(server.URL, "https")

	client, err := New(wsURL, WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Connect
	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if !client.IsConnected() {
		t.Error("IsConnected() = false, want true after Connect()")
	}

	// Cleanup
	client.Close()

	if client.IsConnected() {
		t.Error("IsConnected() = true, want false after Close()")
	}
}

func TestClient_ReadWriteMessage(t *testing.T) {
	// Create test WebSocket server
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo messages
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if err := conn.WriteMessage(mt, message); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "wss" + strings.TrimPrefix(server.URL, "https")

	client, err := New(wsURL, WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	// Write message
	testMsg := []byte("test message")
	if err := client.WriteMessage(testMsg); err != nil {
		t.Fatalf("WriteMessage() error = %v", err)
	}

	// Read echo
	received, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	if string(received) != string(testMsg) {
		t.Errorf("Received = %s, want %s", string(received), string(testMsg))
	}
}

func TestClient_ReadMessage_NotConnected(t *testing.T) {
	client, _ := New("wss://example.com/ws")

	_, err := client.ReadMessage()
	if err == nil {
		t.Error("ReadMessage() should return error when not connected")
	}
}

func TestClient_WriteMessage_NotConnected(t *testing.T) {
	client, _ := New("wss://example.com/ws")

	err := client.WriteMessage([]byte("test"))
	if err == nil {
		t.Error("WriteMessage() should return error when not connected")
	}
}

func TestClient_WithOptions(t *testing.T) {
	headers := make(http.Header)
	headers.Set("X-Custom", "value")

	client, err := New("wss://example.com/ws",
		WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		WithHeaders(headers),
	)

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client.tlsConfig == nil {
		t.Error("TLSConfig not set")
	}

	if client.headers.Get("X-Custom") != "value" {
		t.Error("Custom header not set")
	}
}
