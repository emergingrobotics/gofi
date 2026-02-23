package main

import (
	"encoding/json"
	"fmt"
	"io"
)

const noIPPlaceholder = "-"

// FormatText writes client entries as tab-separated columns for terminal display.
func FormatText(writer io.Writer, entries []ClientEntry) error {
	if len(entries) > 0 {
		if _, err := fmt.Fprintf(writer, "MAC\tIP\tHOSTNAME\tOUI-MANUFACTURER\n"); err != nil {
			return err
		}
	}
	for _, entry := range entries {
		ipDisplay := entry.IP
		if ipDisplay == "" {
			ipDisplay = noIPPlaceholder
		}
		if _, err := fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", entry.MAC, ipDisplay, entry.Hostname, entry.Manufacturer); err != nil {
			return err
		}
	}
	return nil
}

// FormatJSON writes client entries as a JSON array for programmatic consumption.
func FormatJSON(writer io.Writer, entries []ClientEntry) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}
