# gofi - Go UniFi Controller Client

[![Go Reference](https://pkg.go.dev/badge/github.com/unifi-go/gofi.svg)](https://pkg.go.dev/github.com/unifi-go/gofi)
[![Go Report Card](https://goreportcard.com/badge/github.com/unifi-go/gofi)](https://goreportcard.com/report/github.com/unifi-go/gofi)

A comprehensive Go client library for programmatic control of Ubiquiti UniFi UDM Pro devices running UniFi OS 4.x/5.x with Network Application 10.x+.

## Features

- **Complete API Coverage**: All major UniFi Network Application endpoints (v1, v2, REST, WebSocket)
- **Type-Safe**: Full type definitions for all UniFi resources
- **Concurrent-Safe**: Thread-safe operations with proper synchronization
- **Production-Ready**: Comprehensive error handling, retry logic, and connection pooling
- **Well-Tested**: 500+ tests with race detection and high coverage
- **WebSocket Support**: Real-time event streaming
- **Batch Operations**: Concurrent operations for improved performance
- **Mock Server**: Full mock implementation for testing without hardware

## Installation

```bash
go get github.com/unifi-go/gofi
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/unifi-go/gofi"
)

func main() {
    // Create client
    config := &gofi.Config{
        Host:          "192.168.1.1",
        Username:      "admin",
        Password:      "your-password",
        SkipTLSVerify: true, // Only for self-signed certs
    }

    client, err := gofi.New(config)
    if err != nil {
        log.Fatal(err)
    }

    // Connect
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    // List devices
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

## Supported Services

The client provides access to all major UniFi services:

### Core Services
- **Sites**: Site management and health monitoring
- **Devices**: Access points, switches, gateways control
- **Networks**: VLAN and network configuration
- **WLANs**: Wireless network management

### Security & Access
- **Firewall**: Firewall rules and groups (v1 and v2 APIs)
- **Traffic Rules**: QoS and traffic shaping
- **Clients**: Connected client management and guest authorization
- **Users**: Known client management with fixed IPs

### Advanced Features
- **Routing**: Static route management
- **Port Forwarding**: NAT port forwarding rules
- **Port Profiles**: Switch port configuration profiles
- **Settings**: System settings (RADIUS, DNS, NTP, SNMP, etc.)
- **System**: Backups, speed tests, admin management

### Real-Time
- **Events**: WebSocket event streaming for real-time updates

## Examples

See the [examples](./examples/) directory for comprehensive usage examples. Build all examples with `make build`.

All examples require environment variables for authentication:
```bash
export UNIFI_USERNAME=your_username
export UNIFI_PASSWORD=your_password
```

### list

Lists all networks from the controller in table or JSON format.

```bash
./bin/examples/list -H 192.168.1.1 -k           # Table output
./bin/examples/list -H 192.168.1.1 -k -j        # JSON output
./bin/examples/list -H 192.168.1.1 -k -s mysite # Specific site
```

| Flag | Description |
|------|-------------|
| `-H, --host` | UDM Pro host address (required) |
| `-p, --port` | Port (default: 443) |
| `-s, --site` | Site name (default: "default") |
| `-k, --insecure` | Skip TLS certificate verification |
| `-j, --json` | Output in JSON format |
| `-d, --debug` | Enable debug output |

### basic

Demonstrates basic client usage: connecting, listing sites, devices, networks, and health status.

```bash
./bin/examples/basic -H 192.168.1.1 -k
./bin/examples/basic -H 192.168.1.1 -k -d       # With debug output
```

| Flag | Description |
|------|-------------|
| `-H, --host` | UDM Pro host address (required) |
| `-p, --port` | Port (default: 443) |
| `-s, --site` | Site name (default: "default") |
| `-k, --insecure` | Skip TLS certificate verification |
| `-d, --debug` | Enable debug output |
| `-t, --timeout` | Connection timeout (default: 30s) |

### crud

Demonstrates Create, Read, Update, Delete operations for networks and WLANs. Creates a test IoT network and Guest WiFi WLAN, updates them, then cleans up.

```bash
./bin/examples/crud -H 192.168.1.1 -k
```

**Note:** This example modifies your controller configuration. Use with caution.

### concurrent

Demonstrates batch/concurrent operations using `gofi.BatchGet` to fetch multiple devices in parallel.

```bash
./bin/examples/concurrent -H 192.168.1.1 -k
```

### websocket

Subscribes to real-time WebSocket events from the controller. Displays client connect/disconnect events, AP events, and more. Press Ctrl+C to exit.

```bash
./bin/examples/websocket -H 192.168.1.1 -k
```

### errors

Demonstrates error handling patterns including:
- Connection errors (authentication, timeout)
- Resource not found errors
- API error details extraction
- Validation errors
- Automatic retry configuration

```bash
./bin/examples/errors -H 192.168.1.1 -k
```

### fixedips

Lists all clients that have fixed IP addresses assigned. Useful for auditing DHCP reservations.

```bash
./bin/examples/fixedips -H 192.168.1.1 -k           # Table output
./bin/examples/fixedips -H 192.168.1.1 -k -j        # JSON output
./bin/examples/fixedips -H 192.168.1.1 -k -j | jq   # Pipe to jq
```

| Flag | Description |
|------|-------------|
| `-H, --host` | UDM Pro host address (required) |
| `-p, --port` | Port (default: 443) |
| `-s, --site` | Site name (default: "default") |
| `-k, --insecure` | Skip TLS certificate verification |
| `-j, --json` | Output in JSON format |

### addfixedip

Assigns a fixed IP address to a device by MAC address. Checks for conflicts before assignment.

```bash
# Basic usage - auto-detect network from IP
./bin/examples/addfixedip -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff -i 192.168.1.100 -n "My Device"

# Specify network explicitly
./bin/examples/addfixedip -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff -i 192.168.1.100 -n "My Device" -N "LAN"

# Skip conflict checks
./bin/examples/addfixedip -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff -i 192.168.1.100 -n "My Device" -f
```

| Flag | Description |
|------|-------------|
| `-H, --host` | UDM Pro host address (required) |
| `-p, --port` | Port (default: 443) |
| `-s, --site` | Site name (default: "default") |
| `-k, --insecure` | Skip TLS certificate verification |
| `-m, --mac` | MAC address of device (required) |
| `-i, --ip` | Fixed IP address to assign (required) |
| `-n, --name` | Hostname/friendly name (required) |
| `-N, --network` | Network ID or name (auto-detects if not specified) |
| `-f, --force` | Skip conflict checks |

**Conflict checks:**
- Warns if IP is currently in use by an active client
- Warns if IP is already reserved for a different MAC
- Updates existing reservation if MAC already has a fixed IP

### delfixedip

Removes a fixed IP assignment from a device, allowing it to use DHCP for a dynamic address.

```bash
# Remove by MAC address
./bin/examples/delfixedip -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff

# Remove by IP address
./bin/examples/delfixedip -H 192.168.1.1 -k -i 192.168.1.100

# Delete the user entry entirely (not just the fixed IP)
./bin/examples/delfixedip -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff -D
```

| Flag | Description |
|------|-------------|
| `-H, --host` | UDM Pro host address (required) |
| `-p, --port` | Port (default: 443) |
| `-s, --site` | Site name (default: "default") |
| `-k, --insecure` | Skip TLS certificate verification |
| `-m, --mac` | MAC address of device |
| `-i, --ip` | Fixed IP address to look up |
| `-D, --delete` | Delete the user entry entirely (not just the fixed IP) |

**Note:** Either `--mac` or `--ip` must be specified to identify the device.

## API Coverage

### Device Management
```go
// List all devices
devices, err := client.Devices().List(ctx, "default")

// Adopt a device
err = client.Devices().Adopt(ctx, "default", "aa:bb:cc:dd:ee:ff")

// Restart a device
err = client.Devices().Restart(ctx, "default", "aa:bb:cc:dd:ee:ff")

// Upgrade firmware
err = client.Devices().Upgrade(ctx, "default", "aa:bb:cc:dd:ee:ff")

// Locate (flash LED)
err = client.Devices().Locate(ctx, "default", "aa:bb:cc:dd:ee:ff")
```

### Network Management
```go
// Create a network
network := &types.Network{
    Name:         "IoT Network",
    VLANEnabled:  true,
    VLAN:         20,
    IPSubnet:     "192.168.20.1/24",
    DHCPDEnabled: true,
}
created, err := client.Networks().Create(ctx, "default", network)

// Update a network
network.Name = "Updated Name"
updated, err := client.Networks().Update(ctx, "default", network)

// Delete a network
err = client.Networks().Delete(ctx, "default", network.ID)
```

### Wireless Networks
```go
// Create a WLAN
wlan := &types.WLAN{
    Name:       "Guest WiFi",
    Enabled:    true,
    Security:   "wpapsk",
    WPAMode:    "wpa2",
    Passphrase: "guestpassword",
    IsGuest:    true,
}
created, err := client.WLANs().Create(ctx, "default", wlan)

// Enable/Disable
err = client.WLANs().Disable(ctx, "default", wlan.ID)
err = client.WLANs().Enable(ctx, "default", wlan.ID)

// MAC filtering
macs := []string{"aa:bb:cc:dd:ee:ff"}
err = client.WLANs().SetMACFilter(ctx, "default", wlan.ID, "allow", macs)
```

### Client Management
```go
// List active clients
clients, err := client.Clients().ListActive(ctx, "default")

// Block a client
err = client.Clients().Block(ctx, "default", "aa:bb:cc:dd:ee:ff")

// Authorize a guest
err = client.Clients().AuthorizeGuest(ctx, "default", "aa:bb:cc:dd:ee:ff",
    WithDuration(240),      // 4 hours
    WithUploadLimit(5000),  // 5 Mbps
    WithDownloadLimit(10000), // 10 Mbps
)

// Kick (disconnect) a client
err = client.Clients().Kick(ctx, "default", "aa:bb:cc:dd:ee:ff")
```

### Firewall Rules
```go
// List firewall rules
rules, err := client.Firewall().ListRules(ctx, "default")

// Create a rule
rule := &types.FirewallRule{
    Name:        "Block IoT to LAN",
    Enabled:     true,
    Action:      "drop",
    Ruleset:     "LAN_IN",
    SrcNetworkID: iotNetworkID,
    DstNetworkID: lanNetworkID,
}
created, err := client.Firewall().CreateRule(ctx, "default", rule)

// Traffic rules (v2 API)
trafficRules, err := client.Firewall().ListTrafficRules(ctx, "default")
```

### Real-Time Events
```go
// Subscribe to events
eventCh, errorCh, err := client.Events().Subscribe(ctx, "default")
if err != nil {
    log.Fatal(err)
}
defer client.Events().Close()

// Process events
for {
    select {
    case event := <-eventCh:
        fmt.Printf("Event: %s - %s\n", event.Key, event.Message)
    case err := <-errorCh:
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Batch Operations
```go
// Batch get multiple devices
deviceIDs := []string{"id1", "id2", "id3"}
results := gofi.BatchGet(ctx, deviceIDs, func(ctx context.Context, id string) (*types.Device, error) {
    return client.Devices().Get(ctx, "default", id)
})

// Check results
for _, result := range results {
    if result.Error != nil {
        fmt.Printf("Error at index %d: %v\n", result.Index, result.Error)
    } else {
        fmt.Printf("Device: %s\n", result.Item.Name)
    }
}
```

## Configuration

### Basic Configuration

```go
config := &gofi.Config{
    Host:     "192.168.1.1",
    Port:     443, // Default
    Username: "admin",
    Password: "password",
    Site:     "default", // Default site
}
```

### TLS Configuration

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
    SkipTLSVerify: true, // WARNING: Insecure, testing only
}
```

### Advanced Options

```go
client, err := gofi.New(config,
    gofi.WithTimeout(30*time.Second),
    gofi.WithRetry(3, 100*time.Millisecond),
    gofi.WithSite("custom-site"),
    gofi.WithLogger(customLogger),
)
```

### Retry Configuration

```go
config := &gofi.Config{
    // ... other config ...
    RetryConfig: &gofi.RetryConfig{
        MaxRetries:     3,
        InitialBackoff: 100 * time.Millisecond,
        MaxBackoff:     5 * time.Second,
    },
}
```

## Error Handling

The library provides comprehensive error types:

```go
if err := client.Connect(ctx); err != nil {
    // Check for specific errors
    if errors.Is(err, gofi.ErrAuthenticationFailed) {
        // Handle auth failure
    }
    if errors.Is(err, gofi.ErrNotFound) {
        // Handle not found
    }

    // Get API error details
    var apiErr *gofi.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error [%d]: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

### Available Errors

- `ErrNotConnected` - Operation requires connection
- `ErrAlreadyConnected` - Already connected
- `ErrAuthenticationFailed` - Invalid credentials
- `ErrSessionExpired` - Session expired
- `ErrNotFound` - Resource not found
- `ErrPermissionDenied` - Insufficient permissions
- `ErrRateLimited` - Too many requests
- `ErrServerError` - Server error (5xx)

## Testing

The library includes a comprehensive mock server for testing:

```go
import (
    "testing"
    "github.com/unifi-go/gofi"
    "github.com/unifi-go/gofi/mock"
)

func TestYourCode(t *testing.T) {
    // Create mock server
    server := mock.NewServer()
    defer server.Close()

    // Add test data
    server.State().AddDevice(&types.Device{
        ID:   "test-device",
        Name: "Test AP",
    })

    // Create client
    config := &gofi.Config{
        Host:          server.Host(),
        Port:          server.Port(),
        Username:      "admin",
        Password:      "admin",
        SkipTLSVerify: true,
    }

    client, _ := gofi.New(config)
    client.Connect(context.Background())

    // Test your code
    devices, err := client.Devices().List(context.Background(), "default")
    // ...
}
```

## Architecture

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
└── examples/          # Usage examples
```

## Development

```bash
# Run tests
make test

# Run tests with coverage
make coverage

# Run linter
make lint

# Build
make build

# Run all checks
make all
```

## Requirements

- Go 1.21 or later
- UniFi UDM Pro with Network Application 10.x+
- Admin access to the controller

## Compatibility

Tested with:
- UniFi OS 4.x and 5.x
- Network Application 10.x
- UDM Pro, UDM SE, and UDR devices

## Documentation

- See [examples](./examples/) for usage examples
- See [DESIGN.md](./docs/DESIGN.md) for architecture details
- See [GoDoc](https://pkg.go.dev/github.com/unifi-go/gofi) for API reference

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

## Project Status

✅ **Production Ready** - All 21 implementation phases complete

- Phase 0-5: Core foundation (types, utilities, errors, transport, auth)
- Phase 6-7: Mock server and client core
- Phase 8-17: All services (Site, Device, Network, WLAN, Firewall, Client, User, Routing, Ports, Settings, System)
- Phase 18: WebSocket support
- Phase 19: Concurrency and batch operations
- Phase 20: Examples and documentation
- Phase 21: Final testing and polish

**500+ tests passing** with race detection enabled.
