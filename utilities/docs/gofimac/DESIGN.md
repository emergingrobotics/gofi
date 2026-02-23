# gofimac - Client MAC/OUI Lister

A command-line tool for listing connected clients on a UniFi UDM Pro, filtered by connection type, with independent OUI manufacturer lookup.

## Overview

`gofimac` provides filtered client listing with current manufacturer identification via IEEE OUI database lookup. Three filter modes:

- **WiFi** (`--wifi`, `-w`): List only WiFi-connected clients.
- **Wired** (`--wired`, `-e`): List only wired (ethernet) clients.
- **All** (`--all`, `-a`): List all connected clients (default).

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   main.go                        │
│  Flag parsing, mode dispatch, connection setup   │
└──────────┬──────────┬──────────┬────────────────┘
           │          │          │
     ┌─────▼──┐  ┌────▼───┐  ┌──▼──────────┐
     │ oui.go │  │format.go│  │operations.go│
     │        │  │         │  │             │
     │ IEEE   │  │ Text +  │  │ ListClients │
     │ OUI DB │  │ JSON    │  │ with filter │
     │ mgmt   │  │ output  │  │ + sort      │
     └────────┘  └─────────┘  └──────┬──────┘
                                     │
                              ┌──────▼──────┐
                              │  gofi.Client │
                              │  Clients()   │
                              │  ListActive()│
                              └─────────────┘
```

## Data Model

### ClientEntry

Output representation of a single client with OUI lookup:

```go
type ClientEntry struct {
    MAC          string // lowercase colon-separated MAC
    IP           string // IPv4 dotted-quad, or "-" if none
    Hostname     string // Name or Hostname from client, or "unknown"
    Manufacturer string // OUI lookup result, or "unknown"
    IsWired      bool   // true for wired, false for WiFi

    // WiFi-specific fields (nil/zero if wired)
    ESSID        string
    APMAC        string
    Channel      int
    Radio        string
    RadioProto   string
    Signal       int
    Noise        int
    RSSI         int
    Satisfaction int

    // Wired-specific fields (nil/zero if WiFi)
    SWMAC        string
    SWPORT       int

    // Common stats
    RXBytes      int64
    TXBytes      int64
    Uptime       int64
    LastSeen     int64
}
```

### OUIDatabase

In-memory OUI lookup map:

```go
type OUIDatabase struct {
    prefixMap map[string]string // "aa:bb:cc" -> "Manufacturer Name"
    loadedAt  time.Time
}
```

## OUI Database Management

### Storage Location

XDG-compliant data directory:
- `$XDG_DATA_HOME/gofimac/oui.txt` if `XDG_DATA_HOME` is set
- `~/.local/share/gofimac/oui.txt` otherwise

No root access required.

### Freshness Check

On every invocation, before operations:

1. Check if `oui.txt` exists at storage location.
2. If exists, check file modification time:
   - If age > 30 days, attempt re-download.
   - If age <= 30 days, use cached file.
3. If does not exist, download.
4. Download source: `https://standards-oui.ieee.org/oui/oui.txt`
5. Download failure handling:
   - If cached file exists (even if stale), use it and print warning to stderr.
   - If no cached file exists, exit with error.
6. Progress messages go to stderr.

### IEEE OUI File Format

The IEEE OUI database contains entries like:

```
AA-BB-CC   (hex)		Acme Corporation
AABBCC     (base 16)		Acme Corporation
				123 Main Street
				Springfield IL 12345
				US

DD-EE-FF   (hex)		Another Company
DDEEFF     (base 16)		Another Company
				456 Oak Avenue
				...
```

**Parsing rules:**
- Extract only lines matching pattern: `XX-XX-XX   (hex)		<manufacturer>`
- Split on `(hex)`, take the left side for prefix, right side for manufacturer name.
- Normalize prefix: convert hyphens to colons, lowercase (e.g., `AA-BB-CC` -> `aa:bb:cc`).
- Trim whitespace from manufacturer name.
- Ignore all other lines (base 16, addresses, blank lines).
- Build map: `{"aa:bb:cc": "Acme Corporation", ...}`

### Lookup

Given a MAC address:
1. Extract first 3 octets (e.g., `aa:bb:cc:dd:ee:ff` -> `aa:bb:cc`).
2. Normalize to lowercase with colons.
3. Look up in map.
4. Return manufacturer name if found, `unknown` if not.

## Module Design

### oui.go — OUI Database Management

