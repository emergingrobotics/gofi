# gofi - Go UniFi Controller

A Go module for programmatic control of UniFi UDM Pro devices (v10+).

## Project Resources

- **API Reference**: `UNIFI_UDM_PRO_API_DOCUMENTATION.md` - Complete endpoint documentation
- **Architecture**: `docs/DESIGN.md` - Design with mermaid diagrams, service interfaces, type definitions
- **Implementation Plan**: `docs/plan.md` - Phased plan with progress tracking

## Critical Rules

1. **Every function MUST have a test** - No exceptions. Run `make test` to verify.
2. **Every endpoint MUST be supported in the mock server** - Tests use the mock, not real hardware.
3. **No phase advancement without 100% test coverage** - Complete and test each phase before moving on.
4. **Phases are sequential** - Follow `docs/plan.md` in order.

## Concurrency Requirements

- **CSRF tokens**: Use `atomic.Value` for thread-safe storage and updates
- **Session refresh**: Use `sync.Mutex` to prevent concurrent refresh races
- **Rate limiting**: Implement semaphore-based request limiting
- **Connection pooling**: Configure `http.Transport` appropriately

## Architecture Overview

```
gofi/
├── client.go          # Main client, authentication, request handling
├── types.go           # All domain types (Device, Network, WLAN, etc.)
├── errors.go          # Sentinel errors, APIError type
├── services/          # Service implementations
│   ├── site.go
│   ├── device.go
│   ├── network.go
│   ├── wlan.go
│   ├── firewall.go
│   ├── client.go
│   ├── user.go
│   ├── routing.go
│   └── ...
├── mock/              # Mock server for testing
│   ├── server.go
│   ├── handlers.go
│   ├── fixtures/
│   └── scenarios/
└── examples/
```

## Key Technical Details

- **Auth**: Cookie-based session via `POST /api/auth/login`
- **CSRF**: Extract from cookie, send as `X-CSRF-Token` header
- **Base Path**: UDM Pro uses `/proxy/network` prefix
- **API Versions**: v1 (`/api/s/{site}/...`) and v2 (`/v2/api/site/{site}/...`)
- **WebSocket**: Events at `wss://{host}/proxy/network/wss/s/{site}/events`

## Type Patterns

Use flexible types for UniFi's inconsistent JSON:

```go
type FlexInt int64    // Handles "123" or 123
type FlexBool bool    // Handles "true", true, 0, 1
```

## Mock Server

The mock server must:
- Support all endpoints with realistic responses
- Allow fixture loading for consistent test data
- Support error scenarios (auth failures, rate limits, not found)
- Simulate WebSocket events

## gofips Tool

A command-line tool for managing fixed IP + DNS assignments on a UniFi UDM Pro using ISC DHCP host declaration format. Lives in `utilities/gofips/`. Built on the gofi module. Provides hostname support and ISC DHCP format compatibility.

### Purpose

Networks managed by a UDM Pro need a way to bulk-export and bulk-import the complete set of MAC-to-IP-to-hostname bindings. The ISC DHCP `dhcpd.conf` host declaration format is the industry-standard way to represent these bindings. `gofips` reads and writes this format so that administrators can:

- Export the entire UDM fixed-IP table (with hostnames and DNS records) to a file that can be version-controlled, diffed, and edited.
- Import a file of host declarations to provision the UDM in bulk.
- Add or delete individual hosts from the command line using the same format fragment.

### ISC DHCP Host Declaration Format

The canonical format for each host entry is:

```
host myhostname {
    hardware ethernet aa:bb:cc:dd:ee:ff;
    fixed-address 192.168.1.10;
}
```

Rules for parsing and emitting this format:

- `host <hostname> {` opens a declaration. The hostname is a single token (no spaces, DNS-safe characters: `[a-zA-Z0-9._-]`).
- `hardware ethernet <mac>;` specifies the MAC address. Colon-separated hex, case-insensitive on input, lowercase on output.
- `fixed-address <ip>;` specifies the IPv4 address. Only IPv4 is supported.
- `}` closes the declaration.
- Blank lines between declarations are ignored.
- Lines starting with `#` are comments and are ignored on input. On output, a header comment is emitted.
- Whitespace is flexible on input: leading/trailing spaces, tabs, and varying indentation are all accepted. On output, use 4-space indentation as shown above.
- Semicolons after `hardware ethernet` and `fixed-address` values are required.
- Declarations may appear in any order in the file. On output, sort by IP address numerically.
- Other dhcpd.conf directives (subnet, option, etc.) are silently ignored during parsing -- only `host {}` blocks are extracted.

### CLI Interface

```
gofips [connection flags] --get
gofips [connection flags] --set [filename]
gofips [connection flags] --add <host-declaration>
gofips [connection flags] --del --name <hostname>
gofips [connection flags] --del --mac <mac>
gofips [connection flags] --del --ip <ip>
```

