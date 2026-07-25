package main

import (
	"testing"

	"github.com/unifi-go/gofi/types"
)

func TestBuildNetworkEntries_SortsByName(t *testing.T) {
	networks := []types.Network{
		{Name: "guest", IPSubnet: "192.168.20.0/24"},
		{Name: "IoT", IPSubnet: "192.168.30.0/24"},
		{Name: "LAN", IPSubnet: "192.168.4.0/24"},
	}
	entries := buildNetworkEntries(networks)

	want := []string{"IoT", "LAN", "guest"}
	for index, name := range want {
		if entries[index].Name != name {
			t.Errorf("position %d: got %s, want %s", index, entries[index].Name, name)
		}
	}
}

func TestBuildNetworkEntry_DHCPFields(t *testing.T) {
	network := types.Network{
		Name:           "LAN",
		Purpose:        "corporate",
		IPSubnet:       "192.168.4.0/24",
		Enabled:        true,
		DHCPDEnabled:   true,
		DHCPDStart:     "192.168.4.6",
		DHCPDStop:      "192.168.4.99",
		DHCPDLeaseTime: types.FlexInt{Val: 86400},
	}
	entry := buildNetworkEntry(network)

	if !entry.DHCPEnabled {
		t.Error("expected DHCPEnabled true")
	}
	if entry.DHCPStart != "192.168.4.6" || entry.DHCPStop != "192.168.4.99" {
		t.Errorf("unexpected pool: %s - %s", entry.DHCPStart, entry.DHCPStop)
	}
	if entry.DHCPLease != 86400 {
		t.Errorf("expected lease 86400, got %d", entry.DHCPLease)
	}
}

func TestBuildNetworkEntry_VLANOnlyWhenEnabled(t *testing.T) {
	tagged := buildNetworkEntry(types.Network{Name: "guest", VLANEnabled: true, VLAN: types.FlexInt{Val: 20}})
	if tagged.VLAN != 20 {
		t.Errorf("expected VLAN 20, got %d", tagged.VLAN)
	}

	// VLAN value present but not enabled -> treated as untagged (0).
	untagged := buildNetworkEntry(types.Network{Name: "LAN", VLANEnabled: false, VLAN: types.FlexInt{Val: 1}})
	if untagged.VLAN != 0 {
		t.Errorf("expected VLAN 0 when VLAN disabled, got %d", untagged.VLAN)
	}
}

func TestBuildNetworkEntry_GatewayOnlyWhenEnabled(t *testing.T) {
	withGW := buildNetworkEntry(types.Network{Name: "LAN", DHCPDGatewayEnabled: true, DHCPDGateway: "192.168.4.1"})
	if withGW.Gateway != "192.168.4.1" {
		t.Errorf("expected gateway, got %q", withGW.Gateway)
	}

	withoutGW := buildNetworkEntry(types.Network{Name: "LAN", DHCPDGatewayEnabled: false, DHCPDGateway: "192.168.4.1"})
	if withoutGW.Gateway != "" {
		t.Errorf("expected empty gateway when disabled, got %q", withoutGW.Gateway)
	}
}

func TestCollectDNS(t *testing.T) {
	network := types.Network{DHCPDDNS1: "1.1.1.1", DHCPDDNS2: "", DHCPDDNS3: "8.8.8.8"}
	servers := collectDNS(network)
	if len(servers) != 2 || servers[0] != "1.1.1.1" || servers[1] != "8.8.8.8" {
		t.Errorf("expected [1.1.1.1 8.8.8.8], got %v", servers)
	}

	if collectDNS(types.Network{}) != nil {
		t.Error("expected nil for no DNS servers")
	}
}
