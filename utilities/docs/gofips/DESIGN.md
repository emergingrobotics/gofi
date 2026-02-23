# gofips - Fixed IP + DNS Manager (ISC DHCP Format)

A command-line tool for managing DHCP fixed IP assignments with DNS records on a UniFi UDM Pro. Uses the ISC DHCP `dhcpd.conf` host declaration format for import/export.

## Overview

`gofips` provides four operations:

- **Get** (`--get`, `-g`): Export current fixed IP + hostname assignments from the UDM in ISC DHCP format.
- **Set** (`--set`, `-s`): Bulk import host declarations from a file or stdin.
- **Add** (`--add`, `-a`): Add a single host from an ISC DHCP declaration fragment.
- **Del** (`--del`, `-d`): Delete a host by name, MAC, or IP.

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   main.go                        │
│  Flag parsing, mode dispatch, connection setup   │
└──────────┬──────────┬──────────┬────────────────┘
           │          │          │
     ┌─────▼──┐  ┌────▼───┐  ┌──▼──────────┐
     │parse.go│  │format.go│  │operations.go│
     │        │  │         │  │             │
     │ Parse  │  │ Format  │  │ Get/Set/    │
     │ ISC    │  │ ISC     │  │ Add/Del     │
     │ DHCP   │  │ DHCP    │  │ business    │
     │ blocks │  │ output  │  │ logic       │
     └────────┘  └─────────┘  └──────┬──────┘
                                     │
                              ┌──────▼──────┐
                              │  gofi.Client │
                              │  Users()     │
                              │  DNS()       │
                              │  Networks()  │
                              └─────────────┘
```

## Data Model

### HostEntry

Internal representation of a single host declaration:

```go
type HostEntry struct {
    Hostname string // DNS-safe hostname
    MAC      string // lowercase colon-separated MAC
    IP       string // IPv4 dotted-quad
}
```

This maps to:
- UniFi User record: MAC -> FixedIP with Name set to Hostname
- UniFi DNS A record: Key=Hostname, Value=IP

## Module Design

### parse.go — ISC DHCP Parser

Parses ISC DHCP host declarations from an `io.Reader`.

```go
// ParseResult holds parsed entries and any non-fatal warnings.
type ParseResult struct {
    Entries  []HostEntry
    Warnings []string
}

// Parse reads ISC DHCP host declarations from r.
// Returns all parsed entries or an error if validation fails.
// Non-host directives are silently skipped.
func Parse(r io.Reader) (*ParseResult, error)

// ParseSingle parses exactly one host declaration from a string.
// Returns an error if the string contains zero or more than one declaration.
func ParseSingle(s string) (*HostEntry, error)
```

**Parser state machine:**

```
IDLE ──"host <name> {"──▶ IN_BLOCK
IN_BLOCK ──"hardware ethernet <mac>;"──▶ IN_BLOCK (mac set)
IN_BLOCK ──"fixed-address <ip>;"──▶ IN_BLOCK (ip set)
IN_BLOCK ──"}"──▶ IDLE (emit entry if mac+ip present)
```

- Lines starting with `#` are comments, skipped.
- Blank lines are skipped.
- Non-host directives (subnet, option, etc.) are skipped by staying in IDLE.
- Tracks line numbers for error messages.

**Validation (after parsing all entries):**
- Hostname: `[a-zA-Z0-9._-]+`, max 63 chars per label, max 253 total
- MAC: `^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$` (normalized to lowercase)
- IP: valid IPv4 via `net.ParseIP`, must be `.To4()` non-nil
- No duplicate hostnames
- No duplicate MACs
- No duplicate IPs
- All three fields (hostname, MAC, IP) must be present per entry

### format.go — ISC DHCP Formatter

Formats `[]HostEntry` to ISC DHCP host declaration format.

```go
// FormatOptions controls output formatting.
type FormatOptions struct {
    Host string // UDM host for header comment
    Date string // Date string for header comment
}

// Format writes host declarations to w in ISC DHCP format.
// Entries are sorted by IP address numerically before output.
func Format(w io.Writer, entries []HostEntry, opts FormatOptions) error
```

**Output format:**

```
# gofips fixed IP assignments
# exported from UDM at 192.168.1.1
# date: 2026-02-23

host myserver {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 192.168.1.10;
}

host printer {
    hardware ethernet aa:bb:cc:dd:ee:02;
    fixed-address 192.168.1.11;
}
```

- 4-space indentation
- Blank line between entries
- Sorted by IP numerically (uint32 conversion)
- Header comments with export metadata

### operations.go — Business Logic

All operations take a `gofi.Client` and site name. Operations connect/disconnect externally (in main.go).

```go
// DoGet exports all fixed IP assignments with hostnames.
func DoGet(ctx context.Context, client gofi.Client, site string, w io.Writer, opts FormatOptions) error

// DoSet imports host declarations from parsed entries.
func DoSet(ctx context.Context, client gofi.Client, site string, entries []HostEntry, dryRun bool) (*SetResult, error)

// DoAdd adds a single host entry.
func DoAdd(ctx context.Context, client gofi.Client, site string, entry *HostEntry, force bool) error

// DoDel deletes a host by identifier.
func DoDel(ctx context.Context, client gofi.Client, site string, identifier DeleteIdentifier, force, keepDNS bool) error
```

