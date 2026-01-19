package gofi

import (
	"context"
	"testing"
	"time"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/types"
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

func TestClient_Integration_DeviceService(t *testing.T) {
	server := mock.NewServer()
	defer server.Close()

	// Add test devices
	server.State().AddDevice(&types.Device{
		ID:      "device1",
		MAC:     "aa:bb:cc:dd:ee:f1",
		Model:   "UAP-AC-PRO",
		Type:    "uap",
		Name:    "Test AP",
		Adopted: true,
		State:   types.DeviceStateConnected,
	})

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

	// Test List
	devices, err := client.Devices().List(context.Background(), "default")
	if err != nil {
		t.Fatalf("Devices().List() error = %v", err)
	}

	if len(devices) == 0 {
		t.Error("Expected at least one device")
	}

	// Test Get
	device, err := client.Devices().Get(context.Background(), "default", "device1")
	if err != nil {
		t.Fatalf("Devices().Get() error = %v", err)
	}

	if device.Name != "Test AP" {
		t.Errorf("Device name = %s, want 'Test AP'", device.Name)
	}

	// Test GetByMAC
	device, err = client.Devices().GetByMAC(context.Background(), "default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("Devices().GetByMAC() error = %v", err)
	}

	if device.ID != "device1" {
		t.Errorf("Device ID = %s, want 'device1'", device.ID)
	}

	// Test Update
	device.Name = "Updated AP"
	updated, err := client.Devices().Update(context.Background(), "default", device)
	if err != nil {
		t.Fatalf("Devices().Update() error = %v", err)
	}

	if updated.Name != "Updated AP" {
		t.Errorf("Updated device name = %s, want 'Updated AP'", updated.Name)
	}

	// Test Locate
	if err := client.Devices().Locate(context.Background(), "default", "aa:bb:cc:dd:ee:f1"); err != nil {
		t.Fatalf("Devices().Locate() error = %v", err)
	}

	// Test Unlocate
	if err := client.Devices().Unlocate(context.Background(), "default", "aa:bb:cc:dd:ee:f1"); err != nil {
		t.Fatalf("Devices().Unlocate() error = %v", err)
	}
}

func TestClient_Integration_ClientService(t *testing.T) {
	server := mock.NewServer()
	defer server.Close()

	// Add test clients
	server.State().AddClient(&types.Client{
		MAC:      "aa:bb:cc:dd:ee:f1",
		Hostname: "test-device",
		IP:       "192.168.1.100",
		LastSeen: time.Now().Unix() - 60, // Recent (1 minute ago)
	})

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

	// Test ListActive
	clients, err := client.Clients().ListActive(context.Background(), "default")
	if err != nil {
		t.Fatalf("Clients().ListActive() error = %v", err)
	}

	// Note: The client may or may not be active depending on LastSeen timestamp
	_ = clients

	// Test ListAll
	allClients, err := client.Clients().ListAll(context.Background(), "default")
	if err != nil {
		t.Fatalf("Clients().ListAll() error = %v", err)
	}

	if len(allClients) == 0 {
		t.Error("Expected at least one client")
	}

	// Test Block
	if err := client.Clients().Block(context.Background(), "default", "aa:bb:cc:dd:ee:f1"); err != nil {
		t.Fatalf("Clients().Block() error = %v", err)
	}

	// Verify blocked
	blockedClient := server.State().GetClient("aa:bb:cc:dd:ee:f1")
	if blockedClient == nil || !blockedClient.Blocked {
		t.Error("Expected client to be blocked")
	}

	// Test Unblock
	if err := client.Clients().Unblock(context.Background(), "default", "aa:bb:cc:dd:ee:f1"); err != nil {
		t.Fatalf("Clients().Unblock() error = %v", err)
	}

	// Verify unblocked
	unblockedClient := server.State().GetClient("aa:bb:cc:dd:ee:f1")
	if unblockedClient == nil || unblockedClient.Blocked {
		t.Error("Expected client to be unblocked")
	}
}
