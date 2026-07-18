package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestReportProbe_PresentReturnsZero(t *testing.T) {
	entry := &ClientEntry{MAC: "aa:bb:cc:dd:ee:01", IP: "192.168.1.10", Hostname: "here", Manufacturer: "Acme", Status: statusPresent}

	var out, errOut bytes.Buffer
	code, err := reportProbe(&out, &errOut, entry, entry.MAC, false, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit 0 for present device, got %d", code)
	}
	if !strings.Contains(out.String(), "here") {
		t.Errorf("expected device row in output: %s", out.String())
	}
}

func TestReportProbe_GoneReturnsOne(t *testing.T) {
	entry := &ClientEntry{MAC: "aa:bb:cc:dd:ee:02", Hostname: "left", Manufacturer: "Acme", Status: statusGone}

	var out, errOut bytes.Buffer
	code, err := reportProbe(&out, &errOut, entry, entry.MAC, false, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 1 {
		t.Errorf("expected exit 1 for gone device, got %d", code)
	}
}

func TestReportProbe_NotFoundReturnsOne(t *testing.T) {
	var out, errOut bytes.Buffer
	code, err := reportProbe(&out, &errOut, nil, "aa:bb:cc:dd:ee:03", false, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 1 {
		t.Errorf("expected exit 1 for missing device, got %d", code)
	}
	if !strings.Contains(errOut.String(), "not found") {
		t.Errorf("expected 'not found' message on stderr: %s", errOut.String())
	}
}

func TestReportProbe_NotFoundJSONEmptyArray(t *testing.T) {
	var out, errOut bytes.Buffer
	code, err := reportProbe(&out, &errOut, nil, "aa:bb:cc:dd:ee:03", true, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if strings.TrimSpace(out.String()) != "[]" {
		t.Errorf("expected empty JSON array, got %q", out.String())
	}
}
