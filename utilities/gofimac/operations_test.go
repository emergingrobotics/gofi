package main

import (
	"strings"
	"testing"

	"github.com/unifi-go/gofi/types"
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

// listClientsFromSlice is a test helper that processes a slice of clients
// without requiring a UDM connection.
func listClientsFromSlice(clients []types.Client, filter FilterMode, ouiDatabase *OUIDatabase) ([]ClientEntry, error) {
	entries := []ClientEntry{}
	for _, activeClient := range clients {
		if !matchesFilter(activeClient, filter) {
			continue
		}
		entries = append(entries, buildClientEntry(activeClient, ouiDatabase))
	}
	sortClientEntries(entries)
	return entries, nil
}
