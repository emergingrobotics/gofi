package services

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

func newTestSettingTransport(url string) (transport.Transport, error) {
	config := transport.DefaultConfig(url)
	config.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return transport.New(config)
}

func TestSettingService_Get(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test setting
	server.State().AddSetting(&types.Setting{
		Key:    types.SettingKeyMgmt,
		SiteID: "default",
	})

	// Create service
	trans, _ := newTestSettingTransport(server.URL())
	svc := NewSettingService(trans)

	// Test Get
	setting, err := svc.Get(context.Background(), "default", types.SettingKeyMgmt)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if setting == nil {
		t.Fatal("Expected setting, got nil")
	}
}

func TestSettingService_Update(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test setting
	server.State().AddSetting(&types.Setting{
		Key:    types.SettingKeyMgmt,
		SiteID: "default",
	})

	// Create service
	trans, _ := newTestSettingTransport(server.URL())
	svc := NewSettingService(trans)

	// Test Update
	setting := &types.Setting{
		Key:    types.SettingKeyMgmt,
		SiteID: "default",
	}

	err := svc.Update(context.Background(), "default", setting)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

func TestSettingService_ListRadiusProfiles(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test RADIUS profiles
	server.State().AddRADIUSProfile(&types.RADIUSProfile{
		ID:   "radius1",
		Name: "Test RADIUS 1",
	})
	server.State().AddRADIUSProfile(&types.RADIUSProfile{
		ID:   "radius2",
		Name: "Test RADIUS 2",
	})

	// Create service
	trans, _ := newTestSettingTransport(server.URL())
	svc := NewSettingService(trans)

	// Test List
	profiles, err := svc.ListRadiusProfiles(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListRadiusProfiles failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("Expected 2 RADIUS profiles, got %d", len(profiles))
	}
}

func TestSettingService_GetRadiusProfile(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test RADIUS profile
	server.State().AddRADIUSProfile(&types.RADIUSProfile{
		ID:   "radius1",
		Name: "Test RADIUS",
	})

	// Create service
	trans, _ := newTestSettingTransport(server.URL())
	svc := NewSettingService(trans)

	// Test Get
	profile, err := svc.GetRadiusProfile(context.Background(), "default", "radius1")
	if err != nil {
		t.Fatalf("GetRadiusProfile failed: %v", err)
	}

	if profile.Name != "Test RADIUS" {
		t.Errorf("Expected name 'Test RADIUS', got %s", profile.Name)
	}
}

func TestSettingService_CreateRadiusProfile(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestSettingTransport(server.URL())
	svc := NewSettingService(trans)

	// Test Create
	newProfile := &types.RADIUSProfile{
		Name: "New RADIUS",
	}

	created, err := svc.CreateRadiusProfile(context.Background(), "default", newProfile)
	if err != nil {
		t.Fatalf("CreateRadiusProfile failed: %v", err)
	}

	if created.Name != "New RADIUS" {
		t.Errorf("Expected name 'New RADIUS', got %s", created.Name)
	}

	if created.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestSettingService_UpdateRadiusProfile(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test RADIUS profile
	server.State().AddRADIUSProfile(&types.RADIUSProfile{
		ID:   "radius1",
		Name: "Test RADIUS",
	})

	// Create service
	trans, _ := newTestSettingTransport(server.URL())
	svc := NewSettingService(trans)

	// Test Update
	profile, _ := svc.GetRadiusProfile(context.Background(), "default", "radius1")
	profile.Name = "Updated RADIUS"

	updated, err := svc.UpdateRadiusProfile(context.Background(), "default", profile)
	if err != nil {
		t.Fatalf("UpdateRadiusProfile failed: %v", err)
	}

	if updated.Name != "Updated RADIUS" {
		t.Errorf("Expected name 'Updated RADIUS', got %s", updated.Name)
	}
}

func TestSettingService_DeleteRadiusProfile(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test RADIUS profile
	server.State().AddRADIUSProfile(&types.RADIUSProfile{
		ID:   "radius1",
		Name: "Test RADIUS",
	})

	// Create service
	trans, _ := newTestSettingTransport(server.URL())
	svc := NewSettingService(trans)

	// Test Delete
	err := svc.DeleteRadiusProfile(context.Background(), "default", "radius1")
	if err != nil {
		t.Fatalf("DeleteRadiusProfile failed: %v", err)
	}

	// Verify
	_, err = svc.GetRadiusProfile(context.Background(), "default", "radius1")
	if err == nil {
		t.Error("Expected error when getting deleted RADIUS profile")
	}
}

func TestSettingService_GetDynamicDNS(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test Dynamic DNS
	server.State().SetDynamicDNS(&types.DynamicDNS{
		ID:       "ddns1",
		Service:  "dyndns",
		Enabled:  true,
		Hostname: "test.dyndns.org",
	})

	// Create service
	trans, _ := newTestSettingTransport(server.URL())
	svc := NewSettingService(trans)

	// Test Get
	ddns, err := svc.GetDynamicDNS(context.Background(), "default")
	if err != nil {
		t.Fatalf("GetDynamicDNS failed: %v", err)
	}

	if ddns == nil {
		t.Fatal("Expected Dynamic DNS, got nil")
	}

	if ddns.Hostname != "test.dyndns.org" {
		t.Errorf("Expected hostname 'test.dyndns.org', got %s", ddns.Hostname)
	}
}

func TestSettingService_UpdateDynamicDNS(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestSettingTransport(server.URL())
	svc := NewSettingService(trans)

	// Test Update
	newDDNS := &types.DynamicDNS{
		Service:  "dyndns",
		Enabled:  true,
		Hostname: "new.dyndns.org",
	}

	err := svc.UpdateDynamicDNS(context.Background(), "default", newDDNS)
	if err != nil {
		t.Fatalf("UpdateDynamicDNS failed: %v", err)
	}

	// Verify
	ddns := server.State().GetDynamicDNS()
	if ddns == nil {
		t.Fatal("Dynamic DNS not set")
	}

	if ddns.Hostname != "new.dyndns.org" {
		t.Errorf("Expected hostname 'new.dyndns.org', got %s", ddns.Hostname)
	}
}
