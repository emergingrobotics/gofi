package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/services"
	"github.com/unifi-go/gofi/types"
)

// FilterMode selects which connection type to include in the output.
type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterWifi
	FilterWired
)

// SortMode selects the ordering of output entries.
type SortMode int

const (
	SortFirstSeen SortMode = iota
	SortLastSeen
	SortIP
)

const (
	statusPresent = "present"
	statusGone    = "gone"
)

// parseSortMode maps a --sort flag value to a SortMode.
func parseSortMode(value string) (SortMode, error) {
	switch value {
	case "first-seen":
		return SortFirstSeen, nil
	case "last-seen":
		return SortLastSeen, nil
	case "ip":
		return SortIP, nil
	default:
		return 0, fmt.Errorf("invalid sort key %q (want first-seen, last-seen, or ip)", value)
	}
}

// ClientEntry combines UDM client data with an independent OUI manufacturer lookup.
type ClientEntry struct {
	MAC          string `json:"mac"`
	IP           string `json:"ip"`
	Hostname     string `json:"hostname"`
	Manufacturer string `json:"manufacturer"`
	IsWired      bool   `json:"is_wired"`
	Status       string `json:"status"`

	// WiFi fields
	ESSID          string `json:"essid,omitempty"`
	AccessPointMAC string `json:"ap_mac,omitempty"`
	Channel        int    `json:"channel,omitempty"`
	Radio          string `json:"radio,omitempty"`
	RadioProto     string `json:"radio_proto,omitempty"`
	Signal         int    `json:"signal,omitempty"`
	Noise          int    `json:"noise,omitempty"`
	RSSI           int    `json:"rssi,omitempty"`

	// Wired fields
	SwitchMAC  string `json:"sw_mac,omitempty"`
	SwitchPort int    `json:"sw_port,omitempty"`

	// Common stats
	RXBytes      int64 `json:"rx_bytes"`
	TXBytes      int64 `json:"tx_bytes"`
	Uptime       int64 `json:"uptime"`
	FirstSeen    int64 `json:"first_seen"`
	LastSeen     int64 `json:"last_seen"`
	Satisfaction int   `json:"satisfaction,omitempty"`
}

// ListClients fetches active clients from the UDM and enriches them with
// independent OUI manufacturer lookups instead of trusting the UDM's stale OUI field.
func ListClients(ctx context.Context, client gofi.Client, site string, filter FilterMode, sortMode SortMode, ouiDatabase *OUIDatabase) ([]ClientEntry, error) {
	activeClients, err := client.Clients().ListActive(ctx, site)
	if err != nil {
		return nil, err
	}

	entries := []ClientEntry{}
	for _, activeClient := range activeClients {
		if !matchesFilter(activeClient, filter) {
			continue
		}
		entries = append(entries, buildClientEntry(activeClient, ouiDatabase, true))
	}

	sortClientEntries(entries, sortMode)
	return entries, nil
}

// ListClientsHistory fetches the union of currently-active and historical clients
// seen within the given hour window, marking each present or gone. Present devices
// use their live active record for accurate IP/link fields; gone devices use their
// last-known historical record. When goneOnly is set, present devices are excluded.
func ListClientsHistory(ctx context.Context, client gofi.Client, site string, filter FilterMode, sortMode SortMode, withinHours int, goneOnly bool, ouiDatabase *OUIDatabase) ([]ClientEntry, error) {
	activeClients, err := client.Clients().ListActive(ctx, site)
	if err != nil {
		return nil, err
	}
	allClients, err := client.Clients().ListAll(ctx, site, services.WithinHours(withinHours))
	if err != nil {
		return nil, err
	}

	return buildHistoryEntries(activeClients, allClients, filter, sortMode, goneOnly, ouiDatabase), nil
}

// buildHistoryEntries merges active and historical client sets into presence-marked
// entries. Kept separate from ListClientsHistory so it can be tested without a UDM.
func buildHistoryEntries(activeClients, allClients []types.Client, filter FilterMode, sortMode SortMode, goneOnly bool, ouiDatabase *OUIDatabase) []ClientEntry {
	activeByMAC := make(map[string]types.Client, len(activeClients))
	for _, activeClient := range activeClients {
		activeByMAC[strings.ToLower(activeClient.MAC)] = activeClient
	}
	allByMAC := make(map[string]types.Client, len(allClients))
	for _, historicalClient := range allClients {
		allByMAC[strings.ToLower(historicalClient.MAC)] = historicalClient
	}

	entries := []ClientEntry{}
	seen := make(map[string]bool)
	appendEntry := func(mac string) {
		if seen[mac] {
			return
		}
		seen[mac] = true

		source, present := activeByMAC[mac]
		if !present {
			source = allByMAC[mac]
		}
		if goneOnly && present {
			return
		}
		if !matchesFilter(source, filter) {
			return
		}
		entries = append(entries, buildClientEntry(source, ouiDatabase, present))
	}

	for _, historicalClient := range allClients {
		appendEntry(strings.ToLower(historicalClient.MAC))
	}
	for _, activeClient := range activeClients {
		appendEntry(strings.ToLower(activeClient.MAC))
	}

	sortClientEntries(entries, sortMode)
	return entries
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

func buildClientEntry(client types.Client, ouiDatabase *OUIDatabase, present bool) ClientEntry {
	status := statusGone
	if present {
		status = statusPresent
	}
	entry := ClientEntry{
		MAC:          strings.ToLower(client.MAC),
		IP:           client.IP,
		Hostname:     resolveClientHostname(client),
		Manufacturer: ouiDatabase.Lookup(client.MAC),
		IsWired:      client.IsWired,
		Status:       status,
		RXBytes:      client.RXBytes.Int64(),
		TXBytes:      client.TXBytes.Int64(),
		Uptime:       client.Uptime.Int64(),
		FirstSeen:    client.FirstSeen,
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

func sortClientEntries(entries []ClientEntry, mode SortMode) {
	sort.Slice(entries, func(i, j int) bool {
		switch mode {
		case SortFirstSeen:
			if entries[i].FirstSeen != entries[j].FirstSeen {
				return entries[i].FirstSeen > entries[j].FirstSeen
			}
		case SortLastSeen:
			if entries[i].LastSeen != entries[j].LastSeen {
				return entries[i].LastSeen > entries[j].LastSeen
			}
		}
		return lessByIP(entries[i], entries[j])
	})
}

// lessByIP orders entries by IPv4 ascending, placing clients without an IP last
// and breaking ties by MAC. It is the tie-break for time-based sorts and the
// whole ordering for SortIP.
func lessByIP(entryA, entryB ClientEntry) bool {
	ipA := net.ParseIP(entryA.IP)
	ipB := net.ParseIP(entryB.IP)

	if ipA == nil && ipB == nil {
		return entryA.MAC < entryB.MAC
	}
	if ipA == nil {
		return false
	}
	if ipB == nil {
		return true
	}

	return ipToUint32(ipA) < ipToUint32(ipB)
}

func ipToUint32(parsedIP net.IP) uint32 {
	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ipv4)
}
