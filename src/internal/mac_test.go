package internal

import "testing"

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"colon separated", "AA:BB:CC:DD:EE:FF", "aabbccddeeff"},
		{"dash separated", "aa-bb-cc-dd-ee-ff", "aabbccddeeff"},
		{"no separator", "aabbccddeeff", "aabbccddeeff"},
		{"mixed case", "AaBbCcDdEeFf", "aabbccddeeff"},
		{"mixed separators", "AA:BB-CC:DD-EE:FF", "aabbccddeeff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeMAC(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeMAC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatMAC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normalized", "aabbccddeeff", "aa:bb:cc:dd:ee:ff"},
		{"already formatted", "aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:ff"},
		{"uppercase", "AABBCCDDEEFF", "aa:bb:cc:dd:ee:ff"},
		{"dash separated", "aa-bb-cc-dd-ee-ff", "aa:bb:cc:dd:ee:ff"},
		{"invalid length", "aabbcc", "aabbcc"}, // Returns as-is
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatMAC(tt.input)
			if got != tt.want {
				t.Errorf("FormatMAC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateMAC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid colon", "aa:bb:cc:dd:ee:ff", true},
		{"valid dash", "aa-bb-cc-dd-ee-ff", true},
		{"valid no separator", "aabbccddeeff", true},
		{"valid uppercase", "AA:BB:CC:DD:EE:FF", true},
		{"valid mixed", "Aa:Bb:Cc:Dd:Ee:Ff", true},
		{"empty", "", false},
		{"too short", "aa:bb:cc", false},
		{"too long", "aa:bb:cc:dd:ee:ff:11", false},
		{"invalid chars", "zz:bb:cc:dd:ee:ff", false},
		{"missing separator", "aabbccddee:ff", true}, // Mixed is technically valid
		{"wrong format", "aa:bb:cc:dd:ee:", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateMAC(tt.input)
			if got != tt.want {
				t.Errorf("ValidateMAC(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
