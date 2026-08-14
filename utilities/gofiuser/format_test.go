package main

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
	entries := []UserEntry{
		{ID: "u1", Name: "tapo1", Hostname: "plug", MAC: "aa:bb:cc:00:00:01", FixedIP: "192.168.1.10", UseFixedIP: true},
		{ID: "u2", Name: "phone", MAC: "aa:bb:cc:00:00:02"},
	}

	var out bytes.Buffer
	if err := FormatText(&out, entries); err != nil {
		t.Fatalf("FormatText failed: %v", err)
	}

	text := out.String()
	for _, want := range []string{"NAME", "HOSTNAME", "MAC", "FIXED-IP", "FLAGS", "tapo1", "192.168.1.10"} {
		if !strings.Contains(text, want) {
			t.Errorf("Output missing %q:\n%s", want, text)
		}
	}
}

func TestFormatTextMarksBlocked(t *testing.T) {
	entries := []UserEntry{{ID: "u1", Name: "bad", MAC: "aa:bb:cc:00:00:01", Blocked: true}}

	var out bytes.Buffer
	if err := FormatText(&out, entries); err != nil {
		t.Fatalf("FormatText failed: %v", err)
	}

	if !strings.Contains(out.String(), blockedMark) {
		t.Errorf("Blocked entry not marked:\n%s", out.String())
	}
}

func TestFormatJSONEmitsArray(t *testing.T) {
	entries := []UserEntry{{ID: "u1", Name: "tapo1", MAC: "aa:bb:cc:00:00:01"}}

	var out bytes.Buffer
	if err := FormatJSON(&out, entries); err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	var decoded []UserEntry
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("Output was not valid JSON: %v (%s)", err, out.String())
	}
	if len(decoded) != 1 || decoded[0].ID != "u1" {
		t.Errorf("Round trip lost data: %+v", decoded)
	}
}

func TestFormatJSONNilEncodesAsEmptyArray(t *testing.T) {
	var out bytes.Buffer
	if err := FormatJSON(&out, nil); err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	if strings.TrimSpace(out.String()) != "[]" {
		t.Errorf("Expected [], got %q", out.String())
	}
}

// A fixed_ip left behind on a record with use_fixedip false is not a live
// reservation and must not read as one.
func TestFixedIPDisplay(t *testing.T) {
	tests := []struct {
		name  string
		entry UserEntry
		want  string
	}{
		{"active reservation", UserEntry{FixedIP: "192.168.1.10", UseFixedIP: true}, "192.168.1.10"},
		{"stale fixed ip", UserEntry{FixedIP: "192.168.1.10"}, "192.168.1.10 " + dynamicMark},
		{"none", UserEntry{}, placeholder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixedIPDisplay(tt.entry); got != tt.want {
				t.Errorf("fixedIPDisplay = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFlagsDisplay(t *testing.T) {
	if got := flagsDisplay(UserEntry{Blocked: true}); got != blockedMark {
		t.Errorf("Blocked flags = %q, want %q", got, blockedMark)
	}
	if got := flagsDisplay(UserEntry{}); got != placeholder {
		t.Errorf("Unblocked flags = %q, want %q", got, placeholder)
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
