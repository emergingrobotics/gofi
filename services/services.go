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
type NetworkService interface {
	List(ctx context.Context, site string) ([]types.Network, error)
	Get(ctx context.Context, site, id string) (*types.Network, error)
	Create(ctx context.Context, site string, network *types.Network) (*types.Network, error)
	Update(ctx context.Context, site string, network *types.Network) (*types.Network, error)
	Delete(ctx context.Context, site, id string) error
}

// WLANService provides wireless network configuration.
type WLANService interface {
	// WLAN methods
	List(ctx context.Context, site string) ([]types.WLAN, error)
	Get(ctx context.Context, site, id string) (*types.WLAN, error)
	Create(ctx context.Context, site string, wlan *types.WLAN) (*types.WLAN, error)
	Update(ctx context.Context, site string, wlan *types.WLAN) (*types.WLAN, error)
	Delete(ctx context.Context, site, id string) error
	Enable(ctx context.Context, site, id string) error
	Disable(ctx context.Context, site, id string) error
	SetMACFilter(ctx context.Context, site, id, policy string, macs []string) error

	// WLAN Group methods
	ListGroups(ctx context.Context, site string) ([]types.WLANGroup, error)
	GetGroup(ctx context.Context, site, id string) (*types.WLANGroup, error)
	CreateGroup(ctx context.Context, site string, group *types.WLANGroup) (*types.WLANGroup, error)
	UpdateGroup(ctx context.Context, site string, group *types.WLANGroup) (*types.WLANGroup, error)
	DeleteGroup(ctx context.Context, site, id string) error
}

// FirewallService provides firewall rules and groups management.
type FirewallService interface {
	// Firewall Rule methods
	ListRules(ctx context.Context, site string) ([]types.FirewallRule, error)
	GetRule(ctx context.Context, site, id string) (*types.FirewallRule, error)
	CreateRule(ctx context.Context, site string, rule *types.FirewallRule) (*types.FirewallRule, error)
	UpdateRule(ctx context.Context, site string, rule *types.FirewallRule) (*types.FirewallRule, error)
	DeleteRule(ctx context.Context, site, id string) error
	EnableRule(ctx context.Context, site, id string) error
	DisableRule(ctx context.Context, site, id string) error
	ReorderRules(ctx context.Context, site, ruleset string, updates []types.FirewallRuleIndexUpdate) error

	// Firewall Group methods
	ListGroups(ctx context.Context, site string) ([]types.FirewallGroup, error)
	GetGroup(ctx context.Context, site, id string) (*types.FirewallGroup, error)
	CreateGroup(ctx context.Context, site string, group *types.FirewallGroup) (*types.FirewallGroup, error)
	UpdateGroup(ctx context.Context, site string, group *types.FirewallGroup) (*types.FirewallGroup, error)
	DeleteGroup(ctx context.Context, site, id string) error

	// Traffic Rule methods (v2 API)
	ListTrafficRules(ctx context.Context, site string) ([]types.TrafficRule, error)
	GetTrafficRule(ctx context.Context, site, id string) (*types.TrafficRule, error)
	CreateTrafficRule(ctx context.Context, site string, rule *types.TrafficRule) (*types.TrafficRule, error)
	UpdateTrafficRule(ctx context.Context, site string, rule *types.TrafficRule) (*types.TrafficRule, error)
	DeleteTrafficRule(ctx context.Context, site, id string) error
}

// ClientService provides connected client/station operations.
type ClientService interface {
	// ListActive returns all currently connected clients.
	ListActive(ctx context.Context, site string) ([]types.Client, error)

	// ListAll returns all known clients (including historical).
	ListAll(ctx context.Context, site string, opts ...ClientListOption) ([]types.Client, error)

	// Get returns a client by MAC address.
	Get(ctx context.Context, site, mac string) (*types.Client, error)

	// Block blocks a client from the network.
	Block(ctx context.Context, site, mac string) error

	// Unblock unblocks a previously blocked client.
	Unblock(ctx context.Context, site, mac string) error

	// Kick disconnects a client from the network.
	Kick(ctx context.Context, site, mac string) error

	// AuthorizeGuest authorizes a guest client.
	AuthorizeGuest(ctx context.Context, site, mac string, opts ...GuestAuthOption) error

	// UnauthorizeGuest revokes guest authorization.
	UnauthorizeGuest(ctx context.Context, site, mac string) error

	// Forget removes a client from the known clients list.
	Forget(ctx context.Context, site, mac string) error

	// SetFingerprint overrides the device fingerprint.
	SetFingerprint(ctx context.Context, site, mac string, devID int) error
}

// ClientListOption configures client list queries.
type ClientListOption func(*clientListOptions)

// clientListOptions holds options for listing clients.
type clientListOptions struct {
	withinHours int
}

// WithinHours limits results to clients seen within the specified hours.
func WithinHours(hours int) ClientListOption {
	return func(opts *clientListOptions) {
		opts.withinHours = hours
	}
}

// GuestAuthOption configures guest authorization.
type GuestAuthOption func(*guestAuthOptions)

// guestAuthOptions holds options for guest authorization.
type guestAuthOptions struct {
	minutes int
	up      int
	down    int
	bytes   int
	apMAC   string
}

// WithDuration sets the authorization duration in minutes.
func WithDuration(minutes int) GuestAuthOption {
	return func(opts *guestAuthOptions) {
		opts.minutes = minutes
	}
}

