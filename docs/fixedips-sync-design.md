# fixedips-sync Design

Date: 2026-07-16

## Purpose

A round-trip CLI utility (`examples/fixedips-sync`) that:

1. **Dumps** all fixed-IP reservations together with their local DNS names, in
   two formats: a human-readable table and an editable JSON document.
2. **Applies** an edited JSON document back to the controller so that the
   controller's fixed-IP + DNS state matches the file (full reconcile).

It fills the gap left by the existing examples: `fixedips` lists reservations
but omits DNS names and cannot write; `addfixedip`/`delfixedip` mutate a single
reservation at a time.

## Concepts (UniFi data model)

- **site** — a logical grouping of networks and devices under one controller.
  The API path is `/api/s/{site}/...`. Most setups have exactly one, `default`.
- **network** — a LAN/VLAN with a subnet and DHCP scope (e.g. `LAN` =
  192.168.1.0/24). A fixed IP is bound to one network because the reserved IP
  must fall inside that network's subnet. `network_id` is UniFi's internal
  opaque ID (stored on the reservation, `User.NetworkID`); `network_name` is the
  friendly label from `Networks().List`.
- **name** — `User.Name`, the reservation's cosmetic alias in the UniFi clients
  list. It does NOT create DNS resolution.
- **dns_names** — actual local DNS A-records (`DNSRecord.Key` -> IP) served by
  the UDM resolver. A reservation may have zero, one, or many. Multiple A
  records pointing at one IP is fully supported (`DNS().GetByIP` returns a
  slice); each record's identity is the `(hostname, IP)` pair.

`name` and `dns_names` are independent objects in different stores; they are not
required to match.

## Command surface

```
fixedips-sync dump  [-o file.json] [--table]   # read-only
fixedips-sync apply -f file.json [--confirm]   # dry-run unless --confirm
```

Shared flags and env vars follow the other examples:
`-H/--host`, `-p/--port`, `-s/--site`, `-k/--insecure`, and
`UNIFI_USERNAME`, `UNIFI_PASSWORD`, `UNIFI_UDM_IP`.

### dump

- Fetch fixed-IP users: `Users().List`, filter `UseFixedIP && FixedIP != ""`.
- For each reservation, fetch its DNS names: `DNS().GetByIP(fixed_ip)`, collect
  `Key` values of `A` records.
- Resolve `network_name` from `Networks().List` by `network_id`.
- Default output: JSON to **stdout** (redirect to a file yourself); `-o` writes
  to a file instead.
- `--table` prints a human table instead of JSON:
  `NAME  MAC  FIXED IP  NETWORK  DNS NAMES`.
- Reservations sorted by fixed IP for stable output.

### File format (JSON)

```json
{
  "site": "default",
  "reservations": [
    {
      "mac": "aa:bb:cc:dd:ee:ff",
      "name": "Greg's NAS",
      "fixed_ip": "192.168.1.50",
      "network_id": "net123",
      "network_name": "LAN",
      "dns_names": ["nas", "media", "backup"]
    }
  ]
}
```

Identity keys: reservation = `mac`; DNS record = `(hostname, fixed_ip)`.

### apply (full reconcile)

The file is the source of truth. Diff desired (file) against current
(controller):

Reservations (matched by MAC):
- in file, not on controller -> CREATE user (`Users().Create`)
- on controller, not in file -> DELETE reservation (clear/remove fixed IP)
- in both, differ in IP/name/network -> UPDATE (`Users().Update`)

DNS names (per reservation, matched by `(hostname, ip)`):
- in file, not on controller -> CREATE A record
- on controller, not in file -> DELETE A record
- otherwise -> no change

Ordering respects the UniFi constraint `LocalDnsRecordRequiresFixedIp`
(observed in `delfixedip`): on additions, create the fixed IP before its DNS
records; on removals, delete DNS records before clearing/removing the
reservation.

Network resolution on apply: prefer `network_id`; fall back to `network_name`;
fall back to IP-subnet auto-detect (reuse the CIDR logic from `addfixedip`).

Safety:
- **Dry-run by default.** Print the plan (`+ CREATE`, `~ UPDATE`, `- DELETE`)
  and exit without writing. `--confirm` executes.
- **Site guard.** If the file's `site` differs from the `-s` flag, refuse to
  apply (prevents pushing one site's file to another).
- Only fixed-IP users and A-records pointing at those IPs are ever touched;
  regular clients and unrelated DNS records are never modified.

## Scope boundaries

- Only fixed-IP reservations and their `A` DNS records. Standalone DNS records
  (CNAME/MX/TXT, or A records not tied to a fixed IP) are out of scope and left
  untouched.
- IPv4 only, matching `addfixedip`.

## Testing

- Unit tests run against the mock server (per project rule: every function has a
  test, every endpoint is mocked).
- Cover: dump JSON shape, dump `--table`, dump with multiple DNS names per IP,
  apply create/update/delete for reservations, apply create/delete for DNS
  names, apply ordering under the fixed-IP DNS constraint, dry-run makes no
  writes, `--confirm` writes, and site-guard rejection.

## Build integration

Add `fixedips-sync` to `EXAMPLES` in the `Makefile`.
