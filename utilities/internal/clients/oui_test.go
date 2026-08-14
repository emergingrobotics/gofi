package clients

import (
	"strings"
	"testing"
)

const sampleOUIData = `

OUI/MA-L			Organization
company_id			Organization
				Address


00-00-00   (hex)		XEROX CORPORATION
000000     (base 16)		XEROX CORPORATION
				M/S 105-50C
				800 Phillips Road
				Webster NY 14580
				US

00-00-01   (hex)		XEROX CORPORATION
000001     (base 16)		XEROX CORPORATION
				M/S 105-50C
				800 Phillips Road
				Webster NY 14580
				US

AA-BB-CC   (hex)		Acme Corporation
AABBCC     (base 16)		Acme Corporation
				123 Main Street
				Springfield IL 12345
				US

11-22-33   (hex)		Dell Inc.
112233     (base 16)		Dell Inc.
				One Dell Way
				Round Rock TX 78682
				US
`

func TestParseOUIDatabase_ValidData(t *testing.T) {
	database, err := ParseOUIDatabase(strings.NewReader(sampleOUIData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(database.entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(database.entries))
	}

	if database.entries["00:00:00"] != "XEROX CORPORATION" {
		t.Errorf("expected XEROX CORPORATION, got %q", database.entries["00:00:00"])
	}
	if database.entries["aa:bb:cc"] != "Acme Corporation" {
		t.Errorf("expected Acme Corporation, got %q", database.entries["aa:bb:cc"])
	}
	if database.entries["11:22:33"] != "Dell Inc." {
		t.Errorf("expected Dell Inc., got %q", database.entries["11:22:33"])
	}
}

func TestParseOUIDatabase_EmptyInput(t *testing.T) {
	_, err := ParseOUIDatabase(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseOUIDatabase_NoHexEntries(t *testing.T) {
	input := "This is just some random text\nwithout any OUI entries\n"
	_, err := ParseOUIDatabase(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for input with no hex entries")
	}
}

func TestLookup_KnownMAC(t *testing.T) {
	database, err := ParseOUIDatabase(strings.NewReader(sampleOUIData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manufacturer := database.Lookup("aa:bb:cc:dd:ee:ff")
	if manufacturer != "Acme Corporation" {
		t.Errorf("expected Acme Corporation, got %q", manufacturer)
	}
}

func TestLookup_UnknownMAC(t *testing.T) {
	database, err := ParseOUIDatabase(strings.NewReader(sampleOUIData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manufacturer := database.Lookup("ff:ff:ff:dd:ee:ff")
	if manufacturer != "unknown" {
		t.Errorf("expected unknown, got %q", manufacturer)
	}
}

func TestLookup_CaseInsensitive(t *testing.T) {
	database, err := ParseOUIDatabase(strings.NewReader(sampleOUIData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manufacturer := database.Lookup("AA:BB:CC:DD:EE:FF")
	if manufacturer != "Acme Corporation" {
		t.Errorf("expected Acme Corporation, got %q", manufacturer)
	}
}

func TestLookup_HyphenSeparatedMAC(t *testing.T) {
	database, err := ParseOUIDatabase(strings.NewReader(sampleOUIData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manufacturer := database.Lookup("aa-bb-cc-dd-ee-ff")
	if manufacturer != "Acme Corporation" {
		t.Errorf("expected Acme Corporation for hyphen-separated MAC, got %q", manufacturer)
	}
}

func TestLookup_ShortMAC(t *testing.T) {
	database, err := ParseOUIDatabase(strings.NewReader(sampleOUIData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manufacturer := database.Lookup("aa:bb")
	if manufacturer != "unknown" {
		t.Errorf("expected unknown for short MAC, got %q", manufacturer)
	}
}

func TestLookup_EmptyMAC(t *testing.T) {
	database, err := ParseOUIDatabase(strings.NewReader(sampleOUIData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manufacturer := database.Lookup("")
	if manufacturer != "unknown" {
		t.Errorf("expected unknown for empty MAC, got %q", manufacturer)
	}
}

func TestOUIDatabasePath_Default(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	path, err := ouiDatabasePath()
	if err != nil {
		t.Fatalf("ouiDatabasePath: %v", err)
	}
	if !strings.HasSuffix(path, ".local/share/gofi/oui.txt") {
		t.Errorf("unexpected default path: %s", path)
	}
}

func TestOUIDatabasePath_XDGOverride(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/test-xdg")
	path, err := ouiDatabasePath()
	if err != nil {
		t.Fatalf("ouiDatabasePath: %v", err)
	}
	if path != "/tmp/test-xdg/gofi/oui.txt" {
		t.Errorf("expected /tmp/test-xdg/gofi/oui.txt, got %s", path)
	}
}
