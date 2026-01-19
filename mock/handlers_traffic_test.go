package mock

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

func TestHandleListTrafficRules(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	server.State().AddTrafficRule(&types.TrafficRule{
		ID:      "traffic1",
		Name:    "Limit Downloads",
		Enabled: true,
		Action:  types.TrafficActionLimit,
	})

	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/v2/api/site/default/trafficrule", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.TrafficRule]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Errorf("Expected 1 traffic rule, got %d", len(apiResp.Data))
	}
}

func TestHandleCreateTrafficRule(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	newRule := &types.TrafficRule{
		Name:           "Block Social Media",
		Enabled:        true,
		Action:         types.TrafficActionDrop,
		MatchingTarget: types.MatchingTargetAll,
	}

	body, _ := json.Marshal(newRule)
	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/v2/api/site/default/trafficrule", bytes.NewReader(body))
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

func TestHandleUpdateTrafficRule(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	server.State().AddTrafficRule(&types.TrafficRule{
		ID:      "traffic1",
		Name:    "Old Name",
		Enabled: false,
		Action:  types.TrafficActionLimit,
	})

	updatedRule := &types.TrafficRule{
		Name:    "New Name",
		Enabled: true,
		Action:  types.TrafficActionLimit,
	}

	body, _ := json.Marshal(updatedRule)
	req, _ := http.NewRequest("PUT", server.URL()+"/proxy/network/v2/api/site/default/trafficrule/traffic1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Note: v2 API returns 201 for PUT operations
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}
