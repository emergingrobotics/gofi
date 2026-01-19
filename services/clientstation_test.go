package services

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

func newTestClientTransport(url string) (transport.Transport, error) {
	config := transport.DefaultConfig(url)
	config.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return transport.New(config)
}

func TestClientService_ListActive(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test clients with different last seen times
	now := time.Now().Unix()
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:f1",
		Hostname: "active-device",
		LastSeen: now - 60, // Active: 1 minute ago
	})
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:f2",
		Hostname: "inactive-device",
		LastSeen: now - 600, // Inactive: 10 minutes ago
	})

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test ListActive
	clients, err := svc.ListActive(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListActive failed: %v", err)
	}

	// Should only return active clients (seen in last 5 minutes)
	if len(clients) != 1 {
		t.Errorf("Expected 1 active client, got %d", len(clients))
	}

	if len(clients) > 0 && clients[0].MAC != "aa:bb:cc:dd:ee:f1" {
		t.Errorf("Expected MAC aa:bb:cc:dd:ee:f1, got %s", clients[0].MAC)
	}
}

func TestClientService_ListAll(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test clients with different last seen times
	now := time.Now().Unix()
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:f1",
		Hostname: "recent-device",
		LastSeen: now - 3600, // 1 hour ago
	})
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:f2",
		Hostname: "old-device",
		LastSeen: now - 86400*30, // 30 days ago
	})

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test ListAll with 24 hour window
	clients, err := svc.ListAll(context.Background(), "default", WithinHours(24))
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	// Should only return recent client
	if len(clients) != 1 {
		t.Errorf("Expected 1 client within 24 hours, got %d", len(clients))
	}

	if len(clients) > 0 && clients[0].MAC != "aa:bb:cc:dd:ee:f1" {
		t.Errorf("Expected MAC aa:bb:cc:dd:ee:f1, got %s", clients[0].MAC)
	}
}

func TestClientService_Get(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test client
	now := time.Now().Unix()
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:ff",
		Hostname: "test-device",
		IP:       "192.168.1.100",
		LastSeen: now - 60,
	})

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test Get
	client, err := svc.Get(context.Background(), "default", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if client.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("Expected MAC aa:bb:cc:dd:ee:ff, got %s", client.MAC)
	}

	if client.Hostname != "test-device" {
		t.Errorf("Expected hostname test-device, got %s", client.Hostname)
	}
}

func TestClientService_GetNotFound(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test Get non-existent client
	_, err := svc.Get(context.Background(), "default", "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent client")
	}
}

func TestClientService_Block(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test client
	now := time.Now().Unix()
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:ff",
		LastSeen: now,
	})

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test Block
	err := svc.Block(context.Background(), "default", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}

	// Verify client is blocked
	client := server.State().GetClient("aa:bb:cc:dd:ee:ff")
	if client == nil {
		t.Fatal("Client not found after block")
	}

	if !client.Blocked {
		t.Error("Expected client to be blocked")
	}
}

func TestClientService_Unblock(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add blocked client
	now := time.Now().Unix()
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:ff",
		LastSeen: now,
		Blocked:  true,
	})

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test Unblock
	err := svc.Unblock(context.Background(), "default", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Unblock failed: %v", err)
	}

	// Verify client is unblocked
	client := server.State().GetClient("aa:bb:cc:dd:ee:ff")
	if client == nil {
		t.Fatal("Client not found after unblock")
	}

	if client.Blocked {
		t.Error("Expected client to be unblocked")
	}
}

func TestClientService_Kick(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test client
	now := time.Now().Unix()
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:ff",
		LastSeen: now,
	})

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test Kick
	err := svc.Kick(context.Background(), "default", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Kick failed: %v", err)
	}

	// Verify client is kicked
	client := server.State().GetClient("aa:bb:cc:dd:ee:ff")
	if client == nil {
		t.Fatal("Client not found after kick")
	}

	if !client.GuestKicked {
		t.Error("Expected client to be kicked")
	}
}

func TestClientService_Forget(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test client
	now := time.Now().Unix()
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:ff",
		LastSeen: now,
	})

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test Forget
	err := svc.Forget(context.Background(), "default", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Forget failed: %v", err)
	}

	// Verify client is deleted
	client := server.State().GetClient("aa:bb:cc:dd:ee:ff")
	if client != nil {
		t.Error("Expected client to be forgotten")
	}
}

func TestClientService_AuthorizeGuest(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test AuthorizeGuest (creates client if needed)
	err := svc.AuthorizeGuest(context.Background(), "default", "aa:bb:cc:dd:ee:f1",
		WithDuration(60),
		WithUploadLimit(1024),
		WithDownloadLimit(2048),
	)
	if err != nil {
		t.Fatalf("AuthorizeGuest failed: %v", err)
	}

	// Verify guest was created and authorized
	client := server.State().GetClient("aa:bb:cc:dd:ee:f1")
	if client == nil {
		t.Fatal("Expected guest client to be created")
	}

	if !client.GuestAuthorized {
		t.Error("Expected guest to be authorized")
	}

	if !client.Authorized {
		t.Error("Expected client to be authorized")
	}
}

func TestClientService_UnauthorizeGuest(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add authorized guest
	now := time.Now().Unix()
	server.State().AddClient(&types.Client{
		MAC:             "aa:bb:cc:dd:ee:f1",
		LastSeen:        now,
		IsGuest:         true,
		GuestAuthorized: true,
		Authorized:      true,
	})

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test UnauthorizeGuest
	err := svc.UnauthorizeGuest(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("UnauthorizeGuest failed: %v", err)
	}

	// Verify guest was unauthorized
	client := server.State().GetClient("aa:bb:cc:dd:ee:f1")
	if client == nil {
		t.Fatal("Client not found after unauthorize")
	}

	if client.GuestAuthorized {
		t.Error("Expected guest to be unauthorized")
	}
}

func TestClientService_SetFingerprint(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test client
	now := time.Now().Unix()
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:ff",
		LastSeen: now,
	})

	// Create service
	trans, _ := newTestClientTransport(server.URL())
	svc := NewClientService(trans)

	// Test SetFingerprint
	err := svc.SetFingerprint(context.Background(), "default", "aa:bb:cc:dd:ee:ff", 42)
	if err != nil {
		t.Fatalf("SetFingerprint failed: %v", err)
	}

	// Verify fingerprint was set
	client := server.State().GetClient("aa:bb:cc:dd:ee:ff")
	if client == nil {
		t.Fatal("Client not found after set fingerprint")
	}

	if client.DeviceIDOverride != 42 {
		t.Errorf("Expected DeviceIDOverride 42, got %d", client.DeviceIDOverride)
	}
}
