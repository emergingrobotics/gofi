package types

// Setting represents a base settings object.
type Setting struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`
	Key    string `json:"key"`
}

// SettingMgmt represents management/controller settings.
type SettingMgmt struct {
	Setting
	LEDEnabled              bool   `json:"led_enabled,omitempty"`
	AlertEnabled            bool   `json:"alert_enabled,omitempty"`
	AutoUpgrade             bool   `json:"auto_upgrade,omitempty"`
	XSSHEnabled             bool   `json:"x_ssh_enabled,omitempty"`
	XSSHKeys                string `json:"x_ssh_keys,omitempty"`
	XSSHUsername            string `json:"x_ssh_username,omitempty"`
	XSSHPassword            string `json:"x_ssh_password,omitempty"`
	XSSHAuthPasswordEnabled bool   `json:"x_ssh_auth_password_enabled,omitempty"`
}

// SettingConnectivity represents internet connectivity check settings.
type SettingConnectivity struct {
	Setting
	Enabled bool `json:"enabled,omitempty"`
}

// SettingCountry represents country/regulatory domain settings.
type SettingCountry struct {
	Setting
	Code int `json:"code,omitempty"`
}

// SettingGuestAccess represents guest portal settings.
type SettingGuestAccess struct {
	Setting
	Auth              string `json:"auth,omitempty"` // "none", "simple", "hotspot"
	Enabled           bool   `json:"enabled,omitempty"`
	Expire            int    `json:"expire,omitempty"` // Minutes
	ExpireNumber      int    `json:"expire_number,omitempty"`
	ExpireUnit        int    `json:"expire_unit,omitempty"`
	Password          string `json:"password,omitempty"`
	Portal            bool   `json:"portal_enabled,omitempty"`
	PortalCustomized  bool   `json:"portal_customized,omitempty"`
	RedirectEnabled   bool   `json:"redirect_enabled,omitempty"`
	RedirectHTTPS     bool   `json:"redirect_https,omitempty"`
	RedirectURL       string `json:"redirect_url,omitempty"`
}

// SettingDPI represents Deep Packet Inspection settings.
type SettingDPI struct {
	Setting
	Enabled   bool `json:"enabled,omitempty"`
	Fingerprt bool `json:"fingerprt,omitempty"`
}

// SettingIPS represents Intrusion Prevention System settings.
type SettingIPS struct {
	Setting
	Enabled        bool   `json:"enabled,omitempty"`
	RuleCategories []string `json:"rule_categories,omitempty"`
}

// SettingNTP represents NTP server settings.
type SettingNTP struct {
	Setting
	NTPServer1 string `json:"ntp_server_1,omitempty"`
	NTPServer2 string `json:"ntp_server_2,omitempty"`
	NTPServer3 string `json:"ntp_server_3,omitempty"`
	NTPServer4 string `json:"ntp_server_4,omitempty"`
}

// SettingSNMP represents SNMP settings.
type SettingSNMP struct {
	Setting
	Enabled       bool   `json:"enabled,omitempty"`
	Community     string `json:"community,omitempty"`
	Location      string `json:"location,omitempty"`
	Contact       string `json:"contact,omitempty"`
}

// SettingRsyslog represents remote syslog settings.
type SettingRsyslog struct {
	Setting
	Enabled bool   `json:"enabled,omitempty"`
	Host    string `json:"host,omitempty"`
	Port    int    `json:"port,omitempty"`
}

// SettingRadius represents RADIUS settings.
type SettingRadius struct {
	Setting
	Enabled bool `json:"enabled,omitempty"`
}

// RADIUSProfile represents a RADIUS server profile.
type RADIUSProfile struct {
	ID                    string `json:"_id,omitempty"`
	SiteID                string `json:"site_id,omitempty"`
	Name                  string `json:"name"`
	AuthServers           []RADIUSServer `json:"auth_servers,omitempty"`
	AcctServers           []RADIUSServer `json:"acct_servers,omitempty"`
	VLANEnabled           bool   `json:"vlan_enabled,omitempty"`
	VLANWLANMode          string `json:"vlan_wlan_mode,omitempty"`
	InterimUpdateEnabled  bool   `json:"interim_update_enabled,omitempty"`
	InterimUpdateInterval int    `json:"interim_update_interval,omitempty"`
}

// RADIUSServer represents a RADIUS server configuration.
type RADIUSServer struct {
	IP     string `json:"ip"`
	Port   int    `json:"port"`
	Secret string `json:"x_secret"`
}

// DynamicDNS represents Dynamic DNS configuration.
type DynamicDNS struct {
	ID       string `json:"_id,omitempty"`
	SiteID   string `json:"site_id,omitempty"`
	Service  string `json:"service"` // "dyndns", "afraid", "zoneedit", etc.
	Enabled  bool   `json:"enabled"`
	Interface string `json:"interface,omitempty"`
	Hostname string `json:"host"`
	Server   string `json:"server,omitempty"`
	Login    string `json:"login,omitempty"`
	Password string `json:"x_password,omitempty"`
}

// Setting key constants.
const (
	SettingKeyMgmt         = "mgmt"
	SettingKeyConnectivity = "connectivity"
	SettingKeyCountry      = "country"
	SettingKeyGuestAccess  = "guest_access"
	SettingKeyDPI          = "dpi"
	SettingKeyIPS          = "ips"
	SettingKeyNTP          = "ntp"
	SettingKeySNMP         = "snmp"
	SettingKeyRsyslog      = "rsyslog"
	SettingKeyRadius       = "radius"
)
