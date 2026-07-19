# gofips

Command-line tool for managing fixed IP + DNS assignments on a UniFi UDM Pro using ISC DHCP host declaration format. Built on the [gofi](https://github.com/unifi-go/gofi) module.

## Overview

`gofips` reads and writes ISC DHCP `dhcpd.conf`-style host declarations to bulk-export, bulk-import, add, and delete MAC-to-IP-to-hostname bindings on a UniFi UDM Pro.

## Build

```bash
make build
```

The binary is built to `bin/gofips`.

## Usage

### Export assignments

```bash
gofips -H 192.168.1.1 -g > hosts.conf
```

Exports all fixed IP assignments with hostnames to stdout in ISC DHCP format:

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

### Import assignments

```bash
gofips -H 192.168.1.1 -s hosts.conf
```

Or from stdin:

```bash
cat hosts.conf | gofips -H 192.168.1.1 -s
```

Parses all `host {}` blocks, validates them, then creates or updates each entry on the UDM. Unchanged entries are skipped.

### Add a single host

```bash
gofips -H 192.168.1.1 -a 'host mydevice { hardware ethernet aa:bb:cc:dd:ee:ff; fixed-address 192.168.1.50; }'
```

### Delete a host

```bash
gofips -H 192.168.1.1 -d -n mydevice      # by hostname
gofips -H 192.168.1.1 -d -m aa:bb:cc:dd:ee:ff  # by MAC
gofips -H 192.168.1.1 -d -i 192.168.1.50   # by IP
```

Use `--keep-dns` to preserve DNS records when deleting. Use `--force` to delete the user record entirely rather than just clearing the fixed IP.

### Dry run

Add `--dry-run` to `--set` to preview changes without modifying the UDM:

```bash
gofips -H 192.168.1.1 -s --dry-run hosts.conf
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `UNIFI_USERNAME` | Yes | UDM authentication username |
| `UNIFI_PASSWORD` | Yes | UDM authentication password |
| `UNIFI_CONTROLLER_IP` | No | UniFi controller address (fallback if `-H` not given) |

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--get` | `-g` | Export fixed IP assignments to stdout |
| `--set` | `-s` | Import host declarations from file or stdin |
| `--add` | `-a` | Add a single host from ISC DHCP declaration |
| `--del` | `-d` | Delete a host by identifier |
| `--name` | `-n` | Hostname identifier for `--del` |
| `--mac` | `-m` | MAC address identifier for `--del` |
| `--ip` | `-i` | IP address identifier for `--del` |
| `--host` | `-H` | UniFi controller address |
| `--port` | `-p` | UniFi controller port (default 443) |
| `--site` | `-S` | UniFi site name (default "default") |
| `--secure` | `-k` | Enforce TLS certificate verification (default: accept self-signed) |
| `--force` | `-f` | Skip conflict checks |
| `--keep-dns` | `-K` | Preserve DNS records on delete |
| `--dry-run` | | Preview changes without applying |

## Testing

```bash
go test -v -race ./utilities/gofips/
```

## Architecture

See [docs/gofips/DESIGN.md](../docs/gofips/DESIGN.md) for the detailed design document.

```
gofips/
  main.go           # Entry point, flag parsing, mode dispatch
  parse.go          # ISC DHCP format parser
  format.go         # ISC DHCP format output (sorted by IP)
  operations.go     # get/set/add/del business logic
```
