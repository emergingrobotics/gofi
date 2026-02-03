# gofi Examples

This directory contains example programs demonstrating how to use the gofi library to interact with UniFi UDM Pro controllers.

## Prerequisites

All examples require:
- UniFi UDM Pro controller (v10+)
- Controller credentials set via environment variables:
  ```bash
  export UNIFI_USERNAME="your-username"
  export UNIFI_PASSWORD="your-password"
  ```

## Building Examples

Build all examples:
```bash
make build-examples
```

Build individual example:
```bash
go build -o bin/basic examples/basic/main.go
```

## Common Flags

Most examples support these flags:

| Flag | Shorthand | Description | Default |
|------|-----------|-------------|---------|
| `--host` | `-H` | UDM Pro host address | (required) |
| `--port` | `-p` | UDM Pro port | 443 |
| `--site` | `-s` | Site name | default |
| `--insecure` | `-k` | Skip TLS verification | false |
| `--json` | `-j` | Output in JSON format | false |
| `--debug` | `-d` | Enable debug logging | false |

## Example Programs

### 1. basic - Getting Started

**Location**: `examples/basic/main.go`

Basic example demonstrating core gofi functionality. Lists sites, devices, networks, and health status.

**Features**:
- Client creation and authentication
- Site listing
- Device enumeration
- Network discovery
- Health status monitoring
- Custom logger implementation

**Usage**:
```bash
./bin/basic --host 192.168.1.1 --insecure

# With debug logging
./bin/basic --host 192.168.1.1 --insecure --debug

# Different site
./bin/basic --host 192.168.1.1 --site production
```

**Output**:
```
Connected to UniFi controller!

Found 2 site(s):
  - Home (default)
  - Office (office)

Found 5 device(s):
  - UDM Pro (UDMP) - aa:bb:cc:dd:ee:ff - State: Connected
  - Living Room AP (U6-Lite) - aa:bb:cc:dd:ee:01 - State: Connected
  ...

Found 3 network(s):
  - Default (VLAN Enabled: false)
  - IoT (VLAN Enabled: true)
  - Guest (VLAN Enabled: true)

Health Status:
  - wan: up
  - lan: up
  - wlan: up
```

---

### 2. list - Network Listing

**Location**: `examples/list/main.go`

Lists all networks from the controller with detailed information in table or JSON format.

**Features**:
- Network enumeration
- VLAN configuration display
- DHCP status
- Formatted table output
- JSON export

**Usage**:
```bash
# Table format
./bin/list --host 192.168.1.1 --insecure

# JSON format
./bin/list --host 192.168.1.1 --insecure --json

# JSON with jq filtering
./bin/list -H 192.168.1.1 -k -j | jq '.[] | select(.vlan_enabled==true)'
```

**Output** (table):
```
NAME     TYPE       VLAN  SUBNET            DHCP  ENABLED
----     ----       ----  ------            ----  -------
Default  corporate  -     192.168.1.1/24    Yes   Yes
IoT      corporate  20    192.168.20.1/24   Yes   Yes
Guest    guest      30    192.168.30.1/24   Yes   Yes
```

**Output** (JSON):
```json
[
  {
    "id": "507f1f77bcf86cd799439011",
    "name": "Default",
    "purpose": "corporate",
    "vlan_enabled": false,
    "subnet": "192.168.1.1/24",
    "dhcp_enabled": true,
    "enabled": true
  },
  ...
]
```

---

### 3. websocket - Event Streaming

**Location**: `examples/websocket/main.go`

Subscribes to real-time events from the controller via WebSocket.

**Features**:
- WebSocket event subscription
- Real-time event display
- Graceful shutdown (Ctrl+C)
- Event filtering and display

**Usage**:
```bash
./bin/websocket
```

**Note**: This example has hardcoded credentials in the source. Update before use:
```go
config := &gofi.Config{
    Host:          "192.168.1.1",
    Username:      "admin",
    Password:      "your-password",
    SkipTLSVerify: true,
}
```

**Output**:
```
Connected! Subscribing to events...
Listening for events... Press Ctrl+C to exit

[EVENT] EVT_WU_Connected: Client connected
        Client: iPhone, SSID: HomeWiFi

[EVENT] EVT_AP_Connected: Access point connected
        AP: Living Room AP (aa:bb:cc:dd:ee:01)

[EVENT] EVT_WU_Disconnected: Client disconnected
        Client: iPhone, SSID: HomeWiFi
```

---

### 4. crud - Create/Update/Delete Operations

**Location**: `examples/crud/main.go`

Demonstrates full CRUD lifecycle for networks and WLANs.

**Features**:
- Network creation
- WLAN creation with security settings
- Update operations
- Resource listing
- Cleanup (deletion)

