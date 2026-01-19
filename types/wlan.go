package types

// WLAN represents a wireless network (SSID) configuration.
type WLAN struct {
	ID                    string   `json:"_id,omitempty"`
	SiteID                string   `json:"site_id,omitempty"`
	Name                  string   `json:"name"`
	Enabled               bool     `json:"enabled"`
	Security              string   `json:"security"` // "open", "wpapsk", "wpaeap", "wpa3"
	WPAMode               string   `json:"wpa_mode,omitempty"` // "wpa", "wpa2", "wpa3", "both"
	WPAEnc                string   `json:"wpa_enc,omitempty"` // "ccmp", "tkip", "both"
	Passphrase            string   `json:"x_passphrase,omitempty"`
	HideSSID              bool     `json:"hide_ssid"`
	IsGuest               bool     `json:"is_guest"`
	NetworkConfID         string   `json:"networkconf_id,omitempty"`
	UsergroupID           string   `json:"usergroup_id,omitempty"`
	APGroupIDs            []string `json:"ap_group_ids,omitempty"`
	WLANBands             []string `json:"wlan_bands,omitempty"` // ["2g", "5g", "6g"]
	WLANBand              string   `json:"wlan_band,omitempty"` // Legacy single band

	// WPA3 and PMF
	WPA3Support           bool     `json:"wpa3_support"`
	WPA3Transition        bool     `json:"wpa3_transition"`
	WPA3Enhanced          bool     `json:"wpa3_enhanced_192,omitempty"`
	PMFMode               string   `json:"pmf_mode,omitempty"` // "disabled", "optional", "required"

	// Roaming and Performance
	FastRoamingEnabled    bool     `json:"fast_roaming_enabled"`
	UAPSDEnabled          bool     `json:"uapsd_enabled"`
	BSSTransition         bool     `json:"bss_transition,omitempty"`

	// Data Rate Control
	MinrateNGEnabled      bool     `json:"minrate_ng_enabled,omitempty"`
	MinrateNGDataRateKbps int      `json:"minrate_ng_data_rate_kbps,omitempty"`
	MinrateNGAdvEnabled   bool     `json:"minrate_ng_advertising_rates,omitempty"`
	MinrateNGBeaconRateKbps int    `json:"minrate_ng_beacon_rate_kbps,omitempty"`
	MinrateNGMgmtRateKbps int      `json:"minrate_ng_mgmt_rate_kbps,omitempty"`
	MinrateNAEnabled      bool     `json:"minrate_na_enabled,omitempty"`
	MinrateNADataRateKbps int      `json:"minrate_na_data_rate_kbps,omitempty"`
	MinrateNAAdvEnabled   bool     `json:"minrate_na_advertising_rates,omitempty"`
	MinrateNABeaconRateKbps int    `json:"minrate_na_beacon_rate_kbps,omitempty"`
	MinrateNAMgmtRateKbps int      `json:"minrate_na_mgmt_rate_kbps,omitempty"`

	// MAC Filtering
	MACFilterEnabled      bool     `json:"mac_filter_enabled"`
	MACFilterPolicy       string   `json:"mac_filter_policy,omitempty"` // "allow", "deny"
	MACFilterList         []string `json:"mac_filter_list,omitempty"`

	// Schedule
	ScheduleEnabled       bool     `json:"schedule_enabled"`
	Schedule              []string `json:"schedule,omitempty"` // Array of day schedules
	ScheduleWithDuration  []WLANSchedule `json:"schedule_with_duration,omitempty"`

	// DTIM (Delivery Traffic Indication Message)
	DTIMMode              string   `json:"dtim_mode,omitempty"` // "default", "custom"
	DTIMNG                int      `json:"dtim_ng,omitempty"` // 2.4 GHz
	DTIMNA                int      `json:"dtim_na,omitempty"` // 5 GHz

	// Isolation and Security
	IAPPEnabled           bool     `json:"iapp_enabled"`
	L2Isolation           bool     `json:"l2_isolation"`
	ProxyARPEnabled       bool     `json:"proxy_arp,omitempty"`
	GroupRekey            int      `json:"group_rekey,omitempty"` // Seconds

	// RADIUS Settings (for Enterprise)
	RADIUSMACAuthEnabled  bool     `json:"radius_mac_auth_enabled"`
	RADIUSDASEnabled      bool     `json:"radius_das_enabled,omitempty"`
	RADIUSProfileID       string   `json:"radius_profile_id,omitempty"`
	RADIUSOverrideEnabled bool     `json:"radiusprofile_override,omitempty"`

	// Guest Portal
	GuestPortalID         string   `json:"portal_customization_id,omitempty"`
	PortalEnabled         bool     `json:"portal_enabled,omitempty"`
	PortalUseLandingPage  bool     `json:"portal_use_hostname,omitempty"`

	// Bandwidth Limiting
	UsergroupBandwidthLimitEnabled bool `json:"usergroup_bandwidth_limit_enabled,omitempty"`
	UsergroupBandwidthLimitUp      int  `json:"usergroup_bandwidth_limit_up,omitempty"` // kbps
	UsergroupBandwidthLimitDown    int  `json:"usergroup_bandwidth_limit_down,omitempty"` // kbps

	// Advanced Settings
	No2GHzOUI             bool     `json:"no2ghz_oui,omitempty"`
	P2PCrossConnect       bool     `json:"p2p_cross_connect,omitempty"`
	BeaconMode            string   `json:"beacon_mode,omitempty"`
	BCFilterEnabled       bool     `json:"bc_filter_enabled,omitempty"`
	BCFilterList          []string `json:"bc_filter_list,omitempty"`
	UseSavePassphrase     bool     `json:"use_saved_passphrase,omitempty"`

	// Statistics
	NumSTA                int      `json:"num_sta,omitempty"`
	RXBytes               FlexInt  `json:"rx_bytes,omitempty"`
	TXBytes               FlexInt  `json:"tx_bytes,omitempty"`

	// VLAN
	VLANEnabled           bool     `json:"vlan_enabled,omitempty"`
	VLAN                  int      `json:"vlan,omitempty"`
}

