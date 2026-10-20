---
name: gofi-cli
description: Use the gofi CLI to manage a UniFi UDM Pro's fixed IPs, DNS records, networks, clients, and site profiles.
---

# gofi CLI

`gofi` is one binary that manages a UniFi controller (UDM Pro and compatible
consoles, Network Application 10.x+) over its JSON API. Use it for fixed-IP
reservations in ISC DHCP format, local DNS records, networks/VLANs, connected
clients and MAC vendor lookup, known-client records, and site profile
export/import.

```
gofi [global flags] <area> <action> [args] [action flags]
```

Global flags may also come after the action (`gofi network list --output json`
works). Check that it's on `PATH` with `gofi --help`; every command and action
also takes `--help`. `gofi install-skill` installs this skill (see §8).

## 1. Before running anything: which controller will you hit?

**An ambient `UNIFI_API_KEY` wins over everything.** If it is set in the
environment, gofi uses connector mode through `api.ui.com` to whatever console
`UNIFI_CONSOLE_ID` names. It then ignores `-H/--host`, `-p/--port`,
`UNIFI_USERNAME`/`UNIFI_PASSWORD`, and a local-mode `--target`. The only sign
is one stderr line: `Note: UNIFI_API_KEY is set; ignoring UNIFI_USERNAME/UNIFI_PASSWORD.`
Many operators export this key from their shell profile.

So before any **write**:

1. Check `env | grep '^UNIFI_\|^GOFI_'` to see which mode and controller will be used.
2. To reach a different (local or test) controller, strip the key for that one
   command: `env -u UNIFI_API_KEY -u UNIFI_CONSOLE_ID gofi -H 10.0.0.1 ...`.
3. Do a read (`gofi network list`) or a `--dry-run` first, and confirm the
   output is from the controller you meant.

### Connector mode (cloud API key; recommended)

```bash
export UNIFI_API_KEY=...      # cloud key from unifi.ui.com > API (Network scope), not a console-issued key
export UNIFI_CONSOLE_ID=...   # console ID: curl -s -H "X-API-KEY: $UNIFI_API_KEY" https://api.ui.com/v1/hosts
gofi network list
```

