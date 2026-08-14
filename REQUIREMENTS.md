# gofi CLI Requirements

The detailed, numbered requirements for the `gofi` command-line tool. This is the
document to work from when implementing or reviewing a change — per the project's
spec-driven convention, if code and this document disagree, the code is wrong.

- **[`VISION.md`](VISION.md)** carries the *why*: the product vision, the target
  command tree, the area-boundary reasoning, and the record of decisions from the CLI
  design brainstorm. Read it first for context.
- **[`CLAUDE.md`](CLAUDE.md)** carries architecture, concurrency rules, and the
  current (v1) specification of the five flag-mode tools this document supersedes.
- **This document** is the contract: what the finished `gofi` tree must do, testable
  and traceable, one requirement at a time.

## Legend

**Status**, on every requirement:

- **Shipped** — already true today, under one of the five existing binaries
  (`gofips`, `gofimac`, `gofinet`, `gofidns`, `gofiuser`). Carries forward into
  `gofi` unchanged unless noted.
- **Target** — not yet built. Part of the `gofi` tree this document specifies.
- **Blocked** — not yet built, and not startable yet: it needs a write endpoint
  verified against real hardware first, per the project's standing rule against
  typing behavior from API documentation alone (`CLAUDE.md`, Controller API Quirks).

**IDs**: `C-<AREA>-NNN` (constraint), `I-<AREA>-NNN` (invariant), `B-<AREA>-NNN`
(behavior, Given/When/Then), area-prefixed so numbering never shifts when a new area
is added. `GLOBAL` covers requirements that apply across every area.

## GLOBAL

Connection flags, config file, secrets, output, exit codes, write-safety flags. Follows
[`network-cli-convention.md`](../network-cli-convention.md) in full; this section is
gofi's specific commitments under it.

### Constraints

- **C-GLOBAL-001** [Target] — `gofi` is one binary, `gofi <area> <action>`. The five
  existing binaries are removed in the same release this ships, per a published
  old-command → new-command mapping table. No wrapper binaries, no deprecation window.
- **C-GLOBAL-002** [Target] — Every write action that can meaningfully preview its
  effect supports `--dry-run`.
- **C-GLOBAL-003** [Target] — `--force` waives only a state guard (a refusal based on
  current controller/site state). It never waives input validation — a malformed MAC
  is rejected with `--force` exactly as without it.
- **C-GLOBAL-004** [Target] — `--yes` waives only an interactive confirmation prompt.
  It never substitutes for `--force`. An action gated by both requires both, given
  independently.
- **C-GLOBAL-005** [Target] — No `--password`, `--api-key`, or other secret-value flag
  exists on any command. A value passed to a flag that used to exist for this purpose
  is a usage error naming the mistake, not a silent accept.
- **C-GLOBAL-006** [Target] — Secret resolution order is fixed: an environment
  variable (`UNIFI_PASSWORD` or `UNIFI_API_KEY`) first, then a `*_command` named in
  `config.toml` for the resolved target, then an interactive prompt with echo off read
  from `/dev/tty`.
- **C-GLOBAL-007** [Target] — `--secure`/`--insecure` against a connector-mode target
  (one whose config has `console_id`) is a usage error, exit 2 — never a silent no-op.
- **C-GLOBAL-008** [Shipped] — `UNIFI_API_KEY` selects connector mode; when set
  alongside `UNIFI_USERNAME`/`UNIFI_PASSWORD`, the latter are ignored with a note
  printed to stderr, not silently overridden. `UNIFI_CONSOLE_ID` is required whenever
  `UNIFI_API_KEY` is set.
- **C-GLOBAL-009** [Target] — `--target <name>` resolves a controller connection and a
  site together, as one unit. There is no cross-target composition; a controller with
  several sites needs one target entry per site. `-S/--site` may still override the
  resolved target's site for a single invocation.
- **C-GLOBAL-010** [Shipped] — Connector-mode requests always verify TLS
  (`api.ui.com` holds a CA-signed certificate) regardless of any local flag or config
  setting.
