# gofimac Spec Compliance Review

**Date**: 2026-02-23
**Spec source**: `/er/gofi/CLAUDE.md` lines 313-505 (gofimac Tool section)
**Design doc**: `/er/gofi/utilities/docs/gofimac/DESIGN.md`
**Implementation**: `/er/gofi/utilities/gofimac/`

---

## Executive Summary

**Compliance score: 75%**

The core architecture and flow are correctly implemented: OUI database management with XDG paths, freshness checking with 30-day threshold, IEEE OUI parsing, client listing with filtering and sorting, dual output formats, atomic file writes, and custom HTTP client with timeout and user-agent. However, there are two critical spec deviations (JSON `omitempty` on "always present" fields causing missing output, and nil-slice JSON producing `null` instead of `[]`), along with high-severity issues around text output header ambiguity, OUI Lookup not handling hyphen-separated MACs, and empty-output behavior not matching spec.

---

## Critical Issues (blocking)

### C1: JSON "always present" fields use `omitempty`, violating output contract

**Spec** (CLAUDE.md line 463): "Always present: `mac`, `ip`, `hostname`, `manufacturer`, `is_wired`, `rx_bytes`, `tx_bytes`, `uptime`, `last_seen`."

**Location**: `/er/gofi/utilities/gofimac/operations.go` lines 26-49

**Finding**: Several "always present" fields have `omitempty` tags:

```go
IP           string `json:"ip,omitempty"`       // omitted when empty string
RXBytes      int64  `json:"rx_bytes,omitempty"`  // omitted when 0
TXBytes      int64  `json:"tx_bytes,omitempty"`  // omitted when 0
Uptime       int64  `json:"uptime,omitempty"`    // omitted when 0
LastSeen     int64  `json:"last_seen,omitempty"` // omitted when 0
```

When a client has no IP, `ip` will not appear in JSON output. When stats are zero (newly connected clients), `rx_bytes`, `tx_bytes`, `uptime`, and `last_seen` will be missing. Consumers relying on the documented schema will break.

**Fix**: Remove `omitempty` from `ip`, `rx_bytes`, `tx_bytes`, `uptime`, and `last_seen`. For `ip`, either keep as `""` or set to `"-"` when empty (matching the text format convention).

### C2: `FormatJSON` with nil slice produces `null` instead of `[]`

**Spec** (CLAUDE.md line 491): "No clients found matching filter: Print empty output (empty line in text, `[]` in JSON), exit 0."

**Location**: `/er/gofi/utilities/gofimac/operations.go` line 61

**Finding**: `var entries []ClientEntry` starts as nil. If no clients match the filter, the slice remains nil. `json.Encoder.Encode(nil)` produces `null\n`, not `[]\n`.

The test at `format_test.go:142` passes `[]ClientEntry{}` (non-nil empty slice), which masks this bug. The actual code path from `ListClients` returns a nil slice when no clients match.

**Fix**: In `ListClients`, initialize as `entries := []ClientEntry{}` instead of `var entries []ClientEntry`. Or handle nil in `FormatJSON` by converting nil to empty slice before encoding.

---

## High Issues

### H1: OUI `Lookup` does not handle hyphen-separated MAC addresses

**Spec** (CLAUDE.md line 348): "Given a MAC address, extract the first 3 octets and look up the manufacturer."
**Design doc** (DESIGN.md lines 178-179): "Split on `:` or `-`, take first 3 octets."

**Location**: `/er/gofi/utilities/gofimac/oui.go` line 35

**Finding**:
```go
parts := strings.SplitN(strings.ToLower(macAddress), ":", ouiMACOctetCount+1)
```

This only splits on colons. A MAC address in hyphen-separated format (e.g., `aa-bb-cc-dd-ee-ff`) will not be split correctly and `Lookup` will return `"unknown"` for a valid manufacturer. While the UDM likely returns colon-separated MACs, this deviates from the design doc requirement and reduces robustness.

**Fix**: Normalize hyphens to colons before splitting:
```go
normalized := strings.ToLower(strings.ReplaceAll(macAddress, "-", ":"))
parts := strings.SplitN(normalized, ":", ouiMACOctetCount+1)
```

### H2: Text output includes header line -- ambiguous spec compliance

**Spec** (CLAUDE.md lines 395-409): The text output format section shows a format description:
```
MAC              IP              HOSTNAME        OUI-MANUFACTURER
```
followed by a concrete example with only data lines (no header):
```
aa:bb:cc:dd:ee:01	192.168.1.10	myserver	Dell Inc.
aa:bb:cc:dd:ee:02	192.168.1.11	printer 	Hewlett Packard
aa:bb:cc:dd:ee:03	192.168.1.12	unknown 	unknown
```

**Location**: `/er/gofi/utilities/gofimac/format.go` line 14

