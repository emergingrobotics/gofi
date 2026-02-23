# gofimac Test Plan

## Test Strategy

- **Unit tests**: OUI parser, lookup, and output formatters are pure functions — test exhaustively with table-driven tests.
- **Integration tests**: Operations use `gofi.Client` — test with the gofi mock server where available, or with interface-based test doubles.
- **Build test**: Verify `go build` succeeds as part of CI.

All tests run via `go test -v -race -cover ./utilities/gofimac/...`.

## Phase 1: OUI Tests (oui_test.go)

### TestParseOUI_ValidEntries
- Input: sample IEEE OUI file content with multiple (hex) entries
- Expect: map with correct 3-octet prefix keys (lowercase colon-separated) and manufacturer names

### TestParseOUI_MixedFormat
- Input: entries with both `(hex)` and `(base 16)` formats
- Expect: only `(hex)` entries parsed, `(base 16)` ignored

### TestParseOUI_Whitespace
- Input: entries with tabs, multiple spaces, varying indentation
- Expect: parsed correctly with trimmed manufacturer names

### TestParseOUI_EmptyLines
- Input: blank lines between entries
- Expect: blank lines ignored

### TestParseOUI_NoValidEntries
- Input: file with only `(base 16)` entries and comments
- Expect: empty map, no error

### TestParseOUI_MalformedLines
- Input: lines that look like OUI entries but are missing parts
- Expect: malformed lines skipped, valid entries parsed

### TestLookup_ValidPrefix
- Input: MAC `aa:bb:cc:dd:ee:ff` with `aa:bb:cc` in database
- Expect: correct manufacturer name

### TestLookup_UppercaseInput
- Input: MAC `AA:BB:CC:DD:EE:FF`
- Expect: normalized to lowercase prefix, manufacturer found

### TestLookup_UnknownPrefix
- Input: MAC with prefix not in database
- Expect: `unknown` returned

### TestLookup_InvalidMAC
- Input: malformed MAC `zz:zz:zz`
- Expect: `unknown` returned

### TestCheckOUIFreshness_Missing
- Setup: OUI file does not exist
- Expect: download triggered

### TestCheckOUIFreshness_Fresh
- Setup: OUI file exists, modified within 30 days
- Expect: no download, existing file used

### TestCheckOUIFreshness_Stale
- Setup: OUI file exists, modified 31 days ago
- Expect: download triggered

### TestCheckOUIFreshness_DownloadSuccess
- Setup: no existing file, mock HTTP server returns valid OUI data
- Expect: file created at XDG_DATA_HOME location, data written

### TestCheckOUIFreshness_DownloadFailNoCache
- Setup: no existing file, HTTP 404
- Expect: error returned, program exits

### TestCheckOUIFreshness_DownloadFailWithStaleCache
- Setup: stale file exists (40 days old), HTTP fails
- Expect: warning to stderr, stale file used

### TestCheckOUIFreshness_DownloadFailWithFreshCache
- Setup: fresh file exists (10 days old), HTTP fails
- Expect: no warning, existing file used

### TestGetOUIPath_XDGSet
- Setup: XDG_DATA_HOME=/tmp/test
- Expect: path is /tmp/test/gofimac/oui.txt

### TestGetOUIPath_XDGUnset
- Setup: XDG_DATA_HOME unset
- Expect: path is ~/.local/share/gofimac/oui.txt

### TestDownloadOUI_HTTPSuccess
- Setup: mock HTTP server returns valid OUI content
- Expect: content downloaded and saved

### TestDownloadOUI_HTTPError
- Setup: mock HTTP server returns 500
- Expect: error returned

### TestDownloadOUI_NetworkTimeout
- Setup: mock server delays response beyond timeout
- Expect: error returned

### TestDownloadOUI_CreateDirectoryIfNeeded
- Setup: parent directory does not exist
- Expect: directory created, file written

## Phase 2: Operations Tests (operations_test.go)

Operations tests require a test double for `gofi.Client`. Use interface-based test doubles or the gofi mock server if client listing is supported.

### TestListClients_AllMode
- Setup: mock with 5 clients (3 wired, 2 wifi)
- Expect: all 5 clients returned