**Usage**:
```bash
./bin/crud
```

**Note**: Hardcoded credentials - update before use. This example creates resources and then deletes them.

**Output**:
```
Creating new network...
Created network: IoT Network (ID: 507f1f77bcf86cd799439011)

Creating new WLAN...
Created WLAN: Guest WiFi (ID: 507f1f77bcf86cd799439012)

Updating WLAN...
Updated WLAN name to: Guest WiFi (Updated)

Listing all WLANs...
  - HomeWiFi (Security: wpapsk, Enabled: true)
  - Guest WiFi (Updated) (Security: wpapsk, Enabled: true)

Cleaning up...
Deleted WLAN
Deleted network

Done!
```

---

### 5. concurrent - Batch Operations

**Location**: `examples/concurrent/main.go`

Demonstrates concurrent operations using gofi's batch utilities.

**Features**:
- Parallel device fetching
- Batch get operations
- Error handling for partial failures
- Performance optimization

**Usage**:
```bash
./bin/concurrent
```

**Note**: Hardcoded credentials - update before use.

**Output**:
```
Fetching 5 devices concurrently...
  [OK] UDM Pro (aa:bb:cc:dd:ee:ff)
  [OK] Living Room AP (aa:bb:cc:dd:ee:01)
  [OK] Office Switch (aa:bb:cc:dd:ee:02)
  [ERROR] Index 3: device not found
  [OK] Bedroom AP (aa:bb:cc:dd:ee:04)

Batch operation complete: 4 successful, 1 errors

Example: Concurrent device operations (commented out for safety)
// To locate multiple devices:
// macs := []string{"aa:bb:cc:dd:ee:f1", "aa:bb:cc:dd:ee:f2"}
// errors := gofi.BatchDelete(ctx, macs, func(ctx context.Context, mac string) error {
//     return client.Devices().Locate(ctx, site, mac)
// })
```

---

### 6. errors - Error Handling

**Location**: `examples/errors/main.go`

Comprehensive error handling examples using gofi's error types.

**Features**:
- Connection error handling
- Authentication errors
- Not found errors
- API error inspection
- Validation errors
- Retry configuration

**Usage**:
```bash
./bin/errors
```

**Note**: Hardcoded credentials - update before use.

**Output**:
```
=== Example 1: Handling connection errors ===
Connected successfully!

=== Example 2: Handling not found errors ===
Device not found (expected)

=== Example 3: Handling API errors ===
API Error [404]: Resource not found (endpoint: /api/s/default/rest/networkconf/invalid-id)

=== Example 4: Handling validation errors ===
Validation error on field 'host': host is required

=== Example 5: Automatic retry on transient failures ===
Connected with retry configuration

=== Error handling examples complete ===
```

---

### 7. fixedips - List Fixed IP Assignments

**Location**: `examples/fixedips/main.go`

Lists all clients with fixed IP address assignments.

**Features**:
- Fixed IP enumeration
- Client identification (name/hostname/MAC)
- Sorted output by IP
- Table and JSON formats

**Usage**:
```bash
# Table format
./bin/fixedips --host 192.168.1.1 --insecure

# JSON format
./bin/fixedips -H 192.168.1.1 -k -j

# Different site
./bin/fixedips -H 192.168.1.1 -k --site office

# JSON with jq
./bin/fixedips -H 192.168.1.1 -k -j | jq '.[] | select(.fixed_ip | startswith("192.168.1"))'
```

**Output** (table):
```
NAME           HOSTNAME       MAC                 FIXED IP
----           --------       ---                 --------
File Server    fileserver     aa:bb:cc:dd:ee:10   192.168.1.10
Printer        printer        aa:bb:cc:dd:ee:20   192.168.1.20
NAS            nas            aa:bb:cc:dd:ee:30   192.168.1.30

Total: 3 fixed IP assignments
```

---

### 8. addfixedip - Add Fixed IP Assignment

**Location**: `examples/addfixedip/main.go`

Assigns a fixed IP address to a device by MAC address.

**Features**:
- MAC address validation
- IP address validation
- Network auto-detection from IP subnet
- Conflict checking (IP already in use)
- DNS record awareness
- Update existing assignments

**Usage**:
```bash
# Basic usage - auto-detect network
./bin/addfixedip \
  --host 192.168.1.1 \
  --insecure \
  --mac aa:bb:cc:dd:ee:ff \
  --ip 192.168.1.100 \
  --name "My Device"

# Specify network explicitly
./bin/addfixedip \
  -H 192.168.1.1 -k \
  -m aa:bb:cc:dd:ee:ff \
  -i 192.168.1.100 \
  -n "My Device" \
  -N "LAN"

# Force assignment (skip conflict checks)
./bin/addfixedip \
  -H 192.168.1.1 -k \
  -m aa:bb:cc:dd:ee:ff \
  -i 192.168.1.100 \
  -n "My Device" \
  --force
```

