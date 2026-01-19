package types

// Site represents a UniFi site.
type Site struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
	Desc string `json:"desc"`

	// Attributes
	AttrHiddenID    string `json:"attr_hidden_id,omitempty"`
	AttrNoDelete    bool   `json:"attr_no_delete,omitempty"`
	Anonymous Anonymous `json:"anonymous_id,omitempty"`

	// Health subsystems
	Health []HealthData `json:"health,omitempty"`

	// System info
	SysInfo *SysInfo `json:"sysinfo,omitempty"`

	// Role (admin, readonly, etc.)
	Role string `json:"role,omitempty"`
}

// Anonymous represents anonymous site ID.
type Anonymous string

// HealthData represents health information for a subsystem.
type HealthData struct {
	Subsystem          string    `json:"subsystem"`
	Status             string    `json:"status"` // ok, warning, critical
	NumUser            int       `json:"num_user,omitempty"`
	NumGuest           int       `json:"num_guest,omitempty"`
	NumIot             int       `json:"num_iot,omitempty"`
	NumAP              int       `json:"num_ap,omitempty"`
	NumAdopted         int       `json:"num_adopted,omitempty"`
	NumDisabled        int       `json:"num_disabled,omitempty"`
	NumDisconnected    int       `json:"num_disconnected,omitempty"`
	NumPending         int       `json:"num_pending,omitempty"`
	NumGw              int       `json:"num_gw,omitempty"`
	NumSw              int       `json:"num_sw,omitempty"`
	TxBytesR           FlexInt   `json:"tx_bytes-r,omitempty"`
	RxBytesR           FlexInt   `json:"rx_bytes-r,omitempty"`
	RemoteUserEnabled  FlexBool  `json:"remote_user_enabled,omitempty"`
	RemoteUserNumActive int      `json:"remote_user_num_active,omitempty"`
	RemoteUserNumInactive int    `json:"remote_user_num_inactive,omitempty"`
	RemoteUserRxBytes  FlexInt   `json:"remote_user_rx_bytes,omitempty"`
	RemoteUserTxBytes  FlexInt   `json:"remote_user_tx_bytes,omitempty"`
	SiteToSiteEnabled  FlexBool  `json:"site_to_site_enabled,omitempty"`
	WanIP              string    `json:"wan_ip,omitempty"`
	Uptime             FlexInt   `json:"uptime,omitempty"`
	Drops              int       `json:"drops,omitempty"`
	Latency            int       `json:"latency,omitempty"`
	XputUp             FlexInt   `json:"xput_up,omitempty"`
	XputDown           FlexInt   `json:"xput_down,omitempty"`
	SpeedtestStatus    string    `json:"speedtest_status,omitempty"`
	SpeedtestLastRun   int64     `json:"speedtest_lastrun,omitempty"`
	SpeedtestPing      int       `json:"speedtest_ping,omitempty"`
	Gateways           []string  `json:"gateways,omitempty"`
	Netmask            string    `json:"netmask,omitempty"`
	Nameservers        []string  `json:"nameservers,omitempty"`
	LanIP              string    `json:"lan_ip,omitempty"`
	NumSta             int       `json:"num_sta,omitempty"`
	GwMAC              string    `json:"gw_mac,omitempty"`
	GwVersion          string    `json:"gw_version,omitempty"`
	GwName             string    `json:"gw_name,omitempty"`
	ISPName            string    `json:"isp_name,omitempty"`
	ISPOrganization    string    `json:"isp_organization,omitempty"`
	RemoteUserEnabled2 bool      `json:"remote_user_enabled2,omitempty"`
}

// SysInfo represents system information for the controller.
type SysInfo struct {
	Anonymous           Anonymous `json:"anonymous_controller_id,omitempty"`
	Build               string    `json:"build,omitempty"`
	CloudKey            bool      `json:"cloudkey,omitempty"`
	Console             bool      `json:"console,omitempty"`
	ControllerModel     string    `json:"controller_model,omitempty"`
	DataRetentionDays   int       `json:"data_retention_days,omitempty"`
	DataRetentionTimeEnabled bool `json:"data_retention_time_enabled,omitempty"`
	Debug               bool      `json:"debug,omitempty"`
	DebugDevice         string    `json:"debug_device,omitempty"`
	DebugMgmt           string    `json:"debug_mgmt,omitempty"`
	DebugSDN            string    `json:"debug_sdn,omitempty"`
	DebugSystem         string    `json:"debug_system,omitempty"`
	FB                  bool      `json:"fb_registered,omitempty"`
	Hostname            string    `json:"hostname,omitempty"`
	HTTPSPort           int       `json:"https_port,omitempty"`
	InformPort          int       `json:"inform_port,omitempty"`
	IP                  string    `json:"ip_addrs,omitempty"`
	LiveChat            string    `json:"live_chat,omitempty"`
	Name                string    `json:"name,omitempty"`
	PreviousVersion     string    `json:"previous_version,omitempty"`
	TimezoneName        string    `json:"timezone,omitempty"`
	Ubnt                bool      `json:"ubnt_device,omitempty"`
	UDM                 bool      `json:"udm_version,omitempty"`
	Update              string    `json:"update,omitempty"`
	UpdateAvailable     bool      `json:"update_available,omitempty"`
	UpdateDownloaded    bool      `json:"update_downloaded,omitempty"`
	Uptime              int64     `json:"uptime,omitempty"`
	Version             string    `json:"version,omitempty"`
	UnifiGo             bool      `json:"unifi_go_enabled,omitempty"`
	SSHEnabled          bool      `json:"is_ssh_enabled,omitempty"`
	AutoBackup          bool      `json:"autobackup,omitempty"`
	AutoBackupDays      int       `json:"autobackup_days,omitempty"`
	AutoBackupMaxFiles  int       `json:"autobackup_max_files,omitempty"`
	AutoUpgradeEnabled  bool      `json:"auto_upgrade,omitempty"`
	DiscoveredDevices   int       `json:"discovered,omitempty"`
}

// CreateSiteRequest represents a request to create a new site.
type CreateSiteRequest struct {
	Desc string `json:"desc"` // Site description (becomes the name)
	Name string `json:"name,omitempty"` // Optional short name
}

// UpdateSiteRequest represents a request to update a site.
type UpdateSiteRequest struct {
	Desc string `json:"desc,omitempty"`
	Name string `json:"name,omitempty"`
}
