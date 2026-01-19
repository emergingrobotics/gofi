package types

// Client represents a connected client/station.
type Client struct {
	ID              string  `json:"_id,omitempty"`
	SiteID          string  `json:"site_id,omitempty"`
	MAC             string  `json:"mac"`
	Hostname        string  `json:"hostname,omitempty"`
	Name            string  `json:"name,omitempty"`
	OUI             string  `json:"oui,omitempty"`
	IP              string  `json:"ip,omitempty"`
	IsGuest         bool    `json:"is_guest,omitempty"`
	IsWired         bool    `json:"is_wired,omitempty"`
	FirstSeen       int64   `json:"first_seen,omitempty"`
	LastSeen        int64   `json:"last_seen,omitempty"`
	Uptime          FlexInt `json:"uptime,omitempty"`

	// Connection info
	APMA            string  `json:"ap_mac,omitempty"`
	GWMAC           string  `json:"gw_mac,omitempty"`
	SWMAC           string  `json:"sw_mac,omitempty"`
	NetworkID       string  `json:"network_id,omitempty"`
	NetworkName     string  `json:"network,omitempty"`
	ESSID           string  `json:"essid,omitempty"`
	BSSID           string  `json:"bssid,omitempty"`
	Channel         int     `json:"channel,omitempty"`
	Radio           string  `json:"radio,omitempty"`
	RadioProto      string  `json:"radio_proto,omitempty"`

	// Stats
	RXBytes         FlexInt `json:"rx_bytes,omitempty"`
	RXBytesR        FlexInt `json:"rx_bytes-r,omitempty"`
	RXPackets       FlexInt `json:"rx_packets,omitempty"`
	RXRate          FlexInt `json:"rx_rate,omitempty"`
	TXBytes         FlexInt `json:"tx_bytes,omitempty"`
	TXBytesR        FlexInt `json:"tx_bytes-r,omitempty"`
	TXPackets       FlexInt `json:"tx_packets,omitempty"`
	TXRate          FlexInt `json:"tx_rate,omitempty"`
	Signal          FlexInt `json:"signal,omitempty"`
	Noise           FlexInt `json:"noise,omitempty"`
	RSSI            FlexInt `json:"rssi,omitempty"`
	Satisfaction    int     `json:"satisfaction,omitempty"`

	// Status
	Authorized      bool    `json:"authorized,omitempty"`
	Blocked         bool    `json:"blocked,omitempty"`
	Note            string  `json:"note,omitempty"`
	Noted           bool    `json:"noted,omitempty"`
	UseFixedIP      bool    `json:"use_fixedip,omitempty"`
	FixedIP         string  `json:"fixed_ip,omitempty"`
	UsergroupID     string  `json:"usergroup_id,omitempty"`

	// Guest authorization
	GuestAuthorized bool    `json:"guest_authorized,omitempty"`
	GuestKicked     bool    `json:"guest_kicked,omitempty"`
	GuestVoucher    string  `json:"guest_voucher,omitempty"`

	// Device fingerprinting
	DeviceIDOverride int    `json:"dev_id_override,omitempty"`
	DeviceVendor     string `json:"dev_vendor,omitempty"`
	DeviceFamily     string `json:"dev_family,omitempty"`
	OSName           string `json:"os_name,omitempty"`
	OSClass          int    `json:"os_class,omitempty"`

	// Switch port info (for wired clients)
	SWPORT          int    `json:"sw_port,omitempty"`
	SWDepth         int    `json:"sw_depth,omitempty"`

	// Misc
	IdleTime        FlexInt `json:"idletime,omitempty"`
	Anomalies       int     `json:"anomalies,omitempty"`
}