### Mode Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--get` | `-g` | Export all fixed IP assignments with hostnames to stdout in ISC DHCP format |
| `--set` | `-s` | Import host declarations from a file or stdin |
| `--add` | `-a` | Add a single host from a declaration fragment (string argument or stdin) |
| `--del` | `-d` | Delete a host identified by `--name`, `--mac`, or `--ip` |

Exactly one mode flag must be specified. Specifying none, or more than one, prints usage and exits with error.

### Identifier Flags (used with `--del`)

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Hostname to delete |
| `--mac` | `-m` | MAC address to delete |
| `--ip` | `-i` | IP address to delete |

Exactly one identifier must be given with `--del`. If a MAC or IP matches multiple entries (should not happen, but defensive), report all matches and require `--force` to proceed.

### Connection Flags

Connection flags:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--host` | `-H` | `$UNIFI_CONTROLLER_IP` | UDM Pro host address |
| `--port` | `-p` | `443` | UDM Pro port |
| `--site` | `-S` | `default` | UniFi site name |
| `--secure` | `-k` | `false` | Enforce TLS certificate verification (default: accept self-signed) |

### Other Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip conflict checks; proceed even if multiple matches found on delete |
| `--keep-dns` | `-K` | Do not delete associated DNS records when deleting a host |
| `--dry-run` | n/a | Show what would be done without making changes |

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `UNIFI_USERNAME` | Yes | UDM authentication username |
| `UNIFI_PASSWORD` | Yes | UDM authentication password |
| `UNIFI_CONTROLLER_IP` | No | UDM host address (fallback if `-H` not given) |

### Behavior: `--get` Mode

1. Connect to the UDM.
2. List all users via `client.Users().List()`. Filter to `UseFixedIP == true` and `FixedIP != ""`.
3. For each user with a fixed IP, determine the hostname:
   - Use `user.Name` if set and DNS-safe.
   - Fall back to `user.Hostname` if set and DNS-safe.
   - If neither is DNS-safe, use the MAC address with colons replaced by hyphens (e.g., `aa-bb-cc-dd-ee-ff`) as the hostname.
4. Optionally cross-reference DNS records via `client.DNS().GetByIP()` to verify hostname matches a DNS A record. If a DNS record exists with a different hostname than the user record, emit a comment warning above that entry.
5. Sort entries by IP address numerically (uint32 conversion).
6. Output to stdout in ISC DHCP format with a header comment:

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

7. If no assignments exist, output a commented example showing the expected format.
8. Status/progress messages go to stderr.

### Behavior: `--set` Mode

1. Parse the entire input file (or stdin) before connecting. Extract all `host {}` blocks. Ignore any non-host directives.
2. Validate every entry:
   - Hostname must be DNS-safe: `[a-zA-Z0-9._-]+`, max 63 characters per label, max 253 total.
   - MAC must be valid colon-separated hex.
   - IP must be valid IPv4.
   - No duplicate hostnames within the file.
   - No duplicate MACs within the file.
   - No duplicate IPs within the file.
   - On any validation error, print all errors with context (line numbers) and exit before connecting.
3. Connect to the UDM.
4. Fetch existing users and DNS records.
5. For each host declaration:
   - **Skip if unchanged**: MAC already has the same IP and hostname. Print skip to stderr.
   - **Update if changed**: MAC exists but IP or hostname differs. Update the user record and DNS record. Print update to stderr.
   - **Create if new**: MAC has no existing user. Create user with fixed IP, create DNS A record. Print create to stderr.
6. Network auto-detection: determine which UDM network contains each IP by checking subnets.
7. Print summary to stderr: `N processed, N skipped, N created, N updated, N errors`.
8. Exit 1 if any errors occurred.

### Behavior: `--add` Mode

Add a single host using an ISC DHCP declaration fragment. The fragment can be provided as:

- A positional argument (quoted string):
  ```bash
  gofips -H 192.168.1.1 -k --add 'host mydevice {
      hardware ethernet aa:bb:cc:dd:ee:ff;
      fixed-address 192.168.1.50;
  }'
  ```
- Stdin (if no positional argument):
  ```bash
  echo 'host mydevice {
      hardware ethernet aa:bb:cc:dd:ee:ff;
      fixed-address 192.168.1.50;
  }' | gofips -H 192.168.1.1 -k --add
  ```

Behavior:

1. Parse the single host declaration. Validate hostname, MAC, IP.
2. Connect to the UDM.
3. Check for conflicts (unless `--force`):
   - Is the IP already assigned to a different MAC?
   - Is the MAC already assigned a different IP?
   - Does a DNS record with this hostname already point to a different IP?
4. Create or update the user record with fixed IP.
5. Create or update the DNS A record for the hostname.
6. Print the result to stdout:
   ```
   Created: mydevice aa:bb:cc:dd:ee:ff 192.168.1.50
   ```

### Behavior: `--del` Mode

Delete a host identified by one of `--name`, `--mac`, or `--ip`.

1. Connect to the UDM.
2. Find the matching entry:
   - `--name <hostname>`: Search users by name/hostname, and DNS records by key. The user whose name matches AND/OR the DNS record whose key matches identifies the target.
   - `--mac <mac>`: Look up user by MAC via `client.Users().GetByMAC()`.
   - `--ip <ip>`: Search users where `FixedIP == ip`.
3. If no match found, print error and exit 1.
4. If multiple matches found (possible with `--ip` if data is inconsistent), list all matches and exit 1 unless `--force` is set.
5. Display what will be deleted and ask for confirmation (unless stdout is not a terminal, in which case proceed without prompting).
6. Clear the fixed IP assignment from the user record (do not delete the user itself unless `--force` is also set).
7. Delete associated DNS A records pointing to the fixed IP (unless `--keep-dns`).
8. Print the result to stdout:
   ```
   Deleted: mydevice aa:bb:cc:dd:ee:ff 192.168.1.50
     Removed fixed IP assignment
     Deleted DNS record: mydevice -> 192.168.1.50
   ```

### Error Handling

| Condition | Behavior |
|-----------|----------|
| No mode flag or multiple mode flags | Print usage, exit 1 |
| Missing credentials or host | Print error, exit 1 |
| Connection failure | Print error, exit 1 |
| Invalid host declaration syntax | Print error with line number, exit 1 |
| Duplicate hostname/MAC/IP in input file | Print all duplicates, exit 1 before connecting |
| Conflict on `--add` without `--force` | Print conflict details, exit 1 |
| No match on `--del` | Print error, exit 1 |
| Multiple matches on `--del` without `--force` | List matches, exit 1 |
| Network auto-detect failure | Print error for that entry, continue with remaining |
| Individual create/update failure in `--set` | Print error to stderr, continue, exit 1 at end |

### Relationship to Existing Tools

- `gofips` replaces the existing `addfixedip`, `delfixedip`, and `fixedips` examples.
- Those examples remain as reference implementations but `gofips` is the production tool.
- `gofips` adds hostname/DNS management that the flat-format tools lack.
- `gofips` uses ISC DHCP format instead of the flat `IP MAC` format, making it compatible with existing dhcpd.conf files and network documentation.

### Project Layout

```
utilities/
  gofips/
    main.go           # Entry point, flag parsing, mode dispatch
    parse.go          # ISC DHCP format parser
    format.go         # ISC DHCP format output
    operations.go     # get/set/add/del business logic
  docs/
    gofips/
      DESIGN.md       # Detailed design document (generated from this spec)
