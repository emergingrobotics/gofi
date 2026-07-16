<img src="images/gofi-logo.png" width="33%" alt="gofi logo">

# gofi - Go UniFi Controller Client

[![Go Reference](https://pkg.go.dev/badge/github.com/unifi-go/gofi.svg)](https://pkg.go.dev/github.com/unifi-go/gofi)
[![Go Report Card](https://goreportcard.com/badge/github.com/unifi-go/gofi)](https://goreportcard.com/report/github.com/unifi-go/gofi)

A Go module for programmatic control of Ubiquiti UniFi UDM Pro devices, plus command-line utilities built on top of it.

## Utilities

Standalone tools built with the gofi module. Build all utilities with `make utilities` or install to `/usr/local/bin` with `sudo make install`.

All utilities authenticate via environment variables:

```bash
export UNIFI_USERNAME=admin
export UNIFI_PASSWORD=your-password
```

The `UNIFI_UDM_IP` variable is optional and provides a fallback host address when `-H` is not given.

### gofips

Manages fixed IP (DHCP reservation) assignments **with hostnames and DNS records** on a UDM Pro, using the industry-standard ISC DHCP `dhcpd.conf` host-declaration format. This is the recommended tool for bulk fixed-IP management — it supersedes `gofip` and the `fixedips` / `addfixedip` / `delfixedip` examples by adding hostname and DNS support.

**Export current assignments** (dump to disk, edit, push back):

```bash
gofips -H 192.168.1.1 -k --get > hosts.conf
```

Assignments are emitted sorted by IP address in ISC DHCP format, where the `host <name>` label is the DNS name:

```
# gofips fixed IP assignments
# exported from UDM at 192.168.1.1

host myserver {
    hardware ethernet aa:bb:cc:dd:ee:01;
    fixed-address 192.168.1.10;
}

host printer {
    hardware ethernet aa:bb:cc:dd:ee:02;
    fixed-address 192.168.1.11;
}
```

**Import (bulk provision) from a file or stdin:**

```bash
gofips -H 192.168.1.1 -k --set hosts.conf
cat hosts.conf | gofips -H 192.168.1.1 -k --set
```

The input is fully validated before any changes are made; unchanged entries are skipped and the network for each IP is auto-detected from configured subnets. Add `--dry-run` to preview changes without applying them.

**Add or delete a single host:**

```bash
gofips -H 192.168.1.1 -k --add 'host mydev { hardware ethernet aa:bb:cc:dd:ee:ff; fixed-address 192.168.1.50; }'
gofips -H 192.168.1.1 -k --del --name mydev   # or --mac / --ip
```

| Flag | Short | Description |
|------|-------|-------------|
| `--get` | `-g` | Export assignments to stdout in ISC DHCP format |
| `--set` | `-s` | Import host declarations from a file or stdin |
| `--add` | `-a` | Add a single host from an ISC DHCP declaration |
| `--del` | `-d` | Delete a host by `--name`, `--mac`, or `--ip` |
| `--force` | `-f` | Skip conflict checks; force delete of the user record |
| `--keep-dns` | `-K` | Preserve DNS records when deleting |
| `--dry-run` | | Preview changes without applying |
| `--host` | `-H` | UDM Pro host address (or set `UNIFI_UDM_IP`) |
| `--port` | `-p` | Port (default: 443) |
| `--site` | `-S` | Site name (default: "default") |
| `--insecure` | `-k` | Skip TLS certificate verification |

See [utilities/gofips/README.md](./utilities/gofips/README.md) and [utilities/docs/gofips/DESIGN.md](./utilities/docs/gofips/DESIGN.md) for details.

### gofimac

Lists connected clients (wired, WiFi, or all) with manufacturer identification looked up independently from the IEEE OUI database rather than relying on the UDM's built-in fingerprinting.

```bash
gofimac -H 192.168.1.1 -k          # all connected clients (default)
gofimac -H 192.168.1.1 -k --wifi   # WiFi clients only
gofimac -H 192.168.1.1 -k --wired  # wired clients only
gofimac -H 192.168.1.1 -k --json   # JSON output
```

Text output is tab-separated (MAC, IP, hostname, manufacturer), sorted by IP:

```
aa:bb:cc:dd:ee:01	192.168.1.10	myserver	Dell Inc.
aa:bb:cc:dd:ee:02	192.168.1.11	printer	Hewlett Packard
```

The IEEE OUI database is downloaded and cached under `$XDG_DATA_HOME/gofimac/` (default `~/.local/share/gofimac/`) and refreshed automatically when older than 30 days. If a refresh fails, a cached copy is used with a warning.

| Flag | Short | Description |
|------|-------|-------------|
| `--wifi` | `-w` | List only WiFi clients |
| `--wired` | `-e` | List only wired clients |
| `--all` | `-a` | List all clients (default) |
| `--json` | `-j` | Output JSON instead of text |
| `--host` | `-H` | UDM Pro host address (or set `UNIFI_UDM_IP`) |
| `--port` | `-p` | Port (default: 443) |
| `--site` | `-S` | Site name (default: "default") |
| `--insecure` | `-k` | Skip TLS certificate verification |

See [utilities/gofimac/README.md](./utilities/gofimac/README.md) and [utilities/docs/gofimac/DESIGN.md](./utilities/docs/gofimac/DESIGN.md) for details.

### gofip (legacy)

The original fixed-IP tool. Stores assignments as a plain text file — one `IP MAC` pair per line, with no hostname or DNS support. Kept as a reference implementation; **use `gofips` for new work.**

```bash
gofip -H 192.168.1.1 -k --get > hosts.txt   # export "IP MAC" pairs
gofip -H 192.168.1.1 -k --set hosts.txt      # import
```

See [utilities/docs/gofip/DESIGN.md](./utilities/docs/gofip/DESIGN.md) for the full design.

---

## Module

The gofi Go module provides type-safe, concurrent-safe access to all major UniFi Network Application endpoints.

### Features

- **Complete API Coverage**: All major UniFi Network Application endpoints (v1, v2, REST, WebSocket)
- **Type-Safe**: Full type definitions for all UniFi resources
- **Concurrent-Safe**: Thread-safe operations with proper synchronization
- **Production-Ready**: Comprehensive error handling, retry logic, and connection pooling
- **Well-Tested**: 500+ tests with race detection and high coverage
- **WebSocket Support**: Real-time event streaming
- **Batch Operations**: Concurrent operations for improved performance
- **Mock Server**: Full mock implementation for testing without hardware

### Installation

```bash
go get github.com/unifi-go/gofi
```

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/unifi-go/gofi"
)

func main() {
    config := &gofi.Config{
        Host:          "192.168.1.1",
        Username:      "admin",
        Password:      "your-password",
        SkipTLSVerify: true,
    }

    client, err := gofi.New(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    devices, err := client.Devices().List(ctx, "default")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d devices\n", len(devices))
    for _, device := range devices {
        fmt.Printf("- %s (%s)\n", device.Name, device.Model)
    }
}
```

### Supported Services

#### Core Services
- **Sites**: Site management and health monitoring
- **Devices**: Access points, switches, gateways control
- **Networks**: VLAN and network configuration
- **WLANs**: Wireless network management

#### Security & Access
- **Firewall**: Firewall rules and groups (v1 and v2 APIs)
- **Traffic Rules**: QoS and traffic shaping
- **Clients**: Connected client management and guest authorization
- **Users**: Known client management with fixed IPs

#### Advanced Features
- **Routing**: Static route management
- **Port Forwarding**: NAT port forwarding rules
- **Port Profiles**: Switch port configuration profiles
- **Settings**: System settings (RADIUS, DNS, NTP, SNMP, etc.)
- **System**: Backups, speed tests, admin management

#### Real-Time
- **Events**: WebSocket event streaming for real-time updates

### Examples

See the [examples](./examples/) directory for comprehensive usage examples. Build all examples with `make examples`.

All examples require the same environment variables as the utilities above.

| Example | Description |
|---------|-------------|
| `basic` | Connecting, listing sites, devices, networks, health status |
| `list` | List networks in table or JSON format |
| `crud` | Create, Read, Update, Delete operations for networks and WLANs |
| `concurrent` | Batch/concurrent operations with `gofi.BatchGet` |
| `websocket` | Real-time WebSocket event streaming |
| `errors` | Error handling patterns |
| `fixedips` | List all fixed IP assignments (reference; superseded by the `gofips` utility) |
| `addfixedip` | Assign a fixed IP to a device by MAC address (reference; superseded by `gofips`) |
| `delfixedip` | Remove a fixed IP assignment (reference; superseded by `gofips`) |
| `switches` | Switch and PoE management |

### API Coverage

#### Device Management
```go
devices, err := client.Devices().List(ctx, "default")
err = client.Devices().Adopt(ctx, "default", "aa:bb:cc:dd:ee:ff")
err = client.Devices().Restart(ctx, "default", "aa:bb:cc:dd:ee:ff")
err = client.Devices().Upgrade(ctx, "default", "aa:bb:cc:dd:ee:ff")
err = client.Devices().Locate(ctx, "default", "aa:bb:cc:dd:ee:ff")
```

#### Network Management
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

#### Wireless Networks
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
err = client.WLANs().Disable(ctx, "default", wlan.ID)
err = client.WLANs().Enable(ctx, "default", wlan.ID)
macs := []string{"aa:bb:cc:dd:ee:ff"}
err = client.WLANs().SetMACFilter(ctx, "default", wlan.ID, "allow", macs)
```

#### Client Management
```go
clients, err := client.Clients().ListActive(ctx, "default")
err = client.Clients().Block(ctx, "default", "aa:bb:cc:dd:ee:ff")
err = client.Clients().AuthorizeGuest(ctx, "default", "aa:bb:cc:dd:ee:ff",
    WithDuration(240),
    WithUploadLimit(5000),
    WithDownloadLimit(10000),
)
err = client.Clients().Kick(ctx, "default", "aa:bb:cc:dd:ee:ff")
```

#### Firewall Rules
```go
rules, err := client.Firewall().ListRules(ctx, "default")
rule := &types.FirewallRule{
    Name:        "Block IoT to LAN",
    Enabled:     true,
    Action:      "drop",
    Ruleset:     "LAN_IN",
    SrcNetworkID: iotNetworkID,
    DstNetworkID: lanNetworkID,
}
created, err := client.Firewall().CreateRule(ctx, "default", rule)
trafficRules, err := client.Firewall().ListTrafficRules(ctx, "default")
```

#### Real-Time Events
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

#### Batch Operations
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

### Configuration

#### Basic Configuration

```go
config := &gofi.Config{
    Host:     "192.168.1.1",
    Port:     443,
    Username: "admin",
    Password: "password",
    Site:     "default",
}
```

#### TLS Configuration

For production with valid certificates:

```go
config := &gofi.Config{
    Host:      "unifi.example.com",
    Username:  "admin",
    Password:  os.Getenv("UNIFI_PASSWORD"),
    TLSConfig: &tls.Config{
        // Your TLS configuration
    },
}
```

For self-signed certificates (development/testing):

```go
config := &gofi.Config{
    Host:          "192.168.1.1",
    Username:      "admin",
    Password:      "password",
    SkipTLSVerify: true,
}
```

#### Advanced Options

```go
client, err := gofi.New(config,
    gofi.WithTimeout(30*time.Second),
    gofi.WithRetry(3, 100*time.Millisecond),
    gofi.WithSite("custom-site"),
    gofi.WithLogger(customLogger),
)
```

#### Retry Configuration

```go
config := &gofi.Config{
    RetryConfig: &gofi.RetryConfig{
        MaxRetries:     3,
        InitialBackoff: 100 * time.Millisecond,
        MaxBackoff:     5 * time.Second,
    },
}
```

### Error Handling

```go
if err := client.Connect(ctx); err != nil {
    if errors.Is(err, gofi.ErrAuthenticationFailed) {
        // Handle auth failure
    }
    if errors.Is(err, gofi.ErrNotFound) {
        // Handle not found
    }

    var apiErr *gofi.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error [%d]: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

Available sentinel errors: `ErrNotConnected`, `ErrAlreadyConnected`, `ErrAuthenticationFailed`, `ErrSessionExpired`, `ErrNotFound`, `ErrPermissionDenied`, `ErrRateLimited`, `ErrServerError`.

### Testing

The library includes a comprehensive mock server:

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

### Architecture

```
gofi/
├── client.go          # Main client interface
├── types/             # Type definitions for all resources
├── services/          # Service implementations (12 services)
├── auth/              # Authentication and session management
├── transport/         # HTTP transport with retry logic
├── websocket/         # WebSocket client for events
├── mock/              # Mock server for testing
├── internal/          # Internal utilities
├── examples/          # Usage examples
└── utilities/         # Command-line tools
```

---

## Development

```bash
make test          # Run all tests
make coverage      # Generate coverage report
make lint          # Run linter
make build         # Build the module
make examples      # Build all examples to bin/examples/
make utilities     # Build all utilities (gofip, gofips, gofimac) to bin/
sudo make install  # Install utilities to /usr/local/bin
make all           # Run lint, test, and build
```

## Requirements

- Go 1.22 or later
- UniFi UDM Pro with Network Application 10.x+
- Admin access to the controller

## Compatibility

Tested with:
- UniFi OS 4.x and 5.x
- Network Application 10.x
- UDM Pro, UDM SE, and UDR devices

## Documentation

- [Design](./docs/DESIGN.md) - Architecture details
- [Examples](./examples/) - Usage examples
- [GoDoc](https://pkg.go.dev/github.com/unifi-go/gofi) - API reference

## Contributing

Contributions are welcome! Please ensure:
- All tests pass (`make test`)
- Code passes linting (`make lint`)
- New features include tests
- Changes maintain backward compatibility

## License

This project is licensed under the MIT License.

## Acknowledgments

- Inspired by [paultyng/go-unifi](https://github.com/paultyng/go-unifi) (Terraform provider patterns)
- Type patterns from [unpoller/unifi](https://github.com/unpoller/unifi) (FlexInt/FlexBool)
- API patterns from [thib3113/unifi-client](https://github.com/thib3113/unifi-client) (TypeScript)