```go
// EnsureFresh checks the OUI database freshness and downloads if needed.
// Returns an OUIDatabase ready for lookups.
// Progress messages written to stderr.
func EnsureFresh() (*OUIDatabase, error)

// Parse parses the IEEE OUI file from reader into a lookup map.
func Parse(r io.Reader) (*OUIDatabase, error)

// Lookup returns the manufacturer name for a given MAC address.
// Returns "unknown" if not found.
func (db *OUIDatabase) Lookup(mac string) string
```

**EnsureFresh logic:**
1. Resolve storage path: `$XDG_DATA_HOME/gofimac/oui.txt` or `~/.local/share/gofimac/oui.txt`.
2. Check file existence and modification time.
3. If download needed:
   - Create storage directory if missing.
   - Download to temp file in same directory.
   - Atomic rename to final location.
4. Open file, call Parse(), return database.

**Parse logic:**
1. Scan line by line.
2. Match regex: `^([0-9A-Fa-f]{2}-[0-9A-Fa-f]{2}-[0-9A-Fa-f]{2})\s+\(hex\)\s+(.+)$`
3. Extract prefix (group 1) and manufacturer (group 2).
4. Normalize prefix to lowercase colon-separated.
5. Trim manufacturer name.
6. Build map.

**Lookup logic:**
1. Normalize input MAC to lowercase.
2. Split on `:` or `-`, take first 3 octets.
3. Join with `:`.
4. Query map, return value or `"unknown"`.

### operations.go — Client Listing and Filtering

```go
// FilterMode specifies which clients to include.
type FilterMode int

const (
    FilterAll FilterMode = iota
    FilterWiFi
    FilterWired
)

// ListClients fetches active clients, filters by mode, looks up OUI, sorts.
// Returns ClientEntry slice ready for formatting.
func ListClients(ctx context.Context, client gofi.Client, site string, mode FilterMode, db *OUIDatabase) ([]ClientEntry, error)
```

**ListClients flow:**
1. Call `client.Clients().ListActive(ctx, site)` to fetch all active clients.
2. Filter based on mode:
   - `FilterAll`: no filter
   - `FilterWiFi`: `IsWired == false`
   - `FilterWired`: `IsWired == true`
3. For each client:
   - Extract MAC, IP, hostname (Name or Hostname or "unknown").
   - Look up manufacturer via `db.Lookup(mac)`.
   - Build `ClientEntry` struct.
   - Populate WiFi or wired fields based on `IsWired`.
4. Sort entries:
   - Clients with IP addresses sort numerically by IP (uint32 conversion).
   - Clients without IP sort last, alphabetically by MAC.
5. Return sorted slice.

### format.go — Output Formatting

```go
// FormatText writes client entries in tab-separated text format.
func FormatText(w io.Writer, entries []ClientEntry) error

// FormatJSON writes client entries in JSON format.
func FormatJSON(w io.Writer, entries []ClientEntry) error
```

**FormatText output:**
```
MAC              IP              HOSTNAME        OUI-MANUFACTURER
aa:bb:cc:dd:ee:01	192.168.1.10	myserver	Dell Inc.
aa:bb:cc:dd:ee:02	192.168.1.11	printer 	Hewlett Packard
aa:bb:cc:dd:ee:03	-           	unknown 	unknown
```

Rules:
- Header line.
- One entry per line.
- Tab-separated columns.
- MAC lowercase with colons.
- IP as dotted-quad, or `-` if empty.
- Hostname or `unknown`.
- Manufacturer from OUI lookup or `unknown`.

**FormatJSON output:**
```json
[
  {
    "mac": "aa:bb:cc:dd:ee:01",
    "ip": "192.168.1.10",
    "hostname": "myserver",
    "manufacturer": "Dell Inc.",
    "is_wired": false,
    "essid": "MyNetwork",
    "ap_mac": "ff:ee:dd:cc:bb:aa",
    "channel": 36,
    "radio": "na",
    "signal": -42,
    "rx_bytes": 123456789,
    "tx_bytes": 987654321,
    "uptime": 86400,
    "last_seen": 1708700000
  }
]
```

Rules:
- JSON array of objects.
- Always include: `mac`, `ip`, `hostname`, `manufacturer`, `is_wired`, `rx_bytes`, `tx_bytes`, `uptime`, `last_seen`.
- WiFi clients include: `essid`, `ap_mac`, `channel`, `radio`, `radio_proto`, `signal`, `noise`, `rssi`, `satisfaction`.
- Wired clients include: `sw_mac`, `sw_port`.
- Use `omitempty` behavior: omit fields with zero/empty values.

### main.go — Entry Point

