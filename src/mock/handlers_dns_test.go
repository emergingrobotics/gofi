package mock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/src/types"
)

func dnsPath(server *Server, suffix string) string {
	return server.URL() + "/proxy/network/v2/api/site/default/static-dns" + suffix
}

func seedDNSRecord(server *Server, id, key, value string) *types.DNSRecord {
	record := &types.DNSRecord{
		ID:         id,
		Key:        key,
		Value:      value,
		RecordType: types.DNSRecordTypeA,
		Enabled:    true,
	}
	server.state.AddDNSRecord(record)
	return record
}

func doDNSRequest(t *testing.T, method, url string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to encode body: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("Failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := testClientHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	return resp, payload
}

func TestHandleListDNSRecords(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	seedDNSRecord(server, "dns1", "host1.example.com", "192.168.1.10")
	seedDNSRecord(server, "dns2", "host2.example.com", "192.168.1.11")

	resp, payload := doDNSRequest(t, "GET", dnsPath(server, ""), nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// The v2 surface returns a bare array, not a meta/data envelope.
	var records []types.DNSRecord
	if err := json.Unmarshal(payload, &records); err != nil {
		t.Fatalf("Response was not a bare array: %v (%s)", err, payload)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}
}

func TestHandleGetDNSRecordReturnsMethodNotAllowed(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	seedDNSRecord(server, "dns1", "host1.example.com", "192.168.1.10")

	resp, _ := doDNSRequest(t, "GET", dnsPath(server, "/dns1"), nil)

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status 405 for GET by id, got %d", resp.StatusCode)
	}
}

func TestHandleCreateDNSRecord(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	resp, payload := doDNSRequest(t, "POST", dnsPath(server, ""), map[string]interface{}{
		"key":   "new.example.com",
		"value": "192.168.1.50",
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var created types.DNSRecord
	if err := json.Unmarshal(payload, &created); err != nil {
		t.Fatalf("Failed to decode created record: %v (%s)", err, payload)
	}

	if created.ID == "" {
		t.Error("Expected an ID to be generated")
	}
	if created.RecordType != types.DNSRecordTypeA {
		t.Errorf("Expected default record type A, got %q", created.RecordType)
	}
	if server.state.GetDNSRecord(created.ID) == nil {
		t.Error("Created record was not stored in state")
	}
}

func TestHandleCreateDNSRecordValidation(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{name: "missing key", body: map[string]interface{}{"value": "192.168.1.50"}},
		{name: "missing value", body: map[string]interface{}{"key": "new.example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := doDNSRequest(t, "POST", dnsPath(server, ""), tt.body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("Expected status 400, got %d", resp.StatusCode)
			}
		})
	}
}

func TestHandleUpdateDNSRecord(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	seedDNSRecord(server, "dns1", "host1.example.com", "192.168.1.10")

	resp, payload := doDNSRequest(t, "PUT", dnsPath(server, "/dns1"), map[string]interface{}{
		"key":         "host1.example.com",
		"value":       "192.168.1.99",
		"record_type": types.DNSRecordTypeA,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated types.DNSRecord
	if err := json.Unmarshal(payload, &updated); err != nil {
		t.Fatalf("Failed to decode updated record: %v (%s)", err, payload)
	}

	if updated.ID != "dns1" {
		t.Errorf("Expected ID to be preserved, got %q", updated.ID)
	}

	stored := server.state.GetDNSRecord("dns1")
	if stored == nil {
		t.Fatal("Record vanished from state")
	}
	if stored.Value != "192.168.1.99" {
		t.Errorf("Expected value 192.168.1.99, got %q", stored.Value)
	}
}

func TestHandleUpdateDNSRecordNotFound(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	resp, _ := doDNSRequest(t, "PUT", dnsPath(server, "/missing"), map[string]interface{}{
		"key":   "x.example.com",
		"value": "192.168.1.1",
	})

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleDeleteDNSRecord(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	seedDNSRecord(server, "dns1", "host1.example.com", "192.168.1.10")

	resp, _ := doDNSRequest(t, "DELETE", dnsPath(server, "/dns1"), nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	if server.state.GetDNSRecord("dns1") != nil {
		t.Error("Record was not deleted from state")
	}
}

func TestHandleDeleteDNSRecordNotFound(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	resp, _ := doDNSRequest(t, "DELETE", dnsPath(server, "/missing"), nil)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleDNSRecordsRequiresIDForMutations(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	for _, method := range []string{"PUT", "DELETE"} {
		t.Run(method+" without id", func(t *testing.T) {
			resp, _ := doDNSRequest(t, method, dnsPath(server, ""), map[string]interface{}{
				"key":   "x.example.com",
				"value": "192.168.1.1",
			})
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("Expected status 400, got %d", resp.StatusCode)
			}
		})
	}
}

func TestDNSRecordStateAccessors(t *testing.T) {
	state := NewState()

	record := &types.DNSRecord{ID: "dns1", Key: "a.example.com", Value: "192.168.1.1"}
	state.AddDNSRecord(record)

	if got := state.GetDNSRecord("dns1"); got == nil || got.Key != "a.example.com" {
		t.Fatalf("GetDNSRecord returned %+v", got)
	}

	if got := state.ListDNSRecords(); len(got) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(got))
	}

	state.UpdateDNSRecord(&types.DNSRecord{ID: "dns1", Key: "a.example.com", Value: "192.168.1.2"})
	if got := state.GetDNSRecord("dns1"); got.Value != "192.168.1.2" {
		t.Errorf("Expected updated value, got %q", got.Value)
	}

	state.DeleteDNSRecord("dns1")
	if got := state.GetDNSRecord("dns1"); got != nil {
		t.Error("Expected record to be deleted")
	}

	state.Reset()
	if got := state.ListDNSRecords(); len(got) != 0 {
		t.Errorf("Expected Reset to clear DNS records, got %d", len(got))
	}
}
