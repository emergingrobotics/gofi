package dns

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatTextEmptyWritesNothing(t *testing.T) {
	var out bytes.Buffer
	if err := FormatText(&out, nil); err != nil {
		t.Fatalf("FormatText failed: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("Expected no output, got %q", out.String())
	}
}

func TestFormatTextIncludesHeaderAndRows(t *testing.T) {
	entries := []DNSEntry{
		{ID: "dns1", Key: "a.example.com", Value: "192.168.1.1", Type: "A", TTL: 300, Enabled: true},
		{ID: "dns2", Key: "b.example.com", Value: "192.168.1.2", Type: "A", Enabled: true},
	}

	var out bytes.Buffer
	if err := FormatText(&out, entries); err != nil {
		t.Fatalf("FormatText failed: %v", err)
	}

	text := out.String()
	for _, want := range []string{"NAME", "VALUE", "TYPE", "TTL", "ID", "a.example.com", "192.168.1.2", "300"} {
		if !strings.Contains(text, want) {
			t.Errorf("Output missing %q:\n%s", want, text)
		}
	}
}

// A disabled record must not read like an active one in the name column.
func TestFormatTextMarksDisabledRecords(t *testing.T) {
	entries := []DNSEntry{{ID: "dns1", Key: "off.example.com", Value: "192.168.1.1", Enabled: false}}

	var out bytes.Buffer
	if err := FormatText(&out, entries); err != nil {
		t.Fatalf("FormatText failed: %v", err)
	}

	if !strings.Contains(out.String(), disabledMark) {
		t.Errorf("Disabled record not marked:\n%s", out.String())
	}
}

func TestFormatJSONEmitsArray(t *testing.T) {
	entries := []DNSEntry{{ID: "dns1", Key: "a.example.com", Value: "192.168.1.1", Type: "A", Enabled: true}}

	var out bytes.Buffer
	if err := FormatJSON(&out, entries); err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	var decoded []DNSEntry
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("Output was not valid JSON: %v (%s)", err, out.String())
	}
	if len(decoded) != 1 || decoded[0].ID != "dns1" {
		t.Errorf("Round trip lost data: %+v", decoded)
	}
}

// Nil must encode as [] rather than null so consumers can iterate blindly.
func TestFormatJSONNilEncodesAsEmptyArray(t *testing.T) {
	var out bytes.Buffer
	if err := FormatJSON(&out, nil); err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	if strings.TrimSpace(out.String()) != "[]" {
		t.Errorf("Expected [], got %q", out.String())
	}
}

func TestNameDisplay(t *testing.T) {
	if got := nameDisplay(DNSEntry{Key: "a.example.com", Enabled: true}); got != "a.example.com" {
		t.Errorf("Enabled name = %q", got)
	}
	if got := nameDisplay(DNSEntry{Key: "a.example.com"}); got != "a.example.com "+disabledMark {
		t.Errorf("Disabled name = %q", got)
	}
	if got := nameDisplay(DNSEntry{Enabled: true}); got != placeholder {
		t.Errorf("Empty name = %q, want %q", got, placeholder)
	}
}

func TestTTLDisplay(t *testing.T) {
	tests := []struct {
		ttl  int
		want string
	}{
		{300, "300"},
		{0, placeholder},
		{-1, placeholder},
	}

	for _, tt := range tests {
		if got := ttlDisplay(tt.ttl); got != tt.want {
			t.Errorf("ttlDisplay(%d) = %q, want %q", tt.ttl, got, tt.want)
		}
	}
}

func TestOrPlaceholder(t *testing.T) {
	if got := orPlaceholder("value"); got != "value" {
		t.Errorf("orPlaceholder(value) = %q", got)
	}
	if got := orPlaceholder(""); got != placeholder {
		t.Errorf("orPlaceholder(empty) = %q, want %q", got, placeholder)
	}
}
