package types

import (
	"encoding/json"
	"testing"
)

func TestFirewallRule_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "rule123",
		"site_id": "default",
		"name": "Block IoT to LAN",
		"enabled": true,
		"ruleset": "LAN_IN",
		"rule_index": 2000,
		"action": "drop",
		"protocol": "all",
		"protocol_match_excepted": false,
		"logging": true,
		"state_new": false,
		"state_established": false,
		"state_invalid": false,
		"state_related": false,
		"src_firewallgroup_ids": ["group1"],
		"dst_firewallgroup_ids": ["group2"],
		"src_address": "10.0.20.0/24",
		"dst_port": "80,443"
	}`

	var rule FirewallRule
	if err := json.Unmarshal([]byte(jsonData), &rule); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if rule.ID != "rule123" {
		t.Errorf("ID = %v, want rule123", rule.ID)
	}
	if rule.Ruleset != RulesetLANIn {
		t.Errorf("Ruleset = %v, want LAN_IN", rule.Ruleset)
	}
	if rule.Action != FirewallActionDrop {
		t.Errorf("Action = %v, want drop", rule.Action)
	}
	if !rule.Logging {
		t.Error("Logging should be true")
	}
}

func TestFirewallGroup_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "group123",
		"site_id": "default",
		"name": "Web Servers",
		"group_type": "address-group",
		"group_members": ["192.168.1.10", "192.168.1.11"]
	}`

	var group FirewallGroup
	if err := json.Unmarshal([]byte(jsonData), &group); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if group.Name != "Web Servers" {
		t.Errorf("Name = %v, want Web Servers", group.Name)
	}
	if group.GroupType != GroupTypeAddress {
		t.Errorf("GroupType = %v, want address-group", group.GroupType)
	}
	if len(group.GroupMembers) != 2 {
		t.Errorf("GroupMembers length = %v, want 2", len(group.GroupMembers))
	}
}

func TestFirewallRuleIndexUpdate_MarshalJSON(t *testing.T) {
	update := FirewallRuleIndexUpdate{
		ID:        "rule123",
		RuleIndex: 3000,
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var update2 FirewallRuleIndexUpdate
	if err := json.Unmarshal(data, &update2); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if update2.ID != update.ID {
		t.Errorf("ID = %v, want %v", update2.ID, update.ID)
	}
	if update2.RuleIndex != update.RuleIndex {
		t.Errorf("RuleIndex = %v, want %v", update2.RuleIndex, update.RuleIndex)
	}
}

func TestFirewallConstants(t *testing.T) {
	rulesets := []string{RulesetWANIn, RulesetLANIn, RulesetGuestIn}
	actions := []string{FirewallActionAccept, FirewallActionDrop, FirewallActionReject}
	protocols := []string{ProtocolAll, ProtocolTCP, ProtocolUDP}

	for _, r := range rulesets {
		if r == "" {
			t.Error("Ruleset constant should not be empty")
		}
	}
	for _, a := range actions {
		if a == "" {
			t.Error("Action constant should not be empty")
		}
	}
	for _, p := range protocols {
		if p == "" {
			t.Error("Protocol constant should not be empty")
		}
	}
}
