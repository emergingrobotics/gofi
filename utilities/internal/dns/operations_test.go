package dns

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

func seed(server *mock.Server, id, key, value string) {
	server.State().AddDNSRecord(&types.DNSRecord{
		ID:         id,
		Key:        key,
		Value:      value,
		RecordType: types.DNSRecordTypeA,
		Enabled:    true,
	})
}

func TestBuildDNSEntriesSortsByKeyThenValue(t *testing.T) {
	records := []types.DNSRecord{
		{ID: "3", Key: "b.example.com", Value: "192.168.1.3"},
		{ID: "1", Key: "a.example.com", Value: "192.168.1.2"},
		{ID: "2", Key: "a.example.com", Value: "192.168.1.1"},
	}

	entries := buildDNSEntries(records)

	want := []string{"a.example.com/192.168.1.1", "a.example.com/192.168.1.2", "b.example.com/192.168.1.3"}
	for i, expected := range want {
		got := entries[i].Key + "/" + entries[i].Value
		if got != expected {
			t.Errorf("entry %d = %q, want %q", i, got, expected)
		}
	}
}

func TestBuildDNSEntriesEmpty(t *testing.T) {
	if entries := buildDNSEntries(nil); len(entries) != 0 {
		t.Errorf("Expected no entries, got %d", len(entries))
	}
}

func TestBuildDNSEntriesCopiesFields(t *testing.T) {
	entries := buildDNSEntries([]types.DNSRecord{{
		ID:         "dns1",
		Key:        "a.example.com",
		Value:      "192.168.1.1",
		RecordType: types.DNSRecordTypeA,
		TTL:        300,
		Enabled:    true,
	}})

	entry := entries[0]
	if entry.ID != "dns1" || entry.Type != "A" || entry.TTL != 300 || !entry.Enabled {
		t.Errorf("Fields not copied faithfully: %+v", entry)
	}
}

func TestMatchRecords(t *testing.T) {
	records := []types.DNSRecord{
		{ID: "dns1", Key: "a.example.com", Value: "192.168.1.1"},
		{ID: "dns2", Key: "alias.example.com", Value: "192.168.1.1"},
		{ID: "dns3", Key: "b.example.com", Value: "192.168.1.2"},
	}

	tests := []struct {
		name       string
		identifier DeleteIdentifier
		wantIDs    []string
	}{
		{name: "by id", identifier: DeleteIdentifier{ID: "dns2"}, wantIDs: []string{"dns2"}},
		{name: "by name", identifier: DeleteIdentifier{Name: "b.example.com"}, wantIDs: []string{"dns3"}},
		{name: "by name is case insensitive", identifier: DeleteIdentifier{Name: "B.EXAMPLE.COM"}, wantIDs: []string{"dns3"}},
		{name: "by ip matches several", identifier: DeleteIdentifier{IP: "192.168.1.1"}, wantIDs: []string{"dns1", "dns2"}},
		{name: "no match", identifier: DeleteIdentifier{ID: "nope"}, wantIDs: nil},
		{name: "empty identifier", identifier: DeleteIdentifier{}, wantIDs: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matchRecords(records, tt.identifier)
			if len(matches) != len(tt.wantIDs) {
				t.Fatalf("Got %d matches, want %d", len(matches), len(tt.wantIDs))
			}
			for i, id := range tt.wantIDs {
				if matches[i].ID != id {
					t.Errorf("match %d = %q, want %q", i, matches[i].ID, id)
				}
			}
		})
	}
}

func TestDescribeIdentifier(t *testing.T) {
	tests := []struct {
		identifier DeleteIdentifier
		want       string
	}{
		{DeleteIdentifier{ID: "dns1"}, "id dns1"},
		{DeleteIdentifier{Name: "a.example.com"}, "name a.example.com"},
		{DeleteIdentifier{IP: "192.168.1.1"}, "ip 192.168.1.1"},
		{DeleteIdentifier{}, "no identifier"},
	}

	for _, tt := range tests {
		if got := describeIdentifier(tt.identifier); got != tt.want {
			t.Errorf("describeIdentifier(%+v) = %q, want %q", tt.identifier, got, tt.want)
		}
	}
}

