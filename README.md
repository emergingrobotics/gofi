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

### gofip

Manages fixed IP (DHCP reservation) assignments on a UDM Pro. Replaces editing `dhcpd.conf` and DNS zone files for small networks. Assignments are stored as a simple text file — one `IP MAC` pair per line — that can be version-controlled, diffed, and shared.

**Export current assignments:**

```bash
gofip -H 192.168.1.1 -k --get > hosts.txt
```

If no assignments exist, the output contains commented examples showing the file format. If assignments exist, they are printed sorted by IP address:

```
# gofip fixed IP assignments
# format: IP MAC
192.168.1.10 aa:bb:cc:dd:ee:01
192.168.1.11 aa:bb:cc:dd:ee:02
192.168.1.20 11:22:33:44:55:66
```

**Import assignments from a file:**

```bash
gofip -H 192.168.1.1 -k --set hosts.txt
```

**Import from stdin:**

```bash
echo "192.168.1.50 aa:bb:cc:dd:ee:ff" | gofip -H 192.168.1.1 -k --set
```

Existing assignments (same MAC with the same IP) are skipped. The input file is fully validated before any changes are made to the controller. The network for each IP is auto-detected from configured subnets.

| Flag | Short | Description |
|------|-------|-------------|
| `--get` | `-g` | Export assignments to stdout |
| `--set` | `-s` | Import assignments from file or stdin |
| `--host` | `-H` | UDM Pro host address (or set `UNIFI_UDM_IP`) |
| `--port` | `-p` | Port (default: 443) |
| `--site` | `-S` | Site name (default: "default") |
| `--insecure` | `-k` | Skip TLS certificate verification |

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
| `fixedips` | List all fixed IP assignments |
| `addfixedip` | Assign a fixed IP to a device by MAC address |
| `delfixedip` | Remove a fixed IP assignment |
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
make utilities     # Build all utilities to bin/utilities/
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
