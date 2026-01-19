package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// APIResponse is a generic wrapper for UniFi API responses.
type APIResponse[T any] struct {
	Meta ResponseMeta `json:"meta"`
	Data []T          `json:"data"`
}

// ResponseMeta contains metadata about the API response.
type ResponseMeta struct {
	RC      string `json:"rc"`       // Response code ("ok" for success)
	Message string `json:"msg,omitempty"`
	Count   int    `json:"count,omitempty"`
}

// CommandRequest is a generic command request structure.
type CommandRequest struct {
	Cmd string `json:"cmd"`

	// Additional fields for various commands
	MAC      string `json:"mac,omitempty"`
	Duration int    `json:"duration,omitempty"`

	// For upgrades
	URL string `json:"url,omitempty"`

	// For LED override
	Mode string `json:"mode,omitempty"`

	// For port power cycle
	PortIdx int `json:"port_idx,omitempty"`

	// For guest authorization
	Minutes int    `json:"minutes,omitempty"`
	Up      int    `json:"up,omitempty"`   // Upload limit in kbps
	Down    int    `json:"down,omitempty"` // Download limit in kbps
	Bytes   int64  `json:"bytes,omitempty"` // Data transfer limit in bytes
	APMAC   string `json:"ap_mac,omitempty"`
}

// MAC represents a MAC address.
type MAC string

// Validate checks if the MAC address is valid.
func (m MAC) Validate() error {
	if m == "" {
		return fmt.Errorf("MAC address cannot be empty")
	}

	// Normalize and validate format
	normalized := strings.ToLower(strings.ReplaceAll(string(m), ":", ""))
	normalized = strings.ReplaceAll(normalized, "-", "")

	if len(normalized) != 12 {
		return fmt.Errorf("invalid MAC address length: %s", m)
	}

	// Check if hex
	matched, _ := regexp.MatchString("^[0-9a-f]{12}$", normalized)
	if !matched {
		return fmt.Errorf("invalid MAC address format: %s", m)
	}

	return nil
}

// String returns the MAC address as a string.
func (m MAC) String() string {
	return string(m)
}

// DeviceState represents the state of a device.
type DeviceState int

const (
	DeviceStateOffline      DeviceState = 0
	DeviceStateConnected    DeviceState = 1
	DeviceStatePending      DeviceState = 2
	DeviceStateDisconnected DeviceState = 3
	DeviceStateFirmware     DeviceState = 4
	DeviceStateProvisioning DeviceState = 5
	DeviceStateHeartbeat    DeviceState = 6
	DeviceStateAdopting     DeviceState = 7
	DeviceStateDeleting     DeviceState = 8
	DeviceStateInformed     DeviceState = 9
	DeviceStateUpgrading    DeviceState = 10
)

// String returns a string representation of the device state.
func (s DeviceState) String() string {
	switch s {
	case DeviceStateOffline:
		return "offline"
	case DeviceStateConnected:
		return "connected"
	case DeviceStatePending:
		return "pending"
	case DeviceStateDisconnected:
		return "disconnected"
	case DeviceStateFirmware:
		return "firmware"
	case DeviceStateProvisioning:
		return "provisioning"
	case DeviceStateHeartbeat:
		return "heartbeat"
	case DeviceStateAdopting:
		return "adopting"
	case DeviceStateDeleting:
		return "deleting"
	case DeviceStateInformed:
		return "informed"
	case DeviceStateUpgrading:
		return "upgrading"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// UnmarshalJSON implements json.Unmarshaler for DeviceState.
func (s *DeviceState) UnmarshalJSON(data []byte) error {
	var val int
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	*s = DeviceState(val)
	return nil
}

// MarshalJSON implements json.Marshaler for DeviceState.
func (s DeviceState) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(s))
}
