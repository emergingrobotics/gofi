package types

// Event represents a UniFi event (device connect/disconnect, client activity, etc.).
type Event struct {
	ID          string `json:"_id"`
	Time        int64  `json:"time"`
	Datetime    string `json:"datetime"`
	Key         string `json:"key"` // Event type key like "EVT_AP_Connected"
	Message     string `json:"msg"`
	SiteID      string `json:"site_id"`
	Subsystem   string `json:"subsystem"` // "wlan", "lan", "wan", etc.

	// Device info
	AP          string `json:"ap,omitempty"`
	APMAC       string `json:"ap_mac,omitempty"`
	APName      string `json:"ap_name,omitempty"`
	SW          string `json:"sw,omitempty"`
	SWMAC       string `json:"sw_mac,omitempty"`
	SWName      string `json:"sw_name,omitempty"`
	GW          string `json:"gw,omitempty"`
	GWMAC       string `json:"gw_mac,omitempty"`
	GWName      string `json:"gw_name,omitempty"`

	// Client info
	Client      string `json:"client,omitempty"`
	User        string `json:"user,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	SSID        string `json:"ssid,omitempty"`

	// Admin/User info
	Admin       string `json:"admin,omitempty"`
	IsAdmin     bool   `json:"is_admin,omitempty"`

	// Network info
	Network     string `json:"network,omitempty"`
	NetworkName string `json:"network_name,omitempty"`

	// Additional details
	Duration    FlexInt `json:"duration,omitempty"`
	Bytes       FlexInt `json:"bytes,omitempty"`
	Channel     int     `json:"channel,omitempty"`
	Radio       string  `json:"radio,omitempty"`
	InnerID     int     `json:"inner_id,omitempty"`
}

// Alarm represents a UniFi alarm/alert.
type Alarm struct {
	ID             string  `json:"_id"`
	Time           int64   `json:"time"`
	Datetime       string  `json:"datetime"`
	Key            string  `json:"key"` // Alarm type key
	Message        string  `json:"msg"`
	SiteID         string  `json:"site_id"`
	Subsystem      string  `json:"subsystem"`
	Archived       bool    `json:"archived"`
	Handled        bool    `json:"handled"`
	HandledBy      string  `json:"handled_by,omitempty"`
	HandledTime    int64   `json:"handled_time,omitempty"`

	// Device info
	AP             string  `json:"ap,omitempty"`
	APMAC          string  `json:"ap_mac,omitempty"`
	APName         string  `json:"ap_name,omitempty"`
	SW             string  `json:"sw,omitempty"`
	SWMAC          string  `json:"sw_mac,omitempty"`
	SWName         string  `json:"sw_name,omitempty"`
	GW             string  `json:"gw,omitempty"`
	GWMAC          string  `json:"gw_mac,omitempty"`
	GWName         string  `json:"gw_name,omitempty"`

	// IPS/IDS specific
	CatNo          int     `json:"catno,omitempty"`
	SrcIP          string  `json:"src_ip,omitempty"`
	DstIP          string  `json:"dst_ip,omitempty"`
	Proto          string  `json:"proto,omitempty"`
	SrcPort        int     `json:"src_port,omitempty"`
	DstPort        int     `json:"dst_port,omitempty"`
	InnerAlertID   int     `json:"inner_alert_id,omitempty"`
}

// Common event keys.
const (
	EventAPConnected       = "EVT_AP_Connected"
	EventAPDisconnected    = "EVT_AP_Disconnected"
	EventAPRestarted       = "EVT_AP_Restarted"
	EventAPUpgraded        = "EVT_AP_Upgraded"
	EventWUConnected       = "EVT_WU_Connected"
	EventWUDisconnected    = "EVT_WU_Disconnected"
	EventWURoam            = "EVT_WU_Roam"
	EventLUConnected       = "EVT_LU_Connected"
	EventLUDisconnected    = "EVT_LU_Disconnected"
	EventSWConnected       = "EVT_SW_Connected"
	EventSWDisconnected    = "EVT_SW_Disconnected"
	EventGWConnected       = "EVT_GW_Connected"
	EventGWWANTransition   = "EVT_GW_WANTransition"
	EventIPSAlert          = "EVT_IPS_Alert"
	EventADLogin           = "EVT_AD_Login"
)