```

## gofimac Tool

A command-line tool for listing connected clients on a UniFi UDM Pro, filtered by connection type (wired or WiFi), with independent OUI manufacturer lookup. Lives in `utilities/gofimac/`. Built on the gofi module.

### Purpose

Network administrators need to see what devices are connected and who manufactures them. The UDM provides an OUI field on client records, but it is often stale or inaccurate. `gofimac` performs its own OUI lookup using the IEEE OUI database, ensuring manufacturer identification is always current and trustworthy.

### OUI Database

The IEEE publishes the canonical OUI (Organizationally Unique Identifier) database at `https://standards-oui.ieee.org/oui/oui.txt`. The first 3 octets of a MAC address identify the manufacturer.

**Storage location**: `$XDG_DATA_HOME/gofimac/oui.txt`, falling back to `~/.local/share/gofimac/oui.txt` if `$XDG_DATA_HOME` is not set. This follows the XDG Base Directory Specification and requires no root access.

**Freshness check**: On every invocation, before performing any lookups:

1. Check if the OUI file exists at the storage location.
2. If it exists, check the file modification time. If the file is older than 30 days, re-download it.
3. If it does not exist, download it.
4. Download from `https://standards-oui.ieee.org/oui/oui.txt`.
5. If the download fails and a cached file exists (even if stale), use the cached file and print a warning to stderr.
6. If the download fails and no cached file exists, exit with an error.

**Parsing**: The IEEE OUI file format has entries like:

```
AA-BB-CC   (hex)		Acme Corporation
AABBCC     (base 16)		Acme Corporation
				123 Main Street
				Springfield IL 12345
				US
```

Parse only the `(hex)` lines. Extract the 3-octet prefix (normalized to lowercase colon-separated, e.g., `aa:bb:cc`) and the manufacturer name.

**Lookup**: Given a MAC address, extract the first 3 octets and look up the manufacturer. If not found, return `unknown`.

### CLI Interface

```
gofimac [connection flags] --wifi
gofimac [connection flags] --wired
gofimac [connection flags] --all
gofimac [connection flags] --wifi -j
```

