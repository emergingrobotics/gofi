package mock

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/unifi-go/gofi/src/types"
)

var testClientHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func TestHandleClientStat(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test clients
	now := time.Now().Unix()
	activeClient := &types.Client{
		MAC:      "aa:bb:cc:dd:ee:f1",
		Hostname: "active-device",
		LastSeen: now - 60, // 1 minute ago (active)
	}
	inactiveClient := &types.Client{
		MAC:      "aa:bb:cc:dd:ee:f2",
		Hostname: "inactive-device",
		LastSeen: now - 600, // 10 minutes ago (inactive)
	}

	server.state.AddClient(activeClient)
	server.state.AddClient(inactiveClient)

	// Test active clients list
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/stat/sta", nil)
	resp, err := testClientHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get active clients: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.Client `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should only return active client
	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 active client, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].MAC != activeClient.MAC {
		t.Errorf("Expected MAC %s, got %s", activeClient.MAC, apiResp.Data[0].MAC)
	}
}

func TestHandleAllUserStat(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test clients with different last seen times
	now := time.Now().Unix()
	recentClient := &types.Client{
		MAC:      "aa:bb:cc:dd:ee:f1",
		Hostname: "recent-device",
		LastSeen: now - 3600, // 1 hour ago
	}
	oldClient := &types.Client{
		MAC:      "aa:bb:cc:dd:ee:f2",
		Hostname: "old-device",
		LastSeen: now - 86400*30, // 30 days ago
	}

	server.state.AddClient(recentClient)
	server.state.AddClient(oldClient)

	// Test with 24 hour window
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/stat/alluser?within=24", nil)
	resp, err := testClientHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.Client `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should only return recent client
	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 client within 24 hours, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].MAC != recentClient.MAC {
		t.Errorf("Expected MAC %s, got %s", recentClient.MAC, apiResp.Data[0].MAC)
	}
}

func TestHandleClientCommand(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test client
	testClient := &types.Client{
		MAC:      "aa:bb:cc:dd:ee:ff",
		Hostname: "test-device",
		LastSeen: time.Now().Unix(),
	}
	server.state.AddClient(testClient)

	tests := []struct {
		name        string
		cmd         string
		batchMACs   bool
		expectField func(*types.Client) bool
		shouldExist bool
	}{
		{
			name: "block client",
			cmd:  "block-sta",
			expectField: func(c *types.Client) bool {
				return c.Blocked
			},
			shouldExist: true,
		},
		{
			name: "unblock client",
			cmd:  "unblock-sta",
			expectField: func(c *types.Client) bool {
				return !c.Blocked
			},
			shouldExist: true,
		},
		{
			name: "kick client",
			cmd:  "kick-sta",
			expectField: func(c *types.Client) bool {
				return c.GuestKicked
			},
			shouldExist: true,
		},
		{
			name:        "forget client",
			cmd:         "forget-sta",
			batchMACs:   true,
			expectField: nil,
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset client state
			testClient.Blocked = false
			testClient.GuestKicked = false
			server.state.AddClient(testClient)

			// Send command
			cmdBody := map[string]interface{}{
				"cmd": tt.cmd,
			}
			if tt.batchMACs {
				cmdBody["macs"] = []string{testClient.MAC}
			} else {
				cmdBody["mac"] = testClient.MAC
			}
			body, _ := json.Marshal(cmdBody)

			req, _ := http.NewRequest("POST", server.URL()+"/api/s/default/cmd/stamgr", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := testClientHTTPClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to send command: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Expected status 200, got %d", resp.StatusCode)
			}

			// Verify state
			client := server.state.GetClient(testClient.MAC)
			if tt.shouldExist {
				if client == nil {
					t.Fatal("Expected client to exist")
				}
				if tt.expectField != nil && !tt.expectField(client) {
					t.Error("Client state not updated correctly")
				}
			} else {
				if client != nil {
					t.Error("Expected client to be deleted")
				}
			}
		})
	}
}

// Real controllers reject forget-sta carrying the singular "mac" field with
// HTTP 400. Guard the batch-only contract so the singular form cannot creep
// back in and pass here while failing against hardware.
func TestHandleClientCommandForgetRejectsSingularMAC(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	testClient := &types.Client{
		MAC:      "aa:bb:cc:dd:ee:fe",
		Hostname: "forget-singular",
		LastSeen: time.Now().Unix(),
	}
	server.state.AddClient(testClient)

	body, _ := json.Marshal(map[string]interface{}{
		"cmd": "forget-sta",
		"mac": testClient.MAC,
	})

	req, _ := http.NewRequest("POST", server.URL()+"/api/s/default/cmd/stamgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testClientHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for singular mac, got %d", resp.StatusCode)
	}

	if server.state.GetClient(testClient.MAC) == nil {
		t.Error("Client was deleted despite the request being rejected")
	}
}

func TestHandleClientCommandForgetMultipleMACs(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	macs := []string{"aa:bb:cc:dd:ee:11", "aa:bb:cc:dd:ee:22"}
	for _, mac := range macs {
		server.state.AddClient(&types.Client{MAC: mac, LastSeen: time.Now().Unix()})
	}

	body, _ := json.Marshal(map[string]interface{}{
		"cmd":  "forget-sta",
		"macs": macs,
	})

	req, _ := http.NewRequest("POST", server.URL()+"/api/s/default/cmd/stamgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testClientHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	for _, mac := range macs {
		if server.state.GetClient(mac) != nil {
			t.Errorf("Client %s was not deleted", mac)
		}
	}
}

func TestHandleClientGuestAuthorization(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	guestMAC := "aa:bb:cc:dd:ee:f1"

	// Test authorize guest (creates new client if needed)
	cmdBody := map[string]interface{}{
		"cmd":     "authorize-guest",
		"mac":     guestMAC,
		"minutes": 60,
		"up":      1024,
		"down":    2048,
	}
	body, _ := json.Marshal(cmdBody)

	req, _ := http.NewRequest("POST", server.URL()+"/api/s/default/cmd/stamgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testClientHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to authorize guest: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify guest was created and authorized
	client := server.state.GetClient(guestMAC)
	if client == nil {
		t.Fatal("Expected guest client to be created")
	}

	if !client.GuestAuthorized {
		t.Error("Expected guest to be authorized")
	}

	if !client.Authorized {
		t.Error("Expected client to be authorized")
	}

	// Test unauthorize guest
	cmdBody = map[string]interface{}{
		"cmd": "unauthorize-guest",
		"mac": guestMAC,
	}
	body, _ = json.Marshal(cmdBody)

	req, _ = http.NewRequest("POST", server.URL()+"/api/s/default/cmd/stamgr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = testClientHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to unauthorize guest: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify guest was unauthorized
	client = server.state.GetClient(guestMAC)
	if client.GuestAuthorized {
		t.Error("Expected guest to be unauthorized")
	}
}
