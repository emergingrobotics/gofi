package clients

import (
	"strings"
	"testing"

	"github.com/unifi-go/gofi/src/types"
)

func newTestOUIDatabase(t *testing.T) *OUIDatabase {
	t.Helper()
	database, err := ParseOUIDatabase(strings.NewReader(sampleOUIData))
	if err != nil {
		t.Fatalf("failed to create test OUI database: %v", err)
	}
	return database
}

func TestListClients_FilterAll(t *testing.T) {
	clients := []types.Client{
		{MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10", Name: "wired-host", IsWired: true},
		{MAC: "11:22:33:dd:ee:02", IP: "192.168.1.20", Name: "wifi-host", IsWired: false},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(clients, FilterAll, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestListClients_FilterWifi(t *testing.T) {
	clients := []types.Client{
		{MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10", Name: "wired-host", IsWired: true},
		{MAC: "11:22:33:dd:ee:02", IP: "192.168.1.20", Name: "wifi-host", IsWired: false, ESSID: "MyNetwork", Channel: 36},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(clients, FilterWifi, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Hostname != "wifi-host" {
		t.Errorf("expected wifi-host, got %s", entries[0].Hostname)
	}
	if entries[0].IsWired {
		t.Error("expected IsWired to be false")
	}
	if entries[0].ESSID != "MyNetwork" {
		t.Errorf("expected ESSID MyNetwork, got %s", entries[0].ESSID)
	}
}

func TestListClients_FilterWired(t *testing.T) {
	clients := []types.Client{
		{MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10", Name: "wired-host", IsWired: true, SWMAC: "ff:ff:ff:00:00:01", SWPORT: 4},
		{MAC: "11:22:33:dd:ee:02", IP: "192.168.1.20", Name: "wifi-host", IsWired: false},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(clients, FilterWired, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Hostname != "wired-host" {
		t.Errorf("expected wired-host, got %s", entries[0].Hostname)
	}
	if !entries[0].IsWired {
		t.Error("expected IsWired to be true")
	}
	if entries[0].SwitchPort != 4 {
		t.Errorf("expected switch port 4, got %d", entries[0].SwitchPort)
	}
}

func TestListClients_HostnameResolution(t *testing.T) {
	clients := []types.Client{
		{MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10", Name: "custom-name", Hostname: "device-hostname"},
		{MAC: "aa:bb:cc:dd:ee:02", IP: "192.168.1.11", Hostname: "fallback-hostname"},
		{MAC: "aa:bb:cc:dd:ee:03", IP: "192.168.1.12"},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(clients, FilterAll, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entries[0].Hostname != "custom-name" {
		t.Errorf("expected custom-name, got %s", entries[0].Hostname)
	}
	if entries[1].Hostname != "fallback-hostname" {
		t.Errorf("expected fallback-hostname, got %s", entries[1].Hostname)
	}
	if entries[2].Hostname != "unknown" {
		t.Errorf("expected unknown, got %s", entries[2].Hostname)
	}
}

func TestListClients_OUILookup(t *testing.T) {
	clients := []types.Client{
		{MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10", Name: "known"},
		{MAC: "ff:ee:dd:cc:bb:aa", IP: "192.168.1.11", Name: "unknown-oui"},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(clients, FilterAll, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entries[0].Manufacturer != "Acme Corporation" {
		t.Errorf("expected Acme Corporation, got %s", entries[0].Manufacturer)
	}
	if entries[1].Manufacturer != "unknown" {
		t.Errorf("expected unknown, got %s", entries[1].Manufacturer)
	}
}

func TestListClients_SortByIP(t *testing.T) {
	clients := []types.Client{
		{MAC: "aa:bb:cc:dd:ee:03", IP: "192.168.1.100", Name: "c"},
		{MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.2", Name: "a"},
		{MAC: "aa:bb:cc:dd:ee:02", IP: "192.168.1.10", Name: "b"},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(clients, FilterAll, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOrder := []string{"192.168.1.2", "192.168.1.10", "192.168.1.100"}
	for index, expected := range expectedOrder {
		if entries[index].IP != expected {
			t.Errorf("position %d: expected %s, got %s", index, expected, entries[index].IP)
		}
	}
}

func TestListClients_NoIPSortsLast(t *testing.T) {
	clients := []types.Client{
		{MAC: "aa:bb:cc:dd:ee:01", Name: "no-ip"},
		{MAC: "aa:bb:cc:dd:ee:02", IP: "192.168.1.10", Name: "has-ip"},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(clients, FilterAll, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entries[0].Hostname != "has-ip" {
		t.Errorf("expected has-ip first, got %s", entries[0].Hostname)
	}
	if entries[1].Hostname != "no-ip" {
		t.Errorf("expected no-ip last, got %s", entries[1].Hostname)
	}
}

func TestListClients_MultipleNoIPSortsByMAC(t *testing.T) {
	clients := []types.Client{
		{MAC: "cc:cc:cc:dd:ee:01", Name: "c"},
		{MAC: "aa:aa:aa:dd:ee:02", Name: "a"},
		{MAC: "bb:bb:bb:dd:ee:03", Name: "b"},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(clients, FilterAll, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedMACs := []string{"aa:aa:aa:dd:ee:02", "bb:bb:bb:dd:ee:03", "cc:cc:cc:dd:ee:01"}
	for index, expected := range expectedMACs {
		if entries[index].MAC != expected {
			t.Errorf("position %d: expected %s, got %s", index, expected, entries[index].MAC)
		}
	}
}

func TestListClients_Empty(t *testing.T) {
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(nil, FilterAll, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestListClients_MACLowercase(t *testing.T) {
	clients := []types.Client{
		{MAC: "AA:BB:CC:DD:EE:FF", IP: "192.168.1.10", Name: "test"},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries, err := listClientsFromSlice(clients, FilterAll, ouiDatabase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entries[0].MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected lowercase MAC, got %s", entries[0].MAC)
	}
}

func TestParseSortMode(t *testing.T) {
	cases := []struct {
		input    string
		expected SortMode
	}{
		{"first-seen", SortFirstSeen},
		{"last-seen", SortLastSeen},
		{"ip", SortIP},
	}
	for _, testCase := range cases {
		result, err := ParseSortMode(testCase.input)
		if err != nil {
			t.Errorf("ParseSortMode(%q) unexpected error: %v", testCase.input, err)
			continue
		}
		if result != testCase.expected {
			t.Errorf("ParseSortMode(%q) = %d, want %d", testCase.input, result, testCase.expected)
		}
	}
	if _, err := ParseSortMode("bogus"); err == nil {
		t.Error("ParseSortMode(\"bogus\") expected error, got nil")
	}
}

func TestSortClientEntries_FirstSeenDescending(t *testing.T) {
	entries := []ClientEntry{
		{MAC: "aa:aa:aa:00:00:01", IP: "192.168.1.10", FirstSeen: 100},
		{MAC: "aa:aa:aa:00:00:02", IP: "192.168.1.11", FirstSeen: 300},
		{MAC: "aa:aa:aa:00:00:03", IP: "192.168.1.12", FirstSeen: 200},
	}
	sortClientEntries(entries, SortFirstSeen)

	expected := []int64{300, 200, 100}
	for index, want := range expected {
		if entries[index].FirstSeen != want {
			t.Errorf("position %d: FirstSeen = %d, want %d", index, entries[index].FirstSeen, want)
		}
	}
}

func TestSortClientEntries_LastSeenDescending(t *testing.T) {
	entries := []ClientEntry{
		{MAC: "aa:aa:aa:00:00:01", IP: "192.168.1.10", LastSeen: 500},
		{MAC: "aa:aa:aa:00:00:02", IP: "192.168.1.11", LastSeen: 900},
		{MAC: "aa:aa:aa:00:00:03", IP: "192.168.1.12", LastSeen: 700},
	}
	sortClientEntries(entries, SortLastSeen)

	expected := []int64{900, 700, 500}
	for index, want := range expected {
		if entries[index].LastSeen != want {
			t.Errorf("position %d: LastSeen = %d, want %d", index, entries[index].LastSeen, want)
		}
	}
}

func TestSortClientEntries_FirstSeenTieBreaksByIP(t *testing.T) {
	entries := []ClientEntry{
		{MAC: "aa:aa:aa:00:00:01", IP: "192.168.1.100", FirstSeen: 0},
		{MAC: "aa:aa:aa:00:00:02", IP: "192.168.1.2", FirstSeen: 0},
		{MAC: "aa:aa:aa:00:00:03", IP: "192.168.1.10", FirstSeen: 0},
	}
	sortClientEntries(entries, SortFirstSeen)

	expected := []string{"192.168.1.2", "192.168.1.10", "192.168.1.100"}
	for index, want := range expected {
		if entries[index].IP != want {
			t.Errorf("position %d: IP = %s, want %s", index, entries[index].IP, want)
		}
	}
}

func TestBuildHistoryEntries_MarksPresentAndGone(t *testing.T) {
	active := []types.Client{
		{MAC: "aa:bb:cc:00:00:01", IP: "192.168.1.10", Name: "present-host", FirstSeen: 100, LastSeen: 900},
	}
	all := []types.Client{
		{MAC: "aa:bb:cc:00:00:01", IP: "192.168.1.10", Name: "present-host", FirstSeen: 100, LastSeen: 900},
		{MAC: "aa:bb:cc:00:00:02", IP: "192.168.1.11", Name: "gone-host", FirstSeen: 50, LastSeen: 400},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries := buildHistoryEntries(active, all, FilterAll, SortFirstSeen, false, ouiDatabase)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	statusByHost := map[string]string{}
	for _, entry := range entries {
		statusByHost[entry.Hostname] = entry.Status
	}
	if statusByHost["present-host"] != StatusPresent {
		t.Errorf("present-host status = %q, want %q", statusByHost["present-host"], StatusPresent)
	}
	if statusByHost["gone-host"] != StatusGone {
		t.Errorf("gone-host status = %q, want %q", statusByHost["gone-host"], StatusGone)
	}
}

func TestBuildHistoryEntries_GoneOnlyExcludesPresent(t *testing.T) {
	active := []types.Client{
		{MAC: "aa:bb:cc:00:00:01", IP: "192.168.1.10", Name: "present-host"},
	}
	all := []types.Client{
		{MAC: "aa:bb:cc:00:00:01", IP: "192.168.1.10", Name: "present-host"},
		{MAC: "aa:bb:cc:00:00:02", IP: "192.168.1.11", Name: "gone-host"},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries := buildHistoryEntries(active, all, FilterAll, SortFirstSeen, true, ouiDatabase)
	if len(entries) != 1 {
		t.Fatalf("expected 1 gone entry, got %d", len(entries))
	}
	if entries[0].Hostname != "gone-host" {
		t.Errorf("expected gone-host, got %s", entries[0].Hostname)
	}
	if entries[0].Status != StatusGone {
		t.Errorf("expected status gone, got %s", entries[0].Status)
	}
}

func TestBuildHistoryEntries_PrefersActiveRecordForPresent(t *testing.T) {
	active := []types.Client{
		{MAC: "aa:bb:cc:00:00:01", IP: "192.168.1.10", Name: "live-name"},
	}
	// The historical record for the same MAC carries stale/missing fields.
	all := []types.Client{
		{MAC: "aa:bb:cc:00:00:01", IP: "", Name: "stale-name"},
	}
	ouiDatabase := newTestOUIDatabase(t)

	entries := buildHistoryEntries(active, all, FilterAll, SortFirstSeen, false, ouiDatabase)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].IP != "192.168.1.10" {
		t.Errorf("expected live IP 192.168.1.10, got %q", entries[0].IP)
	}
	if entries[0].Hostname != "live-name" {
		t.Errorf("expected live-name, got %s", entries[0].Hostname)
	}
}

func TestSelectEntryByMAC(t *testing.T) {
	entries := []ClientEntry{
		{MAC: "aa:bb:cc:00:00:01", Hostname: "one", Status: StatusPresent},
		{MAC: "aa:bb:cc:00:00:02", Hostname: "two", Status: StatusGone},
	}

	found := selectEntryByMAC(entries, "aa:bb:cc:00:00:02")
	if found == nil || found.Hostname != "two" {
		t.Fatalf("expected to find 'two', got %+v", found)
	}

	// Case-insensitive match.
	upper := selectEntryByMAC(entries, "AA:BB:CC:00:00:01")
	if upper == nil || upper.Hostname != "one" {
		t.Fatalf("expected case-insensitive match for 'one', got %+v", upper)
	}

	if selectEntryByMAC(entries, "ff:ff:ff:ff:ff:ff") != nil {
		t.Error("expected nil for unknown MAC")
	}
}

// listClientsFromSlice is a test helper that processes a slice of clients
// without requiring a UDM connection.
func listClientsFromSlice(clients []types.Client, filter FilterMode, ouiDatabase *OUIDatabase) ([]ClientEntry, error) {
	entries := []ClientEntry{}
	for _, activeClient := range clients {
		if !matchesFilter(activeClient, filter) {
			continue
		}
		entries = append(entries, buildClientEntry(activeClient, ouiDatabase, true))
	}
	sortClientEntries(entries, SortFirstSeen)
	return entries, nil
}
