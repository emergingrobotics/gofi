package profile

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/unifi-go/gofi/src/services"
	"github.com/unifi-go/gofi/src/types"
	"github.com/unifi-go/gofi/utilities/internal/ips"
	"github.com/unifi-go/gofi/utilities/internal/network"
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

func TestApply_skipsNetworksAbsentFromTarget(t *testing.T) {
	client := newMockClientWithNetworks(t, nil) // target site has no networks
	p := &Profile{
		Networks: []network.NetworkEntry{{Name: "Guest"}},
	}
	var progress bytes.Buffer
	if err := Apply(context.Background(), client, p, false, "", &progress); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !strings.Contains(progress.String(), "Guest") {
		t.Error("expected Apply to report the skipped network by name")
	}
}

func TestApply_dryRunMakesNoWrites(t *testing.T) {
	client := newMockClientRecordingWrites(t)
	// DoSet's dry-run path still resolves the DNS domain from the target's
	// networks before classifying entries, so the mock needs one network
	// with a domain name even though it makes no writes.
	client.networks.networks = []types.Network{{Name: "Default", IPSubnet: "192.168.1.0/24", DomainName: "lan.example.com"}}
	p := &Profile{FixedIPs: []ips.HostEntry{{Hostname: "nas", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.13"}}}

	if err := Apply(context.Background(), client, p, true, "", io.Discard); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if client.WriteCount() != 0 {
		t.Errorf("WriteCount() = %d after dry-run, want 0", client.WriteCount())
	}
}

func TestApply_dryRunReportsRealDoSetCounts(t *testing.T) {
	client := newMockClientWithNetworks(t, []types.Network{{Name: "Default", IPSubnet: "192.168.1.0/24", DomainName: "lan.example.com"}})
	client.users = mockUserService{users: []types.User{
		{MAC: "aa:bb:cc:dd:ee:01", Name: "nas", UseFixedIP: true, FixedIP: "192.168.1.13"},
	}}
	p := &Profile{
		Site:     "default",
		FixedIPs: []ips.HostEntry{{Hostname: "nas", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.13"}},
	}
	var progress bytes.Buffer
	if err := Apply(context.Background(), client, p, true, "", &progress); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	// The entry is already unchanged on the target, so a real DoSet dry run
	// classifies it as skipped, not "would import."
	if !strings.Contains(progress.String(), "1 skipped") {
		t.Errorf("expected dry-run preview to reflect DoSet's real skip classification, got: %s", progress.String())
	}
}

func TestApply_skipsWLANsAbsentFromTarget(t *testing.T) {
	client := newMockClientWithWLAN(t, "office-wifi", "")
	p := &Profile{
		WLANs: []WLANEntry{{Name: "guest-wifi", Enabled: true}},
	}
	var progress bytes.Buffer
	if err := Apply(context.Background(), client, p, false, "", &progress); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !strings.Contains(progress.String(), "skipping WLAN \"guest-wifi\"") {
		t.Errorf("expected Apply to report the skipped WLAN by name, got: %s", progress.String())
	}
}

func TestApply_appliesNetworksBeforeFixedIPsBeforeWLANs(t *testing.T) {
	client := newMockClientWithNetworks(t, []types.Network{{Name: "Default", IPSubnet: "192.168.1.0/24", DomainName: "lan.example.com"}})
	client.wlans = mockWLANService{wlans: []types.WLAN{{Name: "guest-wifi", Enabled: true}}}
	p := &Profile{
		Networks: []network.NetworkEntry{{Name: "Default"}},
		FixedIPs: []ips.HostEntry{{Hostname: "nas", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.13"}},
		WLANs:    []WLANEntry{{Name: "guest-wifi", Enabled: true}},
	}
	var progress bytes.Buffer
	if err := Apply(context.Background(), client, p, false, "", &progress); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	out := progress.String()
	networkIdx := strings.Index(out, "Default")
	fixedIPIdx := strings.Index(out, "fixed IPs")
	wlanIdx := strings.Index(out, "guest-wifi")
	if networkIdx == -1 || fixedIPIdx == -1 || wlanIdx == -1 {
		t.Fatalf("expected progress to report all three sections, got: %s", out)
	}
	if !(networkIdx < fixedIPIdx && fixedIPIdx < wlanIdx) {
		t.Errorf("expected order networks < fixed IPs < WLANs in progress output, got: %s", out)
	}
}

func TestApply_dnsDomainOverrideReachesDoSet(t *testing.T) {
	client := newMockClientWithNetworks(t, []types.Network{{Name: "Default", IPSubnet: "192.168.1.0/24"}}) // target network has NO domain configured
	p := &Profile{
		Site:     "default",
		FixedIPs: []ips.HostEntry{{Hostname: "nas", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.13"}},
	}
	var progress bytes.Buffer

	// Without an override, this should fail (no network domain, no override given).
	err := Apply(context.Background(), client, p, false, "", &progress)
	if err == nil {
		t.Fatal("Apply() with no dnsDomainOverride and no network domain: error = nil, want a domain-resolution error")
	}

	// With an override, it should succeed.
	progress.Reset()
	err = Apply(context.Background(), client, p, false, "bench.test", &progress)
	if err != nil {
		t.Fatalf("Apply() with dnsDomainOverride = %v, want nil", err)
	}
}

// newMockClientWithNetworks builds a test client whose target-site network
// list is exactly the given networks (nil means the site has none).
func newMockClientWithNetworks(t *testing.T, networks []types.Network) *mockClient {
	t.Helper()
	return &mockClient{
		networks: mockNetworkService{networks: networks},
	}
}

// recordingMockClient wraps mockClient and counts every write (Create,
// Update, Delete) issued through the user, network, or WLAN services, so
// tests can assert a dry run makes zero real writes.
type recordingMockClient struct {
	*mockClient
	writes *int
}

func (r *recordingMockClient) WriteCount() int { return *r.writes }

// newMockClientRecordingWrites builds a test client identical to an empty
// mockClient, but with write-tracking wired into the user, network, and
// WLAN services.
func newMockClientRecordingWrites(t *testing.T) *recordingMockClient {
	t.Helper()
	writes := new(int)
	client := &mockClient{}
	client.users.writes = writes
	client.networks.writes = writes
	client.wlans.writes = writes
	return &recordingMockClient{mockClient: client, writes: writes}
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
	users  []types.User
	writes *int // optional write counter, set by newMockClientRecordingWrites
}

func (m *mockUserService) countWrite() {
	if m.writes != nil {
		*m.writes++
	}
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
	m.countWrite()
	return user, nil
}

func (m *mockUserService) Update(ctx context.Context, site string, user *types.User) (*types.User, error) {
	m.countWrite()
	return user, nil
}

func (m *mockUserService) Delete(ctx context.Context, site, id string) error {
	m.countWrite()
	return nil
}

func (m *mockUserService) DeleteByMAC(ctx context.Context, site, mac string) error {
	m.countWrite()
	return nil
}

func (m *mockUserService) SetFixedIP(ctx context.Context, site, mac, ip, networkID string) error {
	m.countWrite()
	return nil
}

func (m *mockUserService) ClearFixedIP(ctx context.Context, site, mac string) error {
	m.countWrite()
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
	writes   *int // optional write counter, set by newMockClientRecordingWrites
}

func (m *mockNetworkService) countWrite() {
	if m.writes != nil {
		*m.writes++
	}
}

func (m *mockNetworkService) List(ctx context.Context, site string) ([]types.Network, error) {
	return m.networks, nil
}

func (m *mockNetworkService) Get(ctx context.Context, site, id string) (*types.Network, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockNetworkService) Create(ctx context.Context, site string, network *types.Network) (*types.Network, error) {
	m.countWrite()
	return network, nil
}

func (m *mockNetworkService) Update(ctx context.Context, site string, network *types.Network) (*types.Network, error) {
	m.countWrite()
	return network, nil
}

func (m *mockNetworkService) Delete(ctx context.Context, site, id string) error {
	m.countWrite()
	return nil
}

// mockWLANService implements services.WLANService.
type mockWLANService struct {
	wlans  []types.WLAN
	writes *int // optional write counter, set by newMockClientRecordingWrites
}

func (m *mockWLANService) countWrite() {
	if m.writes != nil {
		*m.writes++
	}
}

func (m *mockWLANService) List(ctx context.Context, site string) ([]types.WLAN, error) {
	return m.wlans, nil
}

func (m *mockWLANService) Get(ctx context.Context, site, id string) (*types.WLAN, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockWLANService) Create(ctx context.Context, site string, wlan *types.WLAN) (*types.WLAN, error) {
	m.countWrite()
	return wlan, nil
}

func (m *mockWLANService) Update(ctx context.Context, site string, wlan *types.WLAN) (*types.WLAN, error) {
	m.countWrite()
	return wlan, nil
}

func (m *mockWLANService) Delete(ctx context.Context, site, id string) error {
	m.countWrite()
	return nil
}

func (m *mockWLANService) Enable(ctx context.Context, site, id string) error {
	m.countWrite()
	return nil
}

func (m *mockWLANService) Disable(ctx context.Context, site, id string) error {
	m.countWrite()
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
