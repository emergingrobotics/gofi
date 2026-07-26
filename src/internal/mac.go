package internal

import (
	"regexp"
	"strings"
)

var macRegex = regexp.MustCompile(`^([0-9a-fA-F]{2}[:-]?){5}([0-9a-fA-F]{2})$`)

// NormalizeMAC normalizes a MAC address to lowercase without separators.
// Example: "AA:BB:CC:DD:EE:FF" -> "aabbccddeeff"
func NormalizeMAC(mac string) string {
	normalized := strings.ToLower(mac)
	normalized = strings.ReplaceAll(normalized, ":", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	return normalized
}

// FormatMAC formats a MAC address with colon separators.
// Example: "aabbccddeeff" -> "aa:bb:cc:dd:ee:ff"
func FormatMAC(mac string) string {
	normalized := NormalizeMAC(mac)
	if len(normalized) != 12 {
		return mac // Return as-is if invalid length
	}

	var formatted strings.Builder
	for i := 0; i < len(normalized); i += 2 {
		if i > 0 {
			formatted.WriteString(":")
		}
		formatted.WriteString(normalized[i : i+2])
	}
	return formatted.String()
}

// ValidateMAC checks if a MAC address is valid.
func ValidateMAC(mac string) bool {
	if mac == "" {
		return false
	}
	return macRegex.MatchString(mac)
}
