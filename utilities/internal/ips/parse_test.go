package ips

import (
	"strings"
	"testing"
)

func TestParse_SingleEntry(t *testing.T) {
	input := `host myserver {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 192.168.1.10;
}`
	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	e := result.Entries[0]
	if e.Hostname != "myserver" {
		t.Errorf("hostname = %q, want %q", e.Hostname, "myserver")
	}
	if e.MAC != "aa:bb:cc:dd:ee:01" {
		t.Errorf("MAC = %q, want %q", e.MAC, "aa:bb:cc:dd:ee:01")
	}
	if e.IP != "192.168.1.10" {
		t.Errorf("IP = %q, want %q", e.IP, "192.168.1.10")
	}
}

func TestParse_MultipleEntries(t *testing.T) {
	input := `host server1 {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 192.168.1.10;
}

host server2 {
    hardware ethernet aa:bb:cc:dd:ee:02;
    fixed-address 192.168.1.11;
}

host server3 {
    hardware ethernet aa:bb:cc:dd:ee:03;
    fixed-address 192.168.1.12;
}`
	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Entries))
	}
	hostnames := []string{"server1", "server2", "server3"}
	for i, want := range hostnames {
		if result.Entries[i].Hostname != want {
			t.Errorf("entry %d hostname = %q, want %q", i, result.Entries[i].Hostname, want)
		}
	}
}

func TestParse_Comments(t *testing.T) {
	input := `# This is a comment
host myhost {
    # Another comment
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address 10.0.0.1;
}
# Trailing comment`
	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
}

func TestParse_BlankLines(t *testing.T) {
	input := `

host h1 {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 10.0.0.1;
}


host h2 {
    hardware ethernet aa:bb:cc:dd:ee:02;
    fixed-address 10.0.0.2;
}

`
	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
}

func TestParse_FlexibleWhitespace(t *testing.T) {
	input := `host myhost {
		hardware ethernet   aa:bb:cc:dd:ee:ff;
		fixed-address   192.168.1.1;
}`
	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	if result.Entries[0].MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("MAC = %q, want %q", result.Entries[0].MAC, "aa:bb:cc:dd:ee:ff")
	}
}

func TestParse_UppercaseMAC(t *testing.T) {
	input := `host myhost {
    hardware ethernet AA:BB:CC:DD:EE:FF;
    fixed-address 10.0.0.1;
}`
	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Entries[0].MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("MAC = %q, want lowercase %q", result.Entries[0].MAC, "aa:bb:cc:dd:ee:ff")
	}
}

func TestParse_NonHostDirectives(t *testing.T) {
	input := `subnet 192.168.1.0 netmask 255.255.255.0 {
    range 192.168.1.100 192.168.1.200;
    option routers 192.168.1.1;
}

host myhost {
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address 192.168.1.50;
}

option domain-name "example.com";`
	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	if result.Entries[0].Hostname != "myhost" {
		t.Errorf("hostname = %q, want %q", result.Entries[0].Hostname, "myhost")
	}
}

