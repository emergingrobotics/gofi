package services

import (
	"context"

	"github.com/unifi-go/gofi/types"
)

// Placeholder service interfaces - will be implemented in later phases

// SiteService provides site management operations.
type SiteService interface {
	List(ctx context.Context) ([]types.Site, error)
	Get(ctx context.Context, id string) (*types.Site, error)
	Create(ctx context.Context, desc, name string) (*types.Site, error)
	Update(ctx context.Context, site *types.Site) (*types.Site, error)
	Delete(ctx context.Context, id string) error
	Health(ctx context.Context, site string) ([]types.HealthData, error)
	SysInfo(ctx context.Context, site string) (*types.SysInfo, error)
}

// DeviceService provides device control and configuration.
type DeviceService interface {
	List(ctx context.Context, site string) ([]types.Device, error)
	ListBasic(ctx context.Context, site string) ([]types.DeviceBasic, error)
	Get(ctx context.Context, site, id string) (*types.Device, error)
	GetByMAC(ctx context.Context, site, mac string) (*types.Device, error)
	Update(ctx context.Context, site string, device *types.Device) (*types.Device, error)
	Adopt(ctx context.Context, site, mac string) error
	Forget(ctx context.Context, site, mac string) error
	Restart(ctx context.Context, site, mac string) error
	ForceProvision(ctx context.Context, site, mac string) error
	Upgrade(ctx context.Context, site, mac string) error
	UpgradeExternal(ctx context.Context, site, mac, url string) error
	Locate(ctx context.Context, site, mac string) error
	Unlocate(ctx context.Context, site, mac string) error
	PowerCyclePort(ctx context.Context, site, switchMAC string, portIdx int) error
	SetLEDOverride(ctx context.Context, site, mac, mode string) error
	SpectrumScan(ctx context.Context, site, mac string) error
}

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
