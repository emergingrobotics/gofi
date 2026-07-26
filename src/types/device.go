package types

// Device represents a UniFi network device (AP, Switch, Gateway, etc.).
type Device struct {
	ID              string `json:"_id"`
	MAC             string `json:"mac"`
	Model           string `json:"model"`
	ModelInLTS      bool   `json:"model_in_lts"`
	ModelInEOL      bool   `json:"model_in_eol"`
	Type            string `json:"type"` // "uap", "usw", "ugw", "udm", etc.
	Name            string `json:"name"`
	Serial          string `json:"serial"`
	Version         string `json:"version"`
	Adopted         bool   `json:"adopted"`
	SiteID          string `json:"site_id"`
	State           DeviceState `json:"state"`
	InformURL       string `json:"inform_url,omitempty"`
	InformIP        string `json:"inform_ip,omitempty"`
	LastSeen        int64  `json:"last_seen"`
	Uptime          FlexInt `json:"uptime"`
	Upgradable      bool   `json:"upgradable"`
	ConfigVersion   string `json:"cfgversion,omitempty"`
	LicenseState    string `json:"license_state,omitempty"`
	ConnectedAt     int64  `json:"connected_at,omitempty"`
	ProvisionedAt   int64  `json:"provisioned_at,omitempty"`
	LEDOverride     string `json:"led_override,omitempty"`
	LEDOverrideColor string `json:"led_override_color,omitempty"`
	LEDOverrideColorBrightness int `json:"led_override_color_brightness,omitempty"`
	Internet        bool   `json:"internet,omitempty"`
	IP              string `json:"ip,omitempty"`

	// Statistics
	SystemStats     *SystemStats `json:"system-stats,omitempty"`
	SysStats        *SysStats    `json:"sys_stats,omitempty"`
	StatBytes       FlexInt      `json:"stat_bytes,omitempty"`
	RxBytes         FlexInt      `json:"rx_bytes,omitempty"`
	TxBytes         FlexInt      `json:"tx_bytes,omitempty"`
	BytesR          FlexInt      `json:"bytes-r,omitempty"`

	// Uplink information
	Uplink          *DeviceUplink `json:"uplink,omitempty"`
	UplinkTable     []DeviceUplink `json:"uplink_table,omitempty"`

	// Network configuration
	ConfigNetwork   *DeviceConfigNetwork `json:"config_network,omitempty"`
	NetworkTable    []NetworkTable `json:"network_table,omitempty"`

	// AP-specific fields
	RadioTable      []RadioTable `json:"radio_table,omitempty"`
	RadioTableStats []RadioTableStats `json:"radio_table_stats,omitempty"`
	VAPTable        []VAPTable `json:"vap_table,omitempty"`
	NumSTA          int        `json:"num_sta,omitempty"`
	UserNumSTA      int        `json:"user-num_sta,omitempty"`
	GuestNumSTA     int        `json:"guest-num_sta,omitempty"`

	// Switch-specific fields
	PortTable       []PortTable `json:"port_table,omitempty"`
	PortOverrides   []PortOverride `json:"port_overrides,omitempty"`
	TotalMaxPower   int        `json:"total_max_power,omitempty"`

	// Gateway-specific fields
	WANType         string `json:"wan_type,omitempty"`
	Wan1            *WAN   `json:"wan1,omitempty"`
	Wan2            *WAN   `json:"wan2,omitempty"`
	SpeedtestStatus string `json:"speedtest_status,omitempty"`
	SpeedtestStatusSaved bool `json:"speedtest-status-saved,omitempty"`
	SpeedtestPing   FlexInt `json:"speedtest_ping,omitempty"`

	// Temperature monitoring
	Temperatures    []Temperature `json:"temperatures,omitempty"`

	// Storage (for UDM)
	Storage         []Storage `json:"storage,omitempty"`

	// Miscellaneous
	DisplayableVersion string `json:"displayable_version,omitempty"`
	RequiredVersion    string `json:"required_version,omitempty"`
	Satisfaction    int    `json:"satisfaction,omitempty"`
	Isolated        bool   `json:"isolated,omitempty"`
	KernelVersion   string `json:"kernel_version,omitempty"`
	Architecture    string `json:"architecture,omitempty"`
	HashID          string `json:"hash_id,omitempty"`
}

// DeviceBasic represents minimal device information for faster queries.
type DeviceBasic struct {
	MAC   string `json:"mac"`
	Type  string `json:"type"`
	Model string `json:"model"`
	Name  string `json:"name,omitempty"`
	State DeviceState `json:"state,omitempty"`
}

