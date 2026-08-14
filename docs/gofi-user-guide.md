# gofi User Guide

`gofi` manages a UniFi controller (UDM Pro and compatible consoles) over its JSON API:
fixed-IP reservations, DNS records, networks, connected clients, known-client records,
and whole-site profiles. This guide documents every command, the shape of the command
tree, and how authentication and configuration work.

## Command structure

Every invocation has the same shape:

```
gofi [global flags] <area> <action> [args] [action flags]
```

- **area** — the thing being managed: `ips`, `dns`, `network`, `clients`, `users`,
  `profile`, `config`.
- **action** — a verb: `list`, `show`, `add`, `rm`, `export`, `import`, `clear`, `init`,
  `vendor`, `targets`.

This is the same shape `git`, `docker`, and `kubectl` converged on, and the same shape
gofi's sibling tool [`gogl`](https://github.com/emergingrobotics/gogl) uses for GL.iNet
routers — both follow a shared convention
([`network-cli-convention.md`](../../network-cli-convention.md)) so the command grammar,
flag names, and exit-code contract are the same regardless of which backend you're
driving.

### Why the areas are shaped this way

**`ips` and `dns` are split, not one area**, because they're a paired write to two
different backend objects. A fixed-IP reservation touches the user record's `fixed_ip`
field *and* a DNS A record for the hostname — `ips add`/`ips rm` keep both in step in
one command, the way you'd normally want. `dns` exists on its own for the narrower case
where only the DNS side should move without touching a reservation something else still
depends on — correcting a stale record, say, without releasing the address it points at.

