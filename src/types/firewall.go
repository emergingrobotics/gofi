package types

// FirewallRule represents a UniFi firewall rule.
type FirewallRule struct {
	ID                    string   `json:"_id,omitempty"`
	SiteID                string   `json:"site_id,omitempty"`
	Name                  string   `json:"name"`
	Enabled               bool     `json:"enabled"`
	Ruleset               string   `json:"ruleset"` // WAN_IN, LAN_IN, etc.
	RuleIndex             int      `json:"rule_index"`
	Action                string   `json:"action"` // "accept", "drop", "reject"
	Protocol              string   `json:"protocol"` // "all", "tcp", "udp", "icmp"
	ProtocolMatchExcepted bool     `json:"protocol_match_excepted"`
	Logging               bool     `json:"logging"`

	// Connection states
	StateNew          bool     `json:"state_new"`
	StateEstablished  bool     `json:"state_established"`
	StateInvalid      bool     `json:"state_invalid"`
	StateRelated      bool     `json:"state_related"`

	// Source and destination
	SrcFirewallGroupIDs []string `json:"src_firewallgroup_ids,omitempty"`
	DstFirewallGroupIDs []string `json:"dst_firewallgroup_ids,omitempty"`
	SrcMACAddress       string   `json:"src_mac_address,omitempty"`
	SrcAddress          string   `json:"src_address,omitempty"`
	SrcNetworkConfID    string   `json:"src_networkconf_id,omitempty"`
	DstAddress          string   `json:"dst_address,omitempty"`
	DstNetworkConfID    string   `json:"dst_networkconf_id,omitempty"`

	// Ports
	SrcPort     string   `json:"src_port,omitempty"`
	DstPort     string   `json:"dst_port,omitempty"`

	// ICMP
	ICMPTypename string   `json:"icmp_typename,omitempty"`

	// IPSec
	IPSecMatchIPSec      string   `json:"ipsec_match_ipsec,omitempty"`
}

// FirewallGroup represents a firewall group (address group, port group, etc.).
type FirewallGroup struct {
	ID           string   `json:"_id,omitempty"`
	SiteID       string   `json:"site_id,omitempty"`
	Name         string   `json:"name"`
	GroupType    string   `json:"group_type"` // "address-group", "port-group", "ipv6-address-group"
	GroupMembers []string `json:"group_members,omitempty"`
}

// FirewallRuleIndexUpdate is used for reordering firewall rules.
type FirewallRuleIndexUpdate struct {
	ID        string `json:"_id"`
	RuleIndex int    `json:"rule_index"`
}

// Ruleset constants.
const (
	RulesetWANIn     = "WAN_IN"
	RulesetWANOut    = "WAN_OUT"
	RulesetWANLocal  = "WAN_LOCAL"
	RulesetLANIn     = "LAN_IN"
	RulesetLANOut    = "LAN_OUT"
	RulesetLANLocal  = "LAN_LOCAL"
	RulesetGuestIn   = "GUEST_IN"
	RulesetGuestOut  = "GUEST_OUT"
	RulesetGuestLocal = "GUEST_LOCAL"
)

// Action constants.
const (
	FirewallActionAccept = "accept"
	FirewallActionDrop   = "drop"
	FirewallActionReject = "reject"
)

// Protocol constants.
const (
	ProtocolAll  = "all"
	ProtocolTCP  = "tcp"
	ProtocolUDP  = "udp"
	ProtocolICMP = "icmp"
	ProtocolIPv6ICMP = "ipv6-icmp"
)

// Group type constants.
const (
	GroupTypeAddress     = "address-group"
	GroupTypePort        = "port-group"
	GroupTypeIPv6Address = "ipv6-address-group"
)
