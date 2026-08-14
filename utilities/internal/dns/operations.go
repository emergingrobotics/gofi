package dns

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	gofi "github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/src/types"
)

// DNSEntry is the flattened view of a local DNS record that gofidns reports.
type DNSEntry struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	Enabled bool   `json:"enabled"`
}

// DeleteIdentifier selects the record(s) to delete. Exactly one field is set.
type DeleteIdentifier struct {
	ID   string
	Name string
	IP   string
}

// DeleteResult reports what a delete pass did.
type DeleteResult struct {
	Deleted int
	Errors  int
}

// ListRecords fetches all local DNS records and flattens them into entries
// sorted by key.
func ListRecords(ctx context.Context, client gofi.Client, site string) ([]DNSEntry, error) {
	records, err := client.DNS().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}
	return buildDNSEntries(records), nil
}

// buildDNSEntries maps and sorts records. Kept separate from ListRecords so it
// can be tested without a controller.
func buildDNSEntries(records []types.DNSRecord) []DNSEntry {
	entries := make([]DNSEntry, 0, len(records))
	for _, record := range records {
		entries = append(entries, DNSEntry{
			ID:      record.ID,
			Key:     record.Key,
			Value:   record.Value,
			Type:    record.RecordType,
			TTL:     record.TTL,
			Enabled: record.Enabled,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Key != entries[j].Key {
			return entries[i].Key < entries[j].Key
		}
		return entries[i].Value < entries[j].Value
	})

	return entries
}

// matchRecords returns every record the identifier selects. Matching by name or
// IP can legitimately return more than one record, so callers must decide
// whether a multi-match is acceptable.
func matchRecords(records []types.DNSRecord, identifier DeleteIdentifier) []types.DNSRecord {
	var matches []types.DNSRecord

	for _, record := range records {
		switch {
		case identifier.ID != "":
			if record.ID == identifier.ID {
				matches = append(matches, record)
			}
		case identifier.Name != "":
			if strings.EqualFold(record.Key, identifier.Name) {
				matches = append(matches, record)
			}
		case identifier.IP != "":
			if record.Value == identifier.IP {
				matches = append(matches, record)
			}
		}
	}

	return matches
}

// describeIdentifier renders the identifier for error messages.
func describeIdentifier(identifier DeleteIdentifier) string {
	switch {
	case identifier.ID != "":
		return "id " + identifier.ID
	case identifier.Name != "":
		return "name " + identifier.Name
	case identifier.IP != "":
		return "ip " + identifier.IP
	default:
		return "no identifier"
	}
}

// DoGet writes all local DNS records to the writer in the requested format.
func DoGet(ctx context.Context, client gofi.Client, site string, options FormatOptions) error {
	entries, err := ListRecords(ctx, client, site)
	if err != nil {
		return err
	}

	if options.JSON {
		return FormatJSON(options.Writer, entries)
	}
	return FormatText(options.Writer, entries)
}

// DoDel deletes the records selected by the identifier.
//
// A name or IP that selects several records is deleted as a set only when the
// caller passes force; otherwise it is an error, so an ambiguous identifier
// cannot silently remove more than the operator intended.
func DoDel(ctx context.Context, client gofi.Client, site string, identifier DeleteIdentifier, dryRun, force bool, progress io.Writer) (*DeleteResult, error) {
	records, err := client.DNS().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	matches := matchRecords(records, identifier)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no DNS record found for %s", describeIdentifier(identifier))
	}

	if len(matches) > 1 && !force {
		return nil, fmt.Errorf("%s matches %d records; pass --force to delete them all", describeIdentifier(identifier), len(matches))
	}

	result := &DeleteResult{}
	for _, record := range matches {
		if dryRun {
			fmt.Fprintf(progress, "  would delete: %s -> %s (id=%s)\n", record.Key, record.Value, record.ID)
			result.Deleted++
			continue
		}

		if err := client.DNS().Delete(ctx, site, record.ID); err != nil {
			fmt.Fprintf(progress, "  error: %s (id=%s): %v\n", record.Key, record.ID, err)
			result.Errors++
			continue
		}

		fmt.Fprintf(progress, "  deleted: %s -> %s (id=%s)\n", record.Key, record.Value, record.ID)
		result.Deleted++
	}

	return result, nil
}
