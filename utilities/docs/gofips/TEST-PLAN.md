# gofips Test Plan

## Test Strategy

- **Unit tests**: Parser, formatter, and hostname validation are pure functions — test exhaustively with table-driven tests.
- **Integration tests**: Operations use `gofi.Client` — test with the gofi mock server where available, or with interface-based test doubles.
- **Build test**: Verify `go build` succeeds as part of CI.

All tests run via `go test -v -race -cover ./utilities/gofips/...`.

## Phase 1: Parser Tests (parse_test.go)

### TestParse_SingleEntry
- Input: one valid host declaration
- Expect: one HostEntry with correct hostname, MAC, IP

### TestParse_MultipleEntries
- Input: three host declarations
- Expect: three HostEntries in input order

### TestParse_Comments
- Input: declarations with `#` comment lines interspersed
- Expect: comments ignored, entries parsed correctly

### TestParse_BlankLines
- Input: declarations separated by multiple blank lines
- Expect: blank lines ignored

### TestParse_FlexibleWhitespace
- Input: tabs, varying indentation, spaces before/after values
- Expect: parsed correctly

### TestParse_UppercaseMAC
- Input: MAC in uppercase `AA:BB:CC:DD:EE:FF`
- Expect: normalized to lowercase `aa:bb:cc:dd:ee:ff`

### TestParse_NonHostDirectives
- Input: `subnet`, `option`, `group {}` blocks mixed with host blocks
- Expect: non-host content ignored, host blocks parsed

### TestParse_ErrorInvalidMAC
- Input: host with malformed MAC `zz:zz:zz:zz:zz:zz`
- Expect: error with line number

### TestParse_ErrorInvalidIP
- Input: host with invalid IP `999.999.999.999`
- Expect: error with line number

### TestParse_ErrorIPv6
- Input: host with IPv6 address
- Expect: error (IPv4 only)

### TestParse_ErrorMissingSemicolon
- Input: `fixed-address 192.168.1.1` (no semicolon)
- Expect: error with line number

### TestParse_ErrorMissingMAC
- Input: host block with only fixed-address, no hardware ethernet
- Expect: error

### TestParse_ErrorMissingIP
- Input: host block with only hardware ethernet, no fixed-address
- Expect: error

### TestParse_ErrorDuplicateHostnames
- Input: two hosts with same hostname
- Expect: error listing both

### TestParse_ErrorDuplicateMACs
- Input: two hosts with same MAC
- Expect: error listing both

### TestParse_ErrorDuplicateIPs
- Input: two hosts with same IP
- Expect: error listing both

### TestParse_ErrorInvalidHostname
- Input: hostname with spaces or special characters
- Expect: error

### TestParse_EmptyInput
- Input: empty reader
- Expect: zero entries, no error

### TestParse_CommentsOnly
- Input: only comment lines
- Expect: zero entries, no error

### TestParseSingle_Valid
- Input: single valid declaration as string
- Expect: correct HostEntry

### TestParseSingle_ErrorMultiple
- Input: string with two declarations
- Expect: error

### TestParseSingle_ErrorEmpty
- Input: empty string
- Expect: error

### TestIsDNSSafe
Table-driven:
- `"myhost"` → true
- `"my-host"` → true
- `"my.host.local"` → true
- `"host_1"` → true
- `""` → false
- `"-host"` → false
- `"host-"` → false
- `".host"` → false
- `"ho st"` → false
- `"host!"` → false
- 64-char label → false
- 254-char total → false

## Phase 2: Formatter Tests (format_test.go)

### TestFormat_SingleEntry
- Input: one HostEntry
- Expect: correct ISC DHCP format with header, 4-space indent, semicolons

### TestFormat_MultipleEntries
- Input: three entries
- Expect: sorted by IP numerically, blank line between entries

### TestFormat_IPSortOrder
- Input: entries with IPs 192.168.1.10, 192.168.1.2, 192.168.1.100
- Expect: output order is .2, .10, .100

### TestFormat_EmptyEntries
- Input: empty slice
- Expect: header comment only with "no assignments" note

### TestFormat_HeaderComments
- Input: FormatOptions with Host and Date
- Expect: header includes host and date

### TestFormat_LowercaseMAC
- Input: entry with already-lowercase MAC
- Expect: MAC remains lowercase in output

## Phase 3: Operations Tests (operations_test.go)

Operations tests require a test double for `gofi.Client`. Since the gofi mock server may not have DNS support yet, use interface-based test doubles.

### TestDoGet_BasicExport
- Setup: mock with 2 users with fixed IPs, matching DNS records
- Expect: output contains 2 host declarations with correct hostnames

### TestDoGet_NoFixedIPs
- Setup: mock with users but none have fixed IPs
- Expect: output contains commented example

### TestDoGet_HostnameFallback
- Setup: user with empty Name, non-empty Hostname
- Expect: uses Hostname

### TestDoGet_MACFallback
- Setup: user with no Name or Hostname
- Expect: uses MAC-based hostname (aa-bb-cc-dd-ee-ff)

### TestDoGet_DNSMismatchWarning
- Setup: user Name="server1", DNS record Key="server2" pointing to same IP
- Expect: warning comment in output

### TestDoSet_CreateNew
- Setup: empty mock, input with 2 entries
- Expect: 2 created, 0 updated, 0 skipped

### TestDoSet_SkipUnchanged
- Setup: mock with existing user matching input exactly
- Expect: 0 created, 0 updated, 1 skipped

### TestDoSet_UpdateChanged
- Setup: mock with user same MAC but different IP
- Expect: 0 created, 1 updated, 0 skipped

### TestDoSet_NetworkDetection
- Setup: mock with 2 networks with different subnets
- Expect: entries assigned to correct networks based on IP

### TestDoSet_DryRun
- Setup: mock with no entries, input with 2 entries, dryRun=true
- Expect: no actual creates, result shows what would happen

### TestDoAdd_NewEntry
- Setup: empty mock
- Expect: user created with fixed IP, DNS A record created

### TestDoAdd_ConflictIP
- Setup: mock with existing user at same IP, different MAC
- Expect: error without --force

### TestDoAdd_ConflictMAC
- Setup: mock with existing user same MAC, different IP
- Expect: error without --force

### TestDoAdd_ForceOverride
- Setup: mock with conflict
- Expect: success with --force

### TestDoDel_ByName
- Setup: mock with user Name="myhost"
- Expect: fixed IP cleared, DNS deleted

### TestDoDel_ByMAC
- Setup: mock with user
- Expect: fixed IP cleared, DNS deleted

### TestDoDel_ByIP
- Setup: mock with user at 192.168.1.10
- Expect: fixed IP cleared, DNS deleted

### TestDoDel_NotFound
- Setup: empty mock
- Expect: error

### TestDoDel_KeepDNS
- Setup: mock with user + DNS
- Expect: fixed IP cleared, DNS preserved

## Build Verification

After all tests pass:
```bash
go build ./utilities/gofips/
```

Must succeed with zero errors and zero warnings.
