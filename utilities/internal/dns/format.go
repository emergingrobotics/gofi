package dns

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
)

const (
	placeholder   = "-"
	columnPadding = 2
	disabledMark  = "(disabled)"
)

// FormatOptions controls how DoGet renders records.
type FormatOptions struct {
	Writer io.Writer
	JSON   bool
}

// FormatText writes DNS entries as space-aligned columns for terminal display.
func FormatText(writer io.Writer, entries []DNSEntry) error {
	if len(entries) == 0 {
		return nil
	}

	tabWriter := tabwriter.NewWriter(writer, 0, 0, columnPadding, ' ', 0)
	if _, err := fmt.Fprintln(tabWriter, "NAME\tVALUE\tTYPE\tTTL\tID"); err != nil {
		return err
	}

	for _, entry := range entries {
		line := strings.Join([]string{
			nameDisplay(entry),
			orPlaceholder(entry.Value),
			orPlaceholder(entry.Type),
			ttlDisplay(entry.TTL),
			orPlaceholder(entry.ID),
		}, "\t")
		if _, err := fmt.Fprintln(tabWriter, line); err != nil {
			return err
		}
	}

	return tabWriter.Flush()
}

// FormatJSON writes DNS entries as a JSON array for programmatic consumption.
func FormatJSON(writer io.Writer, entries []DNSEntry) error {
	if entries == nil {
		entries = []DNSEntry{}
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

// nameDisplay marks disabled records inline so a disabled entry is not mistaken
// for an active one when scanning the column.
func nameDisplay(entry DNSEntry) string {
	name := orPlaceholder(entry.Key)
	if !entry.Enabled {
		return name + " " + disabledMark
	}
	return name
}

func ttlDisplay(ttl int) string {
	if ttl <= 0 {
		return placeholder
	}
	return strconv.Itoa(ttl)
}

func orPlaceholder(value string) string {
	if value == "" {
		return placeholder
	}
	return value
}