**Finding**: The implementation emits a header line `MAC\tIP\tHOSTNAME\tOUI-MANUFACTURER\n` before data rows. The DESIGN.md (lines 226-232) explicitly includes a header line, but the original spec example in CLAUDE.md does not. Per the global rules ("When code and spec disagree, fix the code"), the CLAUDE.md spec is authoritative.

**Impact**: Tools parsing the output would need to handle (or skip) the extra header line.

**Resolution needed**: Clarify with the spec author whether the header is intended. The format description line could be interpreted either way.

### H3: Empty text output should be "empty line", not empty string

**Spec** (CLAUDE.md line 491): "No clients found matching filter: Print empty output (empty line in text, `[]` in JSON), exit 0."

**Location**: `/er/gofi/utilities/gofimac/format.go` lines 12-28

**Finding**: When `entries` has length 0, the function skips the header and the loop, returning without writing anything. The output is completely empty (zero bytes).

The spec says "empty line" which implies at least a `\n` character. If the header line is meant to be always present, an empty result should show the header with no data rows. If no header, at minimum a newline should be emitted.

---

## Medium Issues

### M1: `Satisfaction` field structurally misplaced in `ClientEntry`

**Spec** (CLAUDE.md line 464): `satisfaction` is listed as a WiFi-specific field: "WiFi clients add: `essid`, `ap_mac`, `channel`, `radio`, `radio_proto`, `signal`, `noise`, `rssi`, `satisfaction`."

**Location**: `/er/gofi/utilities/gofimac/operations.go` line 50

**Finding**: `Satisfaction` is placed outside the "WiFi fields" and "Wired fields" comment blocks, sitting with common stats (`RXBytes`, `TXBytes`, etc.). Functionally correct -- it is only populated for WiFi clients in `buildClientEntry` (line 109) -- but structurally misleading. A future developer could move it to the common population path.

**Fix**: Move the `Satisfaction` field declaration into the WiFi fields section of `ClientEntry`.

### M2: No `main_test.go` -- flag parsing and validation logic untested

**Spec** (CLAUDE.md line 3, Critical Rules): "Every function MUST have a test."

**Location**: `/er/gofi/utilities/gofimac/main.go`

**Finding**: `main()` and `exitError()` have no tests. The mutual exclusivity check (`--wifi && --wired`), host resolution from environment, and credential validation are only exercised by integration testing. The `exitError` function calls `os.Exit(1)` making it hard to test, but the validation logic could be extracted into a testable function.

### M3: `signal` field `omitempty` hides legitimate zero values

**Spec** (CLAUDE.md line 466): "Omit fields with zero/empty values from JSON output (`omitempty` behavior)."

For WiFi signal-related fields (`signal`, `noise`, `rssi`), a value of 0 is technically a valid (if rare) reading. The `omitempty` tag will suppress these. This is spec-compliant (`omitempty` is explicitly required by spec), but creates a data fidelity concern where the absence of a field is ambiguous between "not applicable" and "value is zero."

### M4: Download timeout of 60 seconds may be insufficient for constrained networks

**Location**: `/er/gofi/utilities/gofimac/oui.go` line 22

**Finding**: `ouiDownloadTimeout = 60 * time.Second`. The IEEE OUI file is roughly 4.5 MB. On constrained networks common in SBC/embedded deployments (per the Architecture Awareness section of the global CLAUDE.md), 60 seconds may not suffice. The spec does not specify a timeout value.

### M5: `--all` combined with `--wifi` or `--wired` not validated

**Location**: `/er/gofi/utilities/gofimac/main.go` lines 66-75

**Finding**: Only `--wifi && --wired` mutual exclusivity is checked. A user can specify `--all --wifi` without error; `--wifi` silently wins because it is checked first in the `if/else if` chain. The spec says "If no mode flag is given, `--all` is assumed" but does not explicitly require mutual exclusivity for all three. Still, the behavior is surprising.

---

## Low Issues

### L1: `ESSID` and `RSSI` are abbreviations but are industry-standard acronyms

**Spec** (global CLAUDE.md): "Do not abbreviate names."

**Location**: `/er/gofi/utilities/gofimac/operations.go` lines 32, 39

**Finding**: `ESSID` and `RSSI` are abbreviations per the global code quality rules. However, these are universally recognized networking acronyms. Spelling them out (`ExtendedServiceSetIdentifier`, `ReceivedSignalStrengthIndicator`) would be unusual and reduce readability. This is a judgment call.

### L2: No `tabwriter` for aligned text output

**Location**: `/er/gofi/utilities/gofimac/format.go`

**Finding**: The text output uses raw tab characters. Terminal display depends on tab stop settings, which vary. The spec example shows aligned columns, but alignment is implicitly achieved by consistent field widths in the example. Using `tabwriter` would guarantee alignment regardless of field lengths.

### L3: Test coverage gap for `LoadOUIDatabase` integration path

**Location**: `/er/gofi/utilities/gofimac/oui.go` lines 47-61

