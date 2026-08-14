package profile

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/unifi-go/gofi/src/services"
	"github.com/unifi-go/gofi/src/types"
)

func TestCapture_omitsPassphrasesByDefault(t *testing.T) {
	client := newMockClientWithWLAN(t, "guest-wifi", "supersecret")

	p, err := Capture(context.Background(), client, "default", false)
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	for _, w := range p.WLANs {
		if strings.Contains(w.Passphrase, "supersecret") {
			t.Errorf("WLAN %q leaked its passphrase without --with-keys", w.Name)
		}
	}
}

func TestCapture_withKeysIncludesPassphrases(t *testing.T) {
	client := newMockClientWithWLAN(t, "guest-wifi", "supersecret")

	p, err := Capture(context.Background(), client, "default", true)
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	found := false
	for _, w := range p.WLANs {
		if w.Passphrase == "supersecret" {
			found = true
		}
	}
	if !found {
		t.Error("expected --with-keys to include the passphrase")
	}
}

func TestProfileWriteAndReadRoundTrip(t *testing.T) {
	p := &Profile{Site: "default", Captured: "2026-08-13T00:00:00Z"}

	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got, err := ReadProfile(&buf)
	if err != nil {
		t.Fatalf("ReadProfile() error = %v", err)
	}
	if got.Site != p.Site {
		t.Errorf("Site = %q, want %q", got.Site, p.Site)
	}
}

// newMockClientWithWLAN builds a test client whose WLAN list has a single
// entry with the given name and passphrase, on an otherwise empty site
// (no networks, users, or DNS records).
func newMockClientWithWLAN(t *testing.T, name, passphrase string) *mockClient {
	t.Helper()
	return &mockClient{
		wlans: mockWLANService{
			wlans: []types.WLAN{
				{Name: name, Enabled: true, Passphrase: passphrase},
			},
		},
	}
}

// mockClient implements gofi.Client for testing operations.
type mockClient struct {
	users    mockUserService
	dns      mockDNSService
	networks mockNetworkService
	wlans    mockWLANService
}

func (m *mockClient) Connect(ctx context.Context) error         { return nil }
func (m *mockClient) Disconnect(ctx context.Context) error      { return nil }
func (m *mockClient) IsConnected() bool                         { return true }
func (m *mockClient) Sites() services.SiteService               { return nil }
func (m *mockClient) Devices() services.DeviceService           { return nil }
func (m *mockClient) Networks() services.NetworkService         { return &m.networks }
func (m *mockClient) WLANs() services.WLANService               { return &m.wlans }
func (m *mockClient) Firewall() services.FirewallService        { return nil }
func (m *mockClient) Clients() services.ClientService           { return nil }
func (m *mockClient) Users() services.UserService               { return &m.users }
func (m *mockClient) Routing() services.RoutingService          { return nil }
func (m *mockClient) PortForwards() services.PortForwardService { return nil }
func (m *mockClient) PortProfiles() services.PortProfileService { return nil }
func (m *mockClient) Settings() services.SettingService         { return nil }
func (m *mockClient) System() services.SystemService            { return nil }
func (m *mockClient) Events() services.EventService             { return nil }
func (m *mockClient) DNS() services.DNSService                  { return &m.dns }

// mockUserService implements services.UserService.
type mockUserService struct {
	users []types.User
}

func (m *mockUserService) List(ctx context.Context, site string) ([]types.User, error) {
	return m.users, nil
}

