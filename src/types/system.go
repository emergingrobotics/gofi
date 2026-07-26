package types

// Status represents system status (non-authenticated endpoint).
type Status struct {
	Up             bool   `json:"up"`
	ServerVersion  string `json:"server_version,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	Version        string `json:"version,omitempty"`
}

// AdminUser represents an administrator user account.
type AdminUser struct {
	ID               string   `json:"_id,omitempty"`
	UniqueID         string   `json:"unique_id,omitempty"`
	Name             string   `json:"name,omitempty"`
	Email            string   `json:"email,omitempty"`
	EmailStatus      string   `json:"email_status,omitempty"`
	FirstName        string   `json:"first_name,omitempty"`
	LastName         string   `json:"last_name,omitempty"`
	FullName         string   `json:"full_name,omitempty"`
	Status           string   `json:"status,omitempty"`
	Username         string   `json:"username,omitempty"`
	LocalAccountExist bool    `json:"local_account_exist,omitempty"`
	IsOwner          bool     `json:"isOwner,omitempty"`
	IsSuperAdmin     bool     `json:"isSuperAdmin,omitempty"`
	Roles            []Role   `json:"roles,omitempty"`
	Permissions      map[string]interface{} `json:"permissions,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
	CloudAccessGranted bool   `json:"cloud_access_granted,omitempty"`
	UpdateTime       string   `json:"update_time,omitempty"`
	Avatar           string   `json:"avatar,omitempty"`
}

// Role represents an admin user role.
type Role struct {
	Name       string `json:"name,omitempty"`
	SystemRole bool   `json:"system_role,omitempty"`
	UniqueID   string `json:"unique_id,omitempty"`
	SystemKey  string `json:"system_key,omitempty"`
}

// Backup represents a system backup file.
type Backup struct {
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	Time      int64  `json:"time"`
	Datetime  string `json:"datetime,omitempty"`
}

// SpeedTestStatus represents the status of a speed test.
type SpeedTestStatus struct {
	StatusDownload     int     `json:"status_download,omitempty"`
	StatusLatency      int     `json:"status_latency,omitempty"`
	StatusUpload       int     `json:"status_upload,omitempty"`
	StatusSummary      int     `json:"status_summary,omitempty"`
	Latency            int     `json:"latency,omitempty"`
	XputDownload       FlexInt `json:"xput_download,omitempty"`
	XputUpload         FlexInt `json:"xput_upload,omitempty"`
	Running            bool    `json:"running,omitempty"`
	Runtime            int     `json:"runtime,omitempty"`
	ServerName         string  `json:"server_name,omitempty"`
	ServerCountry      string  `json:"server_country,omitempty"`
	LastRun            int64   `json:"lastrun,omitempty"`
}
