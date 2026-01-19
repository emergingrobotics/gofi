package types

// Network represents a UniFi network configuration (VLAN, subnet, DHCP, etc.).
type Network struct {
	ID              string `json:"_id,omitempty"`
	SiteID          string `json:"site_id,omitempty"`
	Name            string `json:"name"`
	Purpose         string `json:"purpose"` // "corporate", "guest", "wan", "vpn", "vlan-only"
	VLANEnabled     bool   `json:"vlan_enabled"`
	VLAN            int    `json:"vlan,omitempty"`
	IPSubnet        string `json:"ip_subnet"`

	// DHCP Server Configuration
	DHCPDEnabled        bool   `json:"dhcpd_enabled"`
	DHCPDStart          string `json:"dhcpd_start,omitempty"`
	DHCPDStop           string `json:"dhcpd_stop,omitempty"`
	DHCPDLeaseTime      int    `json:"dhcpd_leasetime,omitempty"`
	DHCPDDNSEnabled     bool   `json:"dhcpd_dns_enabled"`
	DHCPDDNS1           string `json:"dhcpd_dns_1,omitempty"`
	DHCPDDNS2           string `json:"dhcpd_dns_2,omitempty"`
	DHCPDDNS3           string `json:"dhcpd_dns_3,omitempty"`
	DHCPDDNS4           string `json:"dhcpd_dns_4,omitempty"`
	DHCPDGatewayEnabled bool   `json:"dhcpd_gateway_enabled"`
	DHCPDGateway        string `json:"dhcpd_gateway,omitempty"`
	DHCPDBootEnabled    bool   `json:"dhcpd_boot_enabled,omitempty"`
	DHCPDBootFilename   string `json:"dhcpd_boot_filename,omitempty"`
	DHCPDBootServer     string `json:"dhcpd_boot_server,omitempty"`
	DHCPDNTPEnabled     bool   `json:"dhcpd_ntp_enabled,omitempty"`
	DHCPDNTPServer1     string `json:"dhcpd_ntp_1,omitempty"`
	DHCPDNTPServer2     string `json:"dhcpd_ntp_2,omitempty"`
	DHCPDTFTPServer     string `json:"dhcpd_tftp_server,omitempty"`
	DHCPDWINSEnabled    bool   `json:"dhcpd_winsserver_enabled,omitempty"`
	DHCPDWINSServer1    string `json:"dhcpd_winsserver_1,omitempty"`
	DHCPDWINSServer2    string `json:"dhcpd_winsserver_2,omitempty"`
	DHCPRelayEnabled    bool   `json:"dhcp_relay_enabled,omitempty"`

	// Network Settings
	DomainName          string `json:"domain_name,omitempty"`
	Enabled             bool   `json:"enabled"`
	IsNAT               bool   `json:"is_nat"`
	NetworkGroup        string `json:"networkgroup"` // "LAN", "WAN", etc.
	IGMPSnooping        bool   `json:"igmp_snooping,omitempty"`
	MulticastDNS        bool   `json:"mdns_enabled,omitempty"`
	DHCPGuardEnabled    bool   `json:"dhcpguard_enabled"`
	ARPInspection       bool   `json:"arp_inspection,omitempty"`

	// IPv6
	IPv6InterfaceType   string `json:"ipv6_interface_type,omitempty"`
	IPv6PDStart         string `json:"ipv6_pd_start,omitempty"`
	IPv6PDStop          string `json:"ipv6_pd_stop,omitempty"`
	IPv6RAEnabled       bool   `json:"ipv6_ra_enabled,omitempty"`
	IPv6RAPriorityLife  FlexInt `json:"ipv6_ra_priority,omitempty"`
	IPv6RAValidLifetime FlexInt `json:"ipv6_ra_valid_lifetime,omitempty"`
	IPv6RAPreferredLife FlexInt `json:"ipv6_ra_preferred_lifetime,omitempty"`

	// WAN Settings (for WAN-type networks)
	WANType             string   `json:"wan_type,omitempty"` // "dhcp", "static", "pppoe"
	WANEgressQOS        int      `json:"wan_egress_qos,omitempty"`
	WANLoadBalanceType  string   `json:"wan_load_balance_type,omitempty"`
	WANLoadBalanceWeight int     `json:"wan_load_balance_weight,omitempty"`
	WANNetworkGroup     string   `json:"wan_networkgroup,omitempty"`
	WANSmartQEnabled    bool     `json:"wan_smartq_enabled,omitempty"`
	WANProviderCaps     *WANProviderCaps `json:"wan_provider_capabilities,omitempty"`
	WANVLANEnabled      bool     `json:"wan_vlan_enabled,omitempty"`
	WANVLAN             int      `json:"wan_vlan,omitempty"`

	// PPPoE Settings
	WANUsername         string `json:"wan_username,omitempty"`
	WANPassword         string `json:"wan_password,omitempty"`

	// Static WAN Settings
	WANIPAddress        string   `json:"wan_ip,omitempty"`
	WANNetmask          string   `json:"wan_netmask,omitempty"`
	WANGateway          string   `json:"wan_gateway,omitempty"`
	WANDNS              []string `json:"wan_dns,omitempty"`

	// VPN Settings
	VPNType             string `json:"vpn_type,omitempty"`
	RadiusProfileID     string `json:"radiusprofile_id,omitempty"`

	// LTE Settings (for LTE WANs)
	LTEExtAnt           int    `json:"lte_ext_ant,omitempty"`

	// Content Filtering
	ContentFilterEnabled bool   `json:"contentfilter_enabled,omitempty"`

	// Auto-Scale
	AutoScaleEnabled    bool `json:"auto_scale_enabled,omitempty"`

	// Settings
	SettingPreference   string `json:"setting_preference,omitempty"`

	// Report/Statistics
	NumSTA              int `json:"num_sta,omitempty"`
	RXBytes             FlexInt `json:"rx_bytes,omitempty"`
	TXBytes             FlexInt `json:"tx_bytes,omitempty"`
	Up                  bool    `json:"up,omitempty"`
}

// WANProviderCaps represents WAN provider capabilities.
type WANProviderCaps struct {
	DownloadKilobitsPerSecond FlexInt `json:"download_kilobits_per_second,omitempty"`
	UploadKilobitsPerSecond   FlexInt `json:"upload_kilobits_per_second,omitempty"`
}

// Network purpose constants.
const (
	NetworkPurposeCorporate = "corporate"
	NetworkPurposeGuest     = "guest"
	NetworkPurposeWAN       = "wan"
	NetworkPurposeVPN       = "vpn"
	NetworkPurposeVLANOnly  = "vlan-only"
	NetworkPurposeRemoteUser = "remote-user-vpn"
)

// Network group constants.
const (
	NetworkGroupLAN = "LAN"
	NetworkGroupWAN = "WAN"
	NetworkGroupVPN = "VPN"
)

// WAN type constants.
const (
	WANTypeDHCP    = "dhcp"
	WANTypeStatic  = "static"
	WANTypePPPoE   = "pppoe"
	WANTypeDisabled = "disabled"
)
