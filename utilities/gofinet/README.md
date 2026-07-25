# gofinet

Command-line tool for listing UniFi networks with their subnet and DHCP dynamic
address pool. Built on the [gofi](https://github.com/unifi-go/gofi) module.

## Overview

`gofinet` answers "what is the range of dynamically-assigned IP addresses?" for
each network on a UniFi controller. It reports every network's subnet, DHCP
dynamic pool (`dhcpd_start`–`dhcpd_stop`), lease time, advertised gateway, and
DNS servers. This is the companion to `gofips`: the pool boundaries tell you
which addresses are safe to hand out as static reservations (anything in the
subnet *outside* the dynamic pool).

## Build

```bash
make utilities
```

The binary is built to `bin/gofinet`.

## Usage

### List all networks

```bash
gofinet -H 192.168.1.1
```

```
NETWORK     VLAN  SUBNET           DHCP-POOL                        LEASE   GATEWAY  DNS
Default     -     192.168.4.1/24   192.168.4.100 - 192.168.4.189    86400s  -        1.1.1.1,8.8.8.8
cj-iot      2     192.168.10.1/24  192.168.10.100 - 192.168.10.200  86400s  -        -
Internet 1  -     -                (disabled)                       -       -        -
```

- Networks with no DHCP server (WAN, vlan-only) show `(disabled)` in the pool column.
- `VLAN` shows `-` for untagged/default networks.

### JSON output

```bash
gofinet -H 192.168.1.1 -j
```

Each object contains `name`, `purpose`, `vlan`, `subnet`, `enabled`,
`dhcp_enabled`, `dhcp_start`, `dhcp_stop`, `dhcp_lease`, `gateway`, and `dns`
(empty/zero fields are omitted).

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `UNIFI_API_KEY` | No | Cloud API key (preferred if set); requires `UNIFI_CONSOLE_ID` |
| `UNIFI_CONSOLE_ID` | With key | Site Manager console ID (connector mode), from `GET https://api.ui.com/v1/hosts` |
| `UNIFI_USERNAME` | If no key | Controller authentication username (fallback) |
| `UNIFI_PASSWORD` | If no key | Controller authentication password (fallback) |
| `UNIFI_CONTROLLER_IP` | No | Controller host address (fallback if `-H` not given; username/password mode only) |

If `UNIFI_API_KEY` is set, it is used and `UNIFI_USERNAME`/`UNIFI_PASSWORD` are ignored.

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output in JSON format |
| `--host` | `-H` | UniFi controller address |
| `--port` | `-p` | UniFi controller port (default 443) |
| `--site` | `-S` | UniFi site name (default "default") |
| `--secure` | `-k` | Enforce TLS certificate verification (default: accept self-signed) |

## Testing

```bash
go test -v -race ./utilities/gofinet/
```

## Architecture

```
gofinet/
  main.go           # Entry point, flag parsing, connection
  operations.go     # Network listing and flattening
  format.go         # Text and JSON output
```
