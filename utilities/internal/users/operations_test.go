package users

import (
	"bytes"
	"context"
	"crypto/tls"
	"strings"
	"testing"

	gofi "github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/src/mock"
	"github.com/unifi-go/gofi/src/types"
)

func newTestClient(t *testing.T) (*mock.Server, gofi.Client) {
	t.Helper()

	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())

	config := &gofi.Config{
		Host:          server.URL(),
		BaseURL:       server.URL(),
		Username:      "admin",
		Password:      "admin",
		Site:          "default",
		SkipTLSVerify: true,
		TLSConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	client, err := gofi.New(config)
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create client: %v", err)
	}

	if err := client.Connect(context.Background()); err != nil {
		server.Close()
		t.Fatalf("Failed to connect: %v", err)
	}

	return server, client
}

// seedUser adds both the user record and the matching client record. Removal
// touches both surfaces, so a test fixture needs both to be realistic.
func seedUser(server *mock.Server, id, name, mac, fixedIP string) {
	server.State().AddKnownClient(&types.User{
		ID:         id,
		Name:       name,
		MAC:        mac,
		FixedIP:    fixedIP,
		UseFixedIP: fixedIP != "",
	})
	server.State().AddClient(&types.Client{MAC: mac})
}

func TestBuildUserEntriesSortsByNameThenMAC(t *testing.T) {
	users := []types.User{
		{ID: "u3", Name: "zeta", MAC: "AA:BB:CC:00:00:03"},
		{ID: "u1", Name: "alpha", MAC: "AA:BB:CC:00:00:02"},
		{ID: "u2", Name: "alpha", MAC: "AA:BB:CC:00:00:01"},
	}

	entries := buildUserEntries(users, "")

	want := []string{"alpha/aa:bb:cc:00:00:01", "alpha/aa:bb:cc:00:00:02", "zeta/aa:bb:cc:00:00:03"}
	for i, expected := range want {
		got := entries[i].Name + "/" + entries[i].MAC
		if got != expected {
			t.Errorf("entry %d = %q, want %q", i, got, expected)
		}
	}
}

// MACs are normalized so a filter or comparison is not defeated by the casing
// the controller happens to return.
func TestBuildUserEntriesLowercasesMAC(t *testing.T) {
	entries := buildUserEntries([]types.User{{ID: "u1", MAC: "AA:BB:CC:DD:EE:FF"}}, "")

	if entries[0].MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("MAC = %q, want lowercase", entries[0].MAC)
	}
}

func TestBuildUserEntriesFilter(t *testing.T) {
	users := []types.User{
		{ID: "u1", Name: "tapo1", MAC: "aa:bb:cc:00:00:01", FixedIP: "192.168.1.10", UseFixedIP: true},
		{ID: "u2", Name: "printer", Hostname: "hp-office", MAC: "aa:bb:cc:00:00:02"},
	}

	tests := []struct {
		name    string
		filter  string
		wantIDs []string
	}{
		{"empty matches all", "", []string{"printer", "tapo1"}},
		{"by name", "tapo", []string{"tapo1"}},
		{"by hostname", "hp-office", []string{"printer"}},
		{"by mac fragment", "00:00:02", []string{"printer"}},
		{"by fixed ip", "192.168.1.10", []string{"tapo1"}},
		{"case insensitive", "TAPO", []string{"tapo1"}},
		{"no match", "nothing", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := buildUserEntries(users, tt.filter)
			if len(entries) != len(tt.wantIDs) {
				t.Fatalf("Got %d entries, want %d", len(entries), len(tt.wantIDs))
			}
			for i, want := range tt.wantIDs {
				if entries[i].Name != want {
					t.Errorf("entry %d = %q, want %q", i, entries[i].Name, want)
				}
			}
		})
	}
}

func TestMatchesFilter(t *testing.T) {
	entry := UserEntry{Name: "tapo1", Hostname: "plug", MAC: "aa:bb:cc:00:00:01", FixedIP: "192.168.1.10"}

	if !matchesFilter(entry, "") {
		t.Error("Empty filter should match")
	}
	if !matchesFilter(entry, "PLUG") {
		t.Error("Filter should be case insensitive")
	}
	if matchesFilter(entry, "absent") {
		t.Error("Unrelated filter should not match")
	}
}

func TestFindUser(t *testing.T) {
	users := []types.User{
		{ID: "u1", Name: "tapo1", MAC: "aa:bb:cc:00:00:01"},
		{ID: "u2", Name: "dupe", MAC: "aa:bb:cc:00:00:02"},
		{ID: "u3", Name: "dupe", MAC: "aa:bb:cc:00:00:03"},
	}

	tests := []struct {
		name       string
		identifier DeleteIdentifier
		wantID     string
		wantErr    bool
	}{
		{"by mac", DeleteIdentifier{MAC: "aa:bb:cc:00:00:01"}, "u1", false},
		{"by mac uppercase", DeleteIdentifier{MAC: "AA:BB:CC:00:00:01"}, "u1", false},
		{"by name", DeleteIdentifier{Name: "tapo1"}, "u1", false},
		{"ambiguous name", DeleteIdentifier{Name: "dupe"}, "", true},
		{"missing", DeleteIdentifier{MAC: "ff:ff:ff:ff:ff:ff"}, "", true},
		{"empty identifier", DeleteIdentifier{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := findUser(users, tt.identifier)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("findUser failed: %v", err)
			}
			if user.ID != tt.wantID {
				t.Errorf("Got %q, want %q", user.ID, tt.wantID)
			}
		})
	}
}

