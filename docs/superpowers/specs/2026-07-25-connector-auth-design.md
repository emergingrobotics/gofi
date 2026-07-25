# Design: Cloud-key + Site Manager Connector auth for gofi

Date: 2026-07-25
Status: Approved for planning
Scope: `github.com/emergingrobotics/gofi` (module path currently `github.com/unifi-go/gofi`)

## Goal

Add a second authentication mode — a **cloud API key** (`X-API-KEY`) that reaches a
UDM through the **UniFi Site Manager connector proxy** — *transparently and in place*,
without removing or altering the existing local username/password (cookie) mode, and
without touching any of the ~40 endpoint call sites in `services/`.

Non-goal (deferred): a local `X-API-KEY` pointed straight at a UDM's LAN address, and
fan-out to multiple consoles in a single CLI invocation.

## Background: two API surfaces, one library

gofi drives the **classic/local UniFi Network API** exclusively:

- Classic v1 — `/proxy/network/api/s/{site}/...`
- Classic v2 — `/proxy/network/v2/api/site/{site}/...`
- UniFi OS  — `/api/self`, `/api/status`, `/api/auth/login`, `/api/logout`

The **cloud Site Manager API** (`https://api.ui.com`, key from unifi.ui.com) has no
native config surface — only `GET /v1/hosts`, `GET /v1/devices`, read-only
sites/ISP-metrics/SD-WAN reporting, and a transparent **connector** at
`/v1/connector/consoles/{id}/proxy/...` that forwards any request to the on-prem
console. The connector is therefore the *only* way to keep gofi's full feature set
(firewall, DNS, fixed IPs, routing, port forwards) while authenticating with a key and
reaching the console from off-LAN.

## Hardware verification (2026-07-25)

Probed live against two enrolled consoles with a cloud key scoped to UniFi
Applications → Network. All paths the three CLI tools depend on returned `200` through
the connector on both consoles; both sites report `internalReference: "default"`.

| Path (dependent tool)              | Console A `E438839D` | Console B `74ACB95F` |
|------------------------------------|----------------------|----------------------|
| `integration/v1/sites`             | 200 (`default`)      | 200 (`default`)      |
| `v1 rest/networkconf` (gofinet)    | 200                  | 200                  |
| `v2 static-dns` (gofips)           | 200                  | 200                  |
| `v1 stat/sta` (gofimac)            | 200                  | 200                  |

Conclusion: a cloud key **does** authenticate the classic v1 and v2 paths through the
connector — not just the integration surface. Connector mode is viable for all three
tools. (`/api/self` and `/api/status` UniFi-OS paths through the connector remain
unverified; probe before relying on them in connector mode.)

## Approach (chosen: A — transport-level base-URL + path-prefix)

Mode is decided once at construction. Every service keeps building the same local path
(e.g. `/proxy/network/v2/api/site/default/static-dns`); the transport, in connector
mode, prepends a prefix and swaps the base URL so the wire request becomes
`https://api.ui.com/v1/connector/consoles/{id}/proxy/network/v2/api/site/default/static-dns`.
Rejected alternatives: a dedicated connector decorator transport (extra layer for a
two-line change) and per-service URL building (touches every call site, defeats
"transparent").

## §1 Config surface & mode selection

New `gofi.Config` fields:

```go
APIKey    string // cloud API key; auth via X-API-KEY, no login request is made
ConsoleID string // Site Manager console ID; selecting it enables connector mode
```

New options (`options.go`):

```go
func WithAPIKey(key string) Option          // sets APIKey
func WithConnector(consoleID string) Option  // sets ConsoleID, BaseURL=https://api.ui.com,
                                             // derives PathPrefix /v1/connector/consoles/{id}
```

Mode selection, resolved in `New()` *after* options are applied:

| Config presented                       | Mode                | Auth manager        |
|-----------------------------------------|---------------------|---------------------|
| `ConsoleID` + `APIKey`                  | connector → cloud   | `auth.NewAPIKey()`  |
| `Username` + `Password` (no `APIKey`)   | local LAN (today)   | `auth.New(...)`     |

Validation (moved to *after* the option loop — see §4 bug fixes):

- Local mode: `Host` required (unchanged).
- Exactly one credential set: `APIKey` **or** (`Username`+`Password`). Zero → error
  naming both. Both → error (do not silently prefer one).
- `Host`/`Port` and `ConsoleID` are mutually exclusive → error if both set.
- Reuse `NewValidationError` from `errors.go`.

## §2 Transport changes (`transport/config.go`, `transport/transport.go`)

Add to `transport.Config`:

```go
APIKey     string // injected as X-API-KEY on every request
PathPrefix string // prepended to every request path (connector routing)
```

In `Do()`:

- Set `X-API-KEY` from the stored key *before* the `req.Headers` loop, so per-request
  header overrides still win. Leave CSRF logic untouched (harmless when unused).
- Join `PathPrefix + req.Path` as strings with exactly one `/` at the boundary, then
  resolve against `baseURL`. Do **not** use `url.Parse` relative resolution — an
  absolute (leading-slash) `req.Path` would discard the prefix.
- Do **not** add `SetAPIKey` to the `Transport` interface. The key is known at
  construction; widening the interface would break `NewRetryTransport` and the mock for
  no benefit.

## §3 Auth package (`auth/auth.go`)

Keep the `Manager` interface unchanged. Add a second implementation (no branching inside
`manager`):

```go
// NewAPIKey returns a Manager for API-key auth. There is no session to establish,
// so Login and Logout are no-ops.
func NewAPIKey() Manager
```

