package services

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

func newTestTransport(url string) (transport.Transport, error) {
	config := transport.DefaultConfig(url)
	config.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return transport.New(config)
}

func TestDeviceService_List(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test devices
	server.State().AddDevice(&types.Device{
		ID:      "device1",
		MAC:     "aa:bb:cc:dd:ee:f1",
		Model:   "UAP-AC-PRO",
		Type:    "uap",
		Name:    "AP 1",
		Adopted: true,
		State:   types.DeviceStateConnected,
	})
	server.State().AddDevice(&types.Device{
		ID:      "device2",
		MAC:     "aa:bb:cc:dd:ee:f2",
		Model:   "USW-24-POE",
		Type:    "usw",
		Name:    "Switch 1",
		Adopted: true,
		State:   types.DeviceStateConnected,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Test List
	devices, err := svc.List(context.Background(), "default")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(devices))
	}
}

func TestDeviceService_ListBasic(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "UAP-AC-PRO",
		Type:  "uap",
		Name:  "AP 1",
		State: types.DeviceStateConnected,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Test ListBasic
	basics, err := svc.ListBasic(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListBasic failed: %v", err)
	}

	if len(basics) != 1 {
		t.Errorf("Expected 1 device, got %d", len(basics))
	}

	if basics[0].MAC != "aa:bb:cc:dd:ee:f1" {
		t.Errorf("Expected MAC aa:bb:cc:dd:ee:f1, got %s", basics[0].MAC)
	}
}

func TestDeviceService_Get(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "UAP-AC-PRO",
		Type:  "uap",
		Name:  "AP 1",
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Test Get - success
	device, err := svc.Get(context.Background(), "default", "device1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if device.MAC != "aa:bb:cc:dd:ee:f1" {
		t.Errorf("Expected MAC aa:bb:cc:dd:ee:f1, got %s", device.MAC)
	}

	// Test Get - not found
	_, err = svc.Get(context.Background(), "default", "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestDeviceService_GetByMAC(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "UAP-AC-PRO",
		Type:  "uap",
		Name:  "AP 1",
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Test GetByMAC - success
	device, err := svc.GetByMAC(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("GetByMAC failed: %v", err)
	}

	if device.ID != "device1" {
		t.Errorf("Expected ID device1, got %s", device.ID)
	}

	// Test GetByMAC - case insensitive and no colons
	device, err = svc.GetByMAC(context.Background(), "default", "AABBCCDDEEF1")
	if err != nil {
		t.Fatalf("GetByMAC with different format failed: %v", err)
	}

	if device.ID != "device1" {
		t.Errorf("Expected ID device1, got %s", device.ID)
	}

	// Test GetByMAC - not found
	_, err = svc.GetByMAC(context.Background(), "default", "ff:ff:ff:ff:ff:ff")
	if err == nil {
		t.Error("Expected error for nonexistent MAC")
	}
}

func TestDeviceService_Update(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "UAP-AC-PRO",
		Type:  "uap",
		Name:  "Old Name",
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Update device
	device := &types.Device{
		ID:   "device1",
		Name: "New Name",
	}

	updated, err := svc.Update(context.Background(), "default", device)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
	}

	// Verify it was saved
	saved, exists := server.State().GetDevice("device1")
	if !exists {
		t.Fatal("Device not found after update")
	}
	if saved.Name != "New Name" {
		t.Errorf("Expected saved name 'New Name', got '%s'", saved.Name)
	}
}

