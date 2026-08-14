package users

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

const (
	placeholder   = "-"
	columnPadding = 2
	blockedMark   = "blocked"
	dynamicMark   = "(dynamic)"
)

// FormatOptions controls how DoList renders entries.
type FormatOptions struct {
	Writer io.Writer
	JSON   bool
}

// FormatText writes user entries as space-aligned columns for terminal display.
func FormatText(writer io.Writer, entries []UserEntry) error {
	if len(entries) == 0 {
		return nil
	}

	tabWriter := tabwriter.NewWriter(writer, 0, 0, columnPadding, ' ', 0)
	if _, err := fmt.Fprintln(tabWriter, "NAME\tHOSTNAME\tMAC\tFIXED-IP\tFLAGS\tID"); err != nil {
		return err
	}

	for _, entry := range entries {
		line := strings.Join([]string{
			orPlaceholder(entry.Name),
			orPlaceholder(entry.Hostname),
			orPlaceholder(entry.MAC),
			fixedIPDisplay(entry),
			flagsDisplay(entry),
			orPlaceholder(entry.ID),
		}, "\t")
		if _, err := fmt.Fprintln(tabWriter, line); err != nil {
			return err
		}
	}

	return tabWriter.Flush()
}

// FormatJSON writes user entries as a JSON array for programmatic consumption.
func FormatJSON(writer io.Writer, entries []UserEntry) error {
	if entries == nil {
		entries = []UserEntry{}
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

// fixedIPDisplay distinguishes a reservation from a stale fixed_ip left on a
// record whose use_fixedip is false, which the controller ignores.
func fixedIPDisplay(entry UserEntry) string {
	if entry.FixedIP == "" {
		return placeholder
	}
	if !entry.UseFixedIP {
		return entry.FixedIP + " " + dynamicMark
	}
	return entry.FixedIP
}

func flagsDisplay(entry UserEntry) string {
	if entry.Blocked {
		return blockedMark
	}
	return placeholder
}

func orPlaceholder(value string) string {
	if value == "" {
		return placeholder
	}
	return value
}
