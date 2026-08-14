package services

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/unifi-go/gofi/src/mock"
	"github.com/unifi-go/gofi/src/transport"
	"github.com/unifi-go/gofi/src/types"
)

func newTestDNSTransport(url string) (transport.Transport, error) {
	config := transport.DefaultConfig(url)
	config.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return transport.New(config)
}

func newTestDNSService(t *testing.T) (*mock.Server, DNSService) {
	t.Helper()

	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())

	trans, err := newTestDNSTransport(server.URL())
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create transport: %v", err)
	}

	return server, NewDNSService(trans)
}

func seedRecord(server *mock.Server, id, key, value string) {
	server.State().AddDNSRecord(&types.DNSRecord{
		ID:         id,
		Key:        key,
		Value:      value,
		RecordType: types.DNSRecordTypeA,
		Enabled:    true,
	})
}

func TestDNSService_List(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	seedRecord(server, "dns1", "a.example.com", "192.168.1.10")
	seedRecord(server, "dns2", "b.example.com", "192.168.1.11")

	records, err := svc.List(context.Background(), "default")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
}

// Get must not issue GET on an individual record: the controller answers that
// with 405. It resolves through the collection instead.
func TestDNSService_Get(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	seedRecord(server, "dns1", "a.example.com", "192.168.1.10")

	record, err := svc.Get(context.Background(), "default", "dns1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if record.Key != "a.example.com" {
		t.Errorf("Expected key a.example.com, got %q", record.Key)
	}
	if record.Value != "192.168.1.10" {
		t.Errorf("Expected value 192.168.1.10, got %q", record.Value)
	}
}

func TestDNSService_GetNotFound(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	seedRecord(server, "dns1", "a.example.com", "192.168.1.10")

	if _, err := svc.Get(context.Background(), "default", "missing"); err == nil {
		t.Fatal("Expected an error for a missing record")
	}
}

func TestDNSService_GetByName(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	seedRecord(server, "dns1", "a.example.com", "192.168.1.10")
	seedRecord(server, "dns2", "b.example.com", "192.168.1.11")

	record, err := svc.GetByName(context.Background(), "default", "b.example.com")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if record.ID != "dns2" {
		t.Errorf("Expected dns2, got %q", record.ID)
	}
}

func TestDNSService_GetByNameNotFound(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	if _, err := svc.GetByName(context.Background(), "default", "nope.example.com"); err == nil {
		t.Fatal("Expected an error for a missing name")
	}
}

func TestDNSService_GetByIP(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	// Two names on one address is legal and both must come back.
	seedRecord(server, "dns1", "a.example.com", "192.168.1.10")
	seedRecord(server, "dns2", "alias.example.com", "192.168.1.10")
	seedRecord(server, "dns3", "b.example.com", "192.168.1.11")

	records, err := svc.GetByIP(context.Background(), "default", "192.168.1.10")
	if err != nil {
		t.Fatalf("GetByIP failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 records for the shared IP, got %d", len(records))
	}
}

func TestDNSService_GetByIPNoMatches(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	seedRecord(server, "dns1", "a.example.com", "192.168.1.10")

	records, err := svc.GetByIP(context.Background(), "default", "10.0.0.1")
	if err != nil {
		t.Fatalf("GetByIP failed: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected no records, got %d", len(records))
	}
}

func TestDNSService_Create(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	created, err := svc.Create(context.Background(), "default", &types.DNSRecord{
		Key:        "new.example.com",
		Value:      "192.168.1.50",
		RecordType: types.DNSRecordTypeA,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ID == "" {
		t.Error("Expected the created record to carry an ID")
	}
	if server.State().GetDNSRecord(created.ID) == nil {
		t.Error("Created record is absent from server state")
	}
}

func TestDNSService_Update(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	seedRecord(server, "dns1", "a.example.com", "192.168.1.10")

	updated, err := svc.Update(context.Background(), "default", &types.DNSRecord{
		ID:         "dns1",
		Key:        "a.example.com",
		Value:      "192.168.1.99",
		RecordType: types.DNSRecordTypeA,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Value != "192.168.1.99" {
		t.Errorf("Expected value 192.168.1.99, got %q", updated.Value)
	}

	stored := server.State().GetDNSRecord("dns1")
	if stored == nil {
		t.Fatal("Record vanished from state")
	}
	if stored.Value != "192.168.1.99" {
		t.Errorf("State not updated, value is %q", stored.Value)
	}
}

func TestDNSService_Delete(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	seedRecord(server, "dns1", "a.example.com", "192.168.1.10")

	if err := svc.Delete(context.Background(), "default", "dns1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if server.State().GetDNSRecord("dns1") != nil {
		t.Error("Record was not deleted")
	}
}

func TestDNSService_DeleteNotFound(t *testing.T) {
	server, svc := newTestDNSService(t)
	defer server.Close()

	if err := svc.Delete(context.Background(), "default", "missing"); err == nil {
		t.Fatal("Expected an error deleting a missing record")
	}
}
