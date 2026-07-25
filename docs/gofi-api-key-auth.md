# gofi: add API-key authentication

Task doc for Claude Code. Repo: `github.com/emergingrobotics/gofi`.

Goal: support UniFi API keys (`X-API-KEY`) as a first-class auth mechanism alongside the
existing local-admin username/password flow, and optionally support reaching a console
through the Site Manager connector proxy instead of directly over the LAN.

Read `CLAUDE.md` before starting. Its rules apply: every function gets a test, every
endpoint is supported in the mock server, `make test` must pass.

---

## Step 0 — Blocking: purge a committed credential

`cookies.txt` is **tracked in git** and contains a live-format UniFi session JWT for
`192.168.4.1`, including `userId`, `passwordRevision`, and an embedded `csrfToken`. The
`exp` claim has passed, so the token itself is dead, but it must not stay in the tree or
the history.

Do this first, independently of the rest:

1. `git rm --cached cookies.txt` and add `cookies.txt` to `.gitignore`.
2. Flag to the user that history rewriting (`git filter-repo` or BFG) is needed if this
   repo is or will be public, and that the admin password for the console at
   `192.168.4.1` should be rotated regardless, since `passwordRevision` is in the payload.
3. Do not rewrite history yourself without explicit confirmation.

While in `.gitignore`: `gofinet` (a ~9 MB compiled binary) and `ssh` (empty file) are
both tracked at the repo root. The existing `.gitignore` anchors names like `/gofips` to
root but misses `gofinet`. Untrack both.

---

## Step 1 — Verification gate (do not skip)

The whole design depends on one unverified fact: **does `X-API-KEY` authenticate the
classic `/proxy/network/api/...` endpoints that this library actually calls?**

gofi does not use the official Integration API anywhere. Every path in `services/` is
either classic v1 (`/proxy/network/api/s/{site}/...`), v2
(`/proxy/network/v2/api/site/{site}/...`), or UniFi OS level (`/api/self`, `/api/status`).
Community reports conflict on whether console-issued API keys are accepted on those
classic routes. Third-party clients that target the classic API have added key support,
which suggests yes on recent firmware, but treat that as unconfirmed.

Ask the user to run these two probes and report the status codes:

```bash
# Local, console-issued key, classic endpoint
curl -sk -o /dev/null -w 'local classic: %{http_code}\n' \
  -H "X-API-KEY: $UNIFI_LOCAL_KEY" \
  "https://$UDM_IP/proxy/network/api/s/default/rest/networkconf"

# Local, console-issued key, v2 endpoint (this is what gofips depends on)
curl -sk -o /dev/null -w 'local v2:      %{http_code}\n' \
  -H "X-API-KEY: $UNIFI_LOCAL_KEY" \
  "https://$UDM_IP/proxy/network/v2/api/site/default/static-dns"
```

Branch on the result:

- **Both 200** — proceed with the full plan below. Key auth replaces cookie auth outright
  and username/password becomes a legacy fallback.
- **Classic 401/403** — key auth only works on `/integration/v1/...`. Do **not** try to
  retrofit keys onto the classic paths. Report back to the user before writing code: the
  work becomes "add an integration-API client alongside the classic one," which is a much
  larger change and needs its own design pass. Note that firewall rules, port forwards,
  and static routes are not exposed on the integration surface at all, so the classic
  path cannot be retired.
- **Mixed** — report the split and stop. Partial support needs a decision from the user
  about which services move.

Record the observed result in the commit message or a short note in `docs/`. Future
readers should not have to re-derive it.

---

## Step 2 — `Config` and option plumbing

`config.go` and `client_impl.go`.

Add to `Config`:

```go
// APIKey authenticates via the X-API-KEY header. When set, Username and
// Password are ignored and no login request is made.
APIKey string
```

Add to `options.go`:

```go
func WithAPIKey(key string) Option {
    return func(c *Config) { c.APIKey = key }
}
```

Then fix two existing bugs in `New()` in `client_impl.go`, both of which will otherwise
break this feature:

**Validation runs before options are applied.** The current order is: validate
`Username`/`Password` non-empty → apply defaults → `for _, opt := range opts { opt(config) }`.
A caller passing `gofi.New(&gofi.Config{Host: h}, gofi.WithAPIKey(k))` gets a validation
error before `WithAPIKey` is ever seen. Move the whole validation block to after the
option loop.

**Default timeout is 15 minutes.** This line multiplies the already-30-second default:

```go
config.Timeout = 30 * transport.DefaultConfig("").Timeout   // = 900s
```

Should be `transport.DefaultConfig("").Timeout`. Fix it in the same pass and add a test
asserting the resolved default is 30s.

New validation logic, after options are applied:

- `Host` required (unchanged).
- Exactly one of `APIKey` or (`Username` + `Password`) must be present. Zero → validation
  error naming both options. Both → validation error; do not silently prefer one.
- Keep using `NewValidationError` from `errors.go` for consistency.

---

## Step 3 — Transport changes

`transport/config.go` and `transport/transport.go`. Both changes belong here rather than
in `services/`, because every service already routes through `httpTransport.Do`.

**Header injection.** Add `APIKey string` to `transport.Config`, store it on
`httpTransport`, and in `Do()` set `X-API-KEY` alongside the existing header setup. Set it
before the `req.Headers` loop so per-request overrides still win. Leave the CSRF logic
alone — it is harmless when unused, and mixing the two is what a real browser session does
anyway.

Do not add a `SetAPIKey` method to the `Transport` interface. That interface is
implemented by `NewRetryTransport` and the mock, and widening it is a breaking change for
no benefit; the key is known at construction time.

**Optional connector prefix.** To support reaching a console through the Site Manager
connector without touching any of the ~40 call sites in `services/`, add:

```go
// PathPrefix is prepended to every request path. Used to route through the
// Site Manager connector proxy, e.g.
// "/v1/connector/consoles/{consoleID}" with BaseURL "https://api.ui.com".
PathPrefix string
```

In `Do()`, prepend it before `t.baseURL.Parse(req.Path)`. Verified working shape, tested
against a live console:

```
https://api.ui.com/v1/connector/consoles/{consoleID}/proxy/network/integration/v1/sites
```

The connector is a transparent pass-through — the local path is appended verbatim, the
same `X-API-KEY` header authenticates it, and `GET` works directly (no POST envelope).
Requires console firmware 5.0.3+.

Surface this in the top-level `Config` as `ConsoleID string` plus a `WithConnector()`
option that sets `BaseURL` to `https://api.ui.com` and derives the prefix, so callers
never hand-assemble the URL. Keep `Host`/`Port` and `ConsoleID` mutually exclusive and
validate that.

Note for the user: `/api/self` and `/api/status` in `services/site.go` and
`services/system.go` are UniFi OS paths, not Network paths. Whether they resolve through
the connector prefix is unverified — probe before relying on them in connector mode.

---

## Step 4 — `auth` package

`auth/auth.go`. Keep the `Manager` interface unchanged and add a second implementation
rather than branching inside `manager`:

```go
// NewAPIKey returns a Manager for API-key auth. There is no session to
// establish, so Login and Logout are no-ops.
func NewAPIKey() Manager
```

Semantics: `Login` and `Logout` return nil; `EnsureAuthenticated` returns nil;
`IsAuthenticated` returns true; `Session` returns nil.

`Session() == nil` while `IsAuthenticated() == true` is a new combination — grep for
callers that assume `IsAuthenticated()` implies a non-nil `Session()` and fix them. Also
check `client.IsConnected()` in `client_impl.go`, which ANDs `connected` with
`auth.IsAuthenticated()`; it should still behave correctly but needs a test in key mode.

In `New()`, select the implementation on `config.APIKey != ""`.

Consider having `Connect()` in key mode issue one cheap read (`/api/self` or the
integration `/info`) so a bad key fails at connect time rather than on first real call.
That preserves the existing contract where `Connect()` surfaces auth failures. If you do
this, make the endpoint choice explicit and testable, and keep it to a single request.

