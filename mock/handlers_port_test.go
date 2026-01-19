package mock

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

var testPortHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func TestHandleListPortForwards(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test port forwards
	server.state.AddPortForward(&types.PortForward{
		ID:       "pf1",
		Name:     "Test Forward 1",
		Enabled:  true,
		Protocol: "tcp",
		DstPort:  "80",
		FwdIP:    "192.168.1.100",
		FwdPort:  "8080",
	})
	server.state.AddPortForward(&types.PortForward{
		ID:       "pf2",
		Name:     "Test Forward 2",
		Enabled:  false,
		Protocol: "udp",
		DstPort:  "53",
		FwdIP:    "192.168.1.101",
		FwdPort:  "5353",
	})

	// Test list port forwards
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/rest/portforward", nil)
	resp, err := testPortHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to list port forwards: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.PortForward `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 2 {
		t.Fatalf("Expected 2 port forwards, got %d", len(apiResp.Data))
	}
}

func TestHandleCreatePortForward(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Create port forward
	newForward := types.PortForward{
		Name:     "New Forward",
		Enabled:  true,
		Protocol: "tcp",
		DstPort:  "443",
		FwdIP:    "192.168.1.100",
		FwdPort:  "8443",
	}

	body, _ := json.Marshal(newForward)
	req, _ := http.NewRequest("POST", server.URL()+"/api/s/default/rest/portforward", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testPortHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to create port forward: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.PortForward `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 port forward, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].ID == "" {
		t.Error("Expected ID to be generated")
	}

	if apiResp.Data[0].Name != "New Forward" {
		t.Errorf("Expected name 'New Forward', got %s", apiResp.Data[0].Name)
	}
}

func TestHandleListPortProfiles(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test port profiles
	server.state.AddPortProfile(&types.PortProfile{
		ID:      "pp1",
		Name:    "Test Profile 1",
		Forward: "all",
		POEMode: "auto",
	})
	server.state.AddPortProfile(&types.PortProfile{
		ID:      "pp2",
		Name:    "Test Profile 2",
		Forward: "native",
		POEMode: "off",
	})

	// Test list port profiles
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/rest/portconf", nil)
	resp, err := testPortHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to list port profiles: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.PortProfile `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 2 {
		t.Fatalf("Expected 2 port profiles, got %d", len(apiResp.Data))
	}
}

func TestHandleGetPortProfile(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test port profile
	server.state.AddPortProfile(&types.PortProfile{
		ID:      "pp1",
		Name:    "Test Profile",
		Forward: "all",
		POEMode: "auto",
	})

	// Test get port profile
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/rest/portconf/pp1", nil)
	resp, err := testPortHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get port profile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.PortProfile `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 port profile, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Name != "Test Profile" {
		t.Errorf("Expected name 'Test Profile', got %s", apiResp.Data[0].Name)
	}
}

func TestHandleCreatePortProfile(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Create port profile
	newProfile := types.PortProfile{
		Name:    "New Profile",
		Forward: "native",
		POEMode: "auto",
	}

	body, _ := json.Marshal(newProfile)
	req, _ := http.NewRequest("POST", server.URL()+"/api/s/default/rest/portconf", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testPortHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to create port profile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.PortProfile `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 port profile, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].ID == "" {
		t.Error("Expected ID to be generated")
	}

	if apiResp.Data[0].Name != "New Profile" {
		t.Errorf("Expected name 'New Profile', got %s", apiResp.Data[0].Name)
	}
}

func TestHandleUpdatePortProfile(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test port profile
	server.state.AddPortProfile(&types.PortProfile{
		ID:      "pp1",
		Name:    "Test Profile",
		Forward: "all",
		POEMode: "auto",
	})

	// Update port profile
	update := types.PortProfile{
		Name:    "Updated Profile",
		Forward: "native",
		POEMode: "off",
	}

	body, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", server.URL()+"/api/s/default/rest/portconf/pp1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testPortHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to update port profile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify
	profile := server.state.GetPortProfile("pp1")
	if profile == nil {
		t.Fatal("Port profile not found")
	}

	if profile.Name != "Updated Profile" {
		t.Errorf("Expected name 'Updated Profile', got %s", profile.Name)
	}
}

func TestHandleDeletePortProfile(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test port profile
	server.state.AddPortProfile(&types.PortProfile{
		ID:      "pp1",
		Name:    "Test Profile",
		Forward: "all",
	})

	// Delete port profile
	req, _ := http.NewRequest("DELETE", server.URL()+"/api/s/default/rest/portconf/pp1", nil)
	resp, err := testPortHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to delete port profile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify
	profile := server.state.GetPortProfile("pp1")
	if profile != nil {
		t.Error("Expected port profile to be deleted")
	}
}
