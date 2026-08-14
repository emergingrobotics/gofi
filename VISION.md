# gofi - Go UniFi Controller Client - Vision

A Go module and CLI for programmatic, reproducible control of a UniFi controller (UDM
Pro and compatible consoles). The purpose is the same one `gogl` serves for GL.iNet
travel routers: turn network configuration that would otherwise live only in an admin
UI's click history into something that can be captured, diffed, reviewed and re-applied.

This document is the **product vision and the target CLI shape**. Architecture rules,
concurrency requirements, the mock server contract, and the current (v1) specification
of each shipped tool live in [`CLAUDE.md`](CLAUDE.md) and are not repeated here. Where
the two disagree on the CLI's shape, this document describes where the CLI is going;
`CLAUDE.md` describes what is built today.

## Target System

| Property | Value |
|---|---|
| Controller | UniFi UDM Pro and compatible consoles (v10+ firmware) |
| Auth modes | Cloud API key via the `api.ui.com` connector (recommended), or local username/password against a directly-reachable controller |
| Scope unit | A **site** within a controller. A controller can host several sites; commands act on one site at a time |
| Object graph | Site → networks (VLANs) → devices (APs, switches, gateways) → clients (stations) and users (known-client identity records) |

This is structurally richer than `gogl`'s target: a GL.iNet travel router is one flat
device with one LAN. A UniFi controller manages a site with multiple networks and
potentially many physical devices. The CLI convention below accounts for that rather
than pretending the two backends are the same shape — see the area-boundary reasoning
below and the [resolved design decisions](#decisions-from-the-cli-design-brainstorm-2026-08-13).

## Scope

Five capability areas are already proven out, each currently shipped as its own
flag-mode binary (`gofips`, `gofimac`, `gofinet`, `gofidns`, `gofiuser`) documented in
full in `CLAUDE.md`:

- **Fixed IP + DNS reservations** (`gofips`) — ISC DHCP host-declaration import/export,
  add, delete. Each host declaration is a paired write: a static bind on the user record,
  and a DNS A record for the hostname.
- **Local DNS records** (`gofidns`) — list and delete DNS records directly, independent
  of the user record that a fixed-IP write would otherwise touch. Exists because
  deleting a user to clear a stale DNS record would take a live reservation with it.
- **Networks** (`gofinet`) — list a site's networks with subnet, VLAN tag and DHCP pool
  boundaries. Read-only today.
- **Connected clients** (`gofimac`) — list currently-connected stations, filtered by
  wired/WiFi, with an independent IEEE OUI manufacturer lookup (never trusting the
  controller's own, often-stale, OUI field).
- **Known-client (user) records** (`gofiuser`) — list and remove the controller's
  persistent identity records for clients it has seen, including the two-step forget
  sequence the raw API requires.

This document proposes bringing those five into one binary under a shared noun-verb
tree, following [`network-cli-convention.md`](../network-cli-convention.md) — the same
convention gogl's own command tree already follows.

## Why one binary

`gogl` went through this exact migration: four separate flag-mode utilities became one
`gogl <area> <action>` tree, because four binaries meant four flag parsers, four help
texts, four copies of the connection flags, and there was nothing forcing `--get` to
mean the same thing in every one of them. gofi's five binaries have the identical
problem today — `gofips --get`, `gofidns --get`, and `gofiuser --list` all mean "list
records" but spell it two different ways, and `--del` on three different tools takes a
different combination of identifier flags with no shared vocabulary enforcing
consistency. Collapsing them removes that drift and gets shell completion, a shared
config file, and a shared exit-code contract for free.

## CLI Convention

Follows [`network-cli-convention.md`](../network-cli-convention.md) in full. This
section is the gofi-specific application of it — the target tree, not yet built.

### The `gofi` command (target — not yet built)

```
gofi
├── ips                          Fixed IP + DNS reservations, ISC DHCP format
│   ├── list                     List current fixed-IP assignments
│   ├── export                   Write every assignment to stdout in ISC DHCP format
│   ├── import [file]            Import host declarations from a file or stdin
│   ├── add [declaration]        Add one host, by flags or as a declaration fragment
│   ├── rm                       Remove one host, by --name/--mac/--ip
│   └── clear                    Remove every fixed-IP assignment (--force required)
├── dns                          Local DNS records, independent of ips
│   ├── list                     List DNS records
│   └── rm                       Remove one record, by --id/--name/--ip
├── network                      Networks (VLANs): subnet, DHCP pool, DNS servers
│   ├── list                     List every network on the site
│   └── show <name>              Report one network in detail
├── clients                      Currently-connected stations, with OUI lookup
│   ├── list                     List connected stations (--wifi, --wired filters)
│   └── vendor <mac>             Look up a MAC's manufacturer (offline)
├── users                        Known-client identity records, connected or not
│   ├── list                     List known clients (--filter substring)
│   └── rm                       Remove a known client, by --mac/--name
├── profile                      Capture networks + WLANs + fixed IPs as JSON, apply back
│   ├── export                   Write a profile to stdout
│   └── import [file]            Apply a profile from a file or stdin
├── config                       gofi's own configuration file (local machine only)
│   ├── show                     Report the config file's location and what it resolves to
│   ├── targets                  List the configured targets
│   └── init                     Write a starting configuration file
└── completion                   Shell completion (bash, zsh, fish, powershell)
```

### Area boundaries

**`ips` and `dns` are split for the same reason `gogl`'s `lan reservations` and
`lan dns` are split**: a fixed-IP write touches two backend objects (the user record's
`fixed_ip` field and a DNS A record), and the two can drift — a hand-deleted user, or a
run that died mid-write. `ips add`/`ips rm` keep both in step, the way `gofips` does
today; `dns` exists for the case where only the DNS side should move, without touching
a reservation something else still depends on.

**`network` and `dns` stay read-only (or delete-only) for now, and that is a "not yet
built" gap, not a declared-out-of-scope decision** the way `gogl` treats WAN or VLANs.
GL.iNet's write endpoints for the LAN were exercised and verified against hardware
before `lan set` shipped; UniFi's network/VLAN write surface and an independent
`dns add`/`dns set` have not been attempted yet. Both get added once verified against a
real controller and recorded — never typed from API documentation alone, which has
been wrong often enough on `gogl` to be a standing rule rather than a one-off caution.

**`clients` and `users` are kept as two areas, permanently**, because they are two
different objects on this backend: `clients` reads live station telemetry (`stamgr`) —
signal, uptime, which AP or switch port — and only exists for a station currently
connected. `users` reads persistent identity records (`rest/user`) that outlive the
connection and carry the fixed-IP relationship. `gogl` has no equivalent split because
a GL.iNet router has no separate "known but not currently connected" client table
distinct from its DHCP lease list. Here, merging the two would mean synthesizing a
combined view that corresponds to no single API call — a bigger unification than a
`--all` flag suggests — so each area's one-line description says plainly what it is:
`clients` is "currently connected," `users` is "known to the controller, whether
connected or not."

**No `radio`/`wifi` split.** GL.iNet's radio/SSID seam exists because one physical
router scopes writes that way at the API level. UniFi's WLANs are configured once and
applied across every AP on a site; there is no `network set` moving devices individually
today, so this seam does not yet apply. If per-AP RF tuning (channel, power) is ever
added, it would need its own area, following the same "does the backend consider this
one write" rule — not folded into `network`.

### Global flags (target)

| Flag | Meaning |
|---|---|
| `--target <name>` | Named controller+site from the config file — a target resolves both together, the same way `gogl --router` resolves one fully-specified destination |
| `-H, --host <addr>` | Controller address, overrides the config file (local-mode targets only) |
| `-p, --port <n>` | Controller port, default `443` |
| `-S, --site <name>` | Site name, overrides the target's configured site |
| `--secure` | Enforce TLS certificate verification (default **off**: local controllers commonly ship self-signed certs). **Usage error** (exit 2) when passed against a connector-mode target — connector traffic to `api.ui.com` is always verified, and a flag that would silently do nothing is rejected instead, the same way `gogl` rejects `--insecure=false` without `--https` |
| `--output <text\|json>` | Output format, default `text` |
| `-h, --help` / `-v, --version` | As in the convention |

`--site` is additive to the convention's base flag set: `gogl` has no site concept, so
`network-cli-convention.md` doesn't define one, but a gofi-specific global flag is
allowed by the convention as long as the shared flags aren't renamed or repurposed.

### Config file (target — new)

`${XDG_CONFIG_HOME:-~/.config}/gofi/config.toml`. As with `gogl`, never holds secrets.
Each named target is a **fully-resolved destination — one controller and one site
together**, not two independently-composed parts. A controller with several sites gets
one target entry per site (`home-main`, `home-guest`), the same way `gogl` would define
a separate named router per physical device rather than expecting `--router` and some
other flag to be combined at call time.

Each target also expresses exactly **one of two connection shapes**, since gofi's two
auth modes are structurally different, not just two ways to supply the same fields:

```toml
default = "home-main"

[targets.home-main]
host             = "192.168.1.1"
site             = "default"
password_command = "pass show unifi/home"

[targets.home-guest]
host             = "192.168.1.1"
site             = "guest"
password_command = "pass show unifi/home"

[targets.cloud]
site              = "default"
console_id        = "6a1b2c3d..."
api_key_command    = "pass show unifi/cloud-key"
```

A local target has `host` and never `console_id`; a connector target has `console_id`
and never `host`, since every connector request goes to `api.ui.com` regardless of
where the console actually lives. `gofi config show <name>` reports which mode a
target resolved to, so the distinction is never left implicit.

### Verb vocabulary

The canonical vocabulary from `network-cli-convention.md` applies unchanged, with one
domain addition already precedented by `gogl`:

| Verb | Meaning | Used by |
|---|---|---|
| `vendor` | Offline lookup unrelated to the connected target | `clients vendor <mac>`, same as `gogl clients vendor` |

`gofips`'s `--set` (bulk import) is `ips import`, not `ips set` — `set` is reserved for
writing fields on a single existing object (there is no single "fixed-IP object" that
`ips set` would write fields onto; each entry is created, updated or left alone by
`import`, per host declaration).

`ips clear` (delete every fixed-IP assignment) is kept for symmetry with
`gogl lan reservations clear`, but hard-gated beyond what `gogl` requires: `--force` is
mandatory with no bare invocation, the full list of what would be removed is printed
before the confirmation prompt, and `--yes` (skip the prompt) never substitutes for
`--force` (skip the guard) — the two stay independent, per the convention. gofi manages
shared, often-production infrastructure, unlike a disposable travel-router bench, so
this verb needs a stronger floor than its `gogl` counterpart.

### Exit codes

The same 0/1/2/3 scheme as the convention. This is a **behavior change** from the
current tools, which exit 1 for a refused delete (ambiguous `--name`/`--ip` matching
several records without `--force`). Under the convention, that refusal is a guard, not
a failure, and should exit 3 — a script retrying "controller unreachable" (1) should not
also retry "you didn't pass `--force`" (3).

### Profile scope

`gofi profile` is deliberately narrower than `gogl profile`. UniFi's object graph
(devices and their adoption state, firewall, routing, port profiles) is large enough
that "capture the whole site" would be a much bigger and more ambiguous claim than
"capture the whole router." A profile captures exactly three sections — **networks,
WLANs, and fixed IPs** — the same things `gogl` captures, translated to UniFi's model.
Devices, firewall, routing and port profiles are explicitly and permanently excluded,
not merely deferred, so "profile" can never silently grow to mean "everything on the
site" as new areas are added later. If device or firewall capture is ever wanted, it is
a new, separately-named capability, not an expansion of `profile`.

### Migration from the five existing binaries

One-release cutover, not a transition period. `gofips`, `gofimac`, `gofinet`,
`gofidns` and `gofiuser` are removed the same release the `gofi` tree ships, with the
README and CHANGELOG carrying an explicit old-command → new-command mapping table
(e.g. `gofips --get` → `gofi ips list`, `gofips --del --mac X` → `gofi ips rm --mac X`).
No wrapper binaries, no deprecation window. Simpler to ship and keeps one implementation
live instead of two, at the cost of breaking any existing script or cron job on
upgrade — acceptable because this project has no external users depending on binary
stability today, unlike a published library API.

## Decisions from the CLI design brainstorm (2026-08-13)

Resolved, in conversation with the project owner, against the open questions this
document originally carried:

1. **`clients` and `users` stay two separate areas, permanently.** They read
   different UniFi objects (live `stamgr` telemetry vs. persistent `rest/user`
   identity), and a merged view would synthesize a join neither backend call actually
   performs. Each area's help text states its scope plainly so the split reads as
   informative rather than arbitrary.
2. **A named target resolves a controller and a site together**, never two
   independently-composed parts. A controller with several sites gets one target entry
   per site. This mirrors `gogl --router` resolving one fully-specified destination
   and keeps every command needing zero extra flags once a default target is set.
3. **`--secure`/`--insecure` is a usage error against a connector-mode target**
   (exit 2), not a silent no-op — connector traffic is always TLS-verified, and a flag
   that would do nothing is rejected rather than left to imply it worked. Same
   principle `gogl` already applies to `--insecure=false` without `--https`.
4. **`ips clear` exists, but hard-gated**: `--force` mandatory with no bare
   invocation, the full deletion list printed before the confirmation prompt, and
   `--yes` never substituting for `--force`. gofi manages shared infrastructure, so
   this verb needs a stronger floor than `gogl lan reservations clear`.
5. **`dns` and `network` gain write verbs opportunistically**, each only once its
   specific write endpoint is verified against real hardware and recorded — following
   `gogl`'s standing rule against typing behavior from API documentation alone. Until
   then the gap is documented as "not yet built."
6. **`profile` ships narrow from day one**: networks, WLANs, and fixed IPs only,
   with devices/firewall/routing/port-profiles permanently out of scope for this area
   rather than deferred into it later. See [Profile scope](#profile-scope).
7. **Migration is a one-release cutover** with a published command-mapping table, not
   deprecated wrapper binaries. See [Migration](#migration-from-the-five-existing-binaries).