- **C-GLOBAL-011** [Target] — Every read command accepts `--output text|json`. JSON
  output renders an empty collection as `[]`, never `null`.
- **C-GLOBAL-012** [Target] — Exit codes are fixed: `0` success, `1` error
  (connectivity, unexpected response), `2` usage error (bad invocation), `3` refused
  by a guard (well-formed request, wrong state).

### Invariants

- **I-GLOBAL-001** [Target] — A config target is either fully local-mode (`host` set,
  `console_id` absent) or fully connector-mode (`console_id` set, `host` absent).
  Never a mix.
- **I-GLOBAL-002** [Shipped] — No password or API key is ever written in cleartext to
  any log, debug, or normal output stream.
- **I-GLOBAL-003** [Target] — No command's flag-wiring layer contains business logic;
  every read or write goes through an `internal/<area>` package, per
  `network-cli-convention.md`'s code-layout rule.

### Behavior

- **B-GLOBAL-001** — Given no config file exists, when any command runs with
  sufficient flags and/or environment variables, then it succeeds — a missing config
  file is never itself an error. [Target]
- **B-GLOBAL-002** — Given `--target unknown-name`, when a command runs, then it exits
  2 naming the unknown target, rather than silently falling back to defaults. [Target]
- **B-GLOBAL-003** — Given a malformed `config.toml`, when any command runs
  (including `config show`), then it fails at load with the parse error and its
  location, not a generic message discovered mid-connection. [Target]

## IPS — Fixed IP + DNS Reservations

