# gofip - Fixed IP Assignment Manager

A command-line tool for managing DHCP fixed IP assignments on a UniFi UDM Pro. Replaces the workflow of manually editing `dhcpd.conf` and DNS zone files for small networks by storing all assignments in the UDM controller directly.

## Overview

`gofip` provides two operations:

- **Get** (`--get`, `-g`): Export current fixed IP assignments from the UDM in a flat-file format, sorted by IP address. If no assignments exist, output a commented example showing the expected format.
- **Set** (`--set`, `-s`): Import fixed IP assignments from a file or stdin. Existing assignments (same MAC already has the same IP) are skipped. New assignments are created.

The file format is deliberately simple so it can be version-controlled, diffed, and edited with any text editor.

## File Format

One assignment per line. Each line contains an IPv4 address and a MAC address separated by whitespace. Lines are ordered by IP address.

```
# gofip fixed IP assignments
# format: IP MAC
192.168.1.10 aa:bb:cc:dd:ee:01
192.168.1.11 aa:bb:cc:dd:ee:02
192.168.1.20 11:22:33:44:55:66
192.168.10.5 de:ad:be:ef:00:01
```

Rules:

- Lines starting with `#` are comments and are ignored on input.
- Blank lines are ignored on input.
- MAC addresses are colon-separated, lowercase hex (e.g., `aa:bb:cc:dd:ee:ff`). Uppercase is accepted on input and normalized to lowercase.
- IP addresses are IPv4 dotted-quad only.
- On output (`--get`), lines are sorted by IP address using numeric comparison (not lexicographic), so `192.168.1.9` sorts before `192.168.1.10`.

## CLI Interface

```
gofip [connection flags] --get
gofip [connection flags] --set [filename]
```

### Mode Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--get` | `-g` | Export assignments to stdout |
| `--set` | `-s` | Import assignments from file or stdin |

Exactly one of `--get` or `--set` must be specified. If both or neither are given, the tool prints usage and exits with an error.

### Connection Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--host` | `-H` | `$UNIFI_UDM_IP` | UDM Pro host address |
| `--port` | `-p` | `443` | UDM Pro port |
| `--site` | `-S` | `default` | UniFi site name |
| `--insecure` | `-k` | `false` | Skip TLS certificate verification |

Note: `-s` is taken by `--set`, so `--site` uses `-S`.

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `UNIFI_USERNAME` | Yes | UDM authentication username |
| `UNIFI_PASSWORD` | Yes | UDM authentication password |
| `UNIFI_UDM_IP` | No | UDM host address (fallback if `-H` not given) |

### Input for `--set`

When `--set` is specified:

- If a filename argument follows `--set` (i.e., a non-flag argument), read from that file.
- If no filename is provided, read from stdin.

This supports both patterns:

```bash
gofip -H 192.168.1.1 -k --set hosts.txt
cat hosts.txt | gofip -H 192.168.1.1 -k --set
```

## Behavior

### `--get` Mode

1. Connect to the UDM.
2. List all users via `client.Users().List()`.
3. Filter to users where `UseFixedIP == true` and `FixedIP != ""`.
4. If no assignments exist, output a commented example to stdout:
   ```
   # gofip fixed IP assignments
   # format: IP MAC
   # No fixed IP assignments found on UDM.
   # Example:
   # 192.168.1.10 aa:bb:cc:dd:ee:ff
   # 192.168.1.11 11:22:33:44:55:66
   ```
5. If assignments exist, output them as uncommented data lines with a header comment:
   ```
   # gofip fixed IP assignments
   # format: IP MAC
   192.168.1.10 aa:bb:cc:dd:ee:01
   192.168.1.11 aa:bb:cc:dd:ee:02
   192.168.1.20 11:22:33:44:55:66
   ```
6. Sorting is numeric by IP octets (split on `.`, compare each octet as an integer).
7. All output goes to stdout. Status/progress messages go to stderr.

### `--set` Mode

