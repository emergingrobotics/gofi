package types

// User represents a known client (saved in the user database).
type User struct {
	ID              string  `json:"_id,omitempty"`
	SiteID          string  `json:"site_id,omitempty"`
	MAC             string  `json:"mac"`
	Hostname        string  `json:"hostname,omitempty"`
	Name            string  `json:"name,omitempty"`
	Note            string  `json:"note,omitempty"`
	Noted           bool    `json:"noted,omitempty"`
	OUI             string  `json:"oui,omitempty"`
	FirstSeen       int64   `json:"first_seen,omitempty"`
	LastSeen        int64   `json:"last_seen,omitempty"`

	// Fixed IP
	UseFixedIP      bool    `json:"use_fixedip,omitempty"`
	NetworkID       string  `json:"network_id,omitempty"`
	FixedIP         string  `json:"fixed_ip,omitempty"`

	// User group
	UsergroupID     string  `json:"usergroup_id,omitempty"`

	// Device fingerprinting override
	DeviceIDOverride int    `json:"dev_id_override,omitempty"`

	// Blocking
	Blocked         bool    `json:"blocked,omitempty"`

	// Stats (when client is connected)
	IsGuest         bool    `json:"is_guest,omitempty"`
	IsWired         bool    `json:"is_wired,omitempty"`
	RXBytes         FlexInt `json:"rx_bytes,omitempty"`
	TXBytes         FlexInt `json:"tx_bytes,omitempty"`
}

// UserGroup represents a user group for grouping clients.
type UserGroup struct {
	ID              string  `json:"_id,omitempty"`
	SiteID          string  `json:"site_id,omitempty"`
	Name            string  `json:"name"`
	QOSRateMaxDown  int     `json:"qos_rate_max_down,omitempty"` // kbps
	QOSRateMaxUp    int     `json:"qos_rate_max_up,omitempty"`   // kbps
	AttrNoDelete    bool    `json:"attr_no_delete,omitempty"`
	AttrHiddenID    string  `json:"attr_hidden_id,omitempty"`
}
