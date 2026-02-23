package main

import (
	"context"
	"encoding/binary"
	"net"
	"sort"
	"strings"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/types"
)

// FilterMode selects which connection type to include in the output.
type FilterMode int

const (
	FilterAll  FilterMode = iota
	FilterWifi
	FilterWired
)

// ClientEntry combines UDM client data with an independent OUI manufacturer lookup.
type ClientEntry struct {
	MAC          string `json:"mac"`
	IP           string `json:"ip"`
	Hostname     string `json:"hostname"`
	Manufacturer string `json:"manufacturer"`
	IsWired      bool   `json:"is_wired"`

	// WiFi fields
	ESSID      string `json:"essid,omitempty"`
	AccessPointMAC string `json:"ap_mac,omitempty"`
	Channel    int    `json:"channel,omitempty"`
	Radio      string `json:"radio,omitempty"`
	RadioProto string `json:"radio_proto,omitempty"`
	Signal     int    `json:"signal,omitempty"`
	Noise      int    `json:"noise,omitempty"`
	RSSI       int    `json:"rssi,omitempty"`

	// Wired fields
	SwitchMAC  string `json:"sw_mac,omitempty"`
	SwitchPort int    `json:"sw_port,omitempty"`

	// Common stats
	RXBytes      int64 `json:"rx_bytes"`
	TXBytes      int64 `json:"tx_bytes"`
	Uptime       int64 `json:"uptime"`
	LastSeen     int64 `json:"last_seen"`
	Satisfaction int   `json:"satisfaction,omitempty"`
}

// ListClients fetches active clients from the UDM and enriches them with
// independent OUI manufacturer lookups instead of trusting the UDM's stale OUI field.
func ListClients(ctx context.Context, client gofi.Client, site string, filter FilterMode, ouiDatabase *OUIDatabase) ([]ClientEntry, error) {
	activeClients, err := client.Clients().ListActive(ctx, site)
	if err != nil {
		return nil, err
	}

	entries := []ClientEntry{}
	for _, activeClient := range activeClients {
		if !matchesFilter(activeClient, filter) {
			continue
		}
		entries = append(entries, buildClientEntry(activeClient, ouiDatabase))
	}

	sortClientEntries(entries)
	return entries, nil
}

func matchesFilter(client types.Client, filter FilterMode) bool {
	switch filter {
	case FilterWifi:
		return !client.IsWired
	case FilterWired:
		return client.IsWired
	default:
		return true
	}
}

func buildClientEntry(client types.Client, ouiDatabase *OUIDatabase) ClientEntry {
	entry := ClientEntry{
		MAC:          strings.ToLower(client.MAC),
		IP:           client.IP,
		Hostname:     resolveClientHostname(client),
		Manufacturer: ouiDatabase.Lookup(client.MAC),
		IsWired:      client.IsWired,
		RXBytes:      client.RXBytes.Int64(),
		TXBytes:      client.TXBytes.Int64(),
		Uptime:       client.Uptime.Int64(),
		LastSeen:     client.LastSeen,
	}

	if client.IsWired {
		entry.SwitchMAC = strings.ToLower(client.SWMAC)
		entry.SwitchPort = client.SWPORT
	} else {
		entry.ESSID = client.ESSID
		entry.AccessPointMAC = strings.ToLower(client.APMA)
		entry.Channel = client.Channel
		entry.Radio = client.Radio
		entry.RadioProto = client.RadioProto
		entry.Signal = client.Signal.Int()
		entry.Noise = client.Noise.Int()
		entry.RSSI = client.RSSI.Int()
		entry.Satisfaction = client.Satisfaction
	}

	return entry
}

func resolveClientHostname(client types.Client) string {
	if client.Name != "" {
		return client.Name
	}
	if client.Hostname != "" {
		return client.Hostname
	}
	return "unknown"
}

func sortClientEntries(entries []ClientEntry) {
	sort.Slice(entries, func(i, j int) bool {
		ipA := net.ParseIP(entries[i].IP)
		ipB := net.ParseIP(entries[j].IP)

		// Clients without IPs sort last
		if ipA == nil && ipB == nil {
			return entries[i].MAC < entries[j].MAC
		}
		if ipA == nil {
			return false
		}
		if ipB == nil {
			return true
		}

		return ipToUint32(ipA) < ipToUint32(ipB)
	})
}

func ipToUint32(parsedIP net.IP) uint32 {
	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ipv4)
}
