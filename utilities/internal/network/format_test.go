package network

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatText_HeaderAndPool(t *testing.T) {
	entries := []NetworkEntry{
		{Name: "LAN", Subnet: "192.168.4.0/24", DHCPEnabled: true, DHCPStart: "192.168.4.6", DHCPStop: "192.168.4.99", DHCPLease: 86400, Gateway: "192.168.4.1", DNS: []string{"192.168.4.1"}},
	}
	var buffer bytes.Buffer
	if err := FormatText(&buffer, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buffer.String()
	for _, column := range []string{"NETWORK", "SUBNET", "DHCP-POOL", "LEASE", "GATEWAY", "DNS"} {
		if !strings.Contains(output, column) {
			t.Errorf("missing header column %s: %s", column, output)
		}
	}
	if !strings.Contains(output, "192.168.4.6 - 192.168.4.99") {
		t.Errorf("expected pool range in output: %s", output)
	}
	if !strings.Contains(output, "86400s") {
		t.Errorf("expected lease in output: %s", output)
	}
}

func TestFormatText_DHCPDisabled(t *testing.T) {
	entries := []NetworkEntry{
		{Name: "IoT", Subnet: "192.168.30.0/24", DHCPEnabled: false},
	}
	var buffer bytes.Buffer
	if err := FormatText(&buffer, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.Split(strings.TrimSpace(buffer.String()), "\n")[1]
	if !strings.Contains(line, "(disabled)") {
		t.Errorf("expected (disabled) marker for DHCP-off network: %s", line)
	}
}

func TestFormatText_Empty(t *testing.T) {
	var buffer bytes.Buffer
	if err := FormatText(&buffer, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buffer.String() != "" {
		t.Errorf("expected empty output, got %q", buffer.String())
	}
}

func TestFormatJSON_Basic(t *testing.T) {
	entries := []NetworkEntry{
		{Name: "LAN", Subnet: "192.168.4.0/24", DHCPEnabled: true, DHCPStart: "192.168.4.6", DHCPStop: "192.168.4.99"},
	}
	var buffer bytes.Buffer
	if err := FormatJSON(&buffer, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []NetworkEntry
	if err := json.Unmarshal(buffer.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(parsed) != 1 || parsed[0].DHCPStart != "192.168.4.6" {
		t.Errorf("unexpected parsed entry: %+v", parsed)
	}
}

func TestFormatJSON_Empty(t *testing.T) {
	var buffer bytes.Buffer
	if err := FormatJSON(&buffer, []NetworkEntry{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(buffer.String()) != "[]" {
		t.Errorf("expected empty JSON array, got %q", buffer.String())
	}
}

func TestVLANDisplay(t *testing.T) {
	if vlanDisplay(0) != placeholder {
		t.Errorf("expected placeholder for VLAN 0, got %q", vlanDisplay(0))
	}
	if vlanDisplay(20) != "20" {
		t.Errorf("expected 20, got %q", vlanDisplay(20))
	}
}