func TestParse_ErrorInvalidMAC(t *testing.T) {
	input := `host bad {
    hardware ethernet zz:zz:zz:zz:zz:zz;
    fixed-address 10.0.0.1;
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for invalid MAC")
	}
	if !strings.Contains(err.Error(), "invalid MAC") {
		t.Errorf("error = %q, want mention of invalid MAC", err.Error())
	}
}

func TestParse_ErrorInvalidIP(t *testing.T) {
	input := `host bad {
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address 999.999.999.999;
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
	if !strings.Contains(err.Error(), "invalid IP") {
		t.Errorf("error = %q, want mention of invalid IP", err.Error())
	}
}

func TestParse_ErrorIPv6(t *testing.T) {
	input := `host bad {
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address ::1;
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for IPv6")
	}
	if !strings.Contains(err.Error(), "IPv4") {
		t.Errorf("error = %q, want mention of IPv4", err.Error())
	}
}

func TestParse_ErrorMissingSemicolon(t *testing.T) {
	input := `host bad {
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address 192.168.1.1
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for missing semicolon")
	}
	if !strings.Contains(err.Error(), "semicolon") {
		t.Errorf("error = %q, want mention of semicolon", err.Error())
	}
}

func TestParse_ErrorMissingMAC(t *testing.T) {
	input := `host bad {
    fixed-address 192.168.1.1;
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for missing MAC")
	}
	if !strings.Contains(err.Error(), "hardware ethernet") {
		t.Errorf("error = %q, want mention of hardware ethernet", err.Error())
	}
}

func TestParse_ErrorMissingIP(t *testing.T) {
	input := `host bad {
    hardware ethernet aa:bb:cc:dd:ee:ff;
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for missing IP")
	}
	if !strings.Contains(err.Error(), "fixed-address") {
		t.Errorf("error = %q, want mention of fixed-address", err.Error())
	}
}

func TestParse_ErrorDuplicateHostnames(t *testing.T) {
	input := `host myhost {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 10.0.0.1;
}

host myhost {
    hardware ethernet aa:bb:cc:dd:ee:02;
    fixed-address 10.0.0.2;
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for duplicate hostname")
	}
	if !strings.Contains(err.Error(), "duplicate hostname") {
		t.Errorf("error = %q, want mention of duplicate hostname", err.Error())
	}
}

func TestParse_ErrorDuplicateMACs(t *testing.T) {
	input := `host host1 {
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address 10.0.0.1;
}

host host2 {
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address 10.0.0.2;
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for duplicate MAC")
	}
	if !strings.Contains(err.Error(), "duplicate MAC") {
		t.Errorf("error = %q, want mention of duplicate MAC", err.Error())
	}
}

func TestParse_ErrorDuplicateIPs(t *testing.T) {
	input := `host host1 {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 10.0.0.1;
}

host host2 {
    hardware ethernet aa:bb:cc:dd:ee:02;
    fixed-address 10.0.0.1;
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for duplicate IP")
	}
	if !strings.Contains(err.Error(), "duplicate IP") {
		t.Errorf("error = %q, want mention of duplicate IP", err.Error())
	}
}

func TestParse_ErrorInvalidHostname(t *testing.T) {
	input := `host "my host" {
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address 10.0.0.1;
}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for invalid hostname")
	}
}

func TestParse_EmptyInput(t *testing.T) {
	result, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}

func TestParse_CommentsOnly(t *testing.T) {
	input := `# Just comments
# Nothing else
# here`
	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}

func TestParseSingle_Valid(t *testing.T) {
	input := `host mydevice {
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address 192.168.1.50;
}`
	entry, err := ParseSingle(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Hostname != "mydevice" {
		t.Errorf("hostname = %q, want %q", entry.Hostname, "mydevice")
	}
}

func TestParseSingle_ErrorMultiple(t *testing.T) {
	input := `host h1 {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 10.0.0.1;
}
host h2 {
    hardware ethernet aa:bb:cc:dd:ee:02;
    fixed-address 10.0.0.2;
}`
	_, err := ParseSingle(input)
	if err == nil {
		t.Fatal("expected error for multiple declarations")
	}
}

func TestParseSingle_ErrorEmpty(t *testing.T) {
	_, err := ParseSingle("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestIsDNSSafe(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"myhost", true},
		{"my-host", true},
		{"my.host.local", true},
		{"host_1", true},
		{"A", true},
		{"a1b2c3", true},
		{"", false},
		{"-host", false},
		{"host-", false},
		{".host", false},
		{"host.", false},
		{"ho st", false},
		{"host!", false},
		{"host@name", false},
		{strings.Repeat("a", 64), false},  // single label > 63 chars
		{strings.Repeat("a", 63), true},   // single label = 63 chars
		{strings.Repeat("a", 50) + "." + strings.Repeat("b", 50) + "." + strings.Repeat("c", 50) + "." + strings.Repeat("d", 50) + "." + strings.Repeat("e", 50), false}, // > 253 total
	}
	for _, tt := range tests {
		got := isDNSSafe(tt.input)
		if got != tt.want {
			t.Errorf("isDNSSafe(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
