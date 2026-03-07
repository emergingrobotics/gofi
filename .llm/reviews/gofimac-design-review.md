# gofimac Design/Architecture Review

Reviewed: 2026-02-23
Source files: `utilities/gofimac/{main.go, oui.go, operations.go, format.go}` + test files
Design doc: `utilities/docs/gofimac/DESIGN.md`
Constraint docs: `CLAUDE.md`, `~/.claude/CLAUDE.md`

## Summary

The implementation is well-structured and closely follows the design document. The code is clean, all tests pass, and the module boundaries are respected. The file layout matches the design exactly. Several deviations from the design and the CLAUDE.md code quality rules are documented below.

---

## Critical

None.

---

## High

### H1. Lookup does not handle hyphen-separated MAC addresses

**File**: `/er/gofi/utilities/gofimac/oui.go:34-44`
**Design ref**: DESIGN.md lines 177-179

The design says: "Split on `:` or `-`, take first 3 octets." The implementation only splits on `:` via `strings.SplitN(... ":")`. A MAC address provided as `AA-BB-CC-DD-EE-FF` will not be parsed correctly -- `SplitN` on `:` returns a single element, and the lookup returns "unknown". This matters because the UDM API may return MACs in either format, and the design explicitly requires supporting both separators.

### H2. FormatJSON outputs `null` for nil slice instead of `[]`

**File**: `/er/gofi/utilities/gofimac/format.go:31-35`
**Design ref**: CLAUDE.md spec ("No clients found matching filter: print `[]` in JSON, exit 0")

When `ListClients` returns zero matches, `entries` will be `nil` (not an empty slice), because `operations.go:61` uses `var entries []ClientEntry` with `append`. `json.Encoder.Encode(nil)` outputs `null`, not `[]`. The test `TestFormatJSON_Empty` passes because it explicitly provides `[]ClientEntry{}` (non-nil empty slice), but the real code path in `main.go` passes the return value of `ListClients`, which remains `nil` when no clients match the filter.

Fix: in `ListClients`, initialize as `entries := make([]ClientEntry, 0)` or in `FormatJSON`, handle the nil case.

### H3. Tests duplicate ListClients logic instead of testing the real function

**File**: `/er/gofi/utilities/gofimac/operations_test.go:226-237`

The test helper `listClientsFromSlice` reimplements the core of `ListClients` rather than testing the actual exported function with a mock `gofi.Client`. This means the actual `ListClients` function (which calls `client.Clients().ListActive()`) has zero test coverage. If a bug is introduced in `ListClients`'s interaction with the gofi client, these tests will not catch it. The CLAUDE.md rule states: "Every function MUST have a test."

A mock implementing `gofi.Client` that returns a canned `[]types.Client` would test the real function end-to-end.

### H4. Missing tests for download and freshness logic

**Files**: `/er/gofi/utilities/gofimac/oui_test.go`

There are no tests for:
- `ensureOUIFreshness` (download failure with existing stale cache -- a key error handling path in the design)
- `downloadOUIDatabase` (using a mock HTTP server to verify atomic write, HTTP error handling, empty response handling)

The design (DESIGN.md Phase 1, line 386) explicitly calls for "Full test coverage (mock HTTP server, fixture OUI files)."

---

## Medium

### M1. OUIDatabase struct deviates from design

**File**: `/er/gofi/utilities/gofimac/oui.go:28-30`
**Design ref**: DESIGN.md lines 78-81

The design specifies:
```go
type OUIDatabase struct {
    prefixMap map[string]string
    loadedAt  time.Time
}
```

The implementation uses `entries` instead of `prefixMap` and omits `loadedAt` entirely. The field name `entries` is less descriptive than `prefixMap` -- it does not convey that the keys are 3-octet OUI prefixes. The `loadedAt` field may be needed for future features (reporting database age).

### M2. Function name mismatch: LoadOUIDatabase vs EnsureFresh

**File**: `/er/gofi/utilities/gofimac/oui.go:47`, `/er/gofi/utilities/gofimac/main.go:96`
**Design ref**: DESIGN.md lines 147-150, 278

The design specifies `EnsureFresh() (*OUIDatabase, error)` as the public API. The implementation uses `LoadOUIDatabase()` (which internally calls `ensureOUIFreshness()`). The main.go call site and design doc both describe this as the entry point for OUI management.

### M3. Design field names APMAC/SWMAC/SWPORT vs AccessPointMAC/SwitchMAC/SwitchPort

**File**: `/er/gofi/utilities/gofimac/operations.go:33,42-43`
**Design ref**: DESIGN.md lines 52, 63-64

