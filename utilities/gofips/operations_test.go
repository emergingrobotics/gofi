package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/unifi-go/gofi/src/services"
	"github.com/unifi-go/gofi/src/types"
)

// setResolver installs a stub UDM resolver for the duration of a test.
func setResolver(t *testing.T, fn dnsResolver) {
	prev := resolveFQDN
	resolveFQDN = fn
	t.Cleanup(func() { resolveFQDN = prev })
}

// unresolvedResolver simulates the UDM answering NXDOMAIN for every name.
func unresolvedResolver(ctx context.Context, fqdn string) ([]string, error) {
	return nil, nil
}

// mockClient implements gofi.Client for testing operations.
type mockClient struct {
	users    mockUserService
	dns      mockDNSService
	networks mockNetworkService
}

func (m *mockClient) Connect(ctx context.Context) error         { return nil }
func (m *mockClient) Disconnect(ctx context.Context) error      { return nil }
func (m *mockClient) IsConnected() bool                         { return true }
func (m *mockClient) Sites() services.SiteService               { return nil }
func (m *mockClient) Devices() services.DeviceService           { return nil }
func (m *mockClient) Networks() services.NetworkService         { return &m.networks }
func (m *mockClient) WLANs() services.WLANService               { return nil }
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
	for i := range m.users {
		if m.users[i].ID == id {
			return &m.users[i], nil
		}
	}
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
	user.ID = fmt.Sprintf("id_%d", len(m.users)+1)
	m.users = append(m.users, *user)
	return user, nil
}