- Needs no LAN route to the controller. TLS is always verified.
- `UNIFI_CONSOLE_ID` (or a connector target's `console_id`) is required.
  Without it: `UNIFI_CONSOLE_ID is required for connector mode`, exit 1.
- The connector rate-limits bulk writes with HTTP 429. gofi retries up to 8
  times with backoff of up to 30 s, so large imports can be slow. That is
  expected, not a hang.

### Local mode (username/password on the LAN)

Used only when `UNIFI_API_KEY` is unset.

```bash
env -u UNIFI_API_KEY UNIFI_USERNAME=admin UNIFI_PASSWORD=... gofi -H 192.168.1.1 network list
```

- Host comes from `-H/--host`, then the target's `host`, then `UNIFI_CONTROLLER_IP`.
- Username comes from `UNIFI_USERNAME`, or the target's `username` key.
- TLS verification is **off** by default, because UDMs use self-signed certs.
  `--secure` turns it on. `--secure` in connector mode is a usage error (exit 2).

### Secrets: never on the command line

There is no `--password` or `--api-key` flag. Passing one is an unknown flag
(exit 2). The secret is looked up in this order:

1. The env var (`UNIFI_PASSWORD` or `UNIFI_API_KEY`).
2. The resolved target's `password_command` / `api_key_command` in
   `config.toml`. Its stdout is the secret.
3. An echo-off prompt on `/dev/tty`. **Only when a config target is resolved.**

Agents have no TTY, so step 3 fails with
`no terminal to prompt on ... /dev/tty`, exit 1. Always provide the env var or
a `*_command`. Never put a secret value in `config.toml` or in a flag.

## 2. Global flags and environment

| Flag | Default | Meaning |
|---|---|---|
| `--target <name>` | config `default` | Named controller+site from `config.toml`. Unknown name exits 2 |
| `-H, --host <addr>` | — | Controller address (local mode only) |
| `-p, --port <n>` | `443` | Controller port (local mode only) |
| `-S, --site <name>` | target's site, else `default` | UniFi site name (not the display name) |
| `--secure` | off | Enforce TLS verification. Local mode only; exit 2 in connector mode |
| `--output text\|json` | `text`, or config `output` | Any other value exits 2 |

| Env var | Effect |
|---|---|
| `UNIFI_API_KEY`, `UNIFI_CONSOLE_ID` | Connector mode (see §1) |
| `UNIFI_USERNAME`, `UNIFI_PASSWORD` | Local mode credentials |
| `UNIFI_CONTROLLER_IP` | Fallback host for local mode |
| `UNIFI_DNS_DOMAIN` | Default for every `--dns-domain` flag |
| `GOFI_CONFIG` | Explicit path to `config.toml` (overrides XDG). Useful for isolated runs |
| `XDG_CONFIG_HOME`, `XDG_DATA_HOME` | Config dir and OUI-cache dir |

## 3. Exit codes, as actually implemented

| Code | Meaning |
|---|---|
| `0` | Success, including a successful `--dry-run` |
| `1` | Error: connectivity, auth, not found, missing credentials, partial failure |
| `2` | Usage error: bad flags or args, unknown command or target, malformed input file, `ips clear` without `--force`, an existing file for `config init` without `--force`, `--secure` in connector mode |
| `3` | Refused by a guard: an `ips add` conflict, an ambiguous `ips rm`/`dns rm`/`users rm` identifier, or `ips clear` not confirmed. The message starts `gofi: refused:` |

Errors go to stderr prefixed `gofi:`. Progress lines (`created:`, `skip:`,
`would ...`) also go to **stderr**. Only listings, JSON, and `ips rm`/`ips add`
result lines go to stdout.

On exit 3, nothing was changed. Fix the input, or add `--force` only if the
user wants to override that specific guard. `users rm` has no `--force`, so
disambiguate with `--mac`.

## 4. Areas and actions

| Area | Actions | Writes? |
|---|---|---|
| `ips` | `list`, `export`, `import`, `add`, `rm` (alias `remove`), `clear` | yes |
| `dns` | `list`, `rm` (alias `remove`) | delete only. **No `dns add`/`dns set`** |
| `network` | `list`, `show <name>` | read-only. **No `network set`** |
| `clients` | `list`, `vendor <mac>` | read-only |
| `users` | `list`, `rm` (alias `remove`) | delete only |
| `profile` | `export`, `import` | only fixed IPs are actually written |
| `config` | `show`, `targets`, `init` | local file only, never touches a controller |
| `completion` | `bash`/`zsh`/`fish`/`powershell` | prints a script to stdout |

`clients` and `users` are different objects. `clients` is live station telemetry,
present only while the device is connected. `users` is the persistent
known-client record, which outlives the connection and holds the fixed IP. To
ask "what's connected now", use `clients`. To ask "what does the controller
know about", use `users`.

### `ips`: fixed-IP reservations plus their DNS A records

Each reservation is a **paired write**: `fixed_ip` on the user record **plus** a
static DNS A record `<hostname>.<domain>`.

**DNS domain resolution (needed by `add` and `import`).** The domain comes from
`--dns-domain`, or `UNIFI_DNS_DOMAIN`, or the one domain name set on the site's
networks. If no network sets a domain, or networks set **different** ones, the
command fails before writing. Pass `--dns-domain`.

**List / export.** `list` and `export` are the same command.

```bash
gofi ips list                 # same as `gofi ips export`
gofi ips export > site.hosts
```

- The output is always ISC DHCP text; `--output json` is ignored. For JSON, use
  `gofi profile export` and read `fixed_ips`.
- Entries are sorted by IP, and the file starts with a `#` header comment.
- DNS-drift warnings go to stderr.

```
host nas {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 192.168.1.13;
}
```

**Import.**

```bash
gofi ips import site.hosts --dry-run    # always preview first
gofi ips import site.hosts
gofi ips import - < site.hosts           # or no file arg: reads stdin
gofi ips import site.hosts --force       # re-push entries that are already unchanged
gofi ips import site.hosts --dns-domain lan.example
```

- The whole file is validated before connecting: DNS-safe hostname, valid MAC,
  valid IPv4, no duplicate host, MAC, or IP. Any violation exits 2 with every
  problem listed, and nothing is written.
- For each entry, an unseen MAC is created. A changed IP or name on an existing
  MAC is updated. An identical entry is skipped. Re-running the same file only
  skips, so import is idempotent.
- It is additive. Reservations that are not in the file are **not** removed.
- The stderr summary looks like
  `N processed, S skipped, C created, U updated, E errors`.
- A per-entry failure (for example `no network found containing IP x`) does not
  stop the other entries. The run still exits 1 if any entry failed, so on exit
  1, read the summary: earlier entries may have been applied.

**Add.**

```bash
gofi ips add --name nas --mac aa:bb:cc:dd:ee:01 --ip 192.168.1.13
gofi ips add 'host nas { hardware ethernet aa:bb:cc:dd:ee:01; fixed-address 192.168.1.13; }'
echo 'host nas { ... }' | gofi ips add          # no args, no flags: reads one declaration from stdin
```

- Pass all three flags, or a declaration, but not both (otherwise exit 2).
- `ips add` has **no `--dry-run`**. Preview with `ips import --dry-run` on a
  one-entry file instead.
- The IP must fall inside an existing network's subnet, or the add exits 1.
- Without `--force`, it refuses (exit 3) when the IP belongs to another MAC,
  the MAC already holds another IP, or the DNS name points elsewhere.
  `--force` overwrites that binding, so only use it when the user means to
  move it.
- Re-adding an identical entry succeeds and changes nothing.
- Changing `--dns-domain` moves the record: the stale `<host>.<old>` is deleted
  and `<host>.<new>` is created.

**Remove one.**

```bash
gofi ips rm --mac aa:bb:cc:dd:ee:01 --dry-run   # exactly one of --name/--mac/--ip, else exit 2
gofi ips rm --mac aa:bb:cc:dd:ee:01             # clears fixed_ip + deletes DNS; keeps the user record
gofi ips rm --mac aa:bb:cc:dd:ee:01 --keep-dns  # leave the DNS record
```

- `ips rm` only releases the reservation. It never deletes the user record;
  to forget a client, use `users rm`.
- **Prefer `--mac`**, which always selects exactly one record.
- `--ip` selects every reservation holding that address.
- `--name` matches name or hostname, case-insensitively. When several records
  share the name, only the ones holding a reservation count. If nothing
  matches, it falls back to the reservation at the name's DNS address.
- More than one match exits 3 and lists them on stderr. `--force` releases
  **every** match (preview with `--dry-run --force`).
- An unknown identifier, or a match with no fixed IP, exits 1. That includes
  `--dry-run`.

**Remove all.**

```bash
gofi ips clear --force --dry-run         # list what would go; exit 0
gofi ips clear --force --yes             # non-interactive form
```

- Without `--force`, it exits 2.
- With `--force` but not `--yes`, it prompts `Proceed? [y/N]` on **stdin**.
  Stdin that isn't `y` (including `/dev/null`) exits 3 `refused: not confirmed`.
- For each host, `ips clear` does what `ips rm --mac` does: it deletes the DNS
  record, then clears the fixed IP. User records are kept.
- It stops at the first host that fails and exits 1. Hosts before it are
  already released, so re-run `ips list` to see what's left.
- It is site-wide and irreversible except by re-import, so snapshot first
  (`ips export > before.hosts`), and only run it when the user explicitly asks.

### `dns`: static local DNS records

Use `dns` for DNS-only fixes, such as a stale record whose reservation must stay.

```bash
gofi dns list                                  # NAME VALUE TYPE TTL ID; --output json works
gofi dns rm --id 6a650dc764d5a843350f2f55       # only --id is guaranteed unique
gofi dns rm --name host1.lan.example --dry-run  # names are FQDN record keys
gofi dns rm --ip 192.168.2.10 --force           # remove every record with that value
```

- Pass exactly one of `--id`/`--name`/`--ip`, otherwise exit 2.
- More than one match without `--force` exits 3 and lists the matches.
- No match, including an unknown `--id`, exits 1.
- `--dry-run` prints `would delete:` lines, but still ends with
  `removed N record(s)`. **Nothing was removed.**
- If one delete in a batch fails, an `error:` line goes to stderr and the
  command **still exits 0**. Re-run `dns list` to verify.
- There is no `dns add`/`dns set`. To create a record, use `ips add`, which also
  reserves the IP.

### `network`: read-only

```bash
gofi network list                  # NETWORK VLAN SUBNET DHCP-POOL LEASE GATEWAY DNS
gofi network show Guest --output json
```

- Names are exact.
- An unknown name exits 1. A missing name argument exits 2.
- `(disabled)` in the pool column means no DHCP server (WAN, VLAN-only, or DHCP off).

### `clients`: who is connected now

```bash
gofi clients list            # sorted by IP; --output json gives []
gofi clients list --wifi     # -w
gofi clients list --wired    # -e   (--wifi plus --wired exits 2)
gofi clients vendor b4:0e:cf:2a:85:6f
```

- Vendor names always come from gofi's IEEE OUI cache at
  `$XDG_DATA_HOME/gofi/oui.txt` (fallback `~/.local/share/gofi/oui.txt`),
  never from the controller's `oui` field.
- The cache is downloaded from the internet on first use and when it is older
  than 30 days. A stale cache is used with a warning. No cache and no internet
  exits 1.
- `clients vendor` needs no credentials or controller, only the cache.
- An unregistered or randomized MAC reports vendor `unknown`, and still exits 0.

### `users`: known-client records

```bash
gofi users list
gofi users list --filter phone --output json  # substring of name, hostname, MAC, or fixed IP
gofi users rm --mac aa:bb:cc:00:00:02 --dry-run
gofi users rm --mac aa:bb:cc:00:00:02
```

- `rm` takes exactly one of `--mac`/`--name`, otherwise exit 2.
- An ambiguous `--name` exits 3 (`use --mac to disambiguate`). There is no `--force`.
- `rm` first clears any fixed IP, then forgets the client. It prints
  `removed fixed IP: <bool>, forgot client: <bool>`.
- If the forget fails after the clear succeeded, it exits 1, but the IP is
  already released.
- A fixed IP shown as `(dynamic)` is not active (`use_fixedip` is false).

### `profile`: networks, WLANs, and fixed IPs as JSON

```bash
gofi profile export > site.json               # always JSON, regardless of --output
gofi profile export --with-keys > site.json   # includes WiFi passphrases in CLEARTEXT
gofi profile import site.json --dry-run
gofi -S othersite profile import site.json --dns-domain lan.example
```

- The export has `site`, `captured`, `networks`, `wlans`, and
  `fixed_ips: [{"Hostname","MAC","IP"}]`. The fixed-IP keys are capitalized.
- Devices, firewall, routing, and port profiles are never included.
- Without `--with-keys`, the export prints a stderr note that passphrases were
  omitted. Treat a `--with-keys` file as a credential.
- `import` applies to the **invocation's** site: `-S`, else the target's site,
  else `default`. If that differs from the profile's site, it prints a note and
  proceeds.
- Only `fixed_ips` is really written, with the same semantics as `ips import`
  without `--force`.
- Networks and WLANs are only matched by name and reported (`no write endpoint
  yet`). A name missing on the target is reported as `skipping`.
- ⚠️ Fixed-IP per-entry errors are counted in the summary line, but the import
  **still exits 0**. Read the `fixed IPs: ... N errors` line.
- A malformed JSON file exits 2.

### `config`: gofi's own config file (local only)

The path is `$GOFI_CONFIG`, else `${XDG_CONFIG_HOME:-~/.config}/gofi/config.toml`.
A missing file is fine. A malformed file fails every command, even offline ones like `clients vendor`
(exit 1, with the parse error and line number).

```bash
gofi config show       # PATH / EXISTS / OUTPUT / TARGET (mode); --output json works
gofi config targets    # NAME (default marker) / MODE; JSON is a bare list of names
gofi config init       # writes a commented starter (mode 0600); exit 2 if the file exists, --force to overwrite
gofi --target cloud network list
```

Each target is one controller **and** one site. Use one entry per site. A
target is either local mode or connector mode, never a mix:

```toml
default = "home"      # used when --target is omitted
output  = "text"      # default --output

[targets.home]        # local mode
host             = "192.168.1.1"
site             = "default"
# port = 443   secure = false   username = "admin"
password_command = "pass show unifi/home"

[targets.cloud]       # connector mode (no host)
site            = "default"
console_id      = "your-console-id"
api_key_command = "pass show unifi/cloud-key"
```

`config init` writes a `home` target with `host = "192.168.1.1"` as the default.
After running it, every command without `--target` resolves to that local
target, unless `UNIFI_API_KEY` is set. Edit it or remove it.

## 5. Recommended agent workflow

1. **Orient.** Check `env | grep ^UNIFI_` and run `gofi config show` to learn
   the mode and target.
2. **Read.** Run `gofi network list` and `gofi ips list` to confirm the right
   controller and site. Use `--output json` for parsing (except for `ips`,
   which is always ISC text).
3. **Snapshot** before bulk changes: `gofi ips export > before.hosts` and
   `gofi profile export > before.json`. Re-importing `before.hosts` restores
   reservations that were removed, but it does not delete ones that were added.
4. **Preview.** Run every write that has `--dry-run` with it first, and show
   the user the preview.
5. **Write**, then **verify** by re-listing. Don't trust the exit code alone
   for `dns rm` or `profile import`, which can exit 0 with errors.
6. Use `--force` only to override the specific guard that refused (exit 3),
   and only when the user wants that. Never run `ips clear` unless the user
   explicitly asks.

## 6. Running against a different or test controller

To aim one command at a controller other than the one your environment selects
(a lab UDM, or a local test server), strip the connector variables and isolate
the config file so neither the ambient API key nor a default target applies:

```bash
env -u UNIFI_API_KEY -u UNIFI_CONSOLE_ID GOFI_CONFIG=/nonexistent/config.toml \
  UNIFI_USERNAME=admin UNIFI_PASSWORD=... gofi -H 127.0.0.1 -p 8443 ips list
```

Without the `env -u`, an exported API key silently sends the command to the
real console (§1). A missing `GOFI_CONFIG` file is fine; gofi then uses only
the flags and environment.

## 7. Controller behaviors behind the command set

These real UDM behaviors explain why some commands exist and others don't:

- Removing a known client uses the station manager's batch `forget-sta`
  command. `DELETE` on a user record answers 404, so gofi never deletes user
  records. `ips rm` and `ips clear` only release the reservation, and
  `users rm` forgets the client.
- A single static DNS record can't be fetched by ID (405); gofi reads the whole
  list. There is no verified standalone DNS write, hence no `dns add`/`dns set`.
- No network/VLAN or WLAN write has been verified against hardware, hence no
  `network set`, and `profile import` writes only fixed IPs.

## 8. Installing this skill

`gofi install-skill` writes this skill, embedded in the binary, to
`$XDG_DATA_HOME/skills/gofi-cli/` (default `~/.local/share/skills/gofi-cli/`).
It needs no config, credentials, or controller, and never prompts.

| Exit | Meaning |
|---|---|
| `0` | Installed, replaced, or already identical (prints `nothing changed`) |
| `3` | A different copy is installed; left alone. Re-run with `-f`/`--force` to replace it |
| `1` | The store isn't writable, or neither `XDG_DATA_HOME` nor `HOME` is usable |

Before using `--force`, check whether the installed copy was edited by hand on
purpose. Agents that read only their own skills directory (such as
`~/.claude/skills`) see the skill once a mirror step copies or links the store
into it, for example `sm sync`.