The design specifies `APMAC`, `SWMAC`, `SWPORT` as struct field names. The implementation uses `AccessPointMAC`, `SwitchMAC`, `SwitchPort`. The CLAUDE.md rule says "do not abbreviate names," which supports the implementation's choice, but the documentation hierarchy says "When code and spec disagree, fix the code." These rules conflict. The JSON tags are correct in all cases. Either update the design or revert the field names.

### M4. Satisfaction field moved from WiFi-specific to common stats

**File**: `/er/gofi/utilities/gofimac/operations.go:50`
**Design ref**: DESIGN.md line 59

The design places `Satisfaction int` in the WiFi-specific section. The implementation places it after `LastSeen` in the common stats section. In `buildClientEntry` (line 109), `Satisfaction` is only populated for WiFi clients, so the struct layout is misleading.

### M5. Text output uses raw tabs without tabwriter

**File**: `/er/gofi/utilities/gofimac/format.go:12-28`
**Design ref**: DESIGN.md lines 228-232

The text output uses raw `\t` characters without `text/tabwriter`. The design example shows aligned columns. With raw tabs, alignment depends on terminal tab stops and will break when field lengths vary (e.g., long hostnames). Consider using `text/tabwriter` for proper column alignment.

### M6. ouiDefaultDataDir conflates XDG default with app name

**File**: `/er/gofi/utilities/gofimac/oui.go:17`

The constant `ouiDefaultDataDir = ".local/share/gofimac"` conflates the XDG default path (`.local/share`) with the application directory (`gofimac`). The `gofimac` part is also hard-coded separately at line 109 (`filepath.Join(dataHome, "gofimac")`). If the application name changes, both locations must be updated.

### M7. ParseOUIDatabase name differs from design

**File**: `/er/gofi/utilities/gofimac/oui.go:65`
**Design ref**: DESIGN.md line 153

The design specifies `Parse(r io.Reader)`. The implementation uses `ParseOUIDatabase(reader io.Reader)`. While the implementation name is more descriptive (good), it does not match the design API.

---

## Low

### L1. FilterWifi vs FilterWiFi casing

**File**: `/er/gofi/utilities/gofimac/operations.go:19`
**Design ref**: DESIGN.md line 191

The design uses `FilterWiFi` (capital F and I). The implementation uses `FilterWifi`. "WiFi" is technically the correct capitalization of the brand name, though Go convention would treat it as a word (`Wifi`).

### L2. exitError prevents testing of main error paths

**File**: `/er/gofi/utilities/gofimac/main.go:143-146`

The `exitError` function calls `os.Exit(1)`, which makes the error path in `main()` untestable. Consider restructuring `main()` to call a `run()` function that returns an error, with `main()` only handling the exit code.

### L3. No validation of --all combined with --wifi or --wired

**File**: `/er/gofi/utilities/gofimac/main.go:66-68`

The code checks `if *wifi && *wired` but does not check `if (*wifi || *wired) && *all`. Specifying `--all --wifi` silently selects WiFi-only because `--wifi` is checked first.

### L4. ouiHexMarker match is not anchored

**File**: `/er/gofi/utilities/gofimac/oui.go:70-71`

The design specifies a regex pattern anchored to line start: `^([0-9A-Fa-f]{2}-...)`. The implementation uses `strings.Contains(line, "(hex)")` which could match lines containing "(hex)" in a manufacturer name. The current IEEE OUI file does not trigger this, but the design regex is more precise.

### L5. Unused context in test helper

**File**: `/er/gofi/utilities/gofimac/operations_test.go:227`

The line `_ = context.Background()` serves no purpose and should be removed.

### L6. Comment on oui.go:80 describes WHAT not WHY

**File**: `/er/gofi/utilities/gofimac/oui.go:80`

The comment `// Normalize MAC prefix from "AA-BB-CC" to "aa:bb:cc"` describes WHAT the code does. Per CLAUDE.md comment rules, it should explain WHY (e.g., "Lowercase colon format ensures consistent map lookups regardless of input formatting").

---

## Positive Observations

- Named constants used consistently throughout (no magic numbers or strings for URLs, timeouts, file names, etc.).
- Functions are small and focused -- each does one thing well.
- Error handling is explicit; no errors are silently ignored.
- File layout matches the design document exactly (main.go, oui.go, operations.go, format.go).
- Variable naming is generally descriptive and avoids abbreviations (e.g., `ouiDatabase`, `databasePath`, `manufacturer`, `temporaryPath`).
- Atomic file writes for the OUI database prevent corruption on download.
- XDG Base Directory Specification is correctly implemented.
- All existing tests pass (28/28).
- JSON `omitempty` tags correctly applied for conditional field inclusion.
- Proper stdout/stderr separation (status messages to stderr, data to stdout).
