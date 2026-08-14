package network

import (
	"context"
	"errors"
	"testing"

	"github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/src/services"
	"github.com/unifi-go/gofi/src/types"
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

func TestFindByName_returnsErrNotFoundForUnknownNetwork(t *testing.T) {
	client := newMockClientWithNetworks(t, []types.Network{{Name: "Default"}})
	_, err := FindByName(context.Background(), client, "default", "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByName() error = %v, want ErrNotFound", err)
	}
}

func TestFindByName_findsExactMatch(t *testing.T) {
	client := newMockClientWithNetworks(t, []types.Network{{Name: "Guest"}, {Name: "Default"}})
	entry, err := FindByName(context.Background(), client, "default", "Guest")
	if err != nil {
		t.Fatalf("FindByName() error = %v", err)
	}
	if entry.Name != "Guest" {
		t.Errorf("Name = %q, want Guest", entry.Name)
	}
}

// newMockClientWithNetworks returns a mock client that returns the given networks list.
func newMockClientWithNetworks(t *testing.T, networks []types.Network) gofi.Client {
	return &mockClient{
		networkService: &mockNetworkService{
			networks: networks,
		},
	}
}

// mockClient implements a minimal gofi.Client for testing.
type mockClient struct {
	networkService *mockNetworkService
}

func (m *mockClient) Connect(ctx context.Context) error {
	return nil
}

func (m *mockClient) Disconnect(ctx context.Context) error {
	return nil
}

func (m *mockClient) IsConnected() bool {
	return true
}

func (m *mockClient) Sites() services.SiteService {
	return nil
}

func (m *mockClient) Devices() services.DeviceService {
	return nil
}

func (m *mockClient) Networks() services.NetworkService {
	return m.networkService
}

func (m *mockClient) WLANs() services.WLANService {
	return nil
}

func (m *mockClient) Firewall() services.FirewallService {
	return nil
}

func (m *mockClient) Clients() services.ClientService {
	return nil
}

func (m *mockClient) Users() services.UserService {
	return nil
}

func (m *mockClient) Routing() services.RoutingService {
	return nil
}

func (m *mockClient) PortForwards() services.PortForwardService {
	return nil
}

func (m *mockClient) PortProfiles() services.PortProfileService {
	return nil
}

func (m *mockClient) Settings() services.SettingService {
	return nil
}

func (m *mockClient) System() services.SystemService {
	return nil
}

func (m *mockClient) Events() services.EventService {
	return nil
}

func (m *mockClient) DNS() services.DNSService {
	return nil
}

// mockNetworkService implements a minimal services.NetworkService for testing.
type mockNetworkService struct {
	networks []types.Network
}

func (m *mockNetworkService) List(ctx context.Context, site string) ([]types.Network, error) {
	return m.networks, nil
}

func (m *mockNetworkService) Get(ctx context.Context, site, id string) (*types.Network, error) {
	return nil, nil
}

func (m *mockNetworkService) Create(ctx context.Context, site string, network *types.Network) (*types.Network, error) {
	return nil, nil
}

func (m *mockNetworkService) Update(ctx context.Context, site string, network *types.Network) (*types.Network, error) {
	return nil, nil
}

func (m *mockNetworkService) Delete(ctx context.Context, site, id string) error {
	return nil
}