**Output**:
```
Created fixed IP assignment:
  Name:    My Device
  MAC:     aa:bb:cc:dd:ee:ff
  IP:      192.168.1.100
  Network: Default
```

**Conflict Detection**:
```
Error: IP 192.168.1.100 is already reserved for Printer (aa:bb:cc:dd:ee:20)
Use --force to skip this check
```

---

### 9. delfixedip - Remove Fixed IP Assignment

**Location**: `examples/delfixedip/main.go`

Removes fixed IP assignments, allowing devices to use DHCP. Handles dependent DNS records.

**Features**:
- Lookup by MAC or IP address
- DNS record dependency checking
- Automatic DNS cleanup
- Option to delete user entry entirely
- Option to preserve DNS records

**Usage**:
```bash
# Remove fixed IP by MAC (clears fixed IP, keeps user)
./bin/delfixedip --host 192.168.1.1 --insecure --mac aa:bb:cc:dd:ee:ff

# Remove fixed IP by IP address
./bin/delfixedip -H 192.168.1.1 -k -i 192.168.1.100

# Keep associated DNS records
./bin/delfixedip -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff --keep-dns

# Delete the entire user entry (not just fixed IP)
./bin/delfixedip -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff --delete
```

**Output**:
```
Found user:
  Name:     My Device
  MAC:      aa:bb:cc:dd:ee:ff
  Fixed IP: 192.168.1.100
  DNS Records:
    - mydevice.local -> 192.168.1.100

Deleting 1 DNS record(s) that depend on the fixed IP...
  Deleted DNS record: mydevice.local

Removed fixed IP assignment. Device will now use DHCP.
```

**DNS Dependency Error**:
```
Error: There are still DNS records depending on this fixed IP.
This can happen if DNS records were added after we checked.
Please try again or manually delete the DNS records.
```

---

### 10. switches - Switch Management & PoE Control

**Location**: `examples/switches/main.go`

Comprehensive switch management including listing and PoE port control.

**Features**:
- List all switches with port counts
- Detailed port information (speed, PoE status)
- PoE enable/disable/cycle operations
- State change confirmation with polling
- JSON and table output formats
- Hardware settle time handling

**Usage**:

#### List Switches

```bash
# Table format
./bin/switches --host 192.168.1.1 --list --insecure

# JSON format with full port details
./bin/switches -H 192.168.1.1 -l -j

# Different site
./bin/switches -H 192.168.1.1 -s office -l
```

**Output** (table):
```
NAME       MODEL     MAC                 IP            STATE      PORTS  POE  MAX POWER
----       -----     ---                 --            -----      -----  ---  ---------
OfficeSW   USW-24    aa:bb:cc:dd:ee:10   192.168.1.2   Connected  24     8    95W
BedroomSW  USW-Flex  aa:bb:cc:dd:ee:20   192.168.1.3   Connected  5      4    46W

Total: 2 switch(es)
```

#### PoE Control

```bash
# Enable PoE on port 5
./bin/switches \
  --host 192.168.1.1 \
  --insecure \
  --poe enable \
  --switch "OfficeSW" \
  --port-num 5

# Disable PoE with confirmation wait
./bin/switches \
  -H 192.168.1.1 -k \
  -P disable \
  -S "OfficeSW" \
  -n 5 \
  --wait

# Power cycle a port (500ms default)
./bin/switches \
  -H 192.168.1.1 -k \
  -P cycle \
  -S "OfficeSW" \
  -n 5

# Power cycle with custom duration
./bin/switches \
  -H 192.168.1.1 -k \
  -P cycle \
  -S "OfficeSW" \
  -n 5 \
  --duration 2s

# Enable PoE by switch MAC address
./bin/switches \
  -H 192.168.1.1 -k \
  -P enable \
  -S "aa:bb:cc:dd:ee:10" \
  -n 6 \
  --wait \
  --wait-timeout 60s

# JSON output
./bin/switches \
  -H 192.168.1.1 -k \
  -P enable \
  -S "OfficeSW" \
  -n 5 \
  -j
```

**PoE Control Output**:
```
Switch:    OfficeSW (aa:bb:cc:dd:ee:10)
Port:      5
Action:    enable
Previous:  disabled
New State: enabled
Status:    Success
```