// WithUploadLimit sets the upload bandwidth limit in Kbps.
func WithUploadLimit(kbps int) GuestAuthOption {
	return func(opts *guestAuthOptions) {
		opts.up = kbps
	}
}

// WithDownloadLimit sets the download bandwidth limit in Kbps.
func WithDownloadLimit(kbps int) GuestAuthOption {
	return func(opts *guestAuthOptions) {
		opts.down = kbps
	}
}

// WithDataLimit sets the total data limit in megabytes.
func WithDataLimit(mb int) GuestAuthOption {
	return func(opts *guestAuthOptions) {
		opts.bytes = mb
	}
}

// WithAPMAC restricts authorization to a specific AP.
func WithAPMAC(mac string) GuestAuthOption {
	return func(opts *guestAuthOptions) {
		opts.apMAC = mac
	}
}

// UserService provides known client/user management.
type UserService interface {
	// User operations
	List(ctx context.Context, site string) ([]types.User, error)
	Get(ctx context.Context, site, id string) (*types.User, error)
	GetByMAC(ctx context.Context, site, mac string) (*types.User, error)
	Create(ctx context.Context, site string, user *types.User) (*types.User, error)
	Update(ctx context.Context, site string, user *types.User) (*types.User, error)
	Delete(ctx context.Context, site, id string) error
	DeleteByMAC(ctx context.Context, site, mac string) error
	SetFixedIP(ctx context.Context, site, mac, ip, networkID string) error
	ClearFixedIP(ctx context.Context, site, mac string) error

	// User group operations
	ListGroups(ctx context.Context, site string) ([]types.UserGroup, error)
	GetGroup(ctx context.Context, site, id string) (*types.UserGroup, error)
	CreateGroup(ctx context.Context, site string, group *types.UserGroup) (*types.UserGroup, error)
	UpdateGroup(ctx context.Context, site string, group *types.UserGroup) (*types.UserGroup, error)
	DeleteGroup(ctx context.Context, site, id string) error
}

// RoutingService provides static route management.
type RoutingService interface {
	List(ctx context.Context, site string) ([]types.Route, error)
	Get(ctx context.Context, site, id string) (*types.Route, error)
	Create(ctx context.Context, site string, route *types.Route) (*types.Route, error)
	Update(ctx context.Context, site string, route *types.Route) (*types.Route, error)
	Delete(ctx context.Context, site, id string) error
	Enable(ctx context.Context, site, id string) error
	Disable(ctx context.Context, site, id string) error
}

// PortForwardService provides port forwarding management.
type PortForwardService interface {
	List(ctx context.Context, site string) ([]types.PortForward, error)
	Get(ctx context.Context, site, id string) (*types.PortForward, error)
	Create(ctx context.Context, site string, forward *types.PortForward) (*types.PortForward, error)
	Update(ctx context.Context, site string, forward *types.PortForward) (*types.PortForward, error)
	Delete(ctx context.Context, site, id string) error
	Enable(ctx context.Context, site, id string) error
	Disable(ctx context.Context, site, id string) error
}

// PortProfileService provides port profile management.
type PortProfileService interface {
	List(ctx context.Context, site string) ([]types.PortProfile, error)
	Get(ctx context.Context, site, id string) (*types.PortProfile, error)
	Create(ctx context.Context, site string, profile *types.PortProfile) (*types.PortProfile, error)
	Update(ctx context.Context, site string, profile *types.PortProfile) (*types.PortProfile, error)
	Delete(ctx context.Context, site, id string) error
}

// SettingService provides system settings management.
type SettingService interface {
	Get(ctx context.Context, site, key string) (interface{}, error)
	Update(ctx context.Context, site string, setting interface{}) error

	// RADIUS profiles
	ListRadiusProfiles(ctx context.Context, site string) ([]types.RADIUSProfile, error)
	GetRadiusProfile(ctx context.Context, site, id string) (*types.RADIUSProfile, error)
	CreateRadiusProfile(ctx context.Context, site string, profile *types.RADIUSProfile) (*types.RADIUSProfile, error)
	UpdateRadiusProfile(ctx context.Context, site string, profile *types.RADIUSProfile) (*types.RADIUSProfile, error)
	DeleteRadiusProfile(ctx context.Context, site, id string) error

	// Dynamic DNS
	GetDynamicDNS(ctx context.Context, site string) (*types.DynamicDNS, error)
	UpdateDynamicDNS(ctx context.Context, site string, ddns *types.DynamicDNS) error
}

// SystemService provides system-level operations.
type SystemService interface {
	Status(ctx context.Context) (*types.Status, error)
	Self(ctx context.Context) (*types.AdminUser, error)
	Reboot(ctx context.Context) error
	SpeedTest(ctx context.Context, site string) error
	SpeedTestStatus(ctx context.Context, site string) (*types.SpeedTestStatus, error)
	ListBackups(ctx context.Context) ([]types.Backup, error)
	CreateBackup(ctx context.Context) error
	DeleteBackup(ctx context.Context, filename string) error
	ListAdmins(ctx context.Context) ([]types.AdminUser, error)
}

// EventService provides real-time event streaming.
type EventService interface {
	Subscribe(ctx context.Context, site string) (<-chan types.Event, <-chan error, error)
	Close() error
}
