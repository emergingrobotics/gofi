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
	routes         map[string]*types.Route
	portForwards   map[string]*types.PortForward
	portProfiles   map[string]*types.PortProfile
	settings       map[string]*types.Setting
	radiusProfiles map[string]*types.RADIUSProfile
	dynamicDNS     *types.DynamicDNS
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
		settings:           make(map[string]*types.Setting),
		radiusProfiles:     make(map[string]*types.RADIUSProfile),
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
	s.settings = make(map[string]*types.Setting)
	s.radiusProfiles = make(map[string]*types.RADIUSProfile)
	s.dynamicDNS = nil

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

// WLAN accessors
func (s *State) GetWLAN(id string) *types.WLAN {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wlans[id]
}

func (s *State) ListWLANs() []*types.WLAN {
	s.mu.RLock()
	defer s.mu.RUnlock()
	wlans := make([]*types.WLAN, 0, len(s.wlans))
	for _, wlan := range s.wlans {
		wlans = append(wlans, wlan)
	}
	return wlans
}

func (s *State) AddWLAN(wlan *types.WLAN) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wlans[wlan.ID] = wlan
}

func (s *State) UpdateWLAN(wlan *types.WLAN) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wlans[wlan.ID] = wlan
}

func (s *State) DeleteWLAN(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.wlans, id)
}

// WLAN Group accessors
func (s *State) GetWLANGroup(id string) *types.WLANGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wlanGroups[id]
}

func (s *State) ListWLANGroups() []*types.WLANGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	groups := make([]*types.WLANGroup, 0, len(s.wlanGroups))
	for _, group := range s.wlanGroups {
		groups = append(groups, group)
	}
	return groups
}

func (s *State) AddWLANGroup(group *types.WLANGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wlanGroups[group.ID] = group
}

func (s *State) UpdateWLANGroup(group *types.WLANGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wlanGroups[group.ID] = group
}

func (s *State) DeleteWLANGroup(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.wlanGroups, id)
}

// Firewall Rule accessors
func (s *State) GetFirewallRule(id string) *types.FirewallRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.firewallRules[id]
}

func (s *State) ListFirewallRules() []*types.FirewallRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rules := make([]*types.FirewallRule, 0, len(s.firewallRules))
	for _, rule := range s.firewallRules {
		rules = append(rules, rule)
	}
	return rules
}

func (s *State) AddFirewallRule(rule *types.FirewallRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.firewallRules[rule.ID] = rule
}

func (s *State) UpdateFirewallRule(rule *types.FirewallRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.firewallRules[rule.ID] = rule
}

func (s *State) DeleteFirewallRule(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.firewallRules, id)
}

// Firewall Group accessors
func (s *State) GetFirewallGroup(id string) *types.FirewallGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.firewallGroups[id]
}

func (s *State) ListFirewallGroups() []*types.FirewallGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	groups := make([]*types.FirewallGroup, 0, len(s.firewallGroups))
	for _, group := range s.firewallGroups {
		groups = append(groups, group)
	}
	return groups
}

func (s *State) AddFirewallGroup(group *types.FirewallGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.firewallGroups[group.ID] = group
}

func (s *State) UpdateFirewallGroup(group *types.FirewallGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.firewallGroups[group.ID] = group
}

func (s *State) DeleteFirewallGroup(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.firewallGroups, id)
}

// Traffic Rule accessors
func (s *State) GetTrafficRule(id string) *types.TrafficRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.trafficRules[id]
}

func (s *State) ListTrafficRules() []*types.TrafficRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rules := make([]*types.TrafficRule, 0, len(s.trafficRules))
	for _, rule := range s.trafficRules {
		rules = append(rules, rule)
	}
	return rules
}

func (s *State) AddTrafficRule(rule *types.TrafficRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trafficRules[rule.ID] = rule
}

func (s *State) UpdateTrafficRule(rule *types.TrafficRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trafficRules[rule.ID] = rule
}

func (s *State) DeleteTrafficRule(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.trafficRules, id)
}

// Client accessors
func (s *State) GetClient(mac string) *types.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[mac]
}

func (s *State) ListClients() []*types.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	clients := make([]*types.Client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients
}

func (s *State) AddClient(client *types.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.MAC] = client
}

func (s *State) UpdateClient(client *types.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.MAC] = client
}

func (s *State) DeleteClient(mac string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, mac)
}

// User accessors (known clients, not auth users)
func (s *State) GetKnownClient(id string) *types.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[id]
}

