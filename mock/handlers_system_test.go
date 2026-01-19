package mock

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

var testSystemHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func TestHandleReboot(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Create reboot request
	req := map[string]string{"cmd": "reboot"}
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", server.URL()+"/api/cmd/system", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := testSystemHTTPClient.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to reboot: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandleBackupList(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test backups
	server.state.AddBackup(&types.Backup{
		Filename: "backup1.unf",
		Size:     1024,
		Time:     1234567890,
	})

	// Test list backups
	req, _ := http.NewRequest("GET", server.URL()+"/api/cmd/backup", nil)
	resp, err := testSystemHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.Backup `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Errorf("Expected 1 backup, got %d", len(apiResp.Data))
	}
}

func TestHandleBackupCreate(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Create backup
	req, _ := http.NewRequest("POST", server.URL()+"/api/cmd/backup", nil)
	resp, err := testSystemHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify
	backups := server.state.ListBackups()
	if len(backups) != 1 {
		t.Errorf("Expected 1 backup, got %d", len(backups))
	}
}

func TestHandleBackupDelete(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test backup
	server.state.AddBackup(&types.Backup{
		Filename: "backup1.unf",
		Size:     1024,
		Time:     1234567890,
	})

	// Delete backup
	req, _ := http.NewRequest("DELETE", server.URL()+"/api/cmd/backup/backup1.unf", nil)
	resp, err := testSystemHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to delete backup: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify
	backups := server.state.ListBackups()
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups, got %d", len(backups))
	}
}

func TestHandleAdminList(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test admin
	server.state.AddAdmin(&types.AdminUser{
		ID:       "admin1",
		Username: "admin",
		Name:     "Administrator",
	})

	// Test list admins
	req, _ := http.NewRequest("GET", server.URL()+"/api/stat/admin", nil)
	resp, err := testSystemHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to list admins: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.AdminUser `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Errorf("Expected 1 admin, got %d", len(apiResp.Data))
	}
}

func TestHandleSpeedTest(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Initiate speed test
	req, _ := http.NewRequest("POST", server.URL()+"/api/s/default/cmd/speedtest", nil)
	resp, err := testSystemHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to start speed test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check status
	req, _ = http.NewRequest("GET", server.URL()+"/api/s/default/stat/speedtest", nil)
	resp, err = testSystemHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get speed test status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.SpeedTestStatus `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Errorf("Expected 1 speed test status, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Running {
		t.Error("Expected speed test to be complete")
	}
}
