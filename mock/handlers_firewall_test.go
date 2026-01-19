package mock

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

func TestHandleListFirewallRules(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	server.State().AddFirewallRule(&types.FirewallRule{
		ID:      "rule1",
		Name:    "Allow SSH",
		Enabled: true,
		Ruleset: types.RulesetWANIn,
	})

	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/rest/firewallrule", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp types.APIResponse[types.FirewallRule]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(apiResp.Data))
	}
}

func TestHandleCreateFirewallRule(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	newRule := &types.FirewallRule{
		Name:     "Block ICMP",
		Enabled:  true,
		Ruleset:  types.RulesetWANIn,
		Action:   types.FirewallActionDrop,
		Protocol: types.ProtocolICMP,
	}

	body, _ := json.Marshal(newRule)
	req, _ := http.NewRequest("POST", server.URL()+"/proxy/network/api/s/default/rest/firewallrule", bytes.NewReader(body))
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

func TestHandleListFirewallGroups(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	server.State().AddFirewallGroup(&types.FirewallGroup{
		ID:        "group1",
		Name:      "Internal IPs",
		GroupType: types.GroupTypeAddress,
	})

	req, _ := http.NewRequest("GET", server.URL()+"/proxy/network/api/s/default/rest/firewallgroup", nil)
	resp, err := testHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
