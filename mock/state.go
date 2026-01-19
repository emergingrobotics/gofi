package mock

import (
	"sync"

	"github.com/unifi-go/gofi/types"
)

// State holds the mock server's in-memory state.
type State struct {
	mu sync.RWMutex

	// Authentication state
	authenticatedUsers map[string]string // username -> password
	sessions           map[string]*Session

	// Data stores
	sites        map[string]*types.Site
	devices      map[string]*types.Device
	networks     map[string]*types.Network
	wlans        map[string]*types.WLAN
	wlanGroups   map[string]*types.WLANGroup
	firewallRules map[string]*types.FirewallRule
	firewallGroups map[string]*types.FirewallGroup
	trafficRules map[string]*types.TrafficRule
	clients      map[string]*types.Client
	users        map[string]*types.User
	userGroups   map[string]*types.UserGroup
	routes       map[string]*types.Route
	portForwards map[string]*types.PortForward
	portProfiles map[string]*types.PortProfile
}

// Session represents a mock authentication session.
type Session struct {
	Username  string
	CSRFToken string
}

// NewState creates a new mock state.
func NewState() *State {
	s := &State{
		authenticatedUsers: make(map[string]string),
		sessions:           make(map[string]*Session),
		sites:              make(map[string]*types.Site),
		devices:            make(map[string]*types.Device),
		networks:           make(map[string]*types.Network),
		wlans:              make(map[string]*types.WLAN),
		wlanGroups:         make(map[string]*types.WLANGroup),
		firewallRules:      make(map[string]*types.FirewallRule),
		firewallGroups:     make(map[string]*types.FirewallGroup),
		trafficRules:       make(map[string]*types.TrafficRule),
		clients:            make(map[string]*types.Client),
		users:              make(map[string]*types.User),
		userGroups:         make(map[string]*types.UserGroup),
		routes:             make(map[string]*types.Route),
		portForwards:       make(map[string]*types.PortForward),
		portProfiles:       make(map[string]*types.PortProfile),
	}

	// Add default admin user
	s.authenticatedUsers["admin"] = "admin"

	// Add default site
	s.sites["default"] = &types.Site{
		ID:   "default",
		Name: "default",
		Desc: "Default Site",
	}

	return s
}

// Reset clears all state.
func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions = make(map[string]*Session)
	s.sites = make(map[string]*types.Site)
	s.devices = make(map[string]*types.Device)
	s.networks = make(map[string]*types.Network)
	s.wlans = make(map[string]*types.WLAN)
	s.wlanGroups = make(map[string]*types.WLANGroup)
	s.firewallRules = make(map[string]*types.FirewallRule)
	s.firewallGroups = make(map[string]*types.FirewallGroup)
	s.trafficRules = make(map[string]*types.TrafficRule)
	s.clients = make(map[string]*types.Client)
	s.users = make(map[string]*types.User)
	s.userGroups = make(map[string]*types.UserGroup)
	s.routes = make(map[string]*types.Route)
	s.portForwards = make(map[string]*types.PortForward)
	s.portProfiles = make(map[string]*types.PortProfile)

	// Re-add default site
	s.sites["default"] = &types.Site{
		ID:   "default",
		Name: "default",
		Desc: "Default Site",
	}
}

// AddUser adds a user for authentication.
func (s *State) AddUser(username, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authenticatedUsers[username] = password
}

// ValidateCredentials checks if credentials are valid.
func (s *State) ValidateCredentials(username, password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	storedPassword, exists := s.authenticatedUsers[username]
	return exists && storedPassword == password
}

// CreateSession creates a new session.
func (s *State) CreateSession(token string, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = session
}

// GetSession retrieves a session.
func (s *State) GetSession(token string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[token]
	return session, exists
}

// DeleteSession removes a session.
func (s *State) DeleteSession(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

// Site accessors
func (s *State) GetSite(id string) (*types.Site, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	site, exists := s.sites[id]
	return site, exists
}

func (s *State) ListSites() []*types.Site {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sites := make([]*types.Site, 0, len(s.sites))
	for _, site := range s.sites {
		sites = append(sites, site)
	}
	return sites
}

func (s *State) AddSite(site *types.Site) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sites[site.ID] = site
}

func (s *State) DeleteSite(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sites, id)
}

// Device accessors
func (s *State) GetDevice(id string) (*types.Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	device, exists := s.devices[id]
	return device, exists
}

func (s *State) ListDevices() []*types.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	devices := make([]*types.Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	return devices
}

func (s *State) AddDevice(device *types.Device) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[device.ID] = device
}

func (s *State) DeleteDevice(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.devices, id)
}

// Network accessors
func (s *State) GetNetwork(id string) (*types.Network, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	network, exists := s.networks[id]
	return network, exists
}

func (s *State) ListNetworks() []*types.Network {
	s.mu.RLock()
	defer s.mu.RUnlock()
	networks := make([]*types.Network, 0, len(s.networks))
	for _, network := range s.networks {
		networks = append(networks, network)
	}
	return networks
}

func (s *State) AddNetwork(network *types.Network) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.networks[network.ID] = network
}

func (s *State) DeleteNetwork(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.networks, id)
}
