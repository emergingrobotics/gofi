package types

// DNSRecord represents a local DNS record (static DNS entry).
type DNSRecord struct {
	ID         string `json:"_id,omitempty"`
	Key        string `json:"key,omitempty"`        // Hostname/record name
	Value      string `json:"value,omitempty"`      // IP address or target
	RecordType string `json:"record_type,omitempty"` // A, AAAA, CNAME, MX, TXT, SRV
	TTL        int    `json:"ttl,omitempty"`        // Time to live
	Port       int    `json:"port,omitempty"`       // For SRV records
	Priority   int    `json:"priority,omitempty"`   // For MX/SRV records
	Weight     int    `json:"weight,omitempty"`     // For SRV records
	Enabled    bool   `json:"enabled,omitempty"`
}

// DNSRecordType constants for DNS record types.
const (
	DNSRecordTypeA     = "A"
	DNSRecordTypeAAAA  = "AAAA"
	DNSRecordTypeCNAME = "CNAME"
	DNSRecordTypeMX    = "MX"
	DNSRecordTypeTXT   = "TXT"
	DNSRecordTypeSRV   = "SRV"
)