func (m *mockUserService) Get(ctx context.Context, site, id string) (*types.User, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockUserService) GetByMAC(ctx context.Context, site, mac string) (*types.User, error) {
	for i := range m.users {
		if strings.EqualFold(m.users[i].MAC, mac) {
			return &m.users[i], nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockUserService) Create(ctx context.Context, site string, user *types.User) (*types.User, error) {
	return user, nil
}

func (m *mockUserService) Update(ctx context.Context, site string, user *types.User) (*types.User, error) {
	return user, nil
}

func (m *mockUserService) Delete(ctx context.Context, site, id string) error {
	return nil
}

func (m *mockUserService) DeleteByMAC(ctx context.Context, site, mac string) error {
	return nil
}

func (m *mockUserService) SetFixedIP(ctx context.Context, site, mac, ip, networkID string) error {
	return nil
}

func (m *mockUserService) ClearFixedIP(ctx context.Context, site, mac string) error {
	return nil
}

func (m *mockUserService) ListGroups(ctx context.Context, site string) ([]types.UserGroup, error) {
	return nil, nil
}

func (m *mockUserService) GetGroup(ctx context.Context, site, id string) (*types.UserGroup, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockUserService) CreateGroup(ctx context.Context, site string, group *types.UserGroup) (*types.UserGroup, error) {
	return group, nil
}

func (m *mockUserService) UpdateGroup(ctx context.Context, site string, group *types.UserGroup) (*types.UserGroup, error) {
	return group, nil
}

func (m *mockUserService) DeleteGroup(ctx context.Context, site, id string) error {
	return nil
}

// mockDNSService implements services.DNSService.
type mockDNSService struct {
	records []types.DNSRecord
}

func (m *mockDNSService) List(ctx context.Context, site string) ([]types.DNSRecord, error) {
	return m.records, nil
}

func (m *mockDNSService) Get(ctx context.Context, site, id string) (*types.DNSRecord, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockDNSService) GetByName(ctx context.Context, site, name string) (*types.DNSRecord, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockDNSService) GetByIP(ctx context.Context, site, ip string) ([]types.DNSRecord, error) {
	return nil, nil
}

func (m *mockDNSService) Create(ctx context.Context, site string, record *types.DNSRecord) (*types.DNSRecord, error) {
	return record, nil
}

func (m *mockDNSService) Update(ctx context.Context, site string, record *types.DNSRecord) (*types.DNSRecord, error) {
	return record, nil
}

func (m *mockDNSService) Delete(ctx context.Context, site, id string) error {
	return nil
}

func (m *mockDNSService) DeleteByName(ctx context.Context, site, name string) error {
	return nil
}

// mockNetworkService implements services.NetworkService.
type mockNetworkService struct {
	networks []types.Network
}

func (m *mockNetworkService) List(ctx context.Context, site string) ([]types.Network, error) {
	return m.networks, nil
}

func (m *mockNetworkService) Get(ctx context.Context, site, id string) (*types.Network, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockNetworkService) Create(ctx context.Context, site string, network *types.Network) (*types.Network, error) {
	return network, nil
}

func (m *mockNetworkService) Update(ctx context.Context, site string, network *types.Network) (*types.Network, error) {
	return network, nil
}

func (m *mockNetworkService) Delete(ctx context.Context, site, id string) error {
	return nil
}

// mockWLANService implements services.WLANService.
type mockWLANService struct {
	wlans []types.WLAN
}

func (m *mockWLANService) List(ctx context.Context, site string) ([]types.WLAN, error) {
	return m.wlans, nil
}

func (m *mockWLANService) Get(ctx context.Context, site, id string) (*types.WLAN, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockWLANService) Create(ctx context.Context, site string, wlan *types.WLAN) (*types.WLAN, error) {
	return wlan, nil
}

func (m *mockWLANService) Update(ctx context.Context, site string, wlan *types.WLAN) (*types.WLAN, error) {
	return wlan, nil
}

func (m *mockWLANService) Delete(ctx context.Context, site, id string) error {
	return nil
}

func (m *mockWLANService) Enable(ctx context.Context, site, id string) error {
	return nil
}

func (m *mockWLANService) Disable(ctx context.Context, site, id string) error {
	return nil
}

func (m *mockWLANService) SetMACFilter(ctx context.Context, site, id, policy string, macs []string) error {
	return nil
}

func (m *mockWLANService) ListGroups(ctx context.Context, site string) ([]types.WLANGroup, error) {
	return nil, nil
}

func (m *mockWLANService) GetGroup(ctx context.Context, site, id string) (*types.WLANGroup, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockWLANService) CreateGroup(ctx context.Context, site string, group *types.WLANGroup) (*types.WLANGroup, error) {
	return group, nil
}

func (m *mockWLANService) UpdateGroup(ctx context.Context, site string, group *types.WLANGroup) (*types.WLANGroup, error) {
	return group, nil
}

func (m *mockWLANService) DeleteGroup(ctx context.Context, site, id string) error {
	return nil
}
