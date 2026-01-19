package mock

import (
	"testing"

	"github.com/unifi-go/gofi/types"
)

func TestNewState(t *testing.T) {
	state := NewState()

	// Should have default admin user
	if !state.ValidateCredentials("admin", "admin") {
		t.Error("Default admin user not created")
	}

	// Should have default site
	site, exists := state.GetSite("default")
	if !exists {
		t.Fatal("Default site not created")
	}

	if site.Name != "default" {
		t.Errorf("Default site name = %s, want default", site.Name)
	}
}

func TestState_AddUser(t *testing.T) {
	state := NewState()

	state.AddUser("testuser", "testpass")

	if !state.ValidateCredentials("testuser", "testpass") {
		t.Error("Added user not found")
	}

	if state.ValidateCredentials("testuser", "wrongpass") {
		t.Error("Invalid password accepted")
	}
}

func TestState_Session(t *testing.T) {
	state := NewState()

	session := &Session{
		Username:  "admin",
		CSRFToken: "test-csrf",
	}

	// Create session
	state.CreateSession("token123", session)

	// Get session
	retrieved, exists := state.GetSession("token123")
	if !exists {
		t.Fatal("Session not found")
	}

	if retrieved.Username != "admin" {
		t.Errorf("Username = %s, want admin", retrieved.Username)
	}

	if retrieved.CSRFToken != "test-csrf" {
		t.Errorf("CSRFToken = %s, want test-csrf", retrieved.CSRFToken)
	}

	// Delete session
	state.DeleteSession("token123")

	_, exists = state.GetSession("token123")
	if exists {
		t.Error("Session should be deleted")
	}
}

func TestState_Sites(t *testing.T) {
	state := NewState()

	// Add site
	site := &types.Site{
		ID:   "test-site",
		Name: "test",
		Desc: "Test Site",
	}
	state.AddSite(site)

	// Get site
	retrieved, exists := state.GetSite("test-site")
	if !exists {
		t.Fatal("Site not found")
	}

	if retrieved.Name != "test" {
		t.Errorf("Name = %s, want test", retrieved.Name)
	}

	// List sites (should have default + test)
	sites := state.ListSites()
	if len(sites) != 2 {
		t.Errorf("len(sites) = %d, want 2", len(sites))
	}

	// Delete site
	state.DeleteSite("test-site")

	_, exists = state.GetSite("test-site")
	if exists {
		t.Error("Site should be deleted")
	}
}

func TestState_Devices(t *testing.T) {
	state := NewState()

	// Add device
	device := &types.Device{
		ID:   "dev1",
		MAC:  "aa:bb:cc:dd:ee:ff",
		Name: "Test AP",
	}
	state.AddDevice(device)

	// Get device
	retrieved, exists := state.GetDevice("dev1")
	if !exists {
		t.Fatal("Device not found")
	}

	if retrieved.Name != "Test AP" {
		t.Errorf("Name = %s, want Test AP", retrieved.Name)
	}

	// List devices
	devices := state.ListDevices()
	if len(devices) != 1 {
		t.Errorf("len(devices) = %d, want 1", len(devices))
	}

	// Delete device
	state.DeleteDevice("dev1")

	_, exists = state.GetDevice("dev1")
	if exists {
		t.Error("Device should be deleted")
	}
}

func TestState_Networks(t *testing.T) {
	state := NewState()

	// Add network
	network := &types.Network{
		ID:   "net1",
		Name: "LAN",
	}
	state.AddNetwork(network)

	// Get network
	retrieved, exists := state.GetNetwork("net1")
	if !exists {
		t.Fatal("Network not found")
	}

	if retrieved.Name != "LAN" {
		t.Errorf("Name = %s, want LAN", retrieved.Name)
	}

	// List networks
	networks := state.ListNetworks()
	if len(networks) != 1 {
		t.Errorf("len(networks) = %d, want 1", len(networks))
	}

	// Delete network
	state.DeleteNetwork("net1")

	_, exists = state.GetNetwork("net1")
	if exists {
		t.Error("Network should be deleted")
	}
}

func TestState_Reset(t *testing.T) {
	state := NewState()

	// Add some data
	state.CreateSession("token", &Session{Username: "admin"})
	state.AddDevice(&types.Device{ID: "dev1"})
	state.AddNetwork(&types.Network{ID: "net1"})

	// Reset
	state.Reset()

	// Sessions should be cleared
	_, exists := state.GetSession("token")
	if exists {
		t.Error("Session should be cleared after reset")
	}

	// Devices should be cleared
	devices := state.ListDevices()
	if len(devices) != 0 {
		t.Errorf("Devices should be cleared after reset, got %d", len(devices))
	}

	// Networks should be cleared
	networks := state.ListNetworks()
	if len(networks) != 0 {
		t.Errorf("Networks should be cleared after reset, got %d", len(networks))
	}

	// Default site should still exist
	_, exists = state.GetSite("default")
	if !exists {
		t.Error("Default site should exist after reset")
	}
}

func TestState_ConcurrentAccess(t *testing.T) {
	state := NewState()

	// Test concurrent reads and writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			device := &types.Device{
				ID:   "dev",
				Name: "Device",
			}
			state.AddDevice(device)
			state.DeleteDevice("dev")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			state.GetDevice("dev")
			state.ListDevices()
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done
}