Supersedes `gofips`. See [Area boundaries](VISION.md#area-boundaries) for why this is
split from `DNS`.

### Constraints

- **C-IPS-001** [Shipped] — ISC DHCP host-declaration parsing and emission follows the
  grammar in `CLAUDE.md` (hostname/MAC/IP fields, required semicolons, 4-space
  indented output, entries sorted numerically by IP on output).
- **C-IPS-002** [Shipped] — `ips import` validates every entry — DNS-safe hostname,
  valid MAC, valid IPv4, no duplicate hostname/MAC/IP within the file — before
  connecting to the controller. Any violation aborts the whole run with every
  violation listed, including line number.
- **C-IPS-003** [Shipped] — For each entry, `ips add`/`ips import` creates a new
  record when the MAC is unseen, updates when the MAC exists but IP or hostname
  differs, and skips (reported) when unchanged.
- **C-IPS-004** [Shipped] — `ips add` without `--force` refuses when the IP is already
  bound to a different MAC, the MAC is already bound to a different IP, or the
  hostname's DNS record already points elsewhere.
- **C-IPS-005** [Target] — `ips rm` accepts exactly one of `--name`/`--mac`/`--ip`.
  Zero or more than one given is a usage error (exit 2), not a silent pick of the
  first-given identifier.
- **C-IPS-006** [Shipped→Target] — An ambiguous `--name`/`--ip` matching more than one
  record on `ips rm` is refused without `--force`, as under `gofips` today. The exit
  code changes from `1` to `3` under C-GLOBAL-012, since this is a guard, not a
  failure.
- **C-IPS-007** [Target] — `ips clear` refuses to run without `--force`. There is no
  bare invocation that deletes every reservation on the site.
- **C-IPS-008** [Target] — `ips clear` prints the full list of reservations that would
  be removed before the confirmation prompt, even when `--yes` is also given.
- **C-IPS-009** [Shipped] — `ips rm` clears the fixed-IP flag on the user record and
  deletes the associated DNS A record, unless `--keep-dns` is given. The user record
  itself is never deleted by `ips rm`.
- **C-IPS-010** [Shipped] — Network auto-detection determines which UniFi network
  contains a given IP before writing. A failure to match during `import` is reported
  per-entry and does not abort the remaining entries.
- **C-IPS-011** [Target] — `ips add` accepts either a raw declaration positional
  argument or `--name`/`--mac`/`--ip` flags together, never both. Giving both is a
  usage error, mirroring `gogl lan reservations add`.

### Invariants

- **I-IPS-001** [Shipped] — A reservation's hostname is DNS-label-valid before any
  write reaches the controller — rejected on failure, never silently rewritten.
- **I-IPS-002** [Target] — `ips rm` and `ips clear` never leave a DNS record pointing
  at an address the controller no longer reserves.
- **I-IPS-003** [Shipped] — Output from `ips export`/`ips list` sorts numerically by
  IP address (uint32 conversion), never lexically.

### Behavior

- **B-IPS-001** — Given a host declaration whose MAC already holds a different IP,
  when `ips add` runs without `--force`, then it exits 3 and names the conflicting
  entry. [Shipped, exit code updated per C-IPS-006's reasoning]
- **B-IPS-002** — Given reservations exist, when `ips clear --force` runs without
  `--yes`, then it prints the full list and prompts, proceeding only on confirmation.
  [Target]
- **B-IPS-003** — Given an import file with a duplicate MAC across two declarations,
  when `ips import` runs, then it exits 2 before contacting the controller, listing
  every duplicate found. [Shipped]
- **B-IPS-004** — Given `ips import` on a file already fully applied, when run a
  second time, then every entry reports "skipped" and controller state is unchanged.
  [Shipped]
- **B-IPS-005** — Given both `--name/--mac/--ip` flags and a raw declaration argument,
  when `ips add` runs, then it exits 2. [Target]

## DNS — Local DNS Records

Supersedes `gofidns`. Exists so a stale DNS record can be corrected without touching
the reservation (user record) that owns the address — see
[Area boundaries](VISION.md#area-boundaries).

### Constraints

- **C-DNS-001** [Shipped] — `dns list` flattens each record to ID, key, value, type,
  TTL, enabled; sorted by key, then value.
- **C-DNS-002** [Shipped] — `dns rm` accepts exactly one of `--id`/`--name`/`--ip`.
  `--id` is the only identifier guaranteed to select a single record.
- **C-DNS-003** [Shipped→Target] — Matching more than one record without `--force` is
  refused. Exit code changes from `1` to `3` under C-GLOBAL-012.
- **C-DNS-004** [Blocked] — `dns add`/`dns set` do not exist until a DNS write
  endpoint independent of the `ips` paired-write side effect is verified against real
  hardware and recorded, per the project's fixture-first rule (`CLAUDE.md`).

### Invariants

- **I-DNS-001** [Shipped] — `dns rm` never deletes more records than the identifier
  unambiguously matches, absent `--force`.

### Behavior

- **B-DNS-001** — Given a name matching two DNS records, when `dns rm --name X` runs
  without `--force`, then it exits 3 and lists both matches. [Shipped]
- **B-DNS-002** — Given `dns rm --id <id>` for an ID that does not exist, when run,
  then it exits 1 — not found is a real error, not a guard refusal, since an ID cannot
  be ambiguous. [Shipped]

## NETWORK — Networks (VLANs)

Supersedes `gofinet`. Read-only today; see
[Area boundaries](VISION.md#area-boundaries) for why that is a "not yet built" gap
rather than a permanent one.

### Constraints

- **C-NETWORK-001** [Shipped] — `network list` flattens each network to name,
  purpose, VLAN tag, subnet, enabled, DHCP enabled, DHCP pool bounds, lease time,
  advertised gateway (only when `dhcpd_gateway_enabled`), and DNS servers; sorted by
  name.
- **C-NETWORK-002** [Shipped] — A network with no active DHCP server (WAN, VLAN-only,
  DHCP disabled) displays `(disabled)` in the pool column, and `-` for lease/gateway/
  DNS, rather than blank or zero values.
- **C-NETWORK-003** [Target] — `network show <name>` reports one network's full
  detail. An unknown name exits 1 (not found).
- **C-NETWORK-004** [Blocked] — `network set` does not exist until a network/VLAN
  write endpoint is verified against real hardware.

### Invariants

- **I-NETWORK-001** [Shipped] — `network list`/`network show` never write to the
  controller; both stay strictly read-only until C-NETWORK-004 is resolved.

### Behavior

- **B-NETWORK-001** — Given a WAN-purpose or DHCP-disabled network, when
  `network list` runs, then that row shows `(disabled)` for the pool and `-` for
  lease/gateway/DNS. [Shipped]

## CLIENTS — Currently-Connected Stations

Supersedes `gofimac`. Reads live station telemetry only — see
[Area boundaries](VISION.md#area-boundaries) for why this is kept separate from
`USERS`.

### Constraints

- **C-CLIENTS-001** [Shipped] — `clients list` filters with `--wifi`/`--wired`; with
  neither, every connected client is listed.
- **C-CLIENTS-002** [Shipped] — Manufacturer lookup always comes from gofi's own
  cached IEEE OUI database, never the controller's own (frequently stale) OUI field.
- **C-CLIENTS-003** [Shipped] — The OUI cache lives at `$XDG_DATA_HOME/gofi/oui.txt`
  (renamed from `gofimac`'s `$XDG_DATA_HOME/gofimac/oui.txt` on migration, per
  C-GLOBAL-001), refreshed when older than 30 days, with a stale-cache fallback on
  download failure.
- **C-CLIENTS-004** [Target] — `clients vendor <mac>` performs the same OUI lookup
  standalone, without opening a controller session — mirrors `gogl clients vendor`.
- **C-CLIENTS-005** [Shipped] — Output sorts by IP numerically; clients without an IP
  sort last.

### Invariants

- **I-CLIENTS-001** [Shipped] — `clients list`/`clients vendor` never write to the
  controller, and touch only the local OUI cache file, never a remote system beyond
  refreshing it.

### Behavior

- **B-CLIENTS-001** — Given the OUI download fails and a stale cache exists, when
  `clients list` runs, then it proceeds using the cached data and prints a warning to
  stderr, rather than failing. [Shipped]
- **B-CLIENTS-002** — Given the OUI download fails and no cache exists, when
  `clients list` runs, then it exits 1, rather than showing every device as
  "unknown." [Shipped]
- **B-CLIENTS-003** — Given a MAC with no registered OUI entry, when
  `clients vendor <mac>` runs, then it reports "no registered manufacturer," not a
  failed lookup. [Target]

## USERS — Known-Client Identity Records

Supersedes `gofiuser`. Reads persistent identity records, connected or not — see
[Area boundaries](VISION.md#area-boundaries) for why this is kept separate from
`CLIENTS`.

### Constraints

- **C-USERS-001** [Shipped] — `users list` supports `--filter <substring>`, matching
  name, hostname, MAC, or fixed IP.
- **C-USERS-002** [Shipped] — `users rm` accepts exactly one of `--mac`/`--name`. An
  ambiguous name is an error, never a first-match pick.
- **C-USERS-003** [Shipped] — `users rm` clears any fixed-IP assignment on the record
  *before* forgetting the client, so the address releases even if the forget step is
  rejected.
- **C-USERS-004** [Shipped] — Client removal uses the batch `forget-sta` command with
  a `"macs": [...]` array. The singular `"mac"` field is rejected by the controller
  with HTTP 400; `DELETE /rest/user/<id>` is never used, since it answers 404
  (`CLAUDE.md`, Controller API Quirks).
- **C-USERS-005** [Shipped] — A `fixed_ip` present on a record whose `use_fixedip` is
  false displays as `(dynamic)`, never presented as a live reservation.

### Invariants

- **I-USERS-001** [Shipped] — `users rm` always attempts the fixed-IP clear and the
  forget as two independently-reported steps; a rejection of one is never hidden by
  the success of the other.

### Behavior

- **B-USERS-001** — Given a record with a fixed IP, when `users rm --mac X` runs,
  then the result reports "removed fixed IP assignment" and the forget outcome as two
  separate lines. [Shipped]
- **B-USERS-002** — Given `--name` matching two known clients, when
  `users rm --name X` runs, then it exits 3 (refused — ambiguous identifier), never
  removing the first match. [Shipped, exit code updated per C-GLOBAL-012]

## PROFILE — Site Capture and Apply

New area; no predecessor tool. Deliberately scoped narrower than `gogl profile` — see
[Profile scope](VISION.md#profile-scope).

### Constraints

- **C-PROFILE-001** [Target] — `profile export` captures exactly three sections:
  networks, WLANs, fixed IPs. Devices, firewall, routing, and port profiles are never
  captured — permanently out of scope for this area, not merely deferred.
- **C-PROFILE-002** [Target] — `profile export` omits WiFi passphrases by default;
  `--with-keys` includes them in cleartext, matching `gogl`'s convention.
- **C-PROFILE-003** [Target] — `profile import` applies in a fixed order: networks
  first (fixed IPs must land inside an existing subnet), then fixed IPs, then WLANs.
- **C-PROFILE-004** [Target] — If applying the network section would change the
  site's addressing such that the acting session would be dropped, the run stops
  after that section and reports how to resume — mirrors `gogl`'s subnet-move
  handling.
- **C-PROFILE-005** [Target] — A profile applied against a different controller model
  or site than it was captured from warns rather than fails, and skips any network or
  WLAN the target lacks.

### Invariants

- **I-PROFILE-001** [Target] — `profile import` never writes a device, firewall, or
  routing change, even if such a field appears in a hand-edited profile file — an
  out-of-scope field is rejected or ignored, never applied.

### Behavior

- **B-PROFILE-001** — Given a profile captured without `--with-keys`, when
  `profile import` runs, then existing WLAN passphrases on the target are left
  untouched, not blanked. [Target]
- **B-PROFILE-002** — Given a profile whose network section doesn't exist on the
  target site, when `profile import` runs, then it reports the gap and continues with
  what it can apply. [Target]

## CONFIG — gofi's Own Configuration

New area; no predecessor tool (env vars only today). Acts on the local machine, never
the controller.

### Constraints

- **C-CONFIG-001** [Target] — Config file at
  `${XDG_CONFIG_HOME:-~/.config}/gofi/config.toml`; a missing file is never an error.
- **C-CONFIG-002** [Target] — Each `[targets.NAME]` block is exactly one of
  local-mode (`host`, `site`, optional `password_command`) or connector-mode
  (`console_id`, `site`, optional `api_key_command`) — never a mix (I-GLOBAL-001).
- **C-CONFIG-003** [Target] — `config show` reports the resolved target's mode
  (local/connector), its host or console ID as applicable, and whether the file
  exists — never a secret value.
- **C-CONFIG-004** [Target] — `config init` writes a starter file with comments and
  no live secrets; refuses to overwrite an existing file without `--force`.
- **C-CONFIG-005** [Target] — `config targets` lists every configured target name,
  its resolved mode, and which one is default.

### Invariants

- **I-CONFIG-001** [Target] — No secret value is ever written into `config.toml` by
  any `gofi` command, including `config init`.

### Behavior

- **B-CONFIG-001** — Given no config file and no `--target`, when a command needing a
  connection runs with `-H`/env vars sufficient to connect, then it succeeds without
  ever requiring `gofi config init`. [Target]
- **B-CONFIG-002** — Given a `config.toml` with a syntax error, when `config show`
  runs, then it reports the parse error and its location, not a generic failure.
  [Target]

## Traceability

Every **Shipped** requirement traces to a section of `CLAUDE.md`'s existing tool
specs (`gofips`/`gofimac`/`gofinet`/`gofidns`/`gofiuser`) and is expected to hold
unchanged through the migration unless its entry says otherwise (exit-code changes
under C-GLOBAL-012 are the only systematic exception). Every **Target** or
**Blocked** requirement traces to `VISION.md`'s proposed tree or its recorded
brainstorm decisions.

When a requirement's status changes — a **Blocked** item gets its hardware
verification, a **Target** item ships — update its tag in place here rather than
writing a new document. This file is the living requirements set: per the project's
spec-driven convention, update it first, then bring the code in line with it.
