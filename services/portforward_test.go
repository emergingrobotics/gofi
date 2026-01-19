package services

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

func newTestPortForwardTransport(url string) (transport.Transport, error) {
	config := transport.DefaultConfig(url)
	config.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return transport.New(config)
}

func TestPortForwardService_List(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test port forwards
	server.State().AddPortForward(&types.PortForward{
		ID:       "pf1",
		Name:     "Test Forward 1",
		Enabled:  true,
		Protocol: "tcp",
		DstPort:  "80",
		FwdIP:    "192.168.1.100",
		FwdPort:  "8080",
	})
	server.State().AddPortForward(&types.PortForward{
		ID:       "pf2",
		Name:     "Test Forward 2",
		Enabled:  false,
		Protocol: "udp",
		DstPort:  "53",
		FwdIP:    "192.168.1.101",
		FwdPort:  "5353",
	})

	// Create service
	trans, _ := newTestPortForwardTransport(server.URL())
	svc := NewPortForwardService(trans)

	// Test List
	forwards, err := svc.List(context.Background(), "default")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(forwards) != 2 {
		t.Errorf("Expected 2 port forwards, got %d", len(forwards))
	}
}

func TestPortForwardService_Get(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test port forward
	server.State().AddPortForward(&types.PortForward{
		ID:       "pf1",
		Name:     "Test Forward",
		Enabled:  true,
		Protocol: "tcp",
		DstPort:  "80",
		FwdIP:    "192.168.1.100",
		FwdPort:  "8080",
	})

	// Create service
	trans, _ := newTestPortForwardTransport(server.URL())
	svc := NewPortForwardService(trans)

	// Test Get
	forward, err := svc.Get(context.Background(), "default", "pf1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if forward.Name != "Test Forward" {
		t.Errorf("Expected name 'Test Forward', got %s", forward.Name)
	}

	if !forward.Enabled {
		t.Error("Expected forward to be enabled")
	}
}

func TestPortForwardService_Create(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestPortForwardTransport(server.URL())
	svc := NewPortForwardService(trans)

	// Test Create
	newForward := &types.PortForward{
		Name:     "New Forward",
		Enabled:  true,
		Protocol: "tcp",
		DstPort:  "443",
		FwdIP:    "192.168.1.100",
		FwdPort:  "8443",
	}

	created, err := svc.Create(context.Background(), "default", newForward)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.Name != "New Forward" {
		t.Errorf("Expected name 'New Forward', got %s", created.Name)
	}

	if created.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestPortForwardService_Update(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test port forward
	server.State().AddPortForward(&types.PortForward{
		ID:       "pf1",
		Name:     "Test Forward",
		Enabled:  true,
		Protocol: "tcp",
		DstPort:  "80",
		FwdIP:    "192.168.1.100",
		FwdPort:  "8080",
	})

	// Create service
	trans, _ := newTestPortForwardTransport(server.URL())
	svc := NewPortForwardService(trans)

	// Test Update
	forward, _ := svc.Get(context.Background(), "default", "pf1")
	forward.Name = "Updated Forward"
	forward.FwdPort = "9090"

	updated, err := svc.Update(context.Background(), "default", forward)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != "Updated Forward" {
		t.Errorf("Expected name 'Updated Forward', got %s", updated.Name)
	}

	if updated.FwdPort != "9090" {
		t.Errorf("Expected fwd port 9090, got %s", updated.FwdPort)
	}
}

func TestPortForwardService_Delete(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test port forward
	server.State().AddPortForward(&types.PortForward{
		ID:       "pf1",
		Name:     "Test Forward",
		Enabled:  true,
		Protocol: "tcp",
		DstPort:  "80",
	})

	// Create service
	trans, _ := newTestPortForwardTransport(server.URL())
	svc := NewPortForwardService(trans)

	// Test Delete
	err := svc.Delete(context.Background(), "default", "pf1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify
	_, err = svc.Get(context.Background(), "default", "pf1")
	if err == nil {
		t.Error("Expected error when getting deleted port forward")
	}
}

func TestPortForwardService_Enable(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add disabled port forward
	server.State().AddPortForward(&types.PortForward{
		ID:       "pf1",
		Name:     "Test Forward",
		Enabled:  false,
		Protocol: "tcp",
		DstPort:  "80",
	})

	// Create service
	trans, _ := newTestPortForwardTransport(server.URL())
	svc := NewPortForwardService(trans)

	// Test Enable
	err := svc.Enable(context.Background(), "default", "pf1")
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}

	// Verify
	forward := server.State().GetPortForward("pf1")
	if forward == nil {
		t.Fatal("Port forward not found")
	}

	if !forward.Enabled {
		t.Error("Expected port forward to be enabled")
	}
}

func TestPortForwardService_Disable(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add enabled port forward
	server.State().AddPortForward(&types.PortForward{
		ID:       "pf1",
		Name:     "Test Forward",
		Enabled:  true,
		Protocol: "tcp",
		DstPort:  "80",
	})

	// Create service
	trans, _ := newTestPortForwardTransport(server.URL())
	svc := NewPortForwardService(trans)

	// Test Disable
	err := svc.Disable(context.Background(), "default", "pf1")
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}

	// Verify
	forward := server.State().GetPortForward("pf1")
	if forward == nil {
		t.Fatal("Port forward not found")
	}

	if forward.Enabled {
		t.Error("Expected port forward to be disabled")
	}
}
