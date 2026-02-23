package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatText_BasicOutput(t *testing.T) {
	entries := []ClientEntry{
		{MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10", Hostname: "myserver", Manufacturer: "Dell Inc."},
		{MAC: "aa:bb:cc:dd:ee:02", IP: "192.168.1.11", Hostname: "printer", Manufacturer: "Hewlett Packard"},
	}

	var buffer bytes.Buffer
	if err := FormatText(&buffer, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buffer.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 data), got %d", len(lines))
	}

	if !strings.Contains(lines[0], "MAC") || !strings.Contains(lines[0], "OUI-MANUFACTURER") {
		t.Errorf("expected header line, got: %s", lines[0])
	}
	if !strings.Contains(lines[1], "aa:bb:cc:dd:ee:01") {
		t.Errorf("expected MAC in first data line: %s", lines[1])
	}
	if !strings.Contains(lines[1], "192.168.1.10") {
		t.Errorf("expected IP in first data line: %s", lines[1])
	}
	if !strings.Contains(lines[1], "myserver") {
		t.Errorf("expected hostname in first data line: %s", lines[1])
	}
	if !strings.Contains(lines[1], "Dell Inc.") {
		t.Errorf("expected manufacturer in first data line: %s", lines[1])
	}
}

func TestFormatText_NoIPShowsDash(t *testing.T) {
	entries := []ClientEntry{
		{MAC: "aa:bb:cc:dd:ee:01", Hostname: "no-ip-host", Manufacturer: "unknown"},
	}

	var buffer bytes.Buffer
	if err := FormatText(&buffer, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + data), got %d", len(lines))
	}
	if !strings.Contains(lines[1], "\t-\t") {
		t.Errorf("expected dash for missing IP: %s", lines[1])
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

func TestFormatJSON_BasicOutput(t *testing.T) {
	entries := []ClientEntry{
		{
			MAC:          "aa:bb:cc:dd:ee:01",
			IP:           "192.168.1.10",
			Hostname:     "myserver",
			Manufacturer: "Dell Inc.",
			IsWired:      false,
			ESSID:        "MyNetwork",
			Channel:      36,
		},
	}

	var buffer bytes.Buffer
	if err := FormatJSON(&buffer, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []ClientEntry
	if err := json.Unmarshal(buffer.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed))
	}
	if parsed[0].MAC != "aa:bb:cc:dd:ee:01" {
		t.Errorf("expected MAC aa:bb:cc:dd:ee:01, got %s", parsed[0].MAC)
	}
	if parsed[0].Manufacturer != "Dell Inc." {
		t.Errorf("expected Dell Inc., got %s", parsed[0].Manufacturer)
	}
	if parsed[0].ESSID != "MyNetwork" {
		t.Errorf("expected ESSID MyNetwork, got %s", parsed[0].ESSID)
	}
}

func TestFormatJSON_WiredFields(t *testing.T) {
	entries := []ClientEntry{
		{
			MAC:          "aa:bb:cc:dd:ee:01",
			IP:           "192.168.1.10",
			Hostname:     "switch-host",
			Manufacturer: "Dell Inc.",
			IsWired:      true,
			SwitchMAC:        "11:22:33:44:55:66",
			SwitchPort:       4,
		},
	}

	var buffer bytes.Buffer
	if err := FormatJSON(&buffer, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []ClientEntry
	if err := json.Unmarshal(buffer.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if parsed[0].SwitchMAC != "11:22:33:44:55:66" {
		t.Errorf("expected switch MAC, got %s", parsed[0].SwitchMAC)
	}
	if parsed[0].SwitchPort != 4 {
		t.Errorf("expected switch port 4, got %d", parsed[0].SwitchPort)
	}
}

func TestFormatJSON_Empty(t *testing.T) {
	var buffer bytes.Buffer
	if err := FormatJSON(&buffer, []ClientEntry{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buffer.String())
	if output != "[]" {
		t.Errorf("expected empty JSON array, got %q", output)
	}
}

func TestFormatJSON_OmitsEmptyFields(t *testing.T) {
	entries := []ClientEntry{
		{
			MAC:          "aa:bb:cc:dd:ee:01",
			IP:           "192.168.1.10",
			Hostname:     "test",
			Manufacturer: "Acme",
			IsWired:      true,
		},
	}

	var buffer bytes.Buffer
	if err := FormatJSON(&buffer, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buffer.String()
	if strings.Contains(output, "essid") {
		t.Errorf("expected essid to be omitted for wired client: %s", output)
	}
	if strings.Contains(output, "ap_mac") {
		t.Errorf("expected ap_mac to be omitted for wired client: %s", output)
	}
}

func TestFormatJSON_AlwaysPresentFields(t *testing.T) {
	entries := []ClientEntry{
		{
			MAC:          "aa:bb:cc:dd:ee:01",
			Hostname:     "test",
			Manufacturer: "Acme",
		},
	}

	var buffer bytes.Buffer
	if err := FormatJSON(&buffer, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buffer.String()
	alwaysPresent := []string{"ip", "rx_bytes", "tx_bytes", "uptime", "last_seen"}
	for _, field := range alwaysPresent {
		if !strings.Contains(output, "\""+field+"\"") {
			t.Errorf("expected always-present field %q in output even when zero: %s", field, output)
		}
	}
}