```go
// SetResult holds the outcome of a set operation.
type SetResult struct {
    Created int
    Updated int
    Skipped int
    Errors  int
}

// DeleteIdentifier specifies how to find the host to delete.
type DeleteIdentifier struct {
    Name string // --name
    MAC  string // --mac
    IP   string // --ip
}
```

#### DoGet Flow

1. `client.Users().List(ctx, site)` — get all users
2. Filter to `UseFixedIP == true && FixedIP != ""`
3. For each user, resolve hostname:
   - `user.Name` if non-empty and DNS-safe
   - `user.Hostname` if non-empty and DNS-safe
   - MAC with colons replaced by hyphens as fallback
4. Cross-reference DNS: `client.DNS().List(ctx, site)` — build IP-to-DNS map
5. If DNS record hostname differs from user hostname, emit warning comment
6. Sort by IP, format, write to `w`

#### DoSet Flow

1. Fetch existing users, build MAC-keyed map
2. Fetch existing DNS records, build hostname-keyed map
3. Fetch networks for subnet detection
4. For each entry:
   - Check if MAC exists with same IP and hostname → skip
   - Check if MAC exists with different IP/hostname → update user + DNS
   - Otherwise → create user + DNS record
5. Network auto-detection via subnet containment
6. Return summary counts

#### DoAdd Flow

1. Fetch existing users and DNS records
2. Conflict check (unless `--force`):
   - IP assigned to different MAC?
   - MAC assigned different IP?
   - Hostname DNS record points elsewhere?
3. Create/update user record with fixed IP
4. Create/update DNS A record

#### DoDel Flow

1. Find target by identifier type:
   - By name: search users by Name/Hostname, search DNS by Key
   - By MAC: `client.Users().GetByMAC()`
   - By IP: search users where `FixedIP == ip`
2. If no match → error
3. If multiple matches without `--force` → list matches, error
4. Clear fixed IP from user (do not delete user unless `--force`)
5. Delete DNS records for the IP (unless `--keep-dns`)

### main.go — Entry Point

Responsibilities:
- Parse flags (mode, connection, identifiers, modifiers)
- Validate mode exclusivity
- Resolve host from env fallback
- Validate credentials
- Create gofi client, connect, defer disconnect
- Dispatch to appropriate operation function
- Handle exit codes

## ISC DHCP Format Compatibility

The parser handles real-world `dhcpd.conf` files by:
- Extracting only `host {}` blocks
- Ignoring `subnet`, `option`, `group`, `shared-network`, and other top-level directives
- Accepting flexible whitespace and indentation
- Requiring semicolons on `hardware ethernet` and `fixed-address` lines
- Treating everything outside a `host {}` block as non-host content to skip

## Error Handling Strategy

| Phase | Strategy |
|-------|----------|
| Input parsing | Collect all errors, report all at once, fail before connecting |
| Connection | Fail immediately with clear message |
| Individual operations in --set | Log error, continue, count errors, exit 1 at end |
| --add conflicts | Fail immediately unless --force |
| --del no match | Fail immediately |
| Network detection | Per-entry error in --set, fatal in --add |

## IP Sort Order

Same as gofip: convert IPv4 to uint32 via `binary.BigEndian.Uint32(ip.To4())` for numeric comparison.

## DNS-Safe Hostname Validation

```go
func isDNSSafe(s string) bool
```

- Characters: `[a-zA-Z0-9._-]`
- Must not start or end with `.` or `-`
- Each label (split by `.`) max 63 characters
- Total max 253 characters
- Must have at least one character

## Project Layout

```
utilities/
  gofips/
    main.go           # Entry point, flag parsing, mode dispatch
    parse.go          # ISC DHCP format parser
    parse_test.go     # Parser tests
    format.go         # ISC DHCP format output
    format_test.go    # Formatter tests
    operations.go     # get/set/add/del business logic
    operations_test.go # Operations tests (with mock client)
  docs/
    gofips/
      DESIGN.md       # This file
```

## Implementation Phases

### Phase 1: Parser (parse.go + parse_test.go)
- HostEntry type
- Parse() function with state machine
- ParseSingle() function
- isDNSSafe() validation
- Duplicate detection
- Full test coverage

### Phase 2: Formatter (format.go + format_test.go)
- Format() function
- IP numeric sort
- Header comments
- Empty-input handling
- Full test coverage

### Phase 3: Operations (operations.go + operations_test.go)
- DoGet, DoSet, DoAdd, DoDel
- DeleteIdentifier type
- SetResult type
- Network auto-detection (reuse pattern from gofip)
- Full test coverage with gofi mock client

### Phase 4: Main + Integration (main.go)
- Flag parsing and validation
- Mode dispatch
- Connection management
- Build verification