Responsibilities:
- Parse flags (mode, output format, connection).
- Validate mode (default to `--all` if none specified).
- Resolve host from env fallback.
- Validate credentials.
- Check/update OUI database via `EnsureFresh()`.
- Create gofi client, connect, defer disconnect.
- Call `ListClients()` with filter mode and OUI database.
- Call appropriate format function (text or JSON).
- Handle exit codes.

## Flow Descriptions

### Main Operation Flow

1. Parse command-line flags.
2. Validate mode and output format.
3. Check OUI database freshness:
   - If missing or stale, download.
   - If download fails and no cache, exit 1.
   - If download fails and cache exists, warn and use cache.
4. Parse OUI database into lookup map.
5. Connect to UDM Pro.
6. Fetch active clients.
7. Filter by connection type (WiFi/wired/all).
8. For each client:
   - Extract fields from `types.Client`.
   - Look up manufacturer from OUI database.
   - Build `ClientEntry`.
9. Sort by IP address (clients without IP sort last).
10. Format output (text or JSON).
11. Write to stdout.
12. Disconnect and exit.

### OUI Update Flow

```
Start
  |
  v
Resolve storage path ($XDG_DATA_HOME/gofimac/oui.txt)
  |
  v
File exists?
  |
  +-- No --> Download --> Parse --> Return DB
  |
  +-- Yes --> Check mtime
               |
               v
            Age > 30 days?
               |
               +-- No --> Parse existing --> Return DB
               |
               +-- Yes --> Download --> Success?
                            |
                            +-- Yes --> Parse new --> Return DB
                            |
                            +-- No --> Parse existing --> Warn --> Return DB (or fail if no cache)
```

## Error Handling Strategy

| Condition | Behavior |
|-----------|----------|
| Missing credentials or host | Print error to stderr, exit 1 |
| Connection failure | Print error to stderr, exit 1 |
| OUI download fails, no cache | Print error to stderr, exit 1 |
| OUI download fails, stale cache exists | Print warning to stderr, use cached data, continue |
| OUI parse error | Print error to stderr, exit 1 |
| No clients matching filter | Print empty output (empty table in text, `[]` in JSON), exit 0 |
| Individual client parse error | Skip client, log warning to stderr, continue |

## IP Sort Order

Same as gofips: convert IPv4 to uint32 via `binary.BigEndian.Uint32(ip.To4())` for numeric comparison. Clients without an IP address sort last and are ordered alphabetically by MAC address.

## Output Destinations

| Stream | Content |
|--------|---------|
| stdout | Client listing (text or JSON) |
| stderr | Status messages (OUI download progress, warnings, errors) |

Never mix status messages into stdout.

## Project Layout

```
utilities/
  gofimac/
    main.go           # Entry point, flag parsing, connection
    oui.go            # OUI database download, parse, lookup
    oui_test.go       # OUI tests
    operations.go     # Client listing and filtering
    operations_test.go # Operations tests (with mock client)
    format.go         # Text and JSON output formatting
    format_test.go    # Format tests
  docs/
    gofimac/
      DESIGN.md       # This file
```

## Implementation Phases

### Phase 1: OUI Management (oui.go + oui_test.go)
- Storage path resolution (XDG Base Directory Specification).
- Freshness check logic (30-day threshold).
- Download with HTTP client (user-agent, timeout).
- Atomic file write (temp file + rename).
- Parse IEEE OUI format with regex.
- Lookup function with 3-octet extraction.
- Full test coverage (mock HTTP server, fixture OUI files).

### Phase 2: Operations + Formatting (operations.go + format.go + tests)
- `ClientEntry` type.
- `FilterMode` enum.
- `ListClients()` implementation:
  - Call `client.Clients().ListActive()`.
  - Filter by `IsWired` field.
  - Extract fields from `types.Client`.
  - OUI lookup per client.
  - IP numeric sort with no-IP handling.
- `FormatText()` implementation:
  - Header line.
  - Tab-separated columns.
  - Handle missing fields (`-`, `unknown`).
- `FormatJSON()` implementation:
  - JSON array of objects.
  - Conditional field inclusion (WiFi vs wired).
  - `omitempty` behavior.
- Full test coverage (mock gofi client, fixture data).

### Phase 3: Main + Integration (main.go)
- Flag parsing and validation (pflag or standard flag).
- Mode default to `--all`.
- Connection flag handling (host, port, site, insecure).
- Credential validation.
- Call `EnsureFresh()` for OUI database.
- Create gofi client, connect, defer disconnect.
- Call `ListClients()` with mode and database.
- Call format function based on output flag.
- Build verification (`go build`).
- Manual integration test against real or mock UDM.