// DeviceUplink represents uplink connection information for a device.
type DeviceUplink struct {
	FullDuplex       bool   `json:"full_duplex"`
	IP               string `json:"ip"`
	MAC              string `json:"mac"`
	Name             string `json:"name"`
	Netmask          string `json:"netmask"`
	NumPort          int    `json:"num_port"`
	RxBytes          FlexInt `json:"rx_bytes"`
	TxBytes          FlexInt `json:"tx_bytes"`
	RxBytesR         FlexInt `json:"rx_bytes-r,omitempty"`
	TxBytesR         FlexInt `json:"tx_bytes-r,omitempty"`
	RxPackets        FlexInt `json:"rx_packets,omitempty"`
	TxPackets        FlexInt `json:"tx_packets,omitempty"`
	Speed            int    `json:"speed"`
	MaxSpeed         int    `json:"max_speed,omitempty"`
	Type             string `json:"type"`
	Up               bool   `json:"up"`
	UplinkMAC        string `json:"uplink_mac"`
	UplinkRemotePort int    `json:"uplink_remote_port"`
	Media            string `json:"media,omitempty"`
	PortIdx          int    `json:"port_idx,omitempty"`
}

// DeviceConfigNetwork represents network configuration for a device.
type DeviceConfigNetwork struct {
	Type           string `json:"type"`
	IP             string `json:"ip"`
	Netmask        string `json:"netmask,omitempty"`
	Gateway        string `json:"gateway,omitempty"`
	DNS1           string `json:"dns1,omitempty"`
	DNS2           string `json:"dns2,omitempty"`
	BondingEnabled bool   `json:"bonding_enabled"`
}

// NetworkTable represents a network interface on a device.
type NetworkTable struct {
	ID       string `json:"_id,omitempty"`
	Name     string `json:"name"`
	MAC      string `json:"mac"`
	IP       string `json:"ip,omitempty"`
	Netmask  string `json:"netmask,omitempty"`
	NumSTA   int    `json:"num_sta,omitempty"`
	Up       bool   `json:"up"`
}

// SystemStats represents system statistics for a device.
type SystemStats struct {
	CPU    FlexInt `json:"cpu,omitempty"`
	Mem    FlexInt `json:"mem,omitempty"`
	Uptime FlexInt `json:"uptime,omitempty"`
}

// SysStats represents detailed system statistics.
type SysStats struct {
	Loadavg1  FlexInt `json:"loadavg_1,omitempty"`
	Loadavg5  FlexInt `json:"loadavg_5,omitempty"`
	Loadavg15 FlexInt `json:"loadavg_15,omitempty"`
	MemUsed   FlexInt `json:"mem_used,omitempty"`
	MemTotal  FlexInt `json:"mem_total,omitempty"`
	MemBuffer FlexInt `json:"mem_buffer,omitempty"`
}

// RadioTable represents a radio (2.4GHz, 5GHz, 6GHz) on an AP.
type RadioTable struct {
	Radio             string  `json:"radio"`
	Name              string  `json:"name"`
	BuiltInAntenna    bool    `json:"builtin_antenna"`
	BuiltInAntennaGain int    `json:"builtin_ant_gain,omitempty"`
	MaxTXPower        int     `json:"max_txpower,omitempty"`
	MinTXPower        int     `json:"min_txpower,omitempty"`
	Nss               int     `json:"nss,omitempty"`
	RadioCaps         int     `json:"radio_caps,omitempty"`
	HasDFS            bool    `json:"has_dfs,omitempty"`
	HasFCCDFS         bool    `json:"has_fccdfs,omitempty"`
	CurrentAntenna    int     `json:"current_antenna_gain,omitempty"`
	SensLevelEnabled  bool    `json:"sens_level_enabled,omitempty"`
}

// RadioTableStats represents statistics for a radio.
type RadioTableStats struct {
	Radio            string  `json:"radio"`
	Name             string  `json:"name"`
	Channel          int     `json:"channel,omitempty"`
	TXPower          int     `json:"tx_power,omitempty"`
	TXPackets        FlexInt `json:"tx_packets,omitempty"`
	RXPackets        FlexInt `json:"rx_packets,omitempty"`
	NumSTA           int     `json:"num_sta,omitempty"`
	Satisfaction     int     `json:"satisfaction,omitempty"`
	State            string  `json:"state,omitempty"`
	ExtChannel       int     `json:"extchannel,omitempty"`
	CuTotal          int     `json:"cu_total,omitempty"`
	CuSelfRX         int     `json:"cu_self_rx,omitempty"`
	CuSelfTX         int     `json:"cu_self_tx,omitempty"`
	GuestNumSTA      int     `json:"guest-num_sta,omitempty"`
	UserNumSTA       int     `json:"user-num_sta,omitempty"`
}