Semantics: `Login`, `Logout`, `EnsureAuthenticated` return nil; `IsAuthenticated`
returns true; `Session` returns nil.

`Session() == nil` while `IsAuthenticated() == true` is a new combination. Grep for
callers assuming `IsAuthenticated() ⇒ non-nil Session()` and fix them.
`client.IsConnected()` (`connected && auth.IsAuthenticated()`) stays correct but gets a
key-mode test.

## §4 Client wiring & standing bug fixes (`client_impl.go` `New()`)

Fix two existing bugs in the same pass:

1. **Validation before options.** Move the `Host`/credential validation block to *after*
   the option loop, so `New(&Config{Host:h}, WithAPIKey(k))` no longer errors before the
   option is seen. (Enables §1.)
2. **Default timeout is 15 min.** `config.Timeout = 30 * transport.DefaultConfig("").Timeout`
   (= 900s) drops the `30 *`; resolved default becomes 30s. Add a regression test.

Then:

- Select `auth.NewAPIKey()` when `config.APIKey != ""`, else `auth.New(...)`.
- Connector mode: `BaseURL = https://api.ui.com`, set `transportConfig.PathPrefix =
  /v1/connector/consoles/{ConsoleID}`. Local mode: build `https://host:port` as today.
- Set `transportConfig.APIKey = config.APIKey`.
- `Connect()` in key mode issues one cheap classic GET (e.g. `.../rest/networkconf`,
  verified 200 above) so a bad key or unreachable console fails at connect time,
  preserving today's contract that `Connect()` surfaces auth failures. Endpoint choice
  is explicit and testable; single request only.

## §5 CLI plumbing (`utilities/{gofips,gofimac,gofinet}/main.go`)

Add env vars (no new credential flags — keys belong in the environment, not argv):

| Env                     | Effect                                            |
|-------------------------|---------------------------------------------------|
| `UNIFI_API_KEY`         | selects key auth; preferred when set              |
| `UNIFI_CONSOLE_ID`      | selects connector mode; host/port ignored         |
| `UNIFI_USERNAME/PASSWORD/CONTROLLER_IP` | local mode (unchanged)            |

- Precedence: `UNIFI_API_KEY` wins over username/password. If both present, use the key
  and write one line to stderr saying so.
- Missing-credentials error text names `UNIFI_API_KEY` first.
- Each tool builds the client with `WithAPIKey` + (when `UNIFI_CONSOLE_ID` set)
  `WithConnector`.
- Update each `flag.Usage` "Environment Variables" section; keep the existing `env*`
  const pattern in `gofips`.
- One console per invocation (target via `UNIFI_CONSOLE_ID`). Fan-out to all consoles is
  out of scope; expose the console-list capability so a future `--all-consoles` is easy.

## §6 Mock server (`mock/server.go`, `mock/handlers_auth.go`, `mock/options.go`)

- Accept a configured key via `X-API-KEY`: treat the request as authenticated without a
  prior login, and bypass the CSRF check for key-authenticated requests.
- `mock.WithAPIKey(key)` option, matching existing option style.
- Wrong key → `403` with the verified live body:
  `{"code":"forbidden","httpStatusCode":403,"message":"insufficient permissions"}`.
- Strip a configured `/v1/connector/consoles/{id}` prefix before dispatch so connector
  mode is testable offline.

## §7 Tests

Offline (mock):

- `client_test.go` — construction succeeds with `APIKey` and no user/pass; fails with
  neither; fails with both; `Host`+`ConsoleID` mutually-exclusive error.
- `client_test.go` — resolved default timeout is 30s (regression guard for §4).
- `auth/auth_test.go` — the `NewAPIKey` manager's five methods.
- `transport/transport_test.go` — `X-API-KEY` sent when configured, absent when not,
  per-request override wins.
- `transport/transport_test.go` — `PathPrefix` prepend: with prefix, no prefix, prefix
  with and without trailing slash, absolute `req.Path`.
- One end-to-end service call per auth mode through the mock.
- `IsConnected()` in key mode.

Live (documented runbook, validated 2026-07-25 — see verification table): the Phase 1
`/integration/v1/sites` discovery and Phase 2 classic/v2 gate probes, per console.

## Docs to update (follow-on, in the plan)

- `README.md` quickstart — lead with `UNIFI_API_KEY`, keep user/pass as fallback; add
  key-generation steps (console UI → profile → API → Create API Key; cloud key from
  unifi.ui.com needs UniFi Applications → Network scope; a key inherits its creator's
  permissions).
- `CLAUDE.md` "Key Technical Details" — describe both mechanisms and which endpoints each
  covers.
- `doc.go`, `EXAMPLES.md`, `examples/*/main.go` — switch credential setup to `APIKey`.

## Out of scope, reported separately

- **Step 0 hygiene (independent, do first):** `cookies.txt` (contains a UniFi session
  JWT with `userId`/`passwordRevision`; `exp` passed) and empty `ssh` are still tracked.
  `git rm --cached` both and add to `.gitignore`. History rewrite (filter-repo/BFG) only
  if the repo is/will be public; rotate the `192.168.4.1` admin password regardless. Do
  not rewrite history without explicit confirmation.
- `go.mod` module path (`github.com/unifi-go/gofi`) ≠ repo (`emergingrobotics/gofi`);
  separate commit, touches every import.
- `client.Events()` returns nil ("Phase 18"); `client.Events().Subscribe(...)` panics.
  Unrelated to auth.