**`clients` and `users` are two separate areas, permanently**, even though both sound
like "devices on my network." They read different objects on the controller: `clients`
is live station telemetry (`stamgr`) — signal, uptime, which AP or switch port — and
only exists for a station that's *currently* connected. `users` is the controller's
persistent identity record for a client, which outlives the connection and carries the
fixed-IP relationship. A device can have a `users` record with no live `clients` entry
(it's not connected right now) or vice versa. Merging them would mean synthesizing a
combined view that corresponds to no single API call, so they stay separate — `clients`
answers "what's connected right now," `users` answers "what does the controller know
about, connected or not."

**`network` is read-only today.** GL.iNet-style LAN writes exist in gofi's sibling tool
`gogl`, but a verified write endpoint for UniFi network/VLAN settings hasn't been
exercised against real hardware yet — `network set` doesn't exist because nothing has
confirmed what it would actually need to send. This is a documented gap, not a
permanent design choice.

**`profile` is its own area**, not folded into any of the above, because it spans three
of them: it captures networks, WLANs, and fixed IPs into one JSON file and applies that
file back. It's deliberately narrow — it does *not* capture devices, firewall rules, or
routing, and never will as currently designed, so "profile" can't silently grow to mean
"the whole site" as new areas get added later.

**`config` is the only area that never touches a controller.** It reads and writes
`~/.config/gofi/config.toml` on your machine — named targets, a default, an output
preference. Every other area acts on the device.

## Authentication

gofi supports two ways to authenticate. **A cloud API key through Ubiquiti's Site
Manager connector is the recommended, primary path** — see [the README's Step
1](../README.md#step-1--get-a-unifi-api-key-recommended) for how to create one. It works
even when the controller isn't directly reachable, needs no session cookies, and is
where most of gofi's testing and daily use happens:

```bash
export UNIFI_API_KEY=...
export UNIFI_CONSOLE_ID=...
gofi network list
```

A **local username/password** path exists as a secondary, less-tested alternative for
controllers with no route to `api.ui.com` — see
[`docs/alternate-local-api.md`](alternate-local-api.md) for the full story, including
TLS/self-signed-certificate handling and how it interacts with `--secure`.

There is no `--password` or `--api-key` flag on any command, ever. A secret on the
command line is visible to every other user through `ps` and lands in shell history.
Resolution order is fixed: an environment variable first, then a `*_command` named in
`config.toml`, then an interactive prompt with echo off.

## Global flags

| Flag | Default | Meaning |
|---|---|---|
| `--target <name>` | — | Named controller+site from the config file — resolves both together, as one unit |
| `-H, --host <addr>` | — | Controller address (local mode; overrides the config file) |
| `-p, --port <n>` | `443` | Controller port |
| `-S, --site <name>` | `default` | Site name — overrides the resolved target's configured site |
| `--secure` | off | Enforce TLS certificate verification (local mode only — see below) |
| `--output <text\|json>` | `text` | Output format |
| `-h, --help` | — | Help for the current command, at any depth |

`--secure` defaults to off because local UniFi controllers commonly serve their web UI
with a self-signed certificate — a tool that can't reach a self-signed device out of the
box is useless. It only makes sense for local-mode targets: passing `--secure` against a
connector-mode target (one authenticated via `UNIFI_API_KEY`/`UNIFI_CONSOLE_ID`, or a
`config.toml` target with `console_id` set) is a **usage error, exit code 2** —
connector traffic to `api.ui.com` is always TLS-verified regardless of any local flag,
so the flag would silently do nothing if it were allowed to pass.

## Config file

`${XDG_CONFIG_HOME:-~/.config}/gofi/config.toml` holds named **targets** and an output
preference — never secrets. A missing file is not an error; gofi works from flags and
environment variables alone.

Each target is **one fully-resolved destination** — a controller and a site together,
not two independently-composed parts. A controller with several sites gets one target
entry per site.

Each target is exactly **one of two shapes**, never a mix:

```toml
default = "home-main"

# Local mode: host + optional password_command.
[targets.home-main]
host             = "192.168.1.1"
site             = "default"
password_command = "pass show unifi/home"

# A second site on the same controller gets its own entry.
[targets.home-guest]
host             = "192.168.1.1"
site             = "guest"
password_command = "pass show unifi/home"

# Connector mode: console_id + optional api_key_command, never host.
[targets.cloud]
site             = "default"
console_id       = "your-console-id"
api_key_command  = "pass show unifi/cloud-key"
```

```bash
gofi config init                 # writes a starter file with comments
gofi config show                 # where the file is, what it resolves to (mode included)
gofi config targets              # list configured targets and their mode
gofi --target home-guest network list
```

`gofi config init` refuses to overwrite an existing file unless you pass `--force`.

## Commands, area by area

### `gofi ips` — fixed IP + DNS reservations

ISC DHCP host-declaration format. Each declaration is a paired write: a static bind on
the user record, and a DNS A record for the hostname.

#### `gofi ips list` / `gofi ips export`

Identical read — `list` and `export` are the same command under two names.

```bash
gofi ips list
gofi ips export > bench.hosts
```

Output is always ISC DHCP text, regardless of `--output` — this is the format's whole
purpose (diffable, version-controllable, hand-editable), not something JSON re-encoding
would improve on.

```
# gofi fixed IP assignments
# exported from UDM at 192.168.1.1
# date: 2026-08-14

host nas {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 192.168.1.13;
}
```

#### `gofi ips import [file]`

Import host declarations from a file, or stdin with no argument. Idempotent — a second
run reports everything skipped and leaves the controller unchanged.

```bash
gofi ips import bench.hosts
gofi ips import bench.hosts --dry-run
gofi ips import bench.hosts --force        # reprocess entries even if unchanged
cat bench.hosts | gofi ips import
gofi ips import bench.hosts --dns-domain bench.test   # override the resolved DNS suffix
```

#### `gofi ips add [declaration]`

Add one host, by flags or as a raw ISC DHCP declaration (not both).

```bash
gofi ips add --name nas --mac aa:bb:cc:dd:ee:01 --ip 192.168.1.13
gofi ips add 'host nas { hardware ethernet aa:bb:cc:dd:ee:01; fixed-address 192.168.1.13; }'
```

#### `gofi ips rm`

Remove one host and its DNS record, identified by exactly one of `--name`/`--mac`/`--ip`.

```bash
gofi ips rm --name nas
gofi ips rm --mac aa:bb:cc:dd:ee:01 --keep-dns   # leave the DNS record alone
gofi ips rm --ip 192.168.1.13 --dry-run
```

An identifier matching more than one record without `--force` is refused with exit code
3 (a guard, not a failure).

#### `gofi ips clear`

Remove **every** fixed-IP assignment on the site. Hard-gated well beyond a normal write:

```bash
gofi ips clear                    # refuses: --force is required
gofi ips clear --force --dry-run  # preview: lists everything that would be removed
gofi ips clear --force            # prints the list, then prompts for confirmation
gofi ips clear --force --yes      # still prints the list; skips only the prompt
```

`--force` is mandatory — there's no bare invocation that deletes everything. The full
removal list prints before the confirmation prompt even when `--yes` is given. `--force`
and `--yes` gate two independent things: `--force` acknowledges the destructive scope of
the operation, `--yes` only skips the interactive prompt. Neither substitutes for the
other.

### `gofi dns` — local DNS records

Independent of `ips` — for the case where only a DNS record should move.

```bash
gofi dns list
gofi dns rm --id 6a650dc764d5a843350f2f55
gofi dns rm --name host1.example.com
gofi dns rm --ip 192.168.2.10 --force   # allow removing every record sharing that IP
```

`--id` is the only identifier guaranteed to select exactly one record; `--name`/`--ip`
matching more than one without `--force` exits 3. `dns add`/`dns set` don't exist yet —
a DNS write endpoint independent of `ips`'s paired-write side effect hasn't been
verified against hardware.

### `gofi network` — networks (VLANs)

Read-only. Subnet, VLAN tag, DHCP pool boundaries, DNS servers.

```bash
gofi network list
gofi network show Default
gofi network show Guest --output json
```

`network show <name>` reports one network in full detail; an unknown name exits 1 (not
found — a real error, not a usage mistake).

### `gofi clients` — currently-connected stations

Live station telemetry, with independent IEEE OUI manufacturer lookup (never trusting
the controller's own, often-stale, OUI field).

```bash
gofi clients list
gofi clients list --wifi              # -w, wireless only
gofi clients list --wired             # -e, wired only
gofi clients vendor b4:0e:cf:2a:85:6f  # offline OUI lookup, no controller session opened
```

`clients vendor` is entirely offline — it reads the cached IEEE OUI registry
(`$XDG_DATA_HOME/gofi/oui.txt`, refreshed automatically when older than 30 days) and
never connects to a controller, so it works with no `--host`/`--target`/credentials at
all.

### `gofi users` — known-client identity records

Persistent records, connected or not — see [above](#why-the-areas-are-shaped-this-way)
for why this is separate from `clients`.

```bash
gofi users list
gofi users list --filter phone
gofi users rm --mac aa:bb:cc:00:00:02
gofi users rm --name old-laptop --dry-run
```

`users rm` clears any fixed-IP assignment on the record *before* forgetting the client,
so the address releases even if the forget step is rejected — the result reports both
steps independently.

### `gofi profile` — whole-site capture and apply

Captures **exactly** networks, WLANs, and fixed IPs as one JSON file — deliberately
excludes devices, firewall, and routing, permanently, so "profile" can't silently grow
to mean "everything on the site."

```bash
gofi profile export > bench.json
gofi profile export --with-keys > bench.json   # include WiFi passphrases in cleartext
gofi profile import bench.json
gofi profile import bench.json --dry-run
gofi profile import bench.json --dns-domain bench.test
cat bench.json | gofi profile import
```

`profile export` omits WiFi passphrases by default; `--with-keys` includes them —
treat an exported profile with `--with-keys` as a secret the same way you'd treat any
other credential file.

`profile import` applies in a fixed order — networks first (fixed IPs must land inside
an existing subnet), then fixed IPs, then WLANs — and reports+skips any network or WLAN
the target site lacks rather than failing the whole run. If the profile's site differs
from the one you're applying to (via `-S`/`--target`), you get a warning, not a silent
mismatch. Networks and WLANs currently have no write endpoint (same gap as `network
set`), so `profile import`'s network/WLAN sections report what they *would* do rather
than writing anything — only the fixed-IPs section is a real write today.

### `gofi config` — gofi's own configuration

Acts on your machine, never a controller. See [Config file](#config-file) above for the
full reference; commands are `config show`, `config targets`, `config init [--force]`.

## Shell completion

`gofi completion <shell>` prints a completion script — see the
[README's Step 3](../README.md#step-3--shell-completion-optional) for how to install it
permanently per shell. Once installed: `gofi <TAB>` lists areas, `gofi ips <TAB>` lists
actions, `gofi ips add --<TAB>` lists flags. Flag *values* (a target name, a MAC address,
a site) don't complete — only the command and flag names themselves.

## Write-safety flags

- **`--dry-run`** — preview a write without making it. Every write command that can
  meaningfully preview its effect supports it.
- **`--force`** — waives a *guard*: a refusal based on the target's current state (an
  ambiguous match, "this deletes everything"). Never waives input validation — a
  malformed MAC is still rejected with `--force`.
- **`--yes`** — skips an interactive *confirmation prompt* for a well-formed but
  destructive action (currently only `ips clear`). Never substitutes for `--force`.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Error — connectivity, an unexpected response, "not found" |
| `2` | Usage error — the command itself was invoked wrong |
| `3` | Refused — a guard blocked an otherwise well-formed request because of the target's current state |

A script can tell "I was blocked" (3) from "it broke" (1) without parsing error text.

## See also

- [`README.md`](../README.md) — install, quick start, the Go SDK
- [`docs/alternate-local-api.md`](alternate-local-api.md) — local username/password auth
- [`REQUIREMENTS.md`](../REQUIREMENTS.md) — the detailed, numbered spec this CLI
  implements
- [`VISION.md`](../VISION.md) — why the CLI is shaped this way
- [`network-cli-convention.md`](../../network-cli-convention.md) — the shared grammar
  gofi and its sibling `gogl` both follow