### TestListClients_WifiOnly
- Setup: mock with 3 wired, 2 wifi clients
- Expect: 2 wifi clients returned (IsWired == false)

### TestListClients_WiredOnly
- Setup: mock with 3 wired, 2 wifi clients
- Expect: 3 wired clients returned (IsWired == true)

### TestListClients_EmptyResult
- Setup: mock with no active clients
- Expect: empty slice, no error

### TestListClients_WithHostnames
- Setup: mock with clients having Name field set
- Expect: Name used as hostname

### TestListClients_HostnameFallback
- Setup: mock with client with empty Name, non-empty Hostname field
- Expect: Hostname used as hostname

### TestListClients_NoHostname
- Setup: mock with client with empty Name and empty Hostname
- Expect: `unknown` used as hostname

### TestListClients_OUILookup
- Setup: mock with clients, OUI database loaded with matching prefixes
- Expect: manufacturer correctly populated for each client

### TestListClients_OUILookupUnknown
- Setup: mock with client MAC not in OUI database
- Expect: manufacturer set to `unknown`

### TestListClients_SortByIP
- Input: clients with IPs 192.168.1.100, 192.168.1.2, 192.168.1.50
- Expect: sorted order is .2, .50, .100

### TestListClients_SortNoIP
- Input: clients with IPs and clients without IPs
- Expect: clients with IPs sorted first, clients without IPs at end

### TestListClients_IPConversion
- Setup: clients with various IPs in different subnets
- Expect: correct uint32 conversion and numeric sorting

### TestEnrichClientData_WiFiFields
- Setup: wifi client with essid, ap_mac, channel, signal, etc.
- Expect: all WiFi-specific fields populated

### TestEnrichClientData_WiredFields
- Setup: wired client with sw_mac, sw_port
- Expect: all wired-specific fields populated

### TestEnrichClientData_CommonFields
- Setup: client with rx_bytes, tx_bytes, uptime, last_seen
- Expect: common fields populated for both wired and wifi

## Phase 3: Format Tests (format_test.go)

### TestFormatText_SingleClient
- Input: one client with all fields
- Expect: one line, tab-separated: MAC IP HOSTNAME MANUFACTURER

### TestFormatText_MultipleClients
- Input: three clients
- Expect: three lines, sorted by IP

### TestFormatText_NoHostname
- Input: client with empty hostname
- Expect: `unknown` in hostname column

### TestFormatText_NoManufacturer
- Input: client with `unknown` manufacturer
- Expect: `unknown` in manufacturer column

### TestFormatText_NoIP
- Input: client without IP address
- Expect: `-` in IP column

### TestFormatText_EmptyList
- Input: empty client slice
- Expect: empty output (no lines)

### TestFormatText_MACLowercase
- Input: client with MAC already lowercase
- Expect: MAC remains lowercase in output

### TestFormatJSON_SingleWiFiClient
- Input: one wifi client with full fields
- Expect: JSON array with one object, includes wifi fields (essid, ap_mac, channel, radio, signal)

### TestFormatJSON_SingleWiredClient
- Input: one wired client with full fields
- Expect: JSON array with one object, includes wired fields (sw_mac, sw_port)

### TestFormatJSON_MultipleClients
- Input: mix of wired and wifi clients
- Expect: JSON array with correct is_wired flag and connection-specific fields

### TestFormatJSON_EmptyList
- Input: empty client slice
- Expect: `[]` (empty JSON array)

### TestFormatJSON_OmitEmpty
- Input: wifi client with some fields zero/empty (e.g., channel=0, signal=0)
- Expect: zero/empty fields omitted from JSON output

### TestFormatJSON_NoIP
- Input: client without IP
- Expect: ip field is empty string in JSON

### TestFormatJSON_AllCommonFields
- Input: client with rx_bytes, tx_bytes, uptime, last_seen
- Expect: all common fields present in JSON

### TestFormatJSON_WiFiFieldsOnlyForWiFi
- Input: wired client
- Expect: essid, ap_mac, channel, radio, signal fields absent

### TestFormatJSON_WiredFieldsOnlyForWired
- Input: wifi client
- Expect: sw_mac, sw_port fields absent

## Build Verification

After all tests pass:
```bash
go build ./utilities/gofimac/
```

Must succeed with zero errors and zero warnings.