func (s *State) GetKnownClientByMAC(mac string) *types.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if user.MAC == mac {
			return user
		}
	}
	return nil
}

func (s *State) ListKnownClients() []*types.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]*types.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

func (s *State) AddKnownClient(user *types.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
}

func (s *State) UpdateKnownClient(user *types.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
}

func (s *State) DeleteKnownClient(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
}

// UserGroup accessors
func (s *State) GetUserGroup(id string) *types.UserGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userGroups[id]
}

func (s *State) ListUserGroups() []*types.UserGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	groups := make([]*types.UserGroup, 0, len(s.userGroups))
	for _, group := range s.userGroups {
		groups = append(groups, group)
	}
	return groups
}

func (s *State) AddUserGroup(group *types.UserGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userGroups[group.ID] = group
}

func (s *State) UpdateUserGroup(group *types.UserGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userGroups[group.ID] = group
}

func (s *State) DeleteUserGroup(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.userGroups, id)
}

// Route accessors
func (s *State) GetRoute(id string) *types.Route {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.routes[id]
}

func (s *State) ListRoutes() []*types.Route {
	s.mu.RLock()
	defer s.mu.RUnlock()
	routes := make([]*types.Route, 0, len(s.routes))
	for _, route := range s.routes {
		routes = append(routes, route)
	}
	return routes
}

func (s *State) AddRoute(route *types.Route) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes[route.ID] = route
}

func (s *State) UpdateRoute(route *types.Route) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes[route.ID] = route
}

func (s *State) DeleteRoute(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.routes, id)
}

// PortForward accessors
func (s *State) GetPortForward(id string) *types.PortForward {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.portForwards[id]
}

func (s *State) ListPortForwards() []*types.PortForward {
	s.mu.RLock()
	defer s.mu.RUnlock()
	forwards := make([]*types.PortForward, 0, len(s.portForwards))
	for _, forward := range s.portForwards {
		forwards = append(forwards, forward)
	}
	return forwards
}

func (s *State) AddPortForward(forward *types.PortForward) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portForwards[forward.ID] = forward
}

func (s *State) UpdatePortForward(forward *types.PortForward) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portForwards[forward.ID] = forward
}

func (s *State) DeletePortForward(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.portForwards, id)
}

// PortProfile accessors
func (s *State) GetPortProfile(id string) *types.PortProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.portProfiles[id]
}

func (s *State) ListPortProfiles() []*types.PortProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	profiles := make([]*types.PortProfile, 0, len(s.portProfiles))
	for _, profile := range s.portProfiles {
		profiles = append(profiles, profile)
	}
	return profiles
}

func (s *State) AddPortProfile(profile *types.PortProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portProfiles[profile.ID] = profile
}

func (s *State) UpdatePortProfile(profile *types.PortProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portProfiles[profile.ID] = profile
}

func (s *State) DeletePortProfile(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.portProfiles, id)
}

// Setting accessors
func (s *State) GetSetting(key string) *types.Setting {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings[key]
}

func (s *State) ListSettings() []*types.Setting {
	s.mu.RLock()
	defer s.mu.RUnlock()
	settings := make([]*types.Setting, 0, len(s.settings))
	for _, setting := range s.settings {
		settings = append(settings, setting)
	}
	return settings
}

func (s *State) AddSetting(setting *types.Setting) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings[setting.Key] = setting
}

func (s *State) UpdateSetting(setting *types.Setting) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings[setting.Key] = setting
}

func (s *State) DeleteSetting(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.settings, key)
}

// RADIUSProfile accessors
func (s *State) GetRADIUSProfile(id string) *types.RADIUSProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.radiusProfiles[id]
}

func (s *State) ListRADIUSProfiles() []*types.RADIUSProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	profiles := make([]*types.RADIUSProfile, 0, len(s.radiusProfiles))
	for _, profile := range s.radiusProfiles {
		profiles = append(profiles, profile)
	}
	return profiles
}

func (s *State) AddRADIUSProfile(profile *types.RADIUSProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.radiusProfiles[profile.ID] = profile
}

func (s *State) UpdateRADIUSProfile(profile *types.RADIUSProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.radiusProfiles[profile.ID] = profile
}

func (s *State) DeleteRADIUSProfile(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.radiusProfiles, id)
}

// DynamicDNS accessors
func (s *State) GetDynamicDNS() *types.DynamicDNS {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dynamicDNS
}

func (s *State) SetDynamicDNS(ddns *types.DynamicDNS) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dynamicDNS = ddns
}
