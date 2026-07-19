package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
)

const (
	placeholder    = "-"
	columnPadding  = 2
	dhcpDisabled   = "(disabled)"
	dnsFieldJoiner = ","
)

// FormatText writes network entries as space-aligned columns for terminal display.
func FormatText(writer io.Writer, entries []NetworkEntry) error {
	if len(entries) == 0 {
		return nil
	}

	tabWriter := tabwriter.NewWriter(writer, 0, 0, columnPadding, ' ', 0)
	if _, err := fmt.Fprintln(tabWriter, "NETWORK\tVLAN\tSUBNET\tDHCP-POOL\tLEASE\tGATEWAY\tDNS"); err != nil {
		return err
	}

	for _, entry := range entries {
		line := strings.Join([]string{
			entry.Name,
			vlanDisplay(entry.VLAN),
			orPlaceholder(entry.Subnet),
			poolDisplay(entry),
			leaseDisplay(entry),
			orPlaceholder(entry.Gateway),
			dnsDisplay(entry.DNS),
		}, "\t")
		if _, err := fmt.Fprintln(tabWriter, line); err != nil {
			return err
		}
	}
	return tabWriter.Flush()
}

// FormatJSON writes network entries as a JSON array for programmatic consumption.
func FormatJSON(writer io.Writer, entries []NetworkEntry) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

func vlanDisplay(vlan int) string {
	if vlan == 0 {
		return placeholder
	}
	return strconv.Itoa(vlan)
}

// poolDisplay renders the dynamic DHCP range, or a marker when DHCP is off or
// the range is unset.
func poolDisplay(entry NetworkEntry) string {
	if !entry.DHCPEnabled {
		return dhcpDisabled
	}
	if entry.DHCPStart == "" || entry.DHCPStop == "" {
		return placeholder
	}
	return entry.DHCPStart + " - " + entry.DHCPStop
}

func leaseDisplay(entry NetworkEntry) string {
	if !entry.DHCPEnabled || entry.DHCPLease == 0 {
		return placeholder
	}
	return strconv.Itoa(entry.DHCPLease) + "s"
}

func dnsDisplay(servers []string) string {
	if len(servers) == 0 {
		return placeholder
	}
	return strings.Join(servers, dnsFieldJoiner)
}

func orPlaceholder(value string) string {
	if value == "" {
		return placeholder
	}
	return value
}
