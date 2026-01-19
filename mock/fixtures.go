package mock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/unifi-go/gofi/types"
)

// Fixtures holds test fixture data.
type Fixtures struct {
	Sites        []types.Site        `json:"sites,omitempty"`
	Devices      []types.Device      `json:"devices,omitempty"`
	Networks     []types.Network     `json:"networks,omitempty"`
	WLANs        []types.WLAN        `json:"wlans,omitempty"`
	Clients      []types.Client      `json:"clients,omitempty"`
	Users        []types.User        `json:"users,omitempty"`
	FirewallRules []types.FirewallRule `json:"firewall_rules,omitempty"`
}

// DefaultFixtures returns a minimal set of fixtures.
func DefaultFixtures() *Fixtures {
	return &Fixtures{
		Sites: []types.Site{
			{
				ID:   "default",
				Name: "default",
				Desc: "Default Site",
			},
		},
	}
}

// LoadFixtures loads fixtures from a directory.
// The directory should contain JSON files named: sites.json, devices.json, etc.
func LoadFixtures(dir string) (*Fixtures, error) {
	fixtures := &Fixtures{}

	files := map[string]interface{}{
		"sites.json":    &fixtures.Sites,
		"devices.json":  &fixtures.Devices,
		"networks.json": &fixtures.Networks,
		"wlans.json":    &fixtures.WLANs,
		"clients.json":  &fixtures.Clients,
		"users.json":    &fixtures.Users,
	}

	for filename, target := range files {
		path := filepath.Join(dir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Optional file
			}
			return nil, fmt.Errorf("failed to read %s: %w", filename, err)
		}

		if err := json.Unmarshal(data, target); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filename, err)
		}
	}

	return fixtures, nil
}

// LoadFixtures loads fixtures into the state.
func (s *State) LoadFixtures(fixtures *Fixtures) {
	if fixtures == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Load sites
	for _, site := range fixtures.Sites {
		siteCopy := site
		s.sites[site.ID] = &siteCopy
	}

	// Load devices
	for _, device := range fixtures.Devices {
		deviceCopy := device
		s.devices[device.ID] = &deviceCopy
	}

	// Load networks
	for _, network := range fixtures.Networks {
		networkCopy := network
		s.networks[network.ID] = &networkCopy
	}

	// Load WLANs
	for _, wlan := range fixtures.WLANs {
		wlanCopy := wlan
		s.wlans[wlan.ID] = &wlanCopy
	}

	// Load clients
	for _, client := range fixtures.Clients {
		clientCopy := client
		s.clients[client.MAC] = &clientCopy
	}

	// Load users
	for _, user := range fixtures.Users {
		userCopy := user
		s.users[user.ID] = &userCopy
	}

	// Load firewall rules
	for _, rule := range fixtures.FirewallRules {
		ruleCopy := rule
		s.firewallRules[rule.ID] = &ruleCopy
	}
}