func TestDeviceService_Adopt(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add unadopted device
	server.State().AddDevice(&types.Device{
		ID:      "device1",
		MAC:     "aa:bb:cc:dd:ee:f1",
		Model:   "UAP-AC-PRO",
		Type:    "uap",
		Adopted: false,
		State:   types.DeviceStatePending,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Adopt device
	err := svc.Adopt(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("Adopt failed: %v", err)
	}

	// Verify device was adopted
	device, _ := server.State().GetDevice("device1")
	if !device.Adopted {
		t.Error("Device should be adopted")
	}
	if device.State != types.DeviceStateConnected {
		t.Errorf("Expected state Connected, got %v", device.State)
	}
}

func TestDeviceService_Restart(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "UAP-AC-PRO",
		Type:  "uap",
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Restart device
	err := svc.Restart(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("Restart failed: %v", err)
	}
}

func TestDeviceService_ForceProvision(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "UAP-AC-PRO",
		Type:  "uap",
		State: types.DeviceStateConnected,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Force provision
	err := svc.ForceProvision(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("ForceProvision failed: %v", err)
	}

	// Verify state changed
	device, _ := server.State().GetDevice("device1")
	if device.State != types.DeviceStateProvisioning {
		t.Errorf("Expected state Provisioning, got %v", device.State)
	}
}

func TestDeviceService_Upgrade(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "UAP-AC-PRO",
		Type:  "uap",
		State: types.DeviceStateConnected,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Upgrade
	err := svc.Upgrade(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("Upgrade failed: %v", err)
	}

	// Verify state changed
	device, _ := server.State().GetDevice("device1")
	if device.State != types.DeviceStateUpgrading {
		t.Errorf("Expected state Upgrading, got %v", device.State)
	}
}

func TestDeviceService_UpgradeExternal(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "UAP-AC-PRO",
		Type:  "uap",
		State: types.DeviceStateConnected,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Upgrade external
	err := svc.UpgradeExternal(context.Background(), "default", "aa:bb:cc:dd:ee:f1", "https://example.com/firmware.bin")
	if err != nil {
		t.Fatalf("UpgradeExternal failed: %v", err)
	}

	// Verify state changed
	device, _ := server.State().GetDevice("device1")
	if device.State != types.DeviceStateUpgrading {
		t.Errorf("Expected state Upgrading, got %v", device.State)
	}
}

func TestDeviceService_Locate(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add device
	server.State().AddDevice(&types.Device{
		ID:          "device1",
		MAC:         "aa:bb:cc:dd:ee:f1",
		Model:       "UAP-AC-PRO",
		Type:        "uap",
		LEDOverride: "default",
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Locate
	err := svc.Locate(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("Locate failed: %v", err)
	}

	// Verify LED override set
	device, _ := server.State().GetDevice("device1")
	if device.LEDOverride != "on" {
		t.Errorf("Expected LEDOverride 'on', got '%s'", device.LEDOverride)
	}
}

func TestDeviceService_Unlocate(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add device with locate enabled
	server.State().AddDevice(&types.Device{
		ID:          "device1",
		MAC:         "aa:bb:cc:dd:ee:f1",
		Model:       "UAP-AC-PRO",
		Type:        "uap",
		LEDOverride: "on",
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Unlocate
	err := svc.Unlocate(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("Unlocate failed: %v", err)
	}

	// Verify LED override cleared
	device, _ := server.State().GetDevice("device1")
	if device.LEDOverride != "default" {
		t.Errorf("Expected LEDOverride 'default', got '%s'", device.LEDOverride)
	}
}

func TestDeviceService_PowerCyclePort(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add switch device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "USW-24-POE",
		Type:  "usw",
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Power cycle port
	err := svc.PowerCyclePort(context.Background(), "default", "aa:bb:cc:dd:ee:f1", 5)
	if err != nil {
		t.Fatalf("PowerCyclePort failed: %v", err)
	}
}

func TestDeviceService_SpectrumScan(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add AP device
	server.State().AddDevice(&types.Device{
		ID:    "device1",
		MAC:   "aa:bb:cc:dd:ee:f1",
		Model: "UAP-AC-PRO",
		Type:  "uap",
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewDeviceService(trans)

	// Spectrum scan
	err := svc.SpectrumScan(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("SpectrumScan failed: %v", err)
	}
}
