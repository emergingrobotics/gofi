package types

// TrafficRule represents a v2 API traffic rule (bandwidth limiting, QoS, etc.).
type TrafficRule struct {
	ID                string          `json:"_id,omitempty"`
	SiteID            string          `json:"site_id,omitempty"`
	Name              string          `json:"name"`
	Enabled           bool            `json:"enabled"`
	Action            string          `json:"action"` // "ACCEPT", "DROP", "LIMIT"
	MatchingTarget    string          `json:"matching_target"` // "CLIENT", "NETWORK", "ALL"
	TargetDevices     []TargetDevice  `json:"target_devices,omitempty"`
	IPRange           *IPRange        `json:"ip_range,omitempty"`
	Regions           []string        `json:"regions,omitempty"`
	Domains           []string        `json:"domains,omitempty"`
	Categories        []string        `json:"categories,omitempty"`
	Schedule          *Schedule       `json:"schedule,omitempty"`
	Bandwidth         *Bandwidth      `json:"bandwidth,omitempty"`
	NetworkIDs        []string        `json:"network_ids,omitempty"`
	AppCategoryIDs    []string        `json:"app_category_ids,omitempty"`
}

// TargetDevice represents a target device for a traffic rule.
type TargetDevice struct {
	ClientMAC  string `json:"client_mac,omitempty"`
	NetworkID  string `json:"network_id,omitempty"`
	Type       string `json:"type"` // "CLIENT", "NETWORK", "ALL"
}

// Schedule represents a time-based schedule for traffic rules.
type Schedule struct {
	Mode         string   `json:"mode"` // "ALWAYS", "TIME_RANGE"
	DateStart    string   `json:"date_start,omitempty"` // "YYYY-MM-DD"
	DateEnd      string   `json:"date_end,omitempty"`
	TimeRanges   []TimeRange `json:"time_ranges,omitempty"`
	DaysOfWeek   []string `json:"days_of_week,omitempty"` // "MON", "TUE", etc.
	RepeatOnDays bool     `json:"repeat_on_days,omitempty"`
}

// TimeRange represents a time range within a day.
type TimeRange struct {
	StartHour int `json:"start_hour"`
	StartMin  int `json:"start_min"`
	EndHour   int `json:"end_hour"`
	EndMin    int `json:"end_min"`
}

// Bandwidth represents bandwidth limiting settings.
type Bandwidth struct {
	DownloadEnabled bool    `json:"download_enabled,omitempty"`
	DownloadLimit   FlexInt `json:"download_limit_kbps,omitempty"`
	UploadEnabled   bool    `json:"upload_enabled,omitempty"`
	UploadLimit     FlexInt `json:"upload_limit_kbps,omitempty"`
}

// IPRange represents an IP address range.
type IPRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Traffic action constants.
const (
	TrafficActionAccept = "ACCEPT"
	TrafficActionDrop   = "DROP"
	TrafficActionLimit  = "LIMIT"
)

// Matching target constants.
const (
	MatchingTargetClient  = "CLIENT"
	MatchingTargetNetwork = "NETWORK"
	MatchingTargetAll     = "ALL"
)

// Schedule mode constants.
const (
	ScheduleModeAlways    = "ALWAYS"
	ScheduleModeTimeRange = "TIME_RANGE"
)