1. Parse the input file (or stdin) into a list of `(IP, MAC)` pairs.
2. Validate every entry before connecting:
   - IP must be a valid IPv4 address.
   - MAC must match `xx:xx:xx:xx:xx:xx` hex format.
   - No duplicate IPs within the file (exit with error listing the duplicates).
   - No duplicate MACs within the file (exit with error listing the duplicates).
3. Connect to the UDM.
4. Fetch existing assignments: `client.Users().List()`, filtered to `UseFixedIP == true`.
5. Build a map of existing assignments keyed by MAC address.
6. Fetch available networks: `client.Networks().List()` for subnet auto-detection.
7. For each `(IP, MAC)` pair in the input:
   - **Skip if unchanged**: If the MAC already has a fixed IP assignment with the same IP, print a skip message to stderr and continue.
   - **Update if changed**: If the MAC exists but has a different fixed IP, update it to the new IP. Print a note to stderr.
   - **Create if new**: If the MAC has no existing user entry or no fixed IP, create/set the assignment.
   - **Network detection**: Determine which network the IP belongs to by checking which network's subnet contains the IP. If no matching network is found, report an error for that entry and continue with the remaining entries.
8. Print a summary to stderr:
   ```
   Summary: 15 processed, 10 skipped (unchanged), 3 created, 2 updated, 0 errors
   ```

## IP Address Sort Order

IP addresses are sorted numerically by octet. This is implemented by converting each IP to a 32-bit integer for comparison:

```
192.168.1.2   -> sort position before
192.168.1.10  -> sort position after 192.168.1.2
192.168.1.100 -> sort position after 192.168.1.10
192.168.10.1  -> sort position after 192.168.1.x
```

## Error Handling

| Condition | Behavior |
|-----------|----------|
| Missing `--get` or `--set` | Print usage, exit 1 |
| Both `--get` and `--set` | Print usage, exit 1 |
| Missing credentials | Print error, exit 1 |
| Connection failure | Print error, exit 1 |
| Invalid line in input | Print error with line number, exit 1 (fail fast, before connecting) |
| Duplicate IP in input | Print error listing duplicates, exit 1 |
| Duplicate MAC in input | Print error listing duplicates, exit 1 |
| Network auto-detect failure | Print warning for that entry to stderr, skip it, continue |
| Individual set failure | Print warning to stderr, continue with remaining entries |

Validation errors in the input file cause an immediate exit before any UDM connection is made. This prevents partial application of a malformed file.

## Project Layout

```
utilities/
  gofip/
    main.go         # Entry point, flag parsing, mode dispatch
  docs/
    gofip/
      DESIGN.md     # This file
```

## Examples

### Export existing assignments to a file

```bash
export UNIFI_USERNAME=admin
export UNIFI_PASSWORD=secret

gofip -H 192.168.1.1 -k --get > hosts.txt
```

### Edit and re-apply

```bash
# Export current state
gofip -H 192.168.1.1 -k -g > hosts.txt

# Edit the file (add new entries, remove is not supported via set)
vim hosts.txt

# Apply changes (existing entries are skipped)
gofip -H 192.168.1.1 -k -s hosts.txt
```

### Pipe from stdin

```bash
echo "192.168.1.50 aa:bb:cc:dd:ee:ff" | gofip -H 192.168.1.1 -k --set
```

### First-time setup with no existing assignments

```bash
# Get the example format
gofip -H 192.168.1.1 -k -g > hosts.txt
# File contains commented examples showing the format

# Fill in your assignments
cat > hosts.txt << 'EOF'
# gofip fixed IP assignments
# format: IP MAC
192.168.1.10 aa:bb:cc:dd:ee:01
192.168.1.11 aa:bb:cc:dd:ee:02
192.168.1.20 11:22:33:44:55:66
EOF

# Apply
gofip -H 192.168.1.1 -k -s hosts.txt
```

### Use with version control

```bash
# Check in your assignments
gofip -H 192.168.1.1 -k -g > hosts.txt
git add hosts.txt
git commit -m "current fixed IP assignments"

# Later, apply from version control
gofip -H 192.168.1.1 -k -s hosts.txt
```
