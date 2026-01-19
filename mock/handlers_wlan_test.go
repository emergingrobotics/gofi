package mock

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

func TestHandleListWLANs(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test WLANs
	testWLAN1 := &types.WLAN{
		ID:       "wlan-1",
		SiteID:   "default",
		Name:     "Test WLAN 1",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
	}
	testWLAN2 := &types.WLAN{
		ID:       "wlan-2",
		SiteID:   "default",
		Name:     "Test WLAN 2",
		Enabled:  false,
		Security: types.SecurityTypeOpen,
	}
	server.State().AddWLAN(testWLAN1)
	server.State().AddWLAN(testWLAN2)

	// Make request
	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/rest/wlanconf", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.WLAN]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if apiResp.Meta.RC != "ok" {
		t.Errorf("Expected RC=ok, got %s", apiResp.Meta.RC)
	}

	if len(apiResp.Data) != 2 {
		t.Errorf("Expected 2 WLANs, got %d", len(apiResp.Data))
	}
}

func TestHandleGetWLAN(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test WLAN
	testWLAN := &types.WLAN{
		ID:       "wlan-1",
		SiteID:   "default",
		Name:     "Test WLAN",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
	}
	server.State().AddWLAN(testWLAN)

	// Make request
	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/rest/wlanconf/wlan-1", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.WLAN]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 WLAN, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Name != "Test WLAN" {
		t.Errorf("Expected name 'Test WLAN', got %s", apiResp.Data[0].Name)
	}
}

func TestHandleGetWLANNotFound(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Make request for non-existent WLAN
	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/rest/wlanconf/nonexistent", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleCreateWLAN(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	newWLAN := &types.WLAN{
		Name:     "New WLAN",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
	}

	body, _ := json.Marshal(newWLAN)
	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/api/s/default/rest/wlanconf", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.WLAN]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 WLAN, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Name != "New WLAN" {
		t.Errorf("Expected name 'New WLAN', got %s", apiResp.Data[0].Name)
	}

	if apiResp.Data[0].ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestHandleUpdateWLAN(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test WLAN
	testWLAN := &types.WLAN{
		ID:       "wlan-1",
		SiteID:   "default",
		Name:     "Old Name",
		Enabled:  false,
		Security: types.SecurityTypeWPAPSK,
	}
	server.State().AddWLAN(testWLAN)

	// Update WLAN
	updatedWLAN := &types.WLAN{
		Name:     "New Name",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
	}

	body, _ := json.Marshal(updatedWLAN)
	req, _ := http.NewRequest("PUT", server.URL()+"/proxy/network/api/s/default/rest/wlanconf/wlan-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.WLAN]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 WLAN, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Name != "New Name" {
		t.Errorf("Expected name 'New Name', got %s", apiResp.Data[0].Name)
	}

	if !apiResp.Data[0].Enabled {
		t.Error("Expected WLAN to be enabled")
	}
}

func TestHandleDeleteWLAN(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test WLAN
	testWLAN := &types.WLAN{
		ID:       "wlan-1",
		SiteID:   "default",
		Name:     "Test WLAN",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
	}
	server.State().AddWLAN(testWLAN)

	// Delete WLAN
	req, _ := http.NewRequest("DELETE", server.URL()+"/proxy/network/api/s/default/rest/wlanconf/wlan-1", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify WLAN is deleted
	if wlan := server.State().GetWLAN("wlan-1"); wlan != nil {
		t.Error("Expected WLAN to be deleted")
	}
}

func TestHandleListWLANGroups(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test WLAN groups
	testGroup1 := &types.WLANGroup{
		ID:      "group-1",
		SiteID:  "default",
		Name:    "Test Group 1",
		Members: []string{"aa:bb:cc:dd:ee:ff"},
	}
	testGroup2 := &types.WLANGroup{
		ID:      "group-2",
		SiteID:  "default",
		Name:    "Test Group 2",
		Members: []string{},
	}
	server.State().AddWLANGroup(testGroup1)
	server.State().AddWLANGroup(testGroup2)

	// Make request
	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/rest/wlangroup", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.WLANGroup]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 2 {
		t.Errorf("Expected 2 WLAN groups, got %d", len(apiResp.Data))
	}
}

func TestHandleCreateWLANGroup(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	newGroup := &types.WLANGroup{
		Name:    "New Group",
		Members: []string{"aa:bb:cc:dd:ee:ff"},
	}

	body, _ := json.Marshal(newGroup)
	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/api/s/default/rest/wlangroup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.WLANGroup]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 WLAN group, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Name != "New Group" {
		t.Errorf("Expected name 'New Group', got %s", apiResp.Data[0].Name)
	}

	if apiResp.Data[0].ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestHandleUpdateWLANGroup(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test WLAN group
	testGroup := &types.WLANGroup{
		ID:      "group-1",
		SiteID:  "default",
		Name:    "Old Name",
		Members: []string{},
	}
	server.State().AddWLANGroup(testGroup)

	// Update group
	updatedGroup := &types.WLANGroup{
		Name:    "New Name",
		Members: []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"},
	}

	body, _ := json.Marshal(updatedGroup)
	req, _ := http.NewRequest("PUT", server.URL()+"/proxy/network/api/s/default/rest/wlangroup/group-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.WLANGroup]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if apiResp.Data[0].Name != "New Name" {
		t.Errorf("Expected name 'New Name', got %s", apiResp.Data[0].Name)
	}

	if len(apiResp.Data[0].Members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(apiResp.Data[0].Members))
	}
}

func TestHandleDeleteWLANGroup(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test WLAN group
	testGroup := &types.WLANGroup{
		ID:      "group-1",
		SiteID:  "default",
		Name:    "Test Group",
		Members: []string{},
	}
	server.State().AddWLANGroup(testGroup)

	// Delete group
	req, _ := http.NewRequest("DELETE", server.URL()+"/proxy/network/api/s/default/rest/wlangroup/group-1", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify group is deleted
	if group := server.State().GetWLANGroup("group-1"); group != nil {
		t.Error("Expected WLAN group to be deleted")
	}
}
