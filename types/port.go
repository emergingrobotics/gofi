package types

// PortForward represents a port forwarding rule.
type PortForward struct {
	ID              string `json:"_id,omitempty"`
	SiteID          string `json:"site_id,omitempty"`
	Name            string `json:"name"`
	Enabled         bool   `json:"enabled"`
	Protocol        string `json:"proto"` // "tcp", "udp", "tcp_udp"
	SrcNetworkID    string `json:"src,omitempty"` // "wan" or network ID
	DstPort         string `json:"dst_port"`
	FwdIP           string `json:"fwd"` // Forward to IP
	FwdPort         string `json:"fwd_port"`
	LogForward      bool   `json:"log,omitempty"`
	PfRule          string `json:"pfrule,omitempty"`
}

// PortProfile represents a switch port profile.
type PortProfile struct {
	ID                      string   `json:"_id,omitempty"`
	SiteID                  string   `json:"site_id,omitempty"`
	Name                    string   `json:"name"`
	Forward                 string   `json:"forward,omitempty"` // "all", "native", "customize"
	NativeNetworkConfID     string   `json:"native_networkconf_id,omitempty"`
	TaggedNetworkConfIDs    []string `json:"tagged_networkconf_ids,omitempty"`
	POEMode                 string   `json:"poe_mode,omitempty"` // "auto", "passthrough", "off"
	STormCtrlBroadcastEnabled bool   `json:"stormctrl_bcast_enabled,omitempty"`
	STormCtrlMcastEnabled   bool     `json:"stormctrl_mcast_enabled,omitempty"`
	STormCtrlUcastEnabled   bool     `json:"stormctrl_ucast_enabled,omitempty"`
	STormCtrlBroadcastLevel int      `json:"stormctrl_bcast_level,omitempty"`
	STormCtrlMcastLevel     int      `json:"stormctrl_mcast_level,omitempty"`
	STormCtrlUcastLevel     int      `json:"stormctrl_ucast_level,omitempty"`
	STormCtrlType           string   `json:"stormctrl_type,omitempty"` // "level", "rate"
	LLDPMedEnabled          bool     `json:"lldpmed_enabled,omitempty"`
	LLDPMedNotifyEnabled    bool     `json:"lldpmed_notify_enabled,omitempty"`
	SpeedDuplex             int      `json:"speed,omitempty"`
	FullDuplex              bool     `json:"full_duplex,omitempty"`
	Dot1xCtrl               string   `json:"dot1x_ctrl,omitempty"` // "auto", "force_authorized", "force_unauthorized", "mac_based", "multi_host"
	Dot1xIdleTimeout        int      `json:"dot1x_idle_timeout,omitempty"`
	IsolationEnabled        bool     `json:"isolation,omitempty"`
	OpMode                  string   `json:"op_mode,omitempty"` // "switch", "mirror", "aggregate"
	AggregateNumPorts       int      `json:"aggregate_num_ports,omitempty"`
	ExcludedNetworkConfIDs  []string `json:"excluded_networkconf_ids,omitempty"`
	VoiceNetworkConfID      string   `json:"voice_networkconf_id,omitempty"`
}

// Protocol constants for port forwarding.
const (
	ProtocolTCPUDP = "tcp_udp"
)
