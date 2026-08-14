package ips

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
)

const (
	hostDeclarationMinFields = 3   // "host", hostname, "{"
	maxDNSHostnameLength     = 253 // RFC 1035 total hostname length
	maxDNSLabelLength        = 63  // RFC 1035 per-label length
)

type HostEntry struct {
	Hostname string
	MAC      string
	IP       string
}

type ParseResult struct {
	Entries  []HostEntry
	Warnings []string
}

var macRegex = regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)

func Parse(reader io.Reader) (*ParseResult, error) {
	scanner := bufio.NewScanner(reader)
	var entries []HostEntry
	var warnings []string
	var parseErrors []string

	type blockState struct {
		hostname       string
		macAddress     string
		ipAddress      string
		startLine      int
		hasMACAddress  bool
		hasIPAddress   bool
	}

	var current *blockState
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if current == nil {
			if strings.HasPrefix(line, "host ") && strings.HasSuffix(line, "{") {
				parts := strings.Fields(line)
				if len(parts) < hostDeclarationMinFields {
					parseErrors = append(parseErrors, fmt.Sprintf("line %d: malformed host declaration: %s", lineNumber, line))
					continue
				}
				hostname := parts[1]
				if !isDNSSafe(hostname) {
					parseErrors = append(parseErrors, fmt.Sprintf("line %d: invalid hostname %q (must match [a-zA-Z0-9._-], not start/end with - or .)", lineNumber, hostname))
					continue
				}
				current = &blockState{hostname: hostname, startLine: lineNumber}
			}
			continue
		}

		if line == "}" {
			if !current.hasMACAddress {
				parseErrors = append(parseErrors, fmt.Sprintf("line %d: host %q (started line %d) missing 'hardware ethernet' directive", lineNumber, current.hostname, current.startLine))
			}
			if !current.hasIPAddress {
				parseErrors = append(parseErrors, fmt.Sprintf("line %d: host %q (started line %d) missing 'fixed-address' directive", lineNumber, current.hostname, current.startLine))
			}
			if current.hasMACAddress && current.hasIPAddress {
				entries = append(entries, HostEntry{
					Hostname: current.hostname,
					MAC:      current.macAddress,
					IP:       current.ipAddress,
				})
			}
			current = nil
			continue
		}

		if strings.HasPrefix(line, "hardware ethernet ") {
			value := strings.TrimPrefix(line, "hardware ethernet ")
			value = strings.TrimSuffix(value, ";")
			value = strings.TrimSpace(value)
			if !strings.HasSuffix(line, ";") {
				parseErrors = append(parseErrors, fmt.Sprintf("line %d: missing semicolon after hardware ethernet value", lineNumber))
				continue
			}
			normalizedMAC := strings.ToLower(value)
			if !macRegex.MatchString(normalizedMAC) {
				parseErrors = append(parseErrors, fmt.Sprintf("line %d: invalid MAC address: %s", lineNumber, value))
				continue
			}
			current.macAddress = normalizedMAC
			current.hasMACAddress = true
			continue
		}

		if strings.HasPrefix(line, "fixed-address ") {
			value := strings.TrimPrefix(line, "fixed-address ")
			value = strings.TrimSuffix(value, ";")
			value = strings.TrimSpace(value)
			if !strings.HasSuffix(line, ";") {
				parseErrors = append(parseErrors, fmt.Sprintf("line %d: missing semicolon after fixed-address value", lineNumber))
				continue
			}
			parsedIP := net.ParseIP(value)
			if parsedIP == nil {
				parseErrors = append(parseErrors, fmt.Sprintf("line %d: invalid IP address: %s", lineNumber, value))
				continue
			}
			if parsedIP.To4() == nil {
				parseErrors = append(parseErrors, fmt.Sprintf("line %d: only IPv4 addresses are supported: %s", lineNumber, value))
				continue
			}
			current.ipAddress = value
			current.hasIPAddress = true
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	if current != nil {
		parseErrors = append(parseErrors, fmt.Sprintf("line %d: unclosed host block %q (started line %d)", lineNumber, current.hostname, current.startLine))
	}

	if len(parseErrors) > 0 {
		return nil, fmt.Errorf("parse errors:\n  %s", strings.Join(parseErrors, "\n  "))
	}

	if err := checkDuplicates(entries); err != nil {
		return nil, err
	}

	return &ParseResult{Entries: entries, Warnings: warnings}, nil
}

func ParseSingle(input string) (*HostEntry, error) {
	result, err := Parse(strings.NewReader(input))
	if err != nil {
		return nil, err
	}
	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("no host declaration found in input")
	}
	if len(result.Entries) > 1 {
		return nil, fmt.Errorf("expected exactly one host declaration, found %d", len(result.Entries))
	}
	return &result.Entries[0], nil
}

func checkDuplicates(entries []HostEntry) error {
	seenHostnames := make(map[string]int)
	seenMACs := make(map[string]int)
	seenIPs := make(map[string]int)
	var dupErrors []string

	for index, entry := range entries {
		entryNumber := index + 1
		if previous, ok := seenHostnames[entry.Hostname]; ok {
			dupErrors = append(dupErrors, fmt.Sprintf("duplicate hostname %q: entries %d and %d", entry.Hostname, previous, entryNumber))
		}
		seenHostnames[entry.Hostname] = entryNumber

		if previous, ok := seenMACs[entry.MAC]; ok {
			dupErrors = append(dupErrors, fmt.Sprintf("duplicate MAC %s: entries %d and %d", entry.MAC, previous, entryNumber))
		}
		seenMACs[entry.MAC] = entryNumber

		if previous, ok := seenIPs[entry.IP]; ok {
			dupErrors = append(dupErrors, fmt.Sprintf("duplicate IP %s: entries %d and %d", entry.IP, previous, entryNumber))
		}
		seenIPs[entry.IP] = entryNumber
	}

	if len(dupErrors) > 0 {
		return fmt.Errorf("duplicate entries:\n  %s", strings.Join(dupErrors, "\n  "))
	}
	return nil
}

// isDNSSafe validates RFC 1035 hostname constraints because UniFi DNS
// requires RFC-compliant names for A record creation to succeed.
func isDNSSafe(hostname string) bool {
	if hostname == "" || len(hostname) > maxDNSHostnameLength {
		return false
	}
	if strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return false
	}
	if strings.HasPrefix(hostname, ".") || strings.HasSuffix(hostname, ".") {
		return false
	}

	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > maxDNSLabelLength {
			return false
		}
	}

	for _, character := range hostname {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '.' || character == '-' || character == '_') {
			return false
		}
	}
	return true
}
