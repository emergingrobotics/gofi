package mock

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

func TestHandleListSites(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL()+"/api/self/sites", nil)
	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var apiResp types.APIResponse[types.Site]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if apiResp.Meta.RC != "ok" {
		t.Errorf("RC = %s, want ok", apiResp.Meta.RC)
	}

	// Should have at least the default site
	if len(apiResp.Data) == 0 {
		t.Error("Expected at least one site")
	}
}

func TestHandleHealth(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/stat/health", nil)
	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var apiResp types.APIResponse[types.HealthData]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(apiResp.Data) == 0 {
		t.Error("Expected health data")
	}
}

func TestHandleSysInfo(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/stat/sysinfo", nil)
	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var apiResp types.APIResponse[types.SysInfo]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(apiResp.Data) == 0 {
		t.Fatal("Expected sysinfo data")
	}

	if apiResp.Data[0].Hostname != "UDM-Pro" {
		t.Errorf("Hostname = %s, want UDM-Pro", apiResp.Data[0].Hostname)
	}
}

func TestHandleCreateSite(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	createReq := types.CreateSiteRequest{
		Desc: "Test Site",
		Name: "test",
	}
	body, _ := json.Marshal(createReq)

	req, _ := http.NewRequest("POST", server.URL()+"/api/self/sites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := newTestClient().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var apiResp types.APIResponse[types.Site]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(apiResp.Data) == 0 {
		t.Fatal("Expected site in response")
	}

	if apiResp.Data[0].Desc != "Test Site" {
		t.Errorf("Desc = %s, want 'Test Site'", apiResp.Data[0].Desc)
	}

	// Verify site was added to state
	site, exists := server.State().GetSite("test")
	if !exists {
		t.Error("Site not added to state")
	}

	if site.Desc != "Test Site" {
		t.Errorf("State site Desc = %s, want 'Test Site'", site.Desc)
	}
}
