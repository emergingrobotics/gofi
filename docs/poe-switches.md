# Working with PoE Switches

This document describes how to manage Power over Ethernet (PoE) on UniFi switches managed by a UDM Pro using the gofi library.

## 1. Discovering Switches on the Network

**API Endpoint**: `GET /proxy/network/api/s/{site}/stat/device`

Filter devices where `type == "usw"` (UniFi Switch). Example models include `USW-24-PoE`, `USW-Pro-24-PoE`, etc.

```go
// List all devices, then filter for switches
devices, _ := client.Devices().List(ctx, "default")
for _, d := range devices {
    if d.Type == "usw" {
        fmt.Printf("Switch: %s (%s) - MAC: %s\n", d.Name, d.Model, d.MAC)
    }
}
```

## 2. Getting PoE Power Consumption Per Port

Each switch device has a `PortTable` field (`port_table` in JSON) containing per-port statistics. The relevant PoE fields in `types.PortTable` are:

| Field | Type | Description |
|-------|------|-------------|
| `PortPoe` | `bool` | Whether port supports PoE |
| `PoeEnable` | `bool` | Whether PoE is currently enabled |
| `PoeGood` | `bool` | Whether PoE is functioning properly |
| `PoePower` | `FlexInt` | Power consumption in watts |
| `PoeCurrent` | `FlexInt` | Current draw (mA) |
| `PoeVoltage` | `FlexInt` | Voltage (V) |
| `PoeClass` | `string` | PoE class (determines max power) |
| `PoeCaps` | `int` | PoE capabilities bitmask |
| `PoeMode` | `string` | Mode: `"auto"`, `"pasv24"`, `"passthrough"`, `"off"` |

```go
// Get a switch and read PoE stats per port
device, _ := client.Devices().GetByMAC(ctx, "default", "aa:bb:cc:dd:ee:ff")
for _, port := range device.PortTable {
    if port.PortPoe {
        fmt.Printf("Port %d: PoE=%v, Power=%.1fW, Voltage=%.1fV, Current=%.0fmA\n",
            port.PortIdx, port.PoeEnable,
            port.PoePower.Float64(),   // Watts
            port.PoeVoltage.Float64(), // Volts
            port.PoeCurrent.Float64()) // Milliamps
    }
}
```

## 3. Enabling/Disabling PoE Per Port

There are two approaches:

### Option A: Update Device with Port Overrides

**API Endpoint**: `PUT /proxy/network/api/s/{site}/rest/device/{device_id}`

Use the `port_overrides` field to set `poe_mode` per port:

```go
// types.PortOverride structure
type PortOverride struct {
    PortIdx    int    `json:"port_idx"`
    PoeMode    string `json:"poe_mode,omitempty"`  // "auto", "off", "pasv24", "passthrough"
    PortconfID string `json:"portconf_id,omitempty"`
    Name       string `json:"name,omitempty"`
}
```

**Example request body** to disable PoE on port 5:
```json
{
    "port_overrides": [
        {"port_idx": 5, "poe_mode": "off"}
    ]
}
```

**To re-enable** PoE on port 5:
```json
{
    "port_overrides": [
        {"port_idx": 5, "poe_mode": "auto"}
    ]
}
```

### Option B: Power Cycle a PoE Port

**API Endpoint**: `POST /proxy/network/api/s/{site}/cmd/devmgr`

This temporarily cycles power (turns off then on), useful for rebooting PoE devices:

```json
{
    "cmd": "power-cycle",
    "mac": "aa:bb:cc:dd:ee:ff",
    "port_idx": 5
}
```

```go
// In the gofi library this maps to:
client.Devices().PowerCyclePort(ctx, "default", "aa:bb:cc:dd:ee:ff", 5)
```

## PoE Mode Values

| Mode | Description |
|------|-------------|
| `"auto"` | Automatic PoE negotiation (802.3af/at/bt) |
| `"off"` | PoE disabled on this port |
| `"pasv24"` | Passive 24V PoE (legacy UniFi devices) |
| `"passthrough"` | Pass PoE from uplink to this port |

## Quick Reference

| Task | Endpoint | Method |
|------|----------|--------|
| List switches | `GET /api/s/{site}/stat/device` | Filter `type=="usw"` |
| Get port power stats | Same - read `port_table[].poe_power` | From device response |
| Disable/enable PoE | `PUT /api/s/{site}/rest/device/{id}` | Set `port_overrides[].poe_mode` |
| Power cycle port | `POST /api/s/{site}/cmd/devmgr` | `cmd: "power-cycle"` |

## Related Types

- `types.Device` - Contains `PortTable` and `PortOverrides` for switches
- `types.PortTable` - Per-port status including PoE statistics
- `types.PortOverride` - Per-port configuration overrides
- `types.PortProfile` - Reusable port profiles with default PoE settings
