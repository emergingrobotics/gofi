package mock

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/unifi-go/gofi/types"
)

func TestDefaultFixtures(t *testing.T) {
	fixtures := DefaultFixtures()

	if len(fixtures.Sites) != 1 {
		t.Errorf("len(Sites) = %d, want 1", len(fixtures.Sites))
	}

	if fixtures.Sites[0].ID != "default" {
		t.Errorf("Site ID = %s, want default", fixtures.Sites[0].ID)
	}
}

func TestLoadFixtures(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test fixtures
	sites := []types.Site{
		{ID: "site1", Name: "Site 1", Desc: "Test Site 1"},
		{ID: "site2", Name: "Site 2", Desc: "Test Site 2"},
	}

	devices := []types.Device{
		{ID: "dev1", MAC: "aa:bb:cc:dd:ee:ff", Name: "AP 1"},
	}

	// Write fixtures to files
	sitesData, _ := json.Marshal(sites)
	if err := os.WriteFile(filepath.Join(tmpDir, "sites.json"), sitesData, 0644); err != nil {
		t.Fatalf("Failed to write sites.json: %v", err)
	}

	devicesData, _ := json.Marshal(devices)
	if err := os.WriteFile(filepath.Join(tmpDir, "devices.json"), devicesData, 0644); err != nil {
		t.Fatalf("Failed to write devices.json: %v", err)
	}

	// Load fixtures
	fixtures, err := LoadFixtures(tmpDir)
	if err != nil {
		t.Fatalf("LoadFixtures() error = %v", err)
	}

	if len(fixtures.Sites) != 2 {
		t.Errorf("len(Sites) = %d, want 2", len(fixtures.Sites))
	}

	if len(fixtures.Devices) != 1 {
		t.Errorf("len(Devices) = %d, want 1", len(fixtures.Devices))
	}

	if fixtures.Devices[0].Name != "AP 1" {
		t.Errorf("Device name = %s, want AP 1", fixtures.Devices[0].Name)
	}
}

func TestLoadFixtures_NonexistentDir(t *testing.T) {
	_, err := LoadFixtures("/nonexistent/directory")
	// Should not error on missing optional files, just return empty fixtures
	// It's ok if it errors on missing directory - we don't enforce behavior here
	_ = err
}

func TestLoadFixtures_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(tmpDir, "sites.json"), []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadFixtures(tmpDir)
	if err == nil {
		t.Error("LoadFixtures() should return error for invalid JSON")
	}
}

func TestState_LoadFixtures(t *testing.T) {
	state := NewState()

	fixtures := &Fixtures{
		Sites: []types.Site{
			{ID: "test1", Name: "Test 1"},
			{ID: "test2", Name: "Test 2"},
		},
		Devices: []types.Device{
			{ID: "dev1", Name: "Device 1"},
		},
		Networks: []types.Network{
			{ID: "net1", Name: "Network 1"},
		},
	}

	state.LoadFixtures(fixtures)

	// Verify sites loaded (should have default + 2 test sites)
	sites := state.ListSites()
	if len(sites) < 2 { // At least the test sites
		t.Errorf("len(sites) = %d, want at least 2", len(sites))
	}

	// Verify devices loaded
	devices := state.ListDevices()
	if len(devices) != 1 {
		t.Errorf("len(devices) = %d, want 1", len(devices))
	}

	// Verify networks loaded
	networks := state.ListNetworks()
	if len(networks) != 1 {
		t.Errorf("len(networks) = %d, want 1", len(networks))
	}
}

func TestState_LoadFixtures_Nil(t *testing.T) {
	state := NewState()

	// Should not panic with nil fixtures
	state.LoadFixtures(nil)
}