func TestDescribeIdentifier(t *testing.T) {
	tests := []struct {
		identifier DeleteIdentifier
		want       string
	}{
		{DeleteIdentifier{MAC: "aa:bb:cc:dd:ee:ff"}, "mac aa:bb:cc:dd:ee:ff"},
		{DeleteIdentifier{Name: "tapo1"}, "name tapo1"},
		{DeleteIdentifier{}, "no identifier"},
	}

	for _, tt := range tests {
		if got := describeIdentifier(tt.identifier); got != tt.want {
			t.Errorf("describeIdentifier(%+v) = %q, want %q", tt.identifier, got, tt.want)
		}
	}
}

func TestListUsers(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seedUser(server, "u1", "zeta", "aa:bb:cc:00:00:02", "")
	seedUser(server, "u2", "alpha", "aa:bb:cc:00:00:01", "192.168.1.10")

	entries, err := ListUsers(context.Background(), client, "default", "")
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "alpha" {
		t.Errorf("Entries not sorted; first is %q", entries[0].Name)
	}
}

func TestDoListText(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seedUser(server, "u1", "tapo1", "aa:bb:cc:00:00:01", "192.168.1.10")

	var out bytes.Buffer
	if err := DoList(context.Background(), client, "default", "", FormatOptions{Writer: &out}); err != nil {
		t.Fatalf("DoList failed: %v", err)
	}

	for _, want := range []string{"NAME", "tapo1", "192.168.1.10"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("Output missing %q:\n%s", want, out.String())
		}
	}
}

func TestDoListJSON(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seedUser(server, "u1", "tapo1", "aa:bb:cc:00:00:01", "192.168.1.10")

	var out bytes.Buffer
	if err := DoList(context.Background(), client, "default", "", FormatOptions{Writer: &out, JSON: true}); err != nil {
		t.Fatalf("DoList failed: %v", err)
	}

	if !strings.Contains(out.String(), `"name": "tapo1"`) {
		t.Errorf("JSON output unexpected: %q", out.String())
	}
}

func TestDoDelClearsFixedIPThenForgets(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seedUser(server, "u1", "tapo1", "aa:bb:cc:00:00:01", "192.168.1.10")

	var progress bytes.Buffer
	result, err := DoDel(context.Background(), client, "default", DeleteIdentifier{MAC: "aa:bb:cc:00:00:01"}, false, &progress)
	if err != nil {
		t.Fatalf("DoDel failed: %v", err)
	}

	if !result.ClearedFixedIP {
		t.Error("Expected the fixed IP to be cleared")
	}
	if !result.Forgot {
		t.Error("Expected the client to be forgotten")
	}
	if server.State().GetClient("aa:bb:cc:00:00:01") != nil {
		t.Error("Client record still present")
	}
}

// A client with no reservation must still be forgotten, without a pointless
// ClearFixedIP call.
func TestDoDelWithoutFixedIP(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seedUser(server, "u1", "plain", "aa:bb:cc:00:00:09", "")

	var progress bytes.Buffer
	result, err := DoDel(context.Background(), client, "default", DeleteIdentifier{MAC: "aa:bb:cc:00:00:09"}, false, &progress)
	if err != nil {
		t.Fatalf("DoDel failed: %v", err)
	}

	if result.ClearedFixedIP {
		t.Error("Should not have cleared a fixed IP that was never set")
	}
	if !result.Forgot {
		t.Error("Expected the client to be forgotten")
	}
}

func TestDoDelDryRunChangesNothing(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seedUser(server, "u1", "tapo1", "aa:bb:cc:00:00:01", "192.168.1.10")

	var progress bytes.Buffer
	result, err := DoDel(context.Background(), client, "default", DeleteIdentifier{MAC: "aa:bb:cc:00:00:01"}, true, &progress)
	if err != nil {
		t.Fatalf("DoDel failed: %v", err)
	}

	if result.ClearedFixedIP || result.Forgot {
		t.Errorf("Dry run reported changes: %+v", result)
	}
	if server.State().GetClient("aa:bb:cc:00:00:01") == nil {
		t.Error("Dry run removed the client")
	}
	if !strings.Contains(progress.String(), "would forget") {
		t.Errorf("Progress missing dry-run notice: %q", progress.String())
	}
}

func TestDoDelNoMatch(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	var progress bytes.Buffer
	if _, err := DoDel(context.Background(), client, "default", DeleteIdentifier{MAC: "ff:ff:ff:ff:ff:ff"}, false, &progress); err == nil {
		t.Fatal("Expected an error when nothing matches")
	}
}
