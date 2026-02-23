# gofimac

Command-line tool for listing connected clients on a UniFi UDM Pro with independent OUI manufacturer lookup. Built on the [gofi](https://github.com/unifi-go/gofi) module.

## Overview

`gofimac` fetches active clients from a UDM Pro and displays their MAC address, IP, hostname, and manufacturer. Manufacturer identification uses the IEEE OUI database, not the UDM's built-in fingerprinting.

## Build

```bash
make utilities
```

The binary is built to `bin/gofimac`.

## Usage

### List all clients

```bash
gofimac -H 192.168.1.1 -k
```

### List WiFi clients only

```bash
gofimac -H 192.168.1.1 -k -w
```

### List wired clients only

```bash
gofimac -H 192.168.1.1 -k -e
```

### JSON output

```bash
gofimac -H 192.168.1.1 -k -j
gofimac -H 192.168.1.1 -k -w -j
```

### Text output format

```
aa:bb:cc:dd:ee:01	192.168.1.10	myserver	Dell Inc.
aa:bb:cc:dd:ee:02	192.168.1.11	printer	Hewlett Packard
aa:bb:cc:dd:ee:03	-	unknown	unknown
```

## OUI Database

The IEEE OUI database is automatically downloaded and cached at:
- `$XDG_DATA_HOME/gofimac/oui.txt` (if `XDG_DATA_HOME` is set)
- `~/.local/share/gofimac/oui.txt` (default)

The database is refreshed automatically if older than 30 days. If a download fails and a cached copy exists, the cached copy is used with a warning.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `UNIFI_USERNAME` | Yes | UDM authentication username |
| `UNIFI_PASSWORD` | Yes | UDM authentication password |
| `UNIFI_UDM_IP` | No | UDM host address (fallback if `-H` not given) |
| `XDG_DATA_HOME` | No | Base directory for OUI data (default `~/.local/share`) |

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--wifi` | `-w` | List only WiFi clients |
| `--wired` | `-e` | List only wired clients |
| `--all` | `-a` | List all clients (default) |
| `--json` | `-j` | Output in JSON format |
| `--host` | `-H` | UDM Pro host address |
| `--port` | `-p` | UDM Pro port (default 443) |
| `--site` | `-S` | UniFi site name (default "default") |
| `--insecure` | `-k` | Skip TLS certificate verification |

## Testing

```bash
go test -v -race ./utilities/gofimac/
```

## Architecture

See [docs/gofimac/DESIGN.md](../docs/gofimac/DESIGN.md) for the detailed design document.

```
gofimac/
  main.go           # Entry point, flag parsing
  oui.go            # OUI database download, parse, lookup
  operations.go     # Client listing, filtering, sorting
  format.go         # Text and JSON output
```
