package ips

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormat_SingleEntry(t *testing.T) {
	entries := []HostEntry{
		{Hostname: "myserver", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
	}
	var buf bytes.Buffer
	err := Format(&buf, entries, FormatOptions{Host: "192.168.1.1", Date: "2026-02-23"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "host myserver {") {
		t.Error("missing host declaration")
	}
	if !strings.Contains(output, "    hardware ethernet aa:bb:cc:dd:ee:01;") {
		t.Error("missing or wrong hardware ethernet line")
	}
	if !strings.Contains(output, "    fixed-address 192.168.1.10;") {
		t.Error("missing or wrong fixed-address line")
	}
	if !strings.Contains(output, "}") {
		t.Error("missing closing brace")
	}
}

func TestFormat_MultipleEntries(t *testing.T) {
	entries := []HostEntry{
		{Hostname: "c", MAC: "aa:bb:cc:dd:ee:03", IP: "192.168.1.30"},
		{Hostname: "a", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
		{Hostname: "b", MAC: "aa:bb:cc:dd:ee:02", IP: "192.168.1.20"},
	}
	var buf bytes.Buffer
	err := Format(&buf, entries, FormatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	// Verify sorted order
	idxA := strings.Index(output, "host a {")
	idxB := strings.Index(output, "host b {")
	idxC := strings.Index(output, "host c {")
	if idxA < 0 || idxB < 0 || idxC < 0 {
		t.Fatal("missing host declarations in output")
	}
	if !(idxA < idxB && idxB < idxC) {
		t.Errorf("entries not sorted by IP: a@%d, b@%d, c@%d", idxA, idxB, idxC)
	}
}

func TestFormat_IPSortOrder(t *testing.T) {
	entries := []HostEntry{
		{Hostname: "h100", MAC: "aa:bb:cc:dd:ee:03", IP: "192.168.1.100"},
		{Hostname: "h2", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.2"},
		{Hostname: "h10", MAC: "aa:bb:cc:dd:ee:02", IP: "192.168.1.10"},
	}
	var buf bytes.Buffer
	err := Format(&buf, entries, FormatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	idx2 := strings.Index(output, "host h2 {")
	idx10 := strings.Index(output, "host h10 {")
	idx100 := strings.Index(output, "host h100 {")
	if !(idx2 < idx10 && idx10 < idx100) {
		t.Errorf("IP sort order wrong: h2@%d, h10@%d, h100@%d", idx2, idx10, idx100)
	}
}

func TestFormat_EmptyEntries(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, nil, FormatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "No fixed IP assignments found") {
		t.Error("expected empty-input message")
	}
	if !strings.Contains(output, "# host mydevice {") {
		t.Error("expected example in comments")
	}
}

func TestFormat_HeaderComments(t *testing.T) {
	entries := []HostEntry{
		{Hostname: "h", MAC: "aa:bb:cc:dd:ee:ff", IP: "10.0.0.1"},
	}
	var buf bytes.Buffer
	err := Format(&buf, entries, FormatOptions{Host: "10.0.0.254", Date: "2026-01-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "# exported from UDM at 10.0.0.254") {
		t.Error("missing host in header")
	}
	if !strings.Contains(output, "# date: 2026-01-01") {
		t.Error("missing date in header")
	}
}

func TestFormat_LowercaseMAC(t *testing.T) {
	entries := []HostEntry{
		{Hostname: "h", MAC: "aa:bb:cc:dd:ee:ff", IP: "10.0.0.1"},
	}
	var buf bytes.Buffer
	err := Format(&buf, entries, FormatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "aa:bb:cc:dd:ee:ff") {
		t.Error("MAC not lowercase in output")
	}
}

func TestFormat_RoundTrip(t *testing.T) {
	entries := []HostEntry{
		{Hostname: "server1", MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10"},
		{Hostname: "server2", MAC: "aa:bb:cc:dd:ee:02", IP: "192.168.1.20"},
	}
	var buf bytes.Buffer
	err := Format(&buf, entries, FormatOptions{})
	if err != nil {
		t.Fatalf("format error: %v", err)
	}

	result, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("round-trip: expected 2 entries, got %d", len(result.Entries))
	}
	for i, e := range result.Entries {
		if e.Hostname != entries[i].Hostname || e.MAC != entries[i].MAC || e.IP != entries[i].IP {
			t.Errorf("round-trip mismatch at %d: got %+v, want %+v", i, e, entries[i])
		}
	}
}