**Finding**: `ensureOUIFreshness` and `downloadOUIDatabase` are not unit-tested (they require filesystem and network access). The freshness logic could be made testable with injected time and filesystem abstractions.

### L4: `os.UserHomeDir()` error fallback to `.` is untested

**Location**: `/er/gofi/utilities/gofimac/oui.go` lines 103-105

**Finding**: When `os.UserHomeDir()` fails, the code falls back to the current directory (`.`). This edge case has no test coverage.

---

## Positive Compliance Notes

1. **OUI storage path**: Correctly implements XDG Base Directory Specification with `$XDG_DATA_HOME` fallback to `~/.local/share/gofimac/oui.txt`. Tests verify both paths.

2. **30-day freshness check**: Correctly implemented with file modification time comparison and `ouiMaxAgeDays` named constant.

3. **Download failure with stale cache**: Correctly falls back to cached file with stderr warning (oui.go lines 130-132).

4. **Download failure without cache**: Correctly returns error (oui.go line 134).

5. **Atomic file write**: Uses temp file + rename pattern as specified (oui.go lines 164-187).

6. **Custom HTTP client**: Uses timeout (60s) and user-agent header (`gofimac/1.0`) per design doc requirements (oui.go lines 147-153).

7. **OUI parsing**: Correctly parses only `(hex)` lines, normalizes to lowercase colon-separated prefixes, trims whitespace from manufacturer names.

8. **Hostname resolution**: Correctly prioritizes `Name` > `Hostname` > `"unknown"` per spec (operations.go lines 115-123).

9. **OUI lookup independence**: Never uses the UDM's `OUI` field; always performs independent lookup via `ouiDatabase.Lookup(client.MAC)`.

10. **IP sort order**: Uses `binary.BigEndian.Uint32(ip.To4())` for numeric IP sorting, matching the spec exactly.

11. **No-IP clients sort last**: Correctly sorts clients without IP addresses after all IP-bearing clients, with MAC-based secondary sort among no-IP clients.

12. **MAC lowercase**: Correctly lowercases all MAC addresses in output via `strings.ToLower(client.MAC)`.

13. **Filter modes**: All three filter modes (`FilterAll`, `FilterWifi`, `FilterWired`) with correct `FilterMode` type name matching design doc.

14. **WiFi vs wired field population**: Correctly populates connection-type-specific fields based on `IsWired` (operations.go lines 97-112).

15. **JSON omitempty for conditional fields**: WiFi and wired-specific fields correctly use `omitempty` to exclude irrelevant fields.

16. **Connection flags**: All four connection flags (`--host/-H`, `--port/-p`, `--site/-S`, `--insecure/-k`) with correct short forms.

17. **Environment variable fallback**: `UNIFI_UDM_IP` correctly used as fallback for `--host`.

18. **Credential validation**: Username and password checked before any connection attempt.

19. **Status messages to stderr**: All progress/warning messages correctly go to stderr (connection status, OUI download, client fetch).

20. **JSON indentation**: Uses 2-space indentation matching the spec example format.

21. **File layout**: Matches spec exactly (`main.go`, `oui.go`, `format.go`, `operations.go` with corresponding test files).

22. **Good test coverage**: Tests exist for OUI parsing, lookup (known/unknown/case-insensitive/short/empty MAC), filtering (all/wifi/wired), sorting (by IP/no-IP/multiple-no-IP), hostname resolution, MAC lowercase, and both output formats (text/JSON with empty/wired/WiFi/omitted fields).

23. **Named constants**: Magic values properly extracted (`ouiDatabaseURL`, `ouiMaxAgeDays`, `ouiFileName`, `ouiHexMarker`, etc.).

24. **Field naming**: `AccessPointMAC`, `SwitchMAC`, `SwitchPort` follow the "do not abbreviate" rule (correctly expanded from the types package's abbreviated names).

---

## Summary of Required Changes by Priority

### Must Fix (Critical)
1. **C1**: Remove `omitempty` from `ip`, `rx_bytes`, `tx_bytes`, `uptime`, `last_seen` JSON tags
2. **C2**: Initialize `entries` as `[]ClientEntry{}` in `ListClients` to avoid nil-slice `null` JSON

### Should Fix (High)
3. **H1**: Handle hyphen-separated MACs in `Lookup`
4. **H2**: Clarify whether text header line is intended; align with spec
5. **H3**: Emit at least a newline for empty text output

### Consider Fixing (Medium)
6. **M1**: Move `Satisfaction` field into WiFi section of `ClientEntry`
7. **M2**: Extract validation logic from `main()` into testable functions
8. **M5**: Validate `--all` mutual exclusivity with `--wifi`/`--wired`

### Nice to Have (Low)
9. **L2**: Use `tabwriter` for aligned text output
10. **L3**: Add tests for freshness/download logic with injected dependencies
