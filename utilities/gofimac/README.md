# gofimac

Command-line tool for listing connected clients on a UniFi UDM Pro with independent OUI manufacturer lookup. Built on the [gofi](https://github.com/unifi-go/gofi) module.

## Overview

`gofimac` fetches active clients from a UDM Pro and displays their MAC address, IP, hostname, and manufacturer. Manufacturer identification uses the IEEE OUI database, not the UDM's built-in fingerprinting.

It also surfaces device history: how long each device has been known (`first_seen`), how recently it was seen (`last_seen`), and — via `--since`/`--gone` — which devices have recently left the network. Output is sorted newest-first by default so new arrivals appear at the top.

## Build

```bash
make utilities
```

The binary is built to `bin/gofimac`.

## Usage

### List all clients

```bash
gofimac -H 192.168.1.1
```

### List WiFi clients only

```bash
gofimac -H 192.168.1.1 -w
```

### List wired clients only

```bash
gofimac -H 192.168.1.1 -e
```

### Recently departed devices

```bash
gofimac -H 192.168.1.1 --gone         # departed within the default 7-day window
gofimac -H 192.168.1.1 --gone=30d     # departed within the last 30 days
```

### Devices seen in a window (present and gone)

```bash
gofimac -H 192.168.1.1 --since 7d     # everything seen in the last 7 days, marked present/gone
```

### Probe a single MAC

Check whether one device is on the network right now. Exit code is `0` if present,
`1` if gone or never seen — so it works like `ping` in scripts. Unlike ARP-based
tools, this asks the UDM, so it works across VLANs/subnets.

```bash
gofimac -H 192.168.1.1 --mac aa:bb:cc:dd:ee:ff   # or -m
```

```
MAC                IP            HOSTNAME  OUI-MANUFACTURER              AGE  LAST-SEEN  STATUS
9c:69:d3:00:7c:5e  192.168.4.30  bs        ASIX Electronics Corporation  6mo  now        present
```

A departed device still reports its last-known record with `STATUS gone` and a
non-zero exit. For a local-segment ARP probe instead (same L2 only, needs root),
use the `make mac-ping MAC=<mac>` target at the repo root.

### Sorting

```bash
gofimac -H 192.168.1.1 --sort ip          # numeric IP order (the pre-history default)
gofimac -H 192.168.1.1 --sort last-seen   # most recently seen first
```

Default sort is `first-seen` descending (newest devices on top).

### JSON output

```bash
gofimac -H 192.168.1.1 -j
gofimac -H 192.168.1.1 -w -j
gofimac -H 192.168.1.1 --gone=30d -j
```

### Text output format

Active view adds `AGE` (relative `first_seen`) and `LAST-SEEN` columns:

```
MAC                 IP             HOSTNAME   OUI-MANUFACTURER   AGE   LAST-SEEN
aa:bb:cc:dd:ee:01   192.168.1.10   myserver   Dell Inc.          2h    now
aa:bb:cc:dd:ee:02   192.168.1.11   printer    Hewlett Packard    5d    now
aa:bb:cc:dd:ee:03   -              unknown    unknown            3mo   now
```

`--since`/`--gone` add a `STATUS` column (`present` or `gone`). Departed devices are
sourced from the UDM's historical record, so their IP and link fields may be blank
or stale.

## OUI Database

The IEEE OUI database is automatically downloaded and cached at:
- `$XDG_DATA_HOME/gofimac/oui.txt` (if `XDG_DATA_HOME` is set)
- `~/.local/share/gofimac/oui.txt` (default)

The database is refreshed automatically if older than 30 days. If a download fails and a cached copy exists, the cached copy is used with a warning.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `UNIFI_API_KEY` | No | Cloud API key (preferred if set); requires `UNIFI_CONSOLE_ID` |
| `UNIFI_CONSOLE_ID` | With key | Site Manager console ID (connector mode), from `GET https://api.ui.com/v1/hosts` |
| `UNIFI_USERNAME` | If no key | UDM authentication username (fallback) |
| `UNIFI_PASSWORD` | If no key | UDM authentication password (fallback) |
| `UNIFI_CONTROLLER_IP` | No | UniFi controller address (fallback if `-H` not given; username/password mode only) |
| `XDG_DATA_HOME` | No | Base directory for OUI data (default `~/.local/share`) |

If `UNIFI_API_KEY` is set, it is used and `UNIFI_USERNAME`/`UNIFI_PASSWORD` are ignored.

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--wifi` | `-w` | List only WiFi clients |
| `--wired` | `-e` | List only wired clients |
| `--all` | `-a` | List all clients (default) |
| `--since` | | Show devices seen within a window (present + gone), e.g. `7d`, `24h`, `3mo` |
| `--gone` | | Show only departed devices; optional window (`--gone=30d`), default `7d` |
| `--sort` | | Sort order: `first-seen` (default), `last-seen`, or `ip` |
| `--mac` | `-m` | Probe one MAC; exit 0 if present, 1 if gone/not found |
| `--json` | `-j` | Output in JSON format |
| `--host` | `-H` | UniFi controller address |
| `--port` | `-p` | UniFi controller port (default 443) |
| `--site` | `-S` | UniFi site name (default "default") |
| `--secure` | `-k` | Enforce TLS certificate verification (default: accept self-signed) |

`--since`, `--gone`, and `--mac` are mutually exclusive. Duration values accept `s`, `m`, `h`,
`d`, `w`, and `mo` units and may be compound (e.g. `1w2d`). Because `--gone` takes an
optional value, its window must be attached with `=` (`--gone=30d`, not `--gone 30d`).

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
  operations.go     # Client listing, filtering, sorting, presence
  duration.go       # Compound duration parser (s/m/h/d/w/mo)
  format.go         # Text and JSON output, relative-time rendering
```
