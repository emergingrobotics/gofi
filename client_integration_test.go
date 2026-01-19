package gofi

import (
	"context"
	"testing"

	"github.com/unifi-go/gofi/mock"
)

func TestClient_Integration_SiteService(t *testing.T) {
	server := mock.NewServer()
	defer server.Close()

	config := &Config{
		Host:          server.Host(),
		Port:          server.Port(),
		Username:      "admin",
		Password:      "admin",
		SkipTLSVerify: true,
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Connect
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Disconnect(context.Background())

	// Test Sites service
	sites, err := client.Sites().List(context.Background())
	if err != nil {
		t.Fatalf("Sites().List() error = %v", err)
	}

	if len(sites) == 0 {
		t.Error("Expected at least one site")
	}

	// Test Health
	health, err := client.Sites().Health(context.Background(), "default")
	if err != nil {
		t.Fatalf("Sites().Health() error = %v", err)
	}

	if len(health) == 0 {
		t.Error("Expected health data")
	}

	// Test SysInfo
	sysInfo, err := client.Sites().SysInfo(context.Background(), "default")
	if err != nil {
		t.Fatalf("Sites().SysInfo() error = %v", err)
	}

	if sysInfo == nil {
		t.Error("Expected sysinfo data")
	}

	// Test Create
	newSite, err := client.Sites().Create(context.Background(), "Test Site", "test-site")
	if err != nil {
		t.Fatalf("Sites().Create() error = %v", err)
	}

	if newSite.Desc != "Test Site" {
		t.Errorf("Created site Desc = %s, want 'Test Site'", newSite.Desc)
	}

	// Verify it exists
	retrieved, err := client.Sites().Get(context.Background(), "test-site")
	if err != nil {
		t.Fatalf("Sites().Get() error = %v", err)
	}

	if retrieved.Desc != "Test Site" {
		t.Errorf("Retrieved site Desc = %s, want 'Test Site'", retrieved.Desc)
	}
}