### Mode Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--wifi` | `-w` | List only WiFi-connected clients |
| `--wired` | `-e` | List only wired (ethernet) clients |
| `--all` | `-a` | List all connected clients (default if no mode specified) |

If no mode flag is given, `--all` is assumed.

### Output Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output in JSON format instead of text |

### Connection Flags

Same as `gofips`:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--host` | `-H` | `$UNIFI_CONTROLLER_IP` | UDM Pro host address |
| `--port` | `-p` | `443` | UDM Pro port |
| `--site` | `-S` | `default` | UniFi site name |
| `--secure` | `-k` | `false` | Enforce TLS certificate verification (default: accept self-signed) |

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `UNIFI_USERNAME` | Yes | UDM authentication username |
| `UNIFI_PASSWORD` | Yes | UDM authentication password |
| `UNIFI_CONTROLLER_IP` | No | UDM host address (fallback if `-H` not given) |
| `XDG_DATA_HOME` | No | Base directory for OUI data (default `~/.local/share`) |

### Text Output Format

Default output is one line per client, tab-separated:

```
MAC              IP              HOSTNAME        OUI-MANUFACTURER
```

Concrete example:

```
aa:bb:cc:dd:ee:01	192.168.1.10	myserver	Dell Inc.
aa:bb:cc:dd:ee:02	192.168.1.11	printer 	Hewlett Packard
aa:bb:cc:dd:ee:03	192.168.1.12	unknown 	unknown
```

Rules:
- MAC addresses are lowercase, colon-separated.
- If the client has a `Name` field set, use it as the hostname. Otherwise fall back to `Hostname`. If neither is set, use `unknown`.
- The OUI manufacturer comes from our own OUI database lookup, never from the UDM's `OUI` field.
- Sort output by IP address numerically (same uint32 conversion as gofips).
- Clients without an IP address are listed at the end with IP shown as `-`.
- Status/progress messages (OUI download progress, etc.) go to stderr.

### JSON Output Format

When `--json` or `-j` is specified, output a JSON array to stdout:

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

For wired clients, include switch-specific fields instead of WiFi fields:

```json
{
  "mac": "aa:bb:cc:dd:ee:02",
  "ip": "192.168.1.11",
  "hostname": "printer",
  "manufacturer": "Hewlett Packard",
  "is_wired": true,
  "sw_mac": "11:22:33:44:55:66",
  "sw_port": 4,
  "rx_bytes": 123456789,
  "tx_bytes": 987654321,
  "uptime": 172800,
  "last_seen": 1708700000
}
```

JSON fields:
- Always present: `mac`, `ip`, `hostname`, `manufacturer`, `is_wired`, `rx_bytes`, `tx_bytes`, `uptime`, `last_seen`.
- WiFi clients add: `essid`, `ap_mac`, `channel`, `radio`, `radio_proto`, `signal`, `noise`, `rssi`, `satisfaction`.
- Wired clients add: `sw_mac`, `sw_port`.
- Omit fields with zero/empty values from JSON output (`omitempty` behavior).

### Behavior

1. Check and update the OUI database (see Freshness check above).
2. Parse the OUI database into a lookup map (3-octet prefix -> manufacturer name).
3. Connect to the UDM.
4. Fetch active clients via `client.Clients().ListActive()`.
5. Filter by connection type based on the `IsWired` field:
   - `--wifi`: `IsWired == false`
   - `--wired`: `IsWired == true`
   - `--all`: no filter
6. For each client, look up the manufacturer from the OUI map using the first 3 octets of the MAC.
7. Sort by IP address numerically (clients without IPs sort last).
8. Output in the requested format (text or JSON).

### Error Handling

| Condition | Behavior |
|-----------|----------|
| Missing credentials or host | Print error, exit 1 |
| Connection failure | Print error, exit 1 |
| OUI download fails, no cache | Print error, exit 1 |
| OUI download fails, stale cache exists | Print warning to stderr, continue with cached data |
| OUI parse error | Print error, exit 1 |
| No clients found matching filter | Print empty output (empty line in text, `[]` in JSON), exit 0 |

### Project Layout

```
utilities/
  gofimac/
    main.go           # Entry point, flag parsing
    oui.go            # OUI database download, parse, lookup
    format.go         # Text and JSON output formatting
    operations.go     # Client listing and filtering
  docs/
    gofimac/
      DESIGN.md       # Detailed design document
```

## Commands

```bash
make test      # Run all tests
make lint      # Run linter
make build     # Build the module
make coverage  # Generate coverage report
```

## Reference Implementations

Study these for patterns:
- `github.com/paultyng/go-unifi` - Terraform provider, CRUD patterns
- `github.com/unpoller/unifi` - FlexInt/FlexBool types
- `github.com/thib3113/unifi-client` - TypeScript, comprehensive types
