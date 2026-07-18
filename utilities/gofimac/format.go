package main

import (
	"encoding/json"
	"fmt"
	"io"
)

const noIPPlaceholder = "-"

const (
	secondsPerMinute = 60
	secondsPerHour   = 3600
	secondsPerDay    = 86400
	secondsPerMonth  = secondsPerDay * 30
	secondsPerYear   = secondsPerDay * 365
)

// FormatText writes client entries as tab-separated columns for terminal display.
// The STATUS column is included only in history views (showStatus). AGE and
// LAST-SEEN are rendered relative to now (unix seconds).
func FormatText(writer io.Writer, entries []ClientEntry, showStatus bool, now int64) error {
	if len(entries) == 0 {
		return nil
	}

	header := "MAC\tIP\tHOSTNAME\tOUI-MANUFACTURER\tAGE\tLAST-SEEN"
	if showStatus {
		header += "\tSTATUS"
	}
	if _, err := fmt.Fprintln(writer, header); err != nil {
		return err
	}

	for _, entry := range entries {
		ipDisplay := entry.IP
		if ipDisplay == "" {
			ipDisplay = noIPPlaceholder
		}
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
			entry.MAC, ipDisplay, entry.Hostname, entry.Manufacturer,
			formatRelativeTime(entry.FirstSeen, now), formatRelativeTime(entry.LastSeen, now))
		if showStatus {
			line += "\t" + entry.Status
		}
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return err
		}
	}
	return nil
}

// formatRelativeTime renders a unix timestamp as a compact age relative to now,
// e.g. "now", "4m", "2h", "5d", "3mo", "2y". Zero/unknown timestamps render as "-".
func formatRelativeTime(epoch, now int64) string {
	if epoch <= 0 {
		return noIPPlaceholder
	}
	delta := now - epoch
	if delta < 0 {
		delta = 0
	}

	switch {
	case delta < secondsPerMinute:
		return "now"
	case delta < secondsPerHour:
		return fmt.Sprintf("%dm", delta/secondsPerMinute)
	case delta < secondsPerDay:
		return fmt.Sprintf("%dh", delta/secondsPerHour)
	case delta < secondsPerMonth:
		return fmt.Sprintf("%dd", delta/secondsPerDay)
	case delta < secondsPerYear:
		return fmt.Sprintf("%dmo", delta/secondsPerMonth)
	default:
		return fmt.Sprintf("%dy", delta/secondsPerYear)
	}
}

// FormatJSON writes client entries as a JSON array for programmatic consumption.
func FormatJSON(writer io.Writer, entries []ClientEntry) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}