---

## Step 5 — CLI utilities

`utilities/gofips/main.go`, `utilities/gofimac/main.go`, `utilities/gofinet/main.go`.

All three read `UNIFI_USERNAME` / `UNIFI_PASSWORD` / `UNIFI_CONTROLLER_IP` and hard-exit
when username or password is empty. Add:

- `UNIFI_API_KEY` — preferred when set.
- `UNIFI_CONSOLE_ID` — optional; when set, use connector mode and ignore host/port.
- Precedence: `UNIFI_API_KEY` wins over username/password. If both are present, use the
  key and write one line to stderr saying so.
- Error text when nothing is set must name `UNIFI_API_KEY` first, since that is now the
  recommended path.
- Update each `flag.Usage` block's "Environment Variables" section. `gofips` documents
  these in its usage text at `main.go` around the `envUsername`/`envPassword` constants —
  keep the existing `env*` const pattern.

Do not change existing flag names or add new flags for credentials. Keys belong in the
environment, not in argv where they land in shell history and `ps` output.

---

## Step 6 — Mock server and tests

`mock/server.go`, `mock/handlers_auth.go`.

`server.go` special-cases `/api/auth/login` at the top of its handler and checks
`X-CSRF-Token` for mutations further down. Add:

- Accept a configured API key via `X-API-KEY` and treat the request as authenticated
  without a prior login, bypassing the CSRF check for key-authenticated requests.
- A `mock.WithAPIKey(key)` option in `mock/options.go` following the existing option style.
- Reject a wrong key with the same shape the real console returns:
  `403` with `{"code":"forbidden","httpStatusCode":403,"message":"insufficient permissions"}`.
  That is the verified live response body.
- Optional connector-prefix support so connector mode is testable: strip a configured
  `/v1/connector/consoles/{id}` prefix before dispatch.

Tests to add:

- `client_test.go` — construction succeeds with `APIKey` and no username/password;
  fails with neither; fails with both.
- `client_test.go` — resolved default timeout is 30s (regression guard for Step 2).
- `auth/auth_test.go` — the `NewAPIKey` manager's five methods.
- `transport/transport_test.go` — `X-API-KEY` is sent when configured, absent when not,
  and a per-request header override wins.
- `transport/transport_test.go` — `PathPrefix` is prepended correctly, including the
  no-prefix case and a prefix with and without a trailing slash.
- One end-to-end test per auth mode through the mock, asserting a service call succeeds.

---

## Step 7 — Docs

- `README.md` — the quickstart at the top exports `UNIFI_USERNAME` / `UNIFI_PASSWORD`.
  Lead with `UNIFI_API_KEY` instead and keep username/password as a documented fallback.
  Add a short section on generating a key: console UI → profile icon (bottom left) →
  **API** → Create API Key, shown once. Note that a key inherits the permissions of the
  account that created it, so use Site Admin or Owner. Note that cloud keys from
  unifi.ui.com are a different credential from console-issued keys, and that a cloud key
  needs the **UniFi Applications → Network** scope for connector access.
- `CLAUDE.md` — the "Key Technical Details" section states auth is cookie-based via
  `POST /api/auth/login`. Update it to describe both mechanisms and note which endpoints
  each covers.
- `doc.go` — the package example uses `Username`/`Password`. Switch to `APIKey`.
- `EXAMPLES.md` and `examples/*/main.go` — update credential setup consistently.

---

## Out of scope, worth reporting

`go.mod` declares `module github.com/unifi-go/gofi` but the repo is
`github.com/emergingrobotics/gofi`. `go get` against the real URL will fail on the path
mismatch. Mention it to the user; do not change it as part of this task, since it touches
every import in the tree and is a separate commit.

`client.Events()` in `client_impl.go` returns `nil` with a "Phase 18" comment. A caller
doing `client.Events().Subscribe(...)` gets a nil-interface panic. Unrelated to auth, but
worth a one-line report.
