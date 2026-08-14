package network

import (
	"context"
	"sort"

	"github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/src/types"
)

// NetworkEntry is the flattened view of a UniFi network that gofinet reports,
// focused on the DHCP dynamic pool and related addressing.
type NetworkEntry struct {
	Name        string   `json:"name"`
	Purpose     string   `json:"purpose,omitempty"`
	VLAN        int      `json:"vlan,omitempty"`
	Subnet      string   `json:"subnet,omitempty"`
	Enabled     bool     `json:"enabled"`
	DHCPEnabled bool     `json:"dhcp_enabled"`
	DHCPStart   string   `json:"dhcp_start,omitempty"`
	DHCPStop    string   `json:"dhcp_stop,omitempty"`
	DHCPLease   int      `json:"dhcp_lease,omitempty"`
	Gateway     string   `json:"gateway,omitempty"`
	DNS         []string `json:"dns,omitempty"`
}

// ListNetworks fetches all networks from the UDM and flattens them into entries
// sorted by name.
func ListNetworks(ctx context.Context, client gofi.Client, site string) ([]NetworkEntry, error) {
	networks, err := client.Networks().List(ctx, site)
	if err != nil {
		return nil, err
	}
	return buildNetworkEntries(networks), nil
}

// buildNetworkEntries maps and sorts networks. Kept separate from ListNetworks so
// it can be tested without a UDM.
func buildNetworkEntries(networks []types.Network) []NetworkEntry {
	entries := make([]NetworkEntry, 0, len(networks))
	for _, network := range networks {
		entries = append(entries, buildNetworkEntry(network))
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries
}

func buildNetworkEntry(network types.Network) NetworkEntry {
	entry := NetworkEntry{
		Name:        network.Name,
		Purpose:     network.Purpose,
		Subnet:      network.IPSubnet,
		Enabled:     network.Enabled,
		DHCPEnabled: network.DHCPDEnabled,
		DHCPStart:   network.DHCPDStart,
		DHCPStop:    network.DHCPDStop,
		DHCPLease:   network.DHCPDLeaseTime.Int(),
		DNS:         collectDNS(network),
	}
	if network.VLANEnabled {
		entry.VLAN = network.VLAN.Int()
	}
	if network.DHCPDGatewayEnabled {
		entry.Gateway = network.DHCPDGateway
	}
	return entry
}

// collectDNS returns the non-empty DHCP-advertised DNS servers in order.
func collectDNS(network types.Network) []string {
	candidates := []string{network.DHCPDDNS1, network.DHCPDDNS2, network.DHCPDDNS3, network.DHCPDDNS4}
	servers := []string{}
	for _, server := range candidates {
		if server != "" {
			servers = append(servers, server)
		}
	}
	if len(servers) == 0 {
		return nil
	}
	return servers
}
