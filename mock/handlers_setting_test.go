package mock

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

var testSettingHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func TestHandleGetSetting(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test setting
	server.state.AddSetting(&types.Setting{
		Key:    types.SettingKeyMgmt,
		SiteID: "default",
	})

	// Test get setting
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/rest/setting/"+types.SettingKeyMgmt, nil)
	resp, err := testSettingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get setting: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.Setting `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 setting, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Key != types.SettingKeyMgmt {
		t.Errorf("Expected key '%s', got %s", types.SettingKeyMgmt, apiResp.Data[0].Key)
	}
}

func TestHandleUpdateSetting(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test setting
	server.state.AddSetting(&types.Setting{
		Key:    types.SettingKeyMgmt,
		SiteID: "default",
	})

	// Update setting
	update := types.Setting{
		Key:    types.SettingKeyMgmt,
		SiteID: "default",
	}

	body, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", server.URL()+"/api/s/default/rest/setting/"+types.SettingKeyMgmt, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testSettingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to update setting: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify
	setting := server.state.GetSetting(types.SettingKeyMgmt)
	if setting == nil {
		t.Fatal("Setting not found")
	}
}

func TestHandleListRADIUSProfiles(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test RADIUS profiles
	server.state.AddRADIUSProfile(&types.RADIUSProfile{
		ID:   "radius1",
		Name: "Test RADIUS 1",
	})
	server.state.AddRADIUSProfile(&types.RADIUSProfile{
		ID:   "radius2",
		Name: "Test RADIUS 2",
	})

	// Test list RADIUS profiles
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/rest/radiusprofile", nil)
	resp, err := testSettingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to list RADIUS profiles: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.RADIUSProfile `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 2 {
		t.Fatalf("Expected 2 RADIUS profiles, got %d", len(apiResp.Data))
	}
}

func TestHandleCreateRADIUSProfile(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Create RADIUS profile
	newProfile := types.RADIUSProfile{
		Name: "New RADIUS",
	}

	body, _ := json.Marshal(newProfile)
	req, _ := http.NewRequest("POST", server.URL()+"/api/s/default/rest/radiusprofile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testSettingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to create RADIUS profile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.RADIUSProfile `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 RADIUS profile, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].ID == "" {
		t.Error("Expected ID to be generated")
	}

	if apiResp.Data[0].Name != "New RADIUS" {
		t.Errorf("Expected name 'New RADIUS', got %s", apiResp.Data[0].Name)
	}
}

func TestHandleGetDynamicDNS(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test Dynamic DNS
	server.state.SetDynamicDNS(&types.DynamicDNS{
		ID:       "ddns1",
		Service:  "dyndns",
		Enabled:  true,
		Hostname: "test.dyndns.org",
	})

	// Test get Dynamic DNS
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/rest/dynamicdns", nil)
	resp, err := testSettingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get Dynamic DNS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.DynamicDNS `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 Dynamic DNS, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Hostname != "test.dyndns.org" {
		t.Errorf("Expected hostname 'test.dyndns.org', got %s", apiResp.Data[0].Hostname)
	}
}

func TestHandleUpdateDynamicDNS(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Update Dynamic DNS
	update := types.DynamicDNS{
		Service:  "dyndns",
		Enabled:  true,
		Hostname: "new.dyndns.org",
	}

	body, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", server.URL()+"/api/s/default/rest/dynamicdns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testSettingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to update Dynamic DNS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify
	ddns := server.state.GetDynamicDNS()
	if ddns == nil {
		t.Fatal("Dynamic DNS not found")
	}

	if ddns.Hostname != "new.dyndns.org" {
		t.Errorf("Expected hostname 'new.dyndns.org', got %s", ddns.Hostname)
	}
}
