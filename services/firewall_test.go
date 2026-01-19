package services

import (
	"context"
	"testing"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/types"
)

func TestFirewallService_ListRules(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	server.State().AddFirewallRule(&types.FirewallRule{
		ID:       "rule1",
		Name:     "Allow SSH",
		Enabled:  true,
		Ruleset:  types.RulesetWANIn,
		Action:   types.FirewallActionAccept,
		Protocol: types.ProtocolTCP,
		DstPort:  "22",
	})

	trans, _ := newTestTransport(server.URL())
	svc := NewFirewallService(trans)

	rules, err := svc.ListRules(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}

	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
}

func TestFirewallService_CreateRule(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	trans, _ := newTestTransport(server.URL())
	svc := NewFirewallService(trans)

	newRule := &types.FirewallRule{
		Name:     "Block ICMP",
		Enabled:  true,
		Ruleset:  types.RulesetWANIn,
		Action:   types.FirewallActionDrop,
		Protocol: types.ProtocolICMP,
	}

	created, err := svc.CreateRule(context.Background(), "default", newRule)
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}

	if created.Name != "Block ICMP" {
		t.Errorf("Expected name 'Block ICMP', got %s", created.Name)
	}
}

func TestFirewallService_EnableRule(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	server.State().AddFirewallRule(&types.FirewallRule{
		ID:      "rule1",
		Name:    "Test Rule",
		Enabled: false,
		Ruleset: types.RulesetWANIn,
		Action:  types.FirewallActionAccept,
	})

	trans, _ := newTestTransport(server.URL())
	svc := NewFirewallService(trans)

	err := svc.EnableRule(context.Background(), "default", "rule1")
	if err != nil {
		t.Fatalf("EnableRule failed: %v", err)
	}

	rule := server.State().GetFirewallRule("rule1")
	if !rule.Enabled {
		t.Error("Expected rule to be enabled")
	}
}

func TestFirewallService_ListGroups(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	server.State().AddFirewallGroup(&types.FirewallGroup{
		ID:           "group1",
		Name:         "Internal IPs",
		GroupType:    types.GroupTypeAddress,
		GroupMembers: []string{"192.168.1.0/24"},
	})

	trans, _ := newTestTransport(server.URL())
	svc := NewFirewallService(trans)

	groups, err := svc.ListGroups(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}
}

func TestFirewallService_ListTrafficRules(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	server.State().AddTrafficRule(&types.TrafficRule{
		ID:             "traffic1",
		Name:           "Limit Downloads",
		Enabled:        true,
		Action:         types.TrafficActionLimit,
		MatchingTarget: types.MatchingTargetClient,
	})

	trans, _ := newTestTransport(server.URL())
	svc := NewFirewallService(trans)

	rules, err := svc.ListTrafficRules(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListTrafficRules failed: %v", err)
	}

	if len(rules) != 1 {
		t.Errorf("Expected 1 traffic rule, got %d", len(rules))
	}
}

func TestFirewallService_CreateTrafficRule(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	trans, _ := newTestTransport(server.URL())
	svc := NewFirewallService(trans)

	newRule := &types.TrafficRule{
		Name:           "Block Social Media",
		Enabled:        true,
		Action:         types.TrafficActionDrop,
		MatchingTarget: types.MatchingTargetAll,
	}

	created, err := svc.CreateTrafficRule(context.Background(), "default", newRule)
	if err != nil {
		t.Fatalf("CreateTrafficRule failed: %v", err)
	}

	if created.Name != "Block Social Media" {
		t.Errorf("Expected name 'Block Social Media', got %s", created.Name)
	}
}
