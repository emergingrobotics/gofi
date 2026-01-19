package services

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/transport"
)

func TestSiteService_List(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	config := transport.DefaultConfig(server.URL())
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	svc := NewSiteService(trans)

	sites, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(sites) == 0 {
		t.Error("Expected at least one site")
	}

	// Should have default site
	foundDefault := false
	for _, site := range sites {
		if site.ID == "default" {
			foundDefault = true
			break
		}
	}

	if !foundDefault {
		t.Error("Default site not found in list")
	}
}

func TestSiteService_Get(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	config := transport.DefaultConfig(server.URL())
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	svc := NewSiteService(trans)

	site, err := svc.Get(context.Background(), "default")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if site.ID != "default" {
		t.Errorf("ID = %s, want default", site.ID)
	}
}

func TestSiteService_Get_NotFound(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	config := transport.DefaultConfig(server.URL())
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	svc := NewSiteService(trans)

	_, err = svc.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Get() should return error for nonexistent site")
	}
}

func TestSiteService_Create(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	config := transport.DefaultConfig(server.URL())
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	svc := NewSiteService(trans)

	site, err := svc.Create(context.Background(), "Test Site", "test")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if site.Desc != "Test Site" {
		t.Errorf("Desc = %s, want 'Test Site'", site.Desc)
	}

	// Verify it was added to the mock state
	created, exists := server.State().GetSite("test")
	if !exists {
		t.Error("Site not found in mock state")
	}

	if created.Desc != "Test Site" {
		t.Errorf("State site Desc = %s, want 'Test Site'", created.Desc)
	}
}

func TestSiteService_Health(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	config := transport.DefaultConfig(server.URL())
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	svc := NewSiteService(trans)

	health, err := svc.Health(context.Background(), "default")
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}

	if len(health) == 0 {
		t.Error("Expected health data")
	}
}

func TestSiteService_SysInfo(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	config := transport.DefaultConfig(server.URL())
	config.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	trans, err := transport.New(config)
	if err != nil {
		t.Fatalf("transport.New() error = %v", err)
	}
	defer trans.Close()

	svc := NewSiteService(trans)

	sysInfo, err := svc.SysInfo(context.Background(), "default")
	if err != nil {
		t.Fatalf("SysInfo() error = %v", err)
	}

	if sysInfo.Hostname != "UDM-Pro" {
		t.Errorf("Hostname = %s, want UDM-Pro", sysInfo.Hostname)
	}
}
