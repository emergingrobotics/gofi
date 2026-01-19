# gofi Examples

This directory contains example programs demonstrating various features of the gofi UniFi controller client library.

## Prerequisites

Before running these examples, you need:
- A UniFi UDM Pro controller (or compatible device)
- Admin credentials
- Network access to the controller

## Examples

### basic - Getting Started

Demonstrates basic client usage:
- Connecting to the controller
- Listing sites
- Listing devices
- Listing networks
- Getting health information

```bash
cd examples/basic
go run main.go
```

### crud - Create/Read/Update/Delete Operations

Shows how to manage resources:
- Creating a network
- Creating a WLAN
- Updating configuration
- Deleting resources

```bash
cd examples/crud
go run main.go
```

### concurrent - Concurrent Operations

Demonstrates concurrent/batch operations:
- Batch device retrieval
- Concurrent command execution
- Error handling with partial failures

```bash
cd examples/concurrent
go run main.go
```

### websocket - Real-Time Events

Shows WebSocket event streaming:
- Subscribing to events
- Processing different event types
- Graceful shutdown

```bash
cd examples/websocket
go run main.go
```

### errors - Error Handling

Comprehensive error handling examples:
- Connection errors
- Resource not found
- API errors
- Validation errors
- Automatic retry configuration

```bash
cd examples/errors
go run main.go
```

## Configuration

Update the following values in each example before running:

```go
config := &gofi.Config{
    Host:          "192.168.1.1",  // Your UDM Pro IP
    Username:      "admin",         // Your admin username
    Password:      "your-password", // Your admin password
    SkipTLSVerify: true,            // Only for self-signed certs
}
```

For production use, configure proper TLS:

```go
config := &gofi.Config{
    Host:     "unifi.yourdomain.com",
    Username: "admin",
    Password: os.Getenv("UNIFI_PASSWORD"),
    TLSConfig: &tls.Config{
        // Your TLS configuration
    },
}
```

## Safety Note

These examples are for demonstration purposes. Some operations (creating/deleting networks, rebooting devices) can affect your network. Review and test carefully before running on production systems.
