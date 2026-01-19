package mock

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

var testHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func TestHandleDeviceStat(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add a test device
	testDevice := &types.Device{
		ID:      "test-device-1",
		MAC:     "aa:bb:cc:dd:ee:ff",
		Model:   "UAP-AC-PRO",
		Type:    "uap",
		Name:    "Test AP",
		Adopted: true,
		State:   types.DeviceStateConnected,
	}
	server.State().AddDevice(testDevice)

	// Make request
	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/stat/device", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.Device]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if apiResp.Meta.RC != "ok" {
		t.Errorf("Expected RC=ok, got %s", apiResp.Meta.RC)
	}

	if len(apiResp.Data) != 1 {
		t.Errorf("Expected 1 device, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].MAC != testDevice.MAC {
		t.Errorf("Expected MAC %s, got %s", testDevice.MAC, apiResp.Data[0].MAC)
	}
}

func TestHandleDeviceBasicStat(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add a test device
	testDevice := &types.Device{
		ID:      "test-device-1",
		MAC:     "aa:bb:cc:dd:ee:ff",
		Model:   "UAP-AC-PRO",
		Type:    "uap",
		Name:    "Test AP",
		State:   types.DeviceStateConnected,
	}
	server.State().AddDevice(testDevice)

	// Make request
	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/basicstat/device", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.DeviceBasic]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Errorf("Expected 1 device, got %d", len(apiResp.Data))
	}

	basic := apiResp.Data[0]
	if basic.MAC != testDevice.MAC {
		t.Errorf("Expected MAC %s, got %s", testDevice.MAC, basic.MAC)
	}
	if basic.Type != testDevice.Type {
		t.Errorf("Expected Type %s, got %s", testDevice.Type, basic.Type)
	}
}

func TestHandleDeviceUpdate(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add a test device
	testDevice := &types.Device{
		ID:      "test-device-1",
		MAC:     "aa:bb:cc:dd:ee:ff",
		Model:   "UAP-AC-PRO",
		Type:    "uap",
		Name:    "Old Name",
	}
	server.State().AddDevice(testDevice)

	// Update device
	updateReq := types.Device{
		Name: "New Name",
	}
	body, _ := json.Marshal(updateReq)

	req, _ := http.NewRequest("PUT", server.URL()+"/proxy/network/api/s/default/rest/device/test-device-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.Device]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 device in response, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", apiResp.Data[0].Name)
	}

	// Verify it was saved
	saved, exists := server.State().GetDevice("test-device-1")
	if !exists {
		t.Fatal("Device not found after update")
	}
	if saved.Name != "New Name" {
		t.Errorf("Expected saved name 'New Name', got '%s'", saved.Name)
	}
}

func TestHandleDeviceCommand_Adopt(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add an unadopted device
	testDevice := &types.Device{
		ID:      "test-device-1",
		MAC:     "aa:bb:cc:dd:ee:ff",
		Model:   "UAP-AC-PRO",
		Type:    "uap",
		Adopted: false,
		State:   types.DeviceStatePending,
	}
	server.State().AddDevice(testDevice)

	// Send adopt command
	cmdReq := types.CommandRequest{
		Cmd: "adopt",
		MAC: "aa:bb:cc:dd:ee:ff",
	}
	body, _ := json.Marshal(cmdReq)

	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/api/s/default/cmd/devmgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify device was adopted
	device, exists := server.State().GetDevice("test-device-1")
	if !exists {
		t.Fatal("Device not found")
	}
	if !device.Adopted {
		t.Error("Device should be adopted")
	}
	if device.State != types.DeviceStateConnected {
		t.Errorf("Expected state Connected, got %v", device.State)
	}
}

func TestHandleDeviceCommand_SetLocate(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add a device
	testDevice := &types.Device{
		ID:          "test-device-1",
		MAC:         "aa:bb:cc:dd:ee:ff",
		Model:       "UAP-AC-PRO",
		Type:        "uap",
		LEDOverride: "default",
	}
	server.State().AddDevice(testDevice)

	// Send set-locate command
	cmdReq := types.CommandRequest{
		Cmd: "set-locate",
		MAC: "aa:bb:cc:dd:ee:ff",
	}
	body, _ := json.Marshal(cmdReq)

	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/api/s/default/cmd/devmgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify LED override was set
	device, _ := server.State().GetDevice("test-device-1")
	if device.LEDOverride != "on" {
		t.Errorf("Expected LEDOverride 'on', got '%s'", device.LEDOverride)
	}
}

func TestHandleDeviceCommand_UnsetLocate(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add a device with locate enabled
	testDevice := &types.Device{
		ID:          "test-device-1",
		MAC:         "aa:bb:cc:dd:ee:ff",
		Model:       "UAP-AC-PRO",
		Type:        "uap",
		LEDOverride: "on",
	}
	server.State().AddDevice(testDevice)

	// Send unset-locate command
	cmdReq := types.CommandRequest{
		Cmd: "unset-locate",
		MAC: "aa:bb:cc:dd:ee:ff",
	}
	body, _ := json.Marshal(cmdReq)

	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/api/s/default/cmd/devmgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify LED override was cleared
	device, _ := server.State().GetDevice("test-device-1")
	if device.LEDOverride != "default" {
		t.Errorf("Expected LEDOverride 'default', got '%s'", device.LEDOverride)
	}
}

func TestHandleDeviceCommand_PowerCycle(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add a switch device
	testDevice := &types.Device{
		ID:    "test-device-1",
		MAC:   "aa:bb:cc:dd:ee:ff",
		Model: "USW-24-POE",
		Type:  "usw",
	}
	server.State().AddDevice(testDevice)

	// Send power-cycle command
	cmdReq := types.CommandRequest{
		Cmd:     "power-cycle",
		MAC:     "aa:bb:cc:dd:ee:ff",
		PortIdx: 5,
	}
	body, _ := json.Marshal(cmdReq)

	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/api/s/default/cmd/devmgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandleDeviceCommand_PowerCycle_MissingPortIdx(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add a switch device
	testDevice := &types.Device{
		ID:    "test-device-1",
		MAC:   "aa:bb:cc:dd:ee:ff",
		Model: "USW-24-POE",
		Type:  "usw",
	}
	server.State().AddDevice(testDevice)

	// Send power-cycle command without port_idx
	cmdReq := types.CommandRequest{
		Cmd: "power-cycle",
		MAC: "aa:bb:cc:dd:ee:ff",
	}
	body, _ := json.Marshal(cmdReq)

	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/api/s/default/cmd/devmgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleDeviceCommand_UnknownCommand(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add a device
	testDevice := &types.Device{
		ID:    "test-device-1",
		MAC:   "aa:bb:cc:dd:ee:ff",
		Model: "UAP-AC-PRO",
		Type:  "uap",
	}
	server.State().AddDevice(testDevice)

	// Send unknown command
	cmdReq := types.CommandRequest{
		Cmd: "unknown-command",
		MAC: "aa:bb:cc:dd:ee:ff",
	}
	body, _ := json.Marshal(cmdReq)

	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/api/s/default/cmd/devmgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}