// VAPTable represents a Virtual AP (SSID) on a radio.
type VAPTable struct {
	ID              string  `json:"_id,omitempty"`
	BSSID           string  `json:"bssid"`
	CCQSAP          int     `json:"ccq,omitempty"`
	Channel         int     `json:"channel,omitempty"`
	Essid           string  `json:"essid,omitempty"`
	ExtChannel      int     `json:"extchannel,omitempty"`
	MapID           string  `json:"map_id,omitempty"`
	Name            string  `json:"name"`
	NumSTA          int     `json:"num_sta,omitempty"`
	Radio           string  `json:"radio"`
	RadioName       string  `json:"radio_name,omitempty"`
	RXBytes         FlexInt `json:"rx_bytes,omitempty"`
	RXCrypts        FlexInt `json:"rx_crypts,omitempty"`
	RXDropped       FlexInt `json:"rx_dropped,omitempty"`
	RXErrors        FlexInt `json:"rx_errors,omitempty"`
	RXFrags         FlexInt `json:"rx_frags,omitempty"`
	RXNWIDs         FlexInt `json:"rx_nwids,omitempty"`
	RXPackets       FlexInt `json:"rx_packets,omitempty"`
	State           string  `json:"state,omitempty"`
	TXBytes         FlexInt `json:"tx_bytes,omitempty"`
	TXDropped       FlexInt `json:"tx_dropped,omitempty"`
	TXErrors        FlexInt `json:"tx_errors,omitempty"`
	TXPackets       FlexInt `json:"tx_packets,omitempty"`
	TXPower         int     `json:"tx_power,omitempty"`
	TXRetries       FlexInt `json:"tx_retries,omitempty"`
	Up              bool    `json:"up"`
	Usage           string  `json:"usage,omitempty"`
	WlanconfID      string  `json:"wlanconf_id,omitempty"`
	IsGuest         bool    `json:"is_guest,omitempty"`
	ApMAC           string  `json:"ap_mac,omitempty"`
	SiteID          string  `json:"site_id,omitempty"`
}

// PortTable represents a network port on a switch or AP.
type PortTable struct {
	PortIdx              int     `json:"port_idx"`
	MediaType            string  `json:"media,omitempty"`
	PortPoe              bool    `json:"port_poe,omitempty"`
	PoeCaps              int     `json:"poe_caps,omitempty"`
	PoeClass             string  `json:"poe_class,omitempty"`
	PoeEnable            bool    `json:"poe_enable,omitempty"`
	PoeCurrent           FlexInt `json:"poe_current,omitempty"`
	PoeGood              bool    `json:"poe_good,omitempty"`
	PoeMode              string  `json:"poe_mode,omitempty"`
	PoePower             FlexInt `json:"poe_power,omitempty"`
	PoeVoltage           FlexInt `json:"poe_voltage,omitempty"`
	PortconfID           string  `json:"portconf_id,omitempty"`
	AggregatedBy         bool    `json:"aggregated_by,omitempty"`
	Autoneg              bool    `json:"autoneg,omitempty"`
	BytesR               FlexInt `json:"bytes-r,omitempty"`
	Dot1xMode            string  `json:"dot1x_mode,omitempty"`
	Dot1xStatus          string  `json:"dot1x_status,omitempty"`
	Enable               bool    `json:"enable"`
	Flowctrl             bool    `json:"flowctrl_rx,omitempty"`
	FullDuplex           bool    `json:"full_duplex"`
	IsUplink             bool    `json:"is_uplink,omitempty"`
	Jumbo                bool    `json:"jumbo,omitempty"`
	MAC                  string  `json:"mac,omitempty"`
	Masked               bool    `json:"masked,omitempty"`
	Name                 string  `json:"name,omitempty"`
	NetworkName          string  `json:"network_name,omitempty"`
	OpMode               string  `json:"op_mode,omitempty"`
	PortDelta            *PortDelta `json:"port_delta,omitempty"`
	RXBroadcast          FlexInt `json:"rx_broadcast,omitempty"`
	RXBytes              FlexInt `json:"rx_bytes,omitempty"`
	RXBytesR             FlexInt `json:"rx_bytes-r,omitempty"`
	RXDropped            FlexInt `json:"rx_dropped,omitempty"`
	RXErrors             FlexInt `json:"rx_errors,omitempty"`
	RXMulticast          FlexInt `json:"rx_multicast,omitempty"`
	RXPackets            FlexInt `json:"rx_packets,omitempty"`
	SFPCompliance        string  `json:"sfp_compliance,omitempty"`
	SFPCurrent           FlexInt `json:"sfp_current,omitempty"`
	SFPFound             bool    `json:"sfp_found,omitempty"`
	SFPPart              string  `json:"sfp_part,omitempty"`
	SFPRev               string  `json:"sfp_rev,omitempty"`
	SFPRXPower           FlexInt `json:"sfp_rxpower,omitempty"`
	SFPSerial            string  `json:"sfp_serial,omitempty"`
	SFPTemperature       FlexInt `json:"sfp_temperature,omitempty"`
	SFPTXPower           FlexInt `json:"sfp_txpower,omitempty"`
	SFPVendor            string  `json:"sfp_vendor,omitempty"`
	SFPVoltage           FlexInt `json:"sfp_voltage,omitempty"`
	Speed                int     `json:"speed"`
	SpeedCaps            int     `json:"speed_caps,omitempty"`
	StpPathCost          int     `json:"stp_pathcost,omitempty"`
	StpState             string  `json:"stp_state,omitempty"`
	TXBroadcast          FlexInt `json:"tx_broadcast,omitempty"`
	TXBytes              FlexInt `json:"tx_bytes,omitempty"`
	TXBytesR             FlexInt `json:"tx_bytes-r,omitempty"`
	TXDropped            FlexInt `json:"tx_dropped,omitempty"`
	TXErrors             FlexInt `json:"tx_errors,omitempty"`
	TXMulticast          FlexInt `json:"tx_multicast,omitempty"`
	TXPackets            FlexInt `json:"tx_packets,omitempty"`
	Type                 string  `json:"type,omitempty"`
	Up                   bool    `json:"up"`
}