func TestListRecords(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seed(server, "dns1", "b.example.com", "192.168.1.2")
	seed(server, "dns2", "a.example.com", "192.168.1.1")

	entries, err := ListRecords(context.Background(), client, "default")
	if err != nil {
		t.Fatalf("ListRecords failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}
	if entries[0].Key != "a.example.com" {
		t.Errorf("Entries not sorted; first is %q", entries[0].Key)
	}
}

func TestDoGetText(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seed(server, "dns1", "a.example.com", "192.168.1.1")

	var out bytes.Buffer
	if err := DoGet(context.Background(), client, "default", FormatOptions{Writer: &out}); err != nil {
		t.Fatalf("DoGet failed: %v", err)
	}

	if !strings.Contains(out.String(), "a.example.com") {
		t.Errorf("Output missing record: %q", out.String())
	}
	if !strings.Contains(out.String(), "NAME") {
		t.Errorf("Output missing header: %q", out.String())
	}
}

func TestDoGetJSON(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seed(server, "dns1", "a.example.com", "192.168.1.1")

	var out bytes.Buffer
	if err := DoGet(context.Background(), client, "default", FormatOptions{Writer: &out, JSON: true}); err != nil {
		t.Fatalf("DoGet failed: %v", err)
	}

	if !strings.Contains(out.String(), `"key": "a.example.com"`) {
		t.Errorf("JSON output unexpected: %q", out.String())
	}
}

func TestDoDelByID(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seed(server, "dns1", "a.example.com", "192.168.1.1")

	var progress bytes.Buffer
	result, err := DoDel(context.Background(), client, "default", DeleteIdentifier{ID: "dns1"}, false, false, &progress)
	if err != nil {
		t.Fatalf("DoDel failed: %v", err)
	}

	if result.Deleted != 1 || result.Errors != 0 {
		t.Errorf("Result = %+v, want 1 deleted 0 errors", result)
	}
	if server.State().GetDNSRecord("dns1") != nil {
		t.Error("Record was not deleted")
	}
}

func TestDoDelDryRunLeavesRecord(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seed(server, "dns1", "a.example.com", "192.168.1.1")

	var progress bytes.Buffer
	result, err := DoDel(context.Background(), client, "default", DeleteIdentifier{ID: "dns1"}, true, false, &progress)
	if err != nil {
		t.Fatalf("DoDel failed: %v", err)
	}

	if result.Deleted != 1 {
		t.Errorf("Expected 1 counted, got %d", result.Deleted)
	}
	if server.State().GetDNSRecord("dns1") == nil {
		t.Error("Dry run deleted the record")
	}
	if !strings.Contains(progress.String(), "would delete") {
		t.Errorf("Progress missing dry-run notice: %q", progress.String())
	}
}

// An identifier that selects several records must not delete any of them
// unless the operator opted in, so an ambiguous --ip cannot quietly overreach.
func TestDoDelMultiMatchRequiresForce(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seed(server, "dns1", "a.example.com", "192.168.1.1")
	seed(server, "dns2", "alias.example.com", "192.168.1.1")

	var progress bytes.Buffer
	_, err := DoDel(context.Background(), client, "default", DeleteIdentifier{IP: "192.168.1.1"}, false, false, &progress)
	if err == nil {
		t.Fatal("Expected an error for an ambiguous identifier")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("Error should point at --force, got %q", err.Error())
	}

	if server.State().GetDNSRecord("dns1") == nil || server.State().GetDNSRecord("dns2") == nil {
		t.Error("Records were deleted despite the refusal")
	}
}

func TestDoDelMultiMatchWithForce(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	seed(server, "dns1", "a.example.com", "192.168.1.1")
	seed(server, "dns2", "alias.example.com", "192.168.1.1")

	var progress bytes.Buffer
	result, err := DoDel(context.Background(), client, "default", DeleteIdentifier{IP: "192.168.1.1"}, false, true, &progress)
	if err != nil {
		t.Fatalf("DoDel failed: %v", err)
	}

	if result.Deleted != 2 {
		t.Errorf("Expected 2 deleted, got %d", result.Deleted)
	}
}

func TestDoDelNoMatch(t *testing.T) {
	server, client := newTestClient(t)
	defer server.Close()

	var progress bytes.Buffer
	if _, err := DoDel(context.Background(), client, "default", DeleteIdentifier{ID: "missing"}, false, false, &progress); err == nil {
		t.Fatal("Expected an error when nothing matches")
	}
}
