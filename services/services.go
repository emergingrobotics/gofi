package services

import (
	"context"
)

// Placeholder service interfaces - will be implemented in later phases

// SiteService provides site management operations.
type SiteService interface{}

// DeviceService provides device control and configuration.
type DeviceService interface{}

// NetworkService provides network and VLAN management.
type NetworkService interface{}

// WLANService provides wireless network configuration.
type WLANService interface{}

// FirewallService provides firewall rules and groups management.
type FirewallService interface{}

// ClientService provides connected client/station operations.
type ClientService interface{}

// UserService provides known client/user management.
type UserService interface{}

// RoutingService provides static route management.
type RoutingService interface{}

// PortForwardService provides port forwarding management.
type PortForwardService interface{}

// PortProfileService provides port profile management.
type PortProfileService interface{}

// SettingService provides system settings management.
type SettingService interface{}

// SystemService provides system-level operations.
type SystemService interface{}

// EventService provides real-time event streaming.
type EventService interface {
	Subscribe(ctx context.Context, site string) error
}
