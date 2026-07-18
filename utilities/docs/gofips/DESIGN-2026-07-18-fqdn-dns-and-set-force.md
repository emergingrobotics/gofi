# gofips: FQDN DNS records + `--force` for `--set`

Date: 2026-07-18
Status: Approved for implementation
Scope: `utilities/gofips/` only. No changes to the `gofi` module or other tools.

## Problem

`gofips` (and its predecessor `gofip`) create fixed-IP reservations whose DNS A
records are keyed on the **bare** hostname (`helios`), never the FQDN
(`helios.herlein.me`). On a network with a local domain (`herlein.me`), Linux
clients running systemd-resolved qualify a bare single-label lookup (`ping
helios`) to the FQDN (`helios.herlein.me`) and never send the bare name
upstream. The UDM answers the FQDN with NXDOMAIN, so `ping helios` fails from
those clients even though `nslookup helios <udm>` (a direct bare query)
succeeds.

Two concrete defects:

1. `ensureDNSRecord` writes `Key = hostname` (bare). See
   `utilities/gofips/operations.go`.
2. `--set` skips an entry entirely when the user record is unchanged
   (`operations.go` skip branch), and the skip path never reaches
   `ensureDNSRecord`. A host whose reservation is already correct but whose DNS
   record is missing or stale is therefore never repaired by `--set`.

## Goals

- DNS A records are keyed on the FQDN (`hostname.domain`) so name resolution
  works for domain-qualifying clients.
- A plain `gofips --set` repairs missing or stale DNS records even when the
  user record is unchanged.
- A new use of the existing `--force` flag makes `--set` re-process every entry
  fully (including redundant user-record rewrites), never skipping.

## Non-goals

- No change to the ISC DHCP file format (`host <name> { ... }` still uses the
  bare hostname; the domain is applied only when writing UDM DNS records).
- No SSH/dnsmasq changes on the UDM. The fix creates an explicit FQDN A record,
  which resolves regardless of the (unconfirmed) theory that a bare-key record
  suppresses the UDM's automatic FQDN registration.

## Design decisions (agreed)

1. **DNS key = FQDN only.** `Key = hostname + "." + domain`. Bare-name
   resolution continues to be served by the UDM's own reservation mechanism.
2. **Replace stale bare records.** When writing the FQDN record for a host that
   still has an old bare-key record, delete the bare record first.
3. **Plain `--set` auto-repairs DNS.** When the user record is unchanged,
   `--set` still ensures the DNS record is correct; it only skips the redundant
   user-record write. `--force` additionally forces the user-record rewrite.

## Domain source

The owning network's `DomainName` field (`types.Network`, JSON `domain_name`).
`detectNetwork()` already maps an IP to its network by subnet. It will be
changed to return the `*types.Network` (callers use `.ID` as before) so the
domain is available without a second lookup. If the owning network has no
`DomainName` (e.g. an IoT VLAN), fall back to a bare-key record.

## Component changes

### `ensureDNSRecord` (operations.go)

New signature:

```go
func ensureDNSRecord(ctx context.Context, client gofi.Client, site, hostname, domain, ipAddress string, dryRun bool) error
```

Logic:

1. Compute `fqdn`: `hostname + "." + domain` when `domain != ""`, else
   `hostname` (bare fallback).
2. If `fqdn != hostname`, look up `GetByName(hostname)`; if a bare-key record
   exists (`Key == hostname`), delete it (dry-run: print `would replace`).
3. `GetByName(fqdn)`:
   - present and `Value == ipAddress`: no-op (return nil).
   - present and `Value != ipAddress`: update value.
   - absent: create `{Key: fqdn, Value: ip, RecordType: A, Enabled: true}`.
4. Each mutation prints an action line to stderr
   (`created/updated/replaced DNS: <fqdn> -> <ip>`). Dry-run prints
   `would ...` and mutates nothing.

### `DoSet` (operations.go) + `main.go`

New signature: `DoSet(ctx, client, site, entries, dryRun, force bool)`.
`main.go` passes the already-parsed `*force`.

Per entry with an existing user (matched by MAC):

| Case | User record | DNS |
|------|-------------|-----|
| unchanged, no `--force` | skip write, count `Skipped` | ensure/repair |
| unchanged, `--force`    | rewrite, count `Updated`   | ensure/repair |
| changed                 | update, count `Updated`    | ensure/repair |

"Unchanged" keeps the current definition: `UseFixedIP && FixedIP == entry.IP &&
(Name == entry.Hostname || Hostname == entry.Hostname)`.

New users (no existing MAC): unchanged from today (create/update user) then
ensure DNS with the FQDN key.

Dry-run: existing "would create/update" user messages retained; DNS drift on
unchanged entries surfaces via `ensureDNSRecord`'s dry-run `would ...` output.

### `DoAdd` + `checkAddConflicts` (operations.go)

Thread `domain` (from the detected network) into the `ensureDNSRecord` call and
into the DNS conflict check, so the conflict check queries the FQDN
(`GetByName(fqdn)`) rather than the bare hostname.

### `DoGet` warning (operations.go)

DNS keys are now FQDNs. The "DNS hostname differs from user hostname" warning
must compare the **first label** of the DNS key against the user hostname
(`helios.herlein.me` vs `helios`) so it does not warn spuriously on every host.

### `DoDel` (operations.go)

No change. Deletion is by IP value (`GetByIP`), which matches the FQDN record
(its `Value` is the IP). Both bare and FQDN records pointing at the IP are
removed.

### Usage text (main.go)

Note that `--force` also forces `--set` to re-process unchanged entries.

## Error handling

Unchanged model: DNS failures during `--set`/`--add` print a warning to stderr
and continue (non-fatal); `--set` exits 1 only if hard (user-record) errors
occurred. Networks without a `DomainName` fall back to bare-key records.

## Testing (per project rule: every function has a test, via mock)

Existing `DoSet` test call sites (5) updated for the new `force` parameter.
Test networks gain a `DomainName` where FQDN behavior is asserted.

New/updated cases:

- `ensureDNSRecord`: builds FQDN key; replaces existing bare record
  (delete + create); updates value in place; no-op when already correct;
  bare fallback when domain empty; dry-run mutates nothing.
- `DoSet`: unchanged + no force repairs/creates DNS but does not rewrite the
  user record (counts `Skipped`); unchanged + force rewrites user and ensures
  DNS (counts `Updated`); changed updates both; counters correct.
- `detectNetwork` returns the owning network (domain accessible).
- `DoGet`: FQDN DNS key matching a host does not trigger the mismatch warning.
- `DoAdd`: creates an FQDN-keyed record; conflict check uses the FQDN.

Run `make test` (mock-backed) to verify.
