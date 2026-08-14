# gofi - Go UniFi Controller

A Go module for programmatic control of UniFi UDM Pro devices (v10+).

## Project Resources

- **API Reference**: `UNIFI_UDM_PRO_API_DOCUMENTATION.md` - Complete endpoint documentation
- **Architecture**: `docs/DESIGN.md` - Design with mermaid diagrams, service interfaces, type definitions
- **Implementation Plan**: `docs/plan.md` - Phased plan with progress tracking

## Critical Rules

1. **Every function MUST have a test** - No exceptions. Run `make test` to verify.
2. **Every endpoint MUST be supported in the mock server** - Tests use the mock, not real hardware.
3. **No phase advancement without 100% test coverage** - Complete and test each phase before moving on.
4. **Phases are sequential** - Follow `docs/plan.md` in order.

## Concurrency Requirements

- **CSRF tokens**: Use `atomic.Value` for thread-safe storage and updates
- **Session refresh**: Use `sync.Mutex` to prevent concurrent refresh races
- **Rate limiting**: Implement semaphore-based request limiting
- **Connection pooling**: Configure `http.Transport` appropriately

## Architecture Overview

The module path stays `github.com/unifi-go/gofi`; the library source lives under
`src/`, so the root package is imported as `github.com/unifi-go/gofi/src` (package
name is still `gofi`). `examples/` and `utilities/` remain at the repo top level.

```
gofi/
├── src/               # Library source (package gofi + sub-packages)
│   ├── client.go      # Main client, authentication, request handling
│   ├── types/         # All domain types (Device, Network, WLAN, etc.)
│   ├── errors.go      # Sentinel errors, APIError type
│   ├── services/      # Service implementations
│   │   ├── site.go
│   │   ├── device.go
│   │   ├── network.go
│   │   ├── wlan.go
│   │   ├── firewall.go
│   │   ├── client.go
│   │   ├── user.go
│   │   ├── routing.go
│   │   └── ...
│   ├── mock/          # Mock server for testing
│   │   ├── server.go
│   │   ├── handlers.go
│   │   ├── fixtures/
│   │   └── scenarios/
│   ├── auth/          # Authentication helpers
│   ├── transport/     # HTTP transport, connector prefixing
│   ├── websocket/     # Event stream client
│   └── internal/      # Internal helpers
├── examples/          # Runnable example programs (consumers)
└── utilities/
    ├── gofi/          # The gofi CLI: one binary, flag wiring only
    ├── internal/      # Per-area business logic behind the CLI
    │   ├── clients/   # Connected stations, IEEE OUI cache and lookup
    │   ├── config/    # config.toml, targets, XDG paths
    │   ├── conn/      # Connection/credential resolution from flags + env
    │   ├── dns/       # Local DNS records
    │   ├── ips/       # Fixed-IP reservations, ISC DHCP format
    │   ├── network/   # Networks, subnets, DHCP pools
    │   ├── profile/   # Site capture/apply
    │   └── users/     # Known-client records
    └── docs/          # Design notes carried over from the predecessor tools
```

## Key Technical Details

- **Auth**: Two mechanisms, selected by which `Config` fields are set:
  - **Cloud API key + Site Manager connector** (`APIKey` + `ConsoleID`, or
    `WithAPIKey`/`WithConnector`): sends `X-API-KEY` on every request; the transport talks to
    `https://api.ui.com` and prepends `/v1/connector/consoles/{ConsoleID}` to the request path.
    Reaches the classic v1 and v2 paths below through the connector (verified against a live
    console) — this is the path to use when the controller isn't directly reachable on the LAN.
  - **Local username/password** (`Username`/`Password`, unchanged): cookie-based session via
    `POST /api/auth/login` directly against the controller.
- **CSRF**: Local username/password mode only — extract from cookie, send as `X-CSRF-Token`
  header. Not applicable in API-key mode (no session cookie is established).