**With Wait Confirmation**:
```
Switch:    OfficeSW (aa:bb:cc:dd:ee:10)
Port:      5
Action:    enable
Previous:  disabled
New State: enabled
Actual:    enabled
Wait Time: 2.3s
Status:    Success
```

**JSON Output**:
```json
{
  "meta": {
    "host": "192.168.1.1",
    "site": "default",
    "timestamp": "2025-02-03T10:30:00Z",
    "version": "1.0"
  },
  "action": "enable",
  "switch": {
    "name": "OfficeSW",
    "mac": "aa:bb:cc:dd:ee:10",
    "id": "507f1f77bcf86cd799439011"
  },
  "port": 5,
  "previous_state": "disabled",
  "new_state": "enabled",
  "actual_state": "enabled",
  "success": true,
  "wait_time": "2.3s"
}
```

**Advanced Options**:

The `switches` example supports advanced PoE control options:

- `--wait` / `-w`: Poll the switch to confirm state change completed
- `--wait-timeout` / `-W`: Maximum time to wait for state change (default 30s)
- `--settle` / `-T`: Hardware settle time after config applies (default 2s)
- `--duration` / `-D`: Power cycle duration (default 500ms)

These options help ensure PoE commands fully complete before the program exits.

---

## Common Patterns

### Error Handling

All examples demonstrate proper error handling:

```go
if err := client.Connect(ctx); err != nil {
    if errors.Is(err, gofi.ErrAuthenticationFailed) {
        // Handle auth failure
    }
    if errors.Is(err, gofi.ErrTimeout) {
        // Handle timeout
    }
    // Generic error handling
}
```

### Resource Cleanup

Use defer for cleanup:

```go
client, err := gofi.New(config)
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Disconnect(ctx)
```

### Custom Logging

Implement the `gofi.Logger` interface:

```go
type debugLogger struct{}

func (l *debugLogger) Debug(msg string, keysAndValues ...interface{}) {
    log.Printf("[DEBUG] %s %v", msg, keysAndValues)
}

func (l *debugLogger) Info(msg string, keysAndValues ...interface{}) {
    log.Printf("[INFO] %s %v", msg, keysAndValues)
}

func (l *debugLogger) Warn(msg string, keysAndValues ...interface{}) {
    log.Printf("[WARN] %s %v", msg, keysAndValues)
}

func (l *debugLogger) Error(msg string, keysAndValues ...interface{}) {
    log.Printf("[ERROR] %s %v", msg, keysAndValues)
}

config := &gofi.Config{
    Host:   "192.168.1.1",
    Logger: &debugLogger{},
}
```

### Context Management

Use contexts with timeouts for all operations:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

devices, err := client.Devices().List(ctx, site)
```

## Environment Setup

Create a `.env` file (not committed to git):

```bash
UNIFI_USERNAME=admin
UNIFI_PASSWORD=your-secure-password
```

Load it before running examples:

```bash
source .env
./bin/basic --host 192.168.1.1 --insecure
```

Or use direnv with `.envrc`:

```bash
export UNIFI_USERNAME=admin
export UNIFI_PASSWORD=your-secure-password
```

## Production Usage Notes

1. **TLS Verification**: Don't use `--insecure` in production. Use valid certificates.

2. **Credentials**: Never hardcode credentials. Use environment variables or secret management.

3. **Error Handling**: All examples check errors. Follow this pattern in production code.

4. **Timeouts**: Always use contexts with timeouts to prevent hanging operations.

5. **Logging**: Implement proper logging using the `Logger` interface.

6. **Rate Limiting**: Use `RetryConfig` to handle rate limiting:
   ```go
   config := &gofi.Config{
       Host:     "192.168.1.1",
       Username: username,
       Password: password,
       RetryConfig: &gofi.RetryConfig{
           MaxRetries:     3,
           InitialBackoff: 100,
           MaxBackoff:     5000,
       },
   }
   ```

## Troubleshooting

### Connection Issues

```bash
# Enable debug logging to see HTTP requests
./bin/basic --host 192.168.1.1 --insecure --debug
```

### Authentication Failures

Verify credentials:
```bash
echo "Username: $UNIFI_USERNAME"
echo "Password: [hidden]"
```

### TLS Errors

Use `--insecure` for self-signed certificates:
```bash
./bin/basic --host 192.168.1.1 --insecure
```

Or install the controller's certificate in your system trust store.

### Timeout Issues

Increase timeout:
```bash
./bin/basic --host 192.168.1.1 --timeout 60s
```

## Next Steps

After exploring these examples:

1. Read the [API documentation](UNIFI_UDM_PRO_API_DOCUMENTATION.md)
2. Review the [design document](docs/DESIGN.md)
3. Check the [implementation plan](docs/plan.md)
4. Start building your own integration using gofi