// PortDelta represents port state changes.
type PortDelta struct {
	TimeMS int64 `json:"time_ms,omitempty"`
}

// PortOverride represents port configuration overrides.
type PortOverride struct {
	PortIdx            int    `json:"port_idx"`
	PortconfID         string `json:"portconf_id,omitempty"`
	PoeMode            string `json:"poe_mode,omitempty"`
	Name               string `json:"name,omitempty"`
	AggregateNumPorts  int    `json:"aggregate_num_ports,omitempty"`
}

// Temperature represents temperature sensor data.
type Temperature struct {
	Name  string  `json:"name"`
	Type  string  `json:"type,omitempty"`
	Value FlexInt `json:"value"`
}

// Storage represents storage device information.
type Storage struct {
	MountPoint string  `json:"mount_point"`
	Name       string  `json:"name"`
	Size       FlexInt `json:"size"`
	Type       string  `json:"type"`
	Used       FlexInt `json:"used"`
}

// WAN represents WAN interface information.
type WAN struct {
	BytesR      FlexInt `json:"bytes-r,omitempty"`
	DNS         []string `json:"dns,omitempty"`
	Enable      bool    `json:"enable,omitempty"`
	FullDuplex  bool    `json:"full_duplex,omitempty"`
	Gateway     string  `json:"gateway,omitempty"`
	IFNAME      string  `json:"ifname,omitempty"`
	IP          string  `json:"ip,omitempty"`
	Latency     int     `json:"latency,omitempty"`
	MAC         string  `json:"mac,omitempty"`
	MaxSpeed    int     `json:"max_speed,omitempty"`
	Media       string  `json:"media,omitempty"`
	Name        string  `json:"name,omitempty"`
	Netmask     string  `json:"netmask,omitempty"`
	NetworkGroup string `json:"networkgroup,omitempty"`
	NumPort     int     `json:"num_port,omitempty"`
	RXBytes     FlexInt `json:"rx_bytes,omitempty"`
	RXBytesR    FlexInt `json:"rx_bytes-r,omitempty"`
	RXDropped   FlexInt `json:"rx_dropped,omitempty"`
	RXErrors    FlexInt `json:"rx_errors,omitempty"`
	RXMulticast FlexInt `json:"rx_multicast,omitempty"`
	RXPackets   FlexInt `json:"rx_packets,omitempty"`
	Speed       int     `json:"speed,omitempty"`
	TXBytes     FlexInt `json:"tx_bytes,omitempty"`
	TXBytesR    FlexInt `json:"tx_bytes-r,omitempty"`
	TXDropped   FlexInt `json:"tx_dropped,omitempty"`
	TXErrors    FlexInt `json:"tx_errors,omitempty"`
	TXPackets   FlexInt `json:"tx_packets,omitempty"`
	Type        string  `json:"type,omitempty"`
	Up          bool    `json:"up,omitempty"`
	Uptime      FlexInt `json:"uptime,omitempty"`
	XputDown    FlexInt `json:"xput_down,omitempty"`
	XputUp      FlexInt `json:"xput_up,omitempty"`
}
