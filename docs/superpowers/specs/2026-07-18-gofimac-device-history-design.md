# gofimac Device History & Tenure — Design

Date: 2026-07-18
Status: Approved

## Overview

Extend `gofimac` from a point-in-time active-client lister into a tool that also
answers *"what's new?"*, *"who left and when?"*, and *"how long has this been
around?"* — using data the UDM already provides (`first_seen`, `last_seen`,
`uptime`). No local state is introduced. Snapshot/diff is explicitly out of scope
and reserved for a future spec.

## CLI Surface

```
gofimac [conn flags] [--wifi|--wired|--all] [-j] [--since DUR] [--gone [DUR]] [--sort KEY]
```

| Flag | Short | Meaning |
|------|-------|---------|
| `--since DUR` | none | Switch data source to `ListAll`; show every device seen within `DUR` (present + gone), each marked by STATUS. |
| `--gone [DUR]` | none | `ListAll` minus currently-active; only departed devices. `DUR` optional, default `7d`. |
| `--sort KEY` | none | Sort order: `first-seen` (default), `last-seen`, or `ip`. |

Rules:

- `--since` and `--gone` are mutually exclusive; specifying both is an error.
- Default (neither `--since` nor `--gone`) keeps today's data source (`ListActive`),
  active clients only. Behavior unchanged except sort default and added columns.
- `--wifi`/`--wired` still filter by link type. When combined with `--gone`/`--since`,
  a stderr warning notes that link fields may be stale or missing for departed devices.
- Invalid `--sort` value is an error before connecting.

## Duration Parser (`duration.go`, new)

`time.ParseDuration` does not support days; a custom parser is required.

- Accepted units: `s`, `m`, `h`, `d`, `w`, `mo`.
- Unit values: `s`=1s, `m`=1min, `h`=1h, `d`=24h, `w`=7d, `mo`=30d.
- Compound durations allowed, e.g. `1w2d`, `36h`, `3mo`.
- `mo` (month) must be matched before `m` (minute) during tokenizing.
- Result converted to whole hours, **rounded up**, passed to `ListAll` as `withinHours`.
- Empty/invalid syntax returns an error.

## Sort — Behavior Change

Default sort changes from IP ascending to **`first_seen` descending** (newest
device to the network on top). This is a deliberate, documented break from
current behavior.

Sort keys:

- `first-seen` (default): `first_seen` descending. Ties broken by IP ascending
  (no-IP last, then MAC), which preserves today's ordering when `first_seen` is
  absent/zero.
- `last-seen`: `last_seen` descending, same tie-break chain.
- `ip`: IP ascending, no-IP last, then MAC — today's exact behavior.

Because existing fixtures carry no `first_seen`, the `first-seen` default reduces
to the IP tie-break for them; existing sort tests remain valid.

## Text Output

Default / active view (6 columns):

```
MAC                IP             HOSTNAME   OUI-MANUFACTURER   AGE   LAST-SEEN
```

`--since` / `--gone` view adds a STATUS column (7 columns):

```
MAC   IP   HOSTNAME   OUI-MANUFACTURER   AGE   LAST-SEEN   STATUS
```

- `AGE` = relative `first_seen`; `LAST-SEEN` = relative `last_seen`.
- `STATUS` = `present` or `gone`; shown only in `--since`/`--gone` views.
- Relative-time format: `now`, `4m`, `2h`, `5d`, `3mo`, `2y`; `-` when the
  timestamp is zero/unknown.
- Header printed only when there is at least one entry (matches current behavior).
- Missing IP still renders as `-`.

## JSON Output (`-j`)

Every object gains:

- `first_seen` (unix int, always present)
- `status` (`"present"` or `"gone"`, always present)

`last_seen` is already present. Existing `omitempty` behavior for link-specific
fields is preserved.

## Presence Computation

- Default mode: `ListActive` only; every entry `status = present`.
- `--since` / `--gone` mode:
  1. Fetch `ListActive`, build a set of active MACs.
  2. Fetch `ListAll(within = parsed duration in hours)`.
  3. Union of MACs. For each MAC: if present in the active set, use the active
     record (richer, current IP/link fields) and mark `present`; otherwise use
     the all-users record and mark `gone`.
  4. `--gone`: keep only `gone` entries.
  5. Apply link-type filter on the chosen source record.

## Code Layout

- `duration.go` (new) + `duration_test.go` — duration parser.
- `operations.go` — `FirstSeen`/`Status` on `ClientEntry`; `SortMode` type and
  `--sort` parsing; presence-set logic; history-mode listing; `buildClientEntry`
  gains a `present bool` param; `sortClientEntries` gains a `SortMode` param.
- `format.go` — `AGE`/`LAST-SEEN`/`STATUS` columns; relative-time formatter
  `formatRelativeTime(epoch, now int64) string`; new JSON fields. `FormatText`
  gains `showStatus bool` and `now int64` params.
- `main.go` — new flags, mutual-exclusion + sort-value validation, mode dispatch,
  stale-field warning.

## Testing

- Duration parser: valid single/compound units, `mo` vs `m`, invalid input,
  round-up-to-hours conversion.
- Relative-time formatter: boundaries (0→`-`, <60s→`now`, 59m, 60m→`1h`, day,
  month, year cutoffs).
- Sort: `first-seen` desc, ties fall back to IP; `last-seen` desc; `ip`; missing
  values sort last. Existing IP-sort tests remain green.
- Presence: overlapping active/all sets → correct `present`/`gone` labels;
  `--gone` filters to departed only; active record preferred for present devices.
- Flag validation: `--since` + `--gone` → error; bad duration → error before
  connect; bad `--sort` → error.
- Text: column counts per mode, AGE/LAST-SEEN rendering, STATUS only in history
  views. JSON: `first_seen` and `status` always present.

Every new function has a test (project rule #1). The mock server already serves
`stat/alluser`.
