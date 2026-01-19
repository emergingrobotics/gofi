package services

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

func newTestPortProfileTransport(url string) (transport.Transport, error) {
	config := transport.DefaultConfig(url)
	config.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return transport.New(config)
}

func TestPortProfileService_List(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test port profiles
	server.State().AddPortProfile(&types.PortProfile{
		ID:      "pp1",
		Name:    "Test Profile 1",
		Forward: "all",
		POEMode: "auto",
	})
	server.State().AddPortProfile(&types.PortProfile{
		ID:      "pp2",
		Name:    "Test Profile 2",
		Forward: "native",
		POEMode: "off",
	})

	// Create service
	trans, _ := newTestPortProfileTransport(server.URL())
	svc := NewPortProfileService(trans)

	// Test List
	profiles, err := svc.List(context.Background(), "default")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("Expected 2 port profiles, got %d", len(profiles))
	}
}

func TestPortProfileService_Get(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test port profile
	server.State().AddPortProfile(&types.PortProfile{
		ID:      "pp1",
		Name:    "Test Profile",
		Forward: "all",
		POEMode: "auto",
	})

	// Create service
	trans, _ := newTestPortProfileTransport(server.URL())
	svc := NewPortProfileService(trans)

	// Test Get
	profile, err := svc.Get(context.Background(), "default", "pp1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if profile.Name != "Test Profile" {
		t.Errorf("Expected name 'Test Profile', got %s", profile.Name)
	}

	if profile.Forward != "all" {
		t.Errorf("Expected forward 'all', got %s", profile.Forward)
	}
}

func TestPortProfileService_Create(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestPortProfileTransport(server.URL())
	svc := NewPortProfileService(trans)

	// Test Create
	newProfile := &types.PortProfile{
		Name:    "New Profile",
		Forward: "native",
		POEMode: "auto",
	}

	created, err := svc.Create(context.Background(), "default", newProfile)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.Name != "New Profile" {
		t.Errorf("Expected name 'New Profile', got %s", created.Name)
	}

	if created.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestPortProfileService_Update(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test port profile
	server.State().AddPortProfile(&types.PortProfile{
		ID:      "pp1",
		Name:    "Test Profile",
		Forward: "all",
		POEMode: "auto",
	})

	// Create service
	trans, _ := newTestPortProfileTransport(server.URL())
	svc := NewPortProfileService(trans)

	// Test Update
	profile, _ := svc.Get(context.Background(), "default", "pp1")
	profile.Name = "Updated Profile"
	profile.POEMode = "off"

	updated, err := svc.Update(context.Background(), "default", profile)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != "Updated Profile" {
		t.Errorf("Expected name 'Updated Profile', got %s", updated.Name)
	}

	if updated.POEMode != "off" {
		t.Errorf("Expected POE mode 'off', got %s", updated.POEMode)
	}
}

func TestPortProfileService_Delete(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test port profile
	server.State().AddPortProfile(&types.PortProfile{
		ID:      "pp1",
		Name:    "Test Profile",
		Forward: "all",
	})

	// Create service
	trans, _ := newTestPortProfileTransport(server.URL())
	svc := NewPortProfileService(trans)

	// Test Delete
	err := svc.Delete(context.Background(), "default", "pp1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify
	_, err = svc.Get(context.Background(), "default", "pp1")
	if err == nil {
		t.Error("Expected error when getting deleted port profile")
	}
}