// WLANSchedule represents a schedule entry with time ranges.
type WLANSchedule struct {
	Day       string `json:"day"` // "sun", "mon", "tue", "wed", "thu", "fri", "sat"
	StartHour int    `json:"start_hour"`
	StartMin  int    `json:"start_min"`
	EndHour   int    `json:"end_hour"`
	EndMin    int    `json:"end_min"`
}

// WLANGroup represents a WLAN group configuration.
type WLANGroup struct {
	ID              string   `json:"_id,omitempty"`
	SiteID          string   `json:"site_id,omitempty"`
	Name            string   `json:"name"`
	Members         []string `json:"attr_hidden_id,omitempty"` // List of device MACs
	AttrNoDelete    bool     `json:"attr_no_delete,omitempty"`
}

// Security type constants for WLAN.
const (
	SecurityTypeOpen   = "open"
	SecurityTypeWPAPSK = "wpapsk"
	SecurityTypeWPAEAP = "wpaeap" // Enterprise
	SecurityTypeWPA3   = "wpa3"
)

// WPA mode constants.
const (
	WPAModeWPA     = "wpa"
	WPAModeWPA2    = "wpa2"
	WPAModeWPA3    = "wpa3"
	WPAModeBoth    = "both" // WPA + WPA2
)

// WPA encryption constants.
const (
	WPAEncCCMP = "ccmp" // AES
	WPAEncTKIP = "tkip"
	WPAEncBoth = "both" // CCMP + TKIP
)

// PMF (Protected Management Frames) mode constants.
const (
	PMFModeDisabled  = "disabled"
	PMFModeOptional  = "optional"
	PMFModeRequired  = "required"
)

// MAC filter policy constants.
const (
	MACFilterPolicyAllow = "allow"
	MACFilterPolicyDeny  = "deny"
)

// WLAN band constants.
const (
	WLANBand2G = "2g"
	WLANBand5G = "5g"
	WLANBand6G = "6g"
)

// DTIM mode constants.
const (
	DTIMModeDefault = "default"
	DTIMModeCustom  = "custom"
)