func (m *mockUserService) Update(ctx context.Context, site string, user *types.User) (*types.User, error) {
	for i := range m.users {
		if m.users[i].ID == user.ID {
			m.users[i] = *user
			return user, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockUserService) Delete(ctx context.Context, site, id string) error {
	for i := range m.users {
		if m.users[i].ID == id {
			m.users = append(m.users[:i], m.users[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("not found")
}

func (m *mockUserService) DeleteByMAC(ctx context.Context, site, mac string) error {
	for i := range m.users {
		if strings.EqualFold(m.users[i].MAC, mac) {
			m.users = append(m.users[:i], m.users[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("not found")
}

func (m *mockUserService) SetFixedIP(ctx context.Context, site, mac, ip, networkID string) error {
	for i := range m.users {
		if strings.EqualFold(m.users[i].MAC, mac) {
			m.users[i].UseFixedIP = true
			m.users[i].FixedIP = ip
			m.users[i].NetworkID = networkID
			return nil
		}
	}
	return fmt.Errorf("not found")
}

func (m *mockUserService) ClearFixedIP(ctx context.Context, site, mac string) error {
	for i := range m.users {
		if strings.EqualFold(m.users[i].MAC, mac) {
			m.users[i].UseFixedIP = false
			m.users[i].FixedIP = ""
			return nil
		}
	}
	return fmt.Errorf("not found")
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
	records   []types.DNSRecord
	createErr error
}

func (m *mockDNSService) List(ctx context.Context, site string) ([]types.DNSRecord, error) {
	return m.records, nil
}

func (m *mockDNSService) Get(ctx context.Context, site, id string) (*types.DNSRecord, error) {
	for i := range m.records {
		if m.records[i].ID == id {
			return &m.records[i], nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockDNSService) GetByName(ctx context.Context, site, name string) (*types.DNSRecord, error) {
	for i := range m.records {
		if m.records[i].Key == name {
			return &m.records[i], nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockDNSService) GetByIP(ctx context.Context, site, ip string) ([]types.DNSRecord, error) {
	var result []types.DNSRecord
	for _, r := range m.records {
		if r.Value == ip {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockDNSService) Create(ctx context.Context, site string, record *types.DNSRecord) (*types.DNSRecord, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	record.ID = fmt.Sprintf("dns_%d", len(m.records)+1)
	m.records = append(m.records, *record)
	return record, nil
}

func (m *mockDNSService) Update(ctx context.Context, site string, record *types.DNSRecord) (*types.DNSRecord, error) {
	for i := range m.records {
		if m.records[i].ID == record.ID {
			m.records[i] = *record
			return record, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockDNSService) Delete(ctx context.Context, site, id string) error {
	for i := range m.records {
		if m.records[i].ID == id {
			m.records = append(m.records[:i], m.records[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("not found")
}

func (m *mockDNSService) DeleteByName(ctx context.Context, site, name string) error {
	for i := range m.records {
		if m.records[i].Key == name {
			m.records = append(m.records[:i], m.records[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("not found")
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

func newTestClient() *mockClient {
	return &mockClient{
		networks: mockNetworkService{
			networks: []types.Network{
				{ID: "net_lan", Name: "LAN", IPSubnet: "192.168.1.0/24", DomainName: "lan.test"},
				{ID: "net_iot", Name: "IoT", IPSubnet: "10.0.0.0/24"},
			},
		},
	}
}

func TestDoGet_BasicExport(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", Name: "server1", UseFixedIP: true, FixedIP: "192.168.1.10"},
		{ID: "u2", MAC: "aa:bb:cc:dd:ee:02", Name: "server2", UseFixedIP: true, FixedIP: "192.168.1.20"},
	}

	var buf bytes.Buffer
	err := DoGet(context.Background(), client, "default", &buf, FormatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "host server1 {") {
		t.Error("missing server1")
	}
	if !strings.Contains(output, "host server2 {") {
		t.Error("missing server2")
	}
	if !strings.Contains(output, "aa:bb:cc:dd:ee:01") {
		t.Error("missing MAC 01")
	}
	if !strings.Contains(output, "192.168.1.10") {
		t.Error("missing IP .10")
	}
}

func TestDoGet_NoFixedIPs(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", Name: "dynamic"},
	}

	var buf bytes.Buffer
	err := DoGet(context.Background(), client, "default", &buf, FormatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "No fixed IP assignments found") {
		t.Error("expected empty message")
	}
}

func TestDoGet_HostnameFallback(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", Hostname: "detected-host", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}

	var buf bytes.Buffer
	err := DoGet(context.Background(), client, "default", &buf, FormatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "host detected-host {") {
		t.Error("expected Hostname fallback")
	}
}

func TestDoGet_MACFallback(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}

	var buf bytes.Buffer
	err := DoGet(context.Background(), client, "default", &buf, FormatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "host aa-bb-cc-dd-ee-01 {") {
		t.Error("expected MAC-based hostname fallback")
	}
}

func TestDoSet_CreateNew(t *testing.T) {
	client := newTestClient()
	entries := []HostEntry{
		{Hostname: "server1", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
		{Hostname: "server2", MAC: "aa:bb:cc:dd:ee:02", IP: "192.168.1.20"},
	}

	result, err := DoSet(context.Background(), client, "default", entries, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created != 2 {
		t.Errorf("created = %d, want 2", result.Created)
	}
	if result.Updated != 0 {
		t.Errorf("updated = %d, want 0", result.Updated)
	}
	if result.Skipped != 0 {
		t.Errorf("skipped = %d, want 0", result.Skipped)
	}
	if len(client.users.users) != 2 {
		t.Errorf("users count = %d, want 2", len(client.users.users))
	}
}

func TestDoSet_SkipUnchanged(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", Name: "server1", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}
	entries := []HostEntry{
		{Hostname: "server1", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
	}

	result, err := DoSet(context.Background(), client, "default", entries, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("skipped = %d, want 1", result.Skipped)
	}
}

func TestDoSet_UpdateChanged(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", Name: "old-name", UseFixedIP: true, FixedIP: "192.168.1.99"},
	}
	entries := []HostEntry{
		{Hostname: "new-name", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
	}

	result, err := DoSet(context.Background(), client, "default", entries, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated != 1 {
		t.Errorf("updated = %d, want 1", result.Updated)
	}
	if client.users.users[0].FixedIP != "192.168.1.10" {
		t.Errorf("fixed IP not updated: %s", client.users.users[0].FixedIP)
	}
	if client.users.users[0].Name != "new-name" {
		t.Errorf("name not updated: %s", client.users.users[0].Name)
	}
}

func TestDoSet_DryRun(t *testing.T) {
	client := newTestClient()
	entries := []HostEntry{
		{Hostname: "server1", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
	}

	result, err := DoSet(context.Background(), client, "default", entries, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created != 1 {
		t.Errorf("created = %d, want 1", result.Created)
	}
	if len(client.users.users) != 0 {
		t.Errorf("dry run should not create users, got %d", len(client.users.users))
	}
	if len(client.dns.records) != 0 {
		t.Errorf("dry run should not create DNS records, got %d", len(client.dns.records))
	}
}

func TestDoAdd_NewEntry(t *testing.T) {
	client := newTestClient()
	entry := &HostEntry{Hostname: "newhost", MAC: "aa:bb:cc:dd:ee:ff", IP: "192.168.1.50"}

	err := DoAdd(context.Background(), client, "default", entry, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.users.users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(client.users.users))
	}
	if client.users.users[0].FixedIP != "192.168.1.50" {
		t.Errorf("IP = %q, want %q", client.users.users[0].FixedIP, "192.168.1.50")
	}
	if len(client.dns.records) != 1 {
		t.Fatalf("expected 1 DNS record, got %d", len(client.dns.records))
	}
}

func TestDoAdd_ConflictIP(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "11:22:33:44:55:66", Name: "existing", UseFixedIP: true, FixedIP: "192.168.1.50"},
	}
	entry := &HostEntry{Hostname: "newhost", MAC: "aa:bb:cc:dd:ee:ff", IP: "192.168.1.50"}

	err := DoAdd(context.Background(), client, "default", entry, false)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "already assigned") {
		t.Errorf("error = %q, want mention of already assigned", err.Error())
	}
}

func TestDoAdd_ForceOverride(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "11:22:33:44:55:66", Name: "existing", UseFixedIP: true, FixedIP: "192.168.1.50"},
	}
	entry := &HostEntry{Hostname: "newhost", MAC: "aa:bb:cc:dd:ee:ff", IP: "192.168.1.50"}

	err := DoAdd(context.Background(), client, "default", entry, true)
	if err != nil {
		t.Fatalf("force should skip conflicts: %v", err)
	}
}

func TestDoDel_ByMAC(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:ff", Name: "myhost", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}
	client.dns.records = []types.DNSRecord{
		{ID: "d1", Key: "myhost", Value: "192.168.1.10", RecordType: "A", Enabled: true},
	}

	id := DeleteIdentifier{MAC: "aa:bb:cc:dd:ee:ff"}
	err := DoDel(context.Background(), client, "default", id, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.users.users[0].UseFixedIP {
		t.Error("fixed IP should be cleared")
	}
	if len(client.dns.records) != 0 {
		t.Error("DNS record should be deleted")
	}
}

func TestDoDel_ByIP(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:ff", Name: "myhost", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}

	id := DeleteIdentifier{IP: "192.168.1.10"}
	err := DoDel(context.Background(), client, "default", id, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoDel_ByName(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:ff", Name: "myhost", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}

	id := DeleteIdentifier{Name: "myhost"}
	err := DoDel(context.Background(), client, "default", id, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoDel_NotFound(t *testing.T) {
	client := newTestClient()
	id := DeleteIdentifier{MAC: "aa:bb:cc:dd:ee:ff"}
	err := DoDel(context.Background(), client, "default", id, false, false)
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestDoDel_KeepDNS(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:ff", Name: "myhost", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}
	client.dns.records = []types.DNSRecord{
		{ID: "d1", Key: "myhost", Value: "192.168.1.10", RecordType: "A", Enabled: true},
	}

	id := DeleteIdentifier{MAC: "aa:bb:cc:dd:ee:ff"}
	err := DoDel(context.Background(), client, "default", id, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.dns.records) != 1 {
		t.Error("DNS record should be preserved with --keep-dns")
	}
}

func TestDoSet_NetworkDetection(t *testing.T) {
	client := newTestClient()
	entries := []HostEntry{
		{Hostname: "lan-host", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.50"},
		{Hostname: "iot-host", MAC: "aa:bb:cc:dd:ee:02", IP: "10.0.0.50"},
	}

	result, err := DoSet(context.Background(), client, "default", entries, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created != 2 {
		t.Errorf("created = %d, want 2", result.Created)
	}

	if client.users.users[0].NetworkID != "net_lan" {
		t.Errorf("first user network = %q, want net_lan", client.users.users[0].NetworkID)
	}
	if client.users.users[1].NetworkID != "net_iot" {
		t.Errorf("second user network = %q, want net_iot", client.users.users[1].NetworkID)
	}
}

func TestDetectNetwork_ReturnsNetworkWithDomain(t *testing.T) {
	client := newTestClient()
	nets, _ := client.networks.List(context.Background(), "default")

	network, err := detectNetwork(nets, "192.168.1.50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if network.ID != "net_lan" {
		t.Errorf("network ID = %q, want net_lan", network.ID)
	}
	if network.DomainName != "lan.test" {
		t.Errorf("domain = %q, want lan.test", network.DomainName)
	}

	if _, err := detectNetwork(nets, "8.8.8.8"); err == nil {
		t.Error("expected error for IP outside all subnets")
	}
}

func TestEnsureDNSRecord_CreatesFQDNKey(t *testing.T) {
	client := newTestClient()

	err := ensureDNSRecord(context.Background(), client, "default", "helios", "herlein.me", "192.168.4.30", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.dns.records) != 1 {
		t.Fatalf("record count = %d, want 1", len(client.dns.records))
	}
	if got := client.dns.records[0].Key; got != "helios.herlein.me" {
		t.Errorf("key = %q, want helios.herlein.me", got)
	}
	if got := client.dns.records[0].Value; got != "192.168.4.30" {
		t.Errorf("value = %q, want 192.168.4.30", got)
	}
}

func TestEnsureDNSRecord_ReplacesBareRecord(t *testing.T) {
	client := newTestClient()
	client.dns.records = []types.DNSRecord{
		{ID: "d1", Key: "helios", Value: "192.168.4.30", RecordType: "A", Enabled: true},
	}

	err := ensureDNSRecord(context.Background(), client, "default", "helios", "herlein.me", "192.168.4.30", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.dns.records) != 1 {
		t.Fatalf("record count = %d, want 1 (bare replaced, not duplicated)", len(client.dns.records))
	}
	if got := client.dns.records[0].Key; got != "helios.herlein.me" {
		t.Errorf("key = %q, want helios.herlein.me", got)
	}
}

func TestEnsureDNSRecord_UpdatesStaleValue(t *testing.T) {
	client := newTestClient()
	client.dns.records = []types.DNSRecord{
		{ID: "d1", Key: "helios.herlein.me", Value: "192.168.4.99", RecordType: "A", Enabled: true},
	}

	err := ensureDNSRecord(context.Background(), client, "default", "helios", "herlein.me", "192.168.4.30", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.dns.records) != 1 {
		t.Fatalf("record count = %d, want 1", len(client.dns.records))
	}
	if got := client.dns.records[0].Value; got != "192.168.4.30" {
		t.Errorf("value = %q, want 192.168.4.30 (updated)", got)
	}
}

func TestEnsureDNSRecord_NoOpWhenCorrect(t *testing.T) {
	client := newTestClient()
	client.dns.records = []types.DNSRecord{
		{ID: "d1", Key: "helios.herlein.me", Value: "192.168.4.30", RecordType: "A", Enabled: true},
	}

	err := ensureDNSRecord(context.Background(), client, "default", "helios", "herlein.me", "192.168.4.30", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.dns.records) != 1 {
		t.Fatalf("record count = %d, want 1", len(client.dns.records))
	}
	if got := client.dns.records[0].ID; got != "d1" {
		t.Errorf("record ID = %q, want d1 (unchanged, not recreated)", got)
	}
}

func TestEnsureDNSRecord_BareFallbackWhenNoDomain(t *testing.T) {
	client := newTestClient()

	err := ensureDNSRecord(context.Background(), client, "default", "iot-host", "", "10.0.0.50", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.dns.records) != 1 {
		t.Fatalf("record count = %d, want 1", len(client.dns.records))
	}
	if got := client.dns.records[0].Key; got != "iot-host" {
		t.Errorf("key = %q, want iot-host (bare fallback)", got)
	}
}

func TestEnsureDNSRecord_DryRunNoMutation(t *testing.T) {
	client := newTestClient()

	err := ensureDNSRecord(context.Background(), client, "default", "helios", "herlein.me", "192.168.4.30", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.dns.records) != 0 {
		t.Errorf("dry run created %d records, want 0", len(client.dns.records))
	}
}

func TestDoSet_UnchangedUserRepairsMissingDNS(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", Name: "server1", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}
	entries := []HostEntry{
		{Hostname: "server1", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
	}

	result, err := DoSet(context.Background(), client, "default", entries, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("skipped = %d, want 1 (user record unchanged)", result.Skipped)
	}
	if len(client.dns.records) != 1 {
		t.Fatalf("DNS record count = %d, want 1 (repaired despite user skip)", len(client.dns.records))
	}
	if got := client.dns.records[0].Key; got != "server1.lan.test" {
		t.Errorf("key = %q, want server1.lan.test", got)
	}
}

func TestDoSet_UnchangedUserReplacesBareDNS(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", Name: "server1", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}
	client.dns.records = []types.DNSRecord{
		{ID: "d1", Key: "server1", Value: "192.168.1.10", RecordType: "A", Enabled: true},
	}
	entries := []HostEntry{
		{Hostname: "server1", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
	}

	result, err := DoSet(context.Background(), client, "default", entries, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("skipped = %d, want 1", result.Skipped)
	}
	if len(client.dns.records) != 1 {
		t.Fatalf("DNS record count = %d, want 1 (bare replaced with FQDN)", len(client.dns.records))
	}
	if got := client.dns.records[0].Key; got != "server1.lan.test" {
		t.Errorf("key = %q, want server1.lan.test", got)
	}
}

func TestDoSet_ForceRewritesUnchanged(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", Name: "server1", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}
	entries := []HostEntry{
		{Hostname: "server1", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
	}

	result, err := DoSet(context.Background(), client, "default", entries, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated != 1 {
		t.Errorf("updated = %d, want 1 (force rewrites unchanged)", result.Updated)
	}
	if result.Skipped != 0 {
		t.Errorf("skipped = %d, want 0 (force never skips)", result.Skipped)
	}
	if len(client.dns.records) != 1 {
		t.Errorf("DNS record count = %d, want 1", len(client.dns.records))
	}
}

func TestDNSHostnameMatches(t *testing.T) {
	cases := []struct {
		dnsKey   string
		hostname string
		want     bool
	}{
		{"helios.herlein.me", "helios", true},
		{"helios", "helios", true},
		{"printer.herlein.me", "helios", false},
		{"heliosx.herlein.me", "helios", false},
	}
	for _, c := range cases {
		if got := dnsHostnameMatches(c.dnsKey, c.hostname); got != c.want {
			t.Errorf("dnsHostnameMatches(%q, %q) = %v, want %v", c.dnsKey, c.hostname, got, c.want)
		}
	}
}

func TestEnsureDNSRecord_SkipsWhenDeviceLocalResolvesCorrectly(t *testing.T) {
	client := newTestClient()
	setResolver(t, func(ctx context.Context, fqdn string) ([]string, error) {
		if fqdn == "helios.herlein.me" {
			return []string{"192.168.4.30"}, nil
		}
		return nil, nil
	})

	err := ensureDNSRecord(context.Background(), client, "default", "helios", "herlein.me", "192.168.4.30", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.dns.records) != 0 {
		t.Errorf("record count = %d, want 0 (device-local DNS already serves it; a static record would be rejected)", len(client.dns.records))
	}
}

func TestEnsureDNSRecord_ReportsDriftWhenDeviceLocalWrong(t *testing.T) {
	client := newTestClient()
	setResolver(t, func(ctx context.Context, fqdn string) ([]string, error) {
		return []string{"192.168.4.99"}, nil
	})

	err := ensureDNSRecord(context.Background(), client, "default", "helios", "herlein.me", "192.168.4.30", false)
	if err == nil {
		t.Fatal("expected drift error when device-local DNS resolves to a different IP")
	}
	if !strings.Contains(err.Error(), "device local DNS") {
		t.Errorf("error = %q, want mention of device local DNS", err.Error())
	}
	if len(client.dns.records) != 0 {
		t.Errorf("record count = %d, want 0 (cannot override device-local DNS)", len(client.dns.records))
	}
}

func TestEnsureDNSRecord_ToleratesOverlapOnCreate(t *testing.T) {
	client := newTestClient()
	client.dns.createErr = fmt.Errorf("create DNS record failed with status 400: {\"code\":\"api.err.StaticDnsOverlapsWithDeviceLocalDns\"}")
	setResolver(t, unresolvedResolver)

	err := ensureDNSRecord(context.Background(), client, "default", "helios", "herlein.me", "192.168.4.30", false)
	if err != nil {
		t.Fatalf("overlap-on-create should be tolerated (name already served by device-local DNS), got: %v", err)
	}
}

func TestContainsString(t *testing.T) {
	if !containsString([]string{"a", "b"}, "b") {
		t.Error("want true for present element")
	}
	if containsString([]string{"a"}, "b") {
		t.Error("want false for absent element")
	}
	if containsString(nil, "b") {
		t.Error("want false for nil slice")
	}
}

func TestDoGet_WarnsOnValueDrift(t *testing.T) {
	client := newTestClient()
	client.users.users = []types.User{
		{ID: "u1", MAC: "aa:bb:cc:dd:ee:01", Name: "server1", UseFixedIP: true, FixedIP: "192.168.1.10"},
	}
	setResolver(t, func(ctx context.Context, fqdn string) ([]string, error) {
		if fqdn == "server1.lan.test" {
			return []string{"192.168.1.99"}, nil
		}
		return nil, nil
	})

	old := os.Stderr
	reader, writer, _ := os.Pipe()
	os.Stderr = writer
	var buf bytes.Buffer
	err := DoGet(context.Background(), client, "default", &buf, FormatOptions{})
	writer.Close()
	os.Stderr = old
	stderr, _ := io.ReadAll(reader)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(stderr), "192.168.1.99") {
		t.Errorf("expected drift warning mentioning the resolved IP, got stderr: %q", string(stderr))
	}
}
