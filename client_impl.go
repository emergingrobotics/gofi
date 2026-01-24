package gofi

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/unifi-go/gofi/auth"
	"github.com/unifi-go/gofi/services"
	"github.com/unifi-go/gofi/transport"
)

// client implements the Client interface.
type client struct {
	config    *Config
	transport transport.Transport
	auth      auth.Manager
	connected atomic.Bool

	// Lazy-initialized services
	mu                  sync.Mutex
	sitesService        services.SiteService
	devicesService      services.DeviceService
	networksService     services.NetworkService
	wlansService        services.WLANService
	firewallService     services.FirewallService
	clientsService      services.ClientService
	usersService        services.UserService
	routingService      services.RoutingService
	portForwardService  services.PortForwardService
	portProfileService  services.PortProfileService
	settingService      services.SettingService
	systemService       services.SystemService
	dnsService          services.DNSService

	logger Logger
}

// New creates a new UniFi client.
func New(config *Config, opts ...Option) (Client, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	// Validate required fields
	if config.Host == "" {
		return nil, NewValidationError("Host", "required")
	}

	if config.Username == "" {
		return nil, NewValidationError("Username", "required")
	}

	if config.Password == "" {
		return nil, NewValidationError("Password", "required")
	}

	// Apply defaults
	if config.Port == 0 {
		config.Port = 443
	}

	if config.Site == "" {
		config.Site = "default"
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * transport.DefaultConfig("").Timeout
	}

	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 10
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Build base URL
	baseURL := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
	}

	// Create transport config
	transportConfig := transport.DefaultConfig(baseURL.String())
	transportConfig.Timeout = config.Timeout
	transportConfig.MaxIdleConns = config.MaxIdleConns
	transportConfig.TLSConfig = config.TLSConfig

	// Apply TLS skip verify if configured
	if config.SkipTLSVerify {
		if transportConfig.TLSConfig == nil {
			transportConfig.TLSConfig = &tls.Config{}
		}
		transportConfig.TLSConfig.InsecureSkipVerify = true
	}

	// Create transport with retry wrapper
	baseTransport, err := transport.New(transportConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Wrap with retry if configured
	var trans transport.Transport
	if config.RetryConfig != nil {
		retryConfig := &transport.RetryConfig{
			MaxRetries:     config.RetryConfig.MaxRetries,
			InitialBackoff: config.RetryConfig.InitialBackoff,
			MaxBackoff:     config.RetryConfig.MaxBackoff,
		}
		trans = transport.NewRetryTransport(baseTransport, retryConfig)
	} else {
		trans = baseTransport
	}

	// Create auth manager
	authMgr := auth.New(trans, config.Username, config.Password)

	c := &client{
		config:    config,
		transport: trans,
		auth:      authMgr,
		logger:    config.Logger,
	}

	return c, nil
}

// Connect establishes a connection to the UniFi controller.
func (c *client) Connect(ctx context.Context) error {
	if c.connected.Load() {
		return ErrAlreadyConnected
	}

	// Perform authentication
	if err := c.auth.Login(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.connected.Store(true)

	if c.logger != nil {
		c.logger.Info("Connected to UniFi controller", "host", c.config.Host)
	}

	return nil
}

// Disconnect closes the connection to the UniFi controller.
func (c *client) Disconnect(ctx context.Context) error {
	if !c.connected.Load() {
		return nil // Already disconnected
	}

	// Logout
	if err := c.auth.Logout(ctx); err != nil {
		if c.logger != nil {
			c.logger.Warn("Logout failed", "error", err)
		}
		// Don't fail disconnect on logout error
	}

	c.connected.Store(false)
	c.transport.Close()

	if c.logger != nil {
		c.logger.Info("Disconnected from UniFi controller")
	}

	return nil
}

// IsConnected returns true if the client is connected.
func (c *client) IsConnected() bool {
	return c.connected.Load() && c.auth.IsAuthenticated()
}

// Sites returns the site service.
func (c *client) Sites() services.SiteService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sitesService == nil {
		c.sitesService = services.NewSiteService(c.transport)
	}

	return c.sitesService
}

// Devices returns the device service.
func (c *client) Devices() services.DeviceService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.devicesService == nil {
		c.devicesService = services.NewDeviceService(c.transport)
	}

	return c.devicesService
}

// Networks returns the network service.
func (c *client) Networks() services.NetworkService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.networksService == nil {
		c.networksService = services.NewNetworkService(c.transport)
	}

	return c.networksService
}

// WLANs returns the WLAN service.
func (c *client) WLANs() services.WLANService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.wlansService == nil {
		c.wlansService = services.NewWLANService(c.transport)
	}

	return c.wlansService
}

// Firewall returns the firewall service.
func (c *client) Firewall() services.FirewallService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.firewallService == nil {
		c.firewallService = services.NewFirewallService(c.transport)
	}

	return c.firewallService
}

// Clients returns the client service.
func (c *client) Clients() services.ClientService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.clientsService == nil {
		c.clientsService = services.NewClientService(c.transport)
	}

	return c.clientsService
}

// Users returns the user service.
func (c *client) Users() services.UserService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.usersService == nil {
		c.usersService = services.NewUserService(c.transport)
	}

	return c.usersService
}

// Routing returns the routing service.
func (c *client) Routing() services.RoutingService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.routingService == nil {
		c.routingService = services.NewRoutingService(c.transport)
	}

	return c.routingService
}

// PortForwards returns the port forward service.
func (c *client) PortForwards() services.PortForwardService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.portForwardService == nil {
		c.portForwardService = services.NewPortForwardService(c.transport)
	}

	return c.portForwardService
}

// PortProfiles returns the port profile service.
func (c *client) PortProfiles() services.PortProfileService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.portProfileService == nil {
		c.portProfileService = services.NewPortProfileService(c.transport)
	}

	return c.portProfileService
}

// Settings returns the settings service.
func (c *client) Settings() services.SettingService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.settingService == nil {
		c.settingService = services.NewSettingService(c.transport)
	}

	return c.settingService
}

// System returns the system service.
func (c *client) System() services.SystemService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.systemService == nil {
		c.systemService = services.NewSystemService(c.transport)
	}

	return c.systemService
}

// Events returns the event service.
func (c *client) Events() services.EventService {
	return nil // Implemented in Phase 18
}

// DNS returns the DNS service.
func (c *client) DNS() services.DNSService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.dnsService == nil {
		c.dnsService = services.NewDNSService(c.transport)
	}

	return c.dnsService
}
