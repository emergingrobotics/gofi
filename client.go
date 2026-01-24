package gofi

import (
	"context"

	"github.com/unifi-go/gofi/services"
)

// Client is the main interface for interacting with a UDM Pro.
type Client interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool

	// Service accessors
	Sites() services.SiteService
	Devices() services.DeviceService
	Networks() services.NetworkService
	WLANs() services.WLANService
	Firewall() services.FirewallService
	Clients() services.ClientService
	Users() services.UserService
	Routing() services.RoutingService
	PortForwards() services.PortForwardService
	PortProfiles() services.PortProfileService
	Settings() services.SettingService
	System() services.SystemService
	Events() services.EventService
	DNS() services.DNSService
}
