# gofi Go SDK Guide

The `gofi` module gives you type-safe, concurrent-safe access to UniFi Network
Application endpoints (v1, v2, REST, and WebSocket) — the same library the `gofi` CLI is
built on. This is the detailed reference; see [the README](../README.md#part-2-using-gofi-in-your-go-program-sdk)
for a minimal starting example.

## Features

- **Complete API coverage** — all major endpoints across the v1, v2, REST, and WebSocket surfaces
- **Type-safe** — full type definitions for every UniFi resource
- **Concurrent-safe** — thread-safe operations with proper synchronization
- **Production-ready** — error handling, retry with backoff, and connection pooling
- **Well-tested** — 500+ tests with race detection and high coverage
- **Mock server** — a full mock implementation for testing without hardware

## Install

```bash
go get github.com/unifi-go/gofi
```

The importable package lives under `src/` and is named `gofi`:

```go
import "github.com/unifi-go/gofi/src"   // used as gofi.New(...)
```

## Authentication and configuration

**API key + connector (recommended).** `APIKey` requires `ConsoleID` (or a `BaseURL`
override, used for tests) — a key with no console ID is a validation error from
`gofi.New()`. When `ConsoleID` is set, requests go to `https://api.ui.com` and `Host` /
`Port` are unused.

```go
config := &gofi.Config{
    APIKey:    os.Getenv("UNIFI_API_KEY"),
    ConsoleID: os.Getenv("UNIFI_CONSOLE_ID"),
}
```

See [the README's Step 1](../README.md#step-1--get-a-unifi-api-key-recommended) for how
to create a cloud API key and find your console ID.

**Local username/password (secondary, less-tested).** Cookie/CSRF session against a
directly-reachable controller, including TLS/self-signed-certificate handling — see
[`docs/alternate-local-api.md`](alternate-local-api.md) for the full setup and code
examples. Use this only when the controller has no route to `api.ui.com`.

**Advanced options and retry.**

```go
client, err := gofi.New(config,
    gofi.WithTimeout(30*time.Second),
    gofi.WithRetry(3, 100*time.Millisecond),
    gofi.WithSite("custom-site"),
    gofi.WithLogger(customLogger),
)

// or via Config:
config.RetryConfig = &gofi.RetryConfig{
    MaxRetries:     3,
    InitialBackoff: 100 * time.Millisecond,
    MaxBackoff:     5 * time.Second,
}
```

## Supported services

| Area | Services |
|------|----------|
| Core | Sites, Devices, Networks, WLANs |
| Security & access | Firewall (v1/v2), Traffic Rules, Clients, Users |
| Advanced | Routing, Port Forwarding, Port Profiles, Settings, System |
| Real-time | Events (WebSocket streaming) |

## Common operations

**Devices**

```go
devices, err := client.Devices().List(ctx, "default")
err = client.Devices().Adopt(ctx, "default", "aa:bb:cc:dd:ee:ff")
err = client.Devices().Restart(ctx, "default", "aa:bb:cc:dd:ee:ff")
err = client.Devices().Upgrade(ctx, "default", "aa:bb:cc:dd:ee:ff")
```

**Networks**

```go
network := &types.Network{
    Name:         "IoT Network",
    VLANEnabled:  true,
    VLAN:         20,
    IPSubnet:     "192.168.20.1/24",
    DHCPDEnabled: true,
}
created, err := client.Networks().Create(ctx, "default", network)
updated, err := client.Networks().Update(ctx, "default", network)
err = client.Networks().Delete(ctx, "default", network.ID)
```

**Wireless networks**

```go
wlan := &types.WLAN{
    Name:       "Guest WiFi",
    Enabled:    true,
    Security:   "wpapsk",
    WPAMode:    "wpa2",
    Passphrase: "guestpassword",
    IsGuest:    true,
}
created, err := client.WLANs().Create(ctx, "default", wlan)
err = client.WLANs().Enable(ctx, "default", wlan.ID)
macs := []string{"aa:bb:cc:dd:ee:ff"}
err = client.WLANs().SetMACFilter(ctx, "default", wlan.ID, "allow", macs)
```

**Clients**

```go
clients, err := client.Clients().ListActive(ctx, "default")
err = client.Clients().Block(ctx, "default", "aa:bb:cc:dd:ee:ff")
err = client.Clients().AuthorizeGuest(ctx, "default", "aa:bb:cc:dd:ee:ff",
    WithDuration(240),
    WithUploadLimit(5000),
    WithDownloadLimit(10000),
)
```

**Firewall**

```go
rules, err := client.Firewall().ListRules(ctx, "default")
rule := &types.FirewallRule{
    Name:         "Block IoT to LAN",
    Enabled:      true,
    Action:       "drop",
    Ruleset:      "LAN_IN",
    SrcNetworkID: iotNetworkID,
    DstNetworkID: lanNetworkID,
}
created, err := client.Firewall().CreateRule(ctx, "default", rule)
```

**Real-time events**

```go
eventCh, errorCh, err := client.Events().Subscribe(ctx, "default")
if err != nil {
    log.Fatal(err)
}
defer client.Events().Close()

for {
    select {
    case event := <-eventCh:
        fmt.Printf("Event: %s - %s\n", event.Key, event.Message)
    case err := <-errorCh:
        fmt.Printf("Error: %v\n", err)
    }
}
```

**Batch operations**

```go
deviceIDs := []string{"id1", "id2", "id3"}
results := gofi.BatchGet(ctx, deviceIDs, func(ctx context.Context, id string) (*types.Device, error) {
    return client.Devices().Get(ctx, "default", id)
})

for _, result := range results {
    if result.Error != nil {
        fmt.Printf("Error at index %d: %v\n", result.Index, result.Error)
    } else {
        fmt.Printf("Device: %s\n", result.Item.Name)
    }
}
```

## Error handling

```go
if err := client.Connect(ctx); err != nil {
    if errors.Is(err, gofi.ErrAuthenticationFailed) {
        // handle auth failure
    }

    var apiErr *gofi.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error [%d]: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

Sentinel errors: `ErrNotConnected`, `ErrAlreadyConnected`, `ErrAuthenticationFailed`,
`ErrSessionExpired`, `ErrNotFound`, `ErrPermissionDenied`, `ErrRateLimited`,
`ErrServerError`.

## Testing with the mock server

The library ships a full mock server so you can test without hardware.

```go
func TestYourCode(t *testing.T) {
    server := mock.NewServer()
    defer server.Close()

    server.State().AddDevice(&types.Device{
        ID:   "test-device",
        Name: "Test AP",
    })

    config := &gofi.Config{
        Host:          server.Host(),
        Port:          server.Port(),
        Username:      "admin",
        Password:      "admin",
        SkipTLSVerify: true,
    }

    client, _ := gofi.New(config)
    client.Connect(context.Background())

    devices, err := client.Devices().List(context.Background(), "default")
    // ...
}
```

## Architecture

```
gofi/
├── src/               # Library source (package gofi + sub-packages)
│   ├── client.go      # Main client interface
│   ├── types/         # Type definitions for all resources
│   ├── services/      # Service implementations
│   ├── auth/          # Authentication and session management
│   ├── transport/     # HTTP transport with retry logic
│   ├── websocket/     # WebSocket client for events
│   ├── mock/          # Mock server for testing
│   └── internal/      # Internal utilities
├── examples/          # Example programs
└── utilities/         # CLI tool and shared utilities
    ├── gofi/          # Command-line interface
    └── internal/      # Shared internal packages
```

## Example programs

The [`examples/`](../examples/) directory holds small, focused programs that demonstrate
the SDK — plain API usage, not command-line tools. Build them with `make examples`
(binaries land in `bin/examples/`).

Auth support varies by example: `concurrent`, `crud`, `errors`, and `websocket` accept
the API-key variables (see [Authentication and configuration](#authentication-and-configuration)
above), while the others currently use `UNIFI_USERNAME`/`UNIFI_PASSWORD`. Check the top
of each `main.go`, or [`examples/README.md`](../examples/README.md), for specifics.

| Example | Description |
|---------|-------------|
| `basic` | Connecting, listing sites, devices, networks, health status |
| `list` | List networks in table or JSON format |
| `crud` | Create, read, update, delete for networks and WLANs |
| `concurrent` | Batch/concurrent operations with `gofi.BatchGet` |
| `websocket` | Real-time WebSocket event streaming |
| `errors` | Error handling patterns |
| `fixedips` | List all fixed IP assignments — the `Users()`/fixed-IP API the `gofi ips` CLI area is built on |
| `addfixedip` | Assign a fixed IP by MAC address |
| `delfixedip` | Remove a fixed IP assignment |
| `switches` | Switch and PoE management |

## See also

- [`README.md`](../README.md) — install, quick start, the `gofi` CLI
- [`docs/alternate-local-api.md`](alternate-local-api.md) — local username/password auth
- [`docs/gofi-user-guide.md`](gofi-user-guide.md) — the `gofi` CLI's full command reference
- [`docs/DESIGN.md`](DESIGN.md) — architecture details
- [GoDoc](https://pkg.go.dev/github.com/unifi-go/gofi/src) — API reference
