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

A command-line tool for managing fixed IP + DNS assignments on a UniFi UDM Pro using ISC DHCP host declaration format. Lives in `utilities/gofips/`. Built on the gofi module. Replaces and extends the existing `gofip` tool by adding hostname support and ISC DHCP format compatibility.

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

Same as `gofip`:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--host` | `-H` | `$UNIFI_UDM_IP` | UDM Pro host address |
| `--port` | `-p` | `443` | UDM Pro port |
| `--site` | `-S` | `default` | UniFi site name |
| `--insecure` | `-k` | `false` | Skip TLS certificate verification |

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
| `UNIFI_UDM_IP` | No | UDM host address (fallback if `-H` not given) |

### Behavior: `--get` Mode

1. Connect to the UDM.
2. List all users via `client.Users().List()`. Filter to `UseFixedIP == true` and `FixedIP != ""`.
3. For each user with a fixed IP, determine the hostname:
   - Use `user.Name` if set and DNS-safe.
   - Fall back to `user.Hostname` if set and DNS-safe.
   - If neither is DNS-safe, use the MAC address with colons replaced by hyphens (e.g., `aa-bb-cc-dd-ee-ff`) as the hostname.
4. Optionally cross-reference DNS records via `client.DNS().GetByIP()` to verify hostname matches a DNS A record. If a DNS record exists with a different hostname than the user record, emit a comment warning above that entry.
5. Sort entries by IP address numerically (same uint32 conversion as `gofip`).
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
6. Network auto-detection: determine which UDM network contains each IP by checking subnets, same as `gofip`.
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

- `gofips` replaces the existing `gofip`, `addfixedip`, `delfixedip`, and `fixedips` examples.
- The existing tools remain as reference implementations but `gofips` is the production tool.
- `gofips` adds hostname/DNS management that `gofip` lacks.
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