- **Base Path**: UDM Pro uses `/proxy/network` prefix (both auth modes); in connector mode the
  `/v1/connector/consoles/{ConsoleID}` prefix is prepended ahead of it.
- **API Versions**: v1 (`/api/s/{site}/...`) and v2 (`/v2/api/site/{site}/...`) — both reachable
  under either auth mode.
- **WebSocket**: Events at `wss://{host}/proxy/network/wss/s/{site}/events` (local
  username/password mode; not exercised through the connector).

## Type Patterns

Use flexible types for UniFi's inconsistent JSON:

```go
type FlexInt int64    // Handles "123" or 123
type FlexBool bool    // Handles "true", true, 0, 1
```

## Mock Server

The mock server must:
- Support all endpoints with realistic responses
- Allow fixture loading for consistent test data
- Support error scenarios (auth failures, rate limits, not found)
- Simulate WebSocket events

## gofi CLI

One binary, `gofi <area> <action>`, living in `utilities/gofi/` and built on the
module in `src/`. It replaces the five flag-mode tools that preceded it —
`gofips`, `gofimac`, `gofinet`, `gofidns`, `gofiuser` — which were removed in the
same release, with no wrapper binaries and no deprecation window. The areas are
`ips`, `dns`, `network`, `clients`, `users`, `profile`, and `config`; each area's
business logic lives in an `utilities/internal/<area>` package, and the files under
`utilities/gofi/` are flag wiring only.

Two documents own the CLI, and this one does not:

- **[`REQUIREMENTS.md`](REQUIREMENTS.md)** — the contract. Every flag, exit code,
  guard, and output rule, numbered and traceable (`C-<AREA>-NNN` constraints,
  `I-<AREA>-NNN` invariants, `B-<AREA>-NNN` behaviors). Work from this file when
  implementing or reviewing a CLI change; if code and it disagree, the code is wrong.
- **[`VISION.md`](VISION.md)** — the why. The area boundaries, the naming decisions,
  and the record of the design brainstorm behind them.

A few module-level facts the CLI depends on, kept here because they are properties of
this repository rather than of the command tree:

- Exit codes are fixed across every command: `0` success, `1` error, `2` usage error,
  `3` refused by a guard (C-GLOBAL-012).
- The IEEE OUI registry is cached at `$XDG_DATA_HOME/gofi/oui.txt`, falling back to
  `~/.local/share/gofi/oui.txt`, and re-downloaded when older than 30 days
  (`utilities/internal/clients/oui.go`). Manufacturer lookups always come from this
  cache, never from the controller's own (frequently stale) `oui` field.
- Secrets never appear in `config.toml` or on any flag: they come from the
  environment, from a `*_command` the config names, or from an interactive prompt.

## Controller API Quirks

Behaviors confirmed against a UDM Pro that differ from what the endpoint shape
suggests. The mock server reproduces each one so code cannot pass against the mock
and fail against hardware.

| Endpoint | Quirk |
|----------|-------|
| `POST cmd/stamgr` `forget-sta` | Batch-only. Requires `"macs": [...]`; the singular `"mac"` field is rejected with 400. Other stamgr commands take `"mac"`. See `batchMACCommands` in `src/services/clientstation.go`. |
| `DELETE /rest/user/<id>` | Answers 404. `UserService.Delete` cannot remove a client; use `ClientService.Forget`. |
| `GET /v2/api/site/<site>/static-dns/<id>` | Answers 405. Only the collection is readable, so `DNSService.Get` resolves through `List`. |
| `/v2/api/site/...` responses | Bare JSON, not the v1 `meta`/`data` envelope. |

## Commands

```bash
make test      # Run all tests
make lint      # Run linter
make build     # Build the module
make coverage  # Generate coverage report
```

## Reference Implementations

Study these for patterns:
- `github.com/paultyng/go-unifi` - Terraform provider, CRUD patterns
- `github.com/unpoller/unifi` - FlexInt/FlexBool types
- `github.com/thib3113/unifi-client` - TypeScript, comprehensive types
