package main

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	gofi "github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/src/types"
)

// UserEntry is the flattened view of a known-client record that gofiuser
// reports.
type UserEntry struct {
	Name       string `json:"name,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	MAC        string `json:"mac"`
	FixedIP    string `json:"fixed_ip,omitempty"`
	UseFixedIP bool   `json:"use_fixed_ip"`
	NetworkID  string `json:"network_id,omitempty"`
	Blocked    bool   `json:"blocked,omitempty"`
	ID         string `json:"id"`
}

// DeleteIdentifier selects the client to remove. Exactly one field is set.
type DeleteIdentifier struct {
	MAC  string
	Name string
}

// DeleteResult reports what a removal did. Clearing the fixed IP and forgetting
// the client are tracked separately because the first can succeed while the
// second fails, which still frees the address.
type DeleteResult struct {
	ClearedFixedIP bool
	Forgot         bool
}

// ListUsers fetches all known-client records and flattens them into entries
// sorted by name, then MAC.
func ListUsers(ctx context.Context, client gofi.Client, site, filter string) ([]UserEntry, error) {
	users, err := client.Users().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return buildUserEntries(users, filter), nil
}

// buildUserEntries maps, filters, and sorts users. Kept separate from ListUsers
// so it can be tested without a controller.
func buildUserEntries(users []types.User, filter string) []UserEntry {
	entries := make([]UserEntry, 0, len(users))

	for _, user := range users {
		entry := UserEntry{
			Name:       user.Name,
			Hostname:   user.Hostname,
			MAC:        strings.ToLower(user.MAC),
			FixedIP:    user.FixedIP,
			UseFixedIP: user.UseFixedIP,
			NetworkID:  user.NetworkID,
			Blocked:    user.Blocked,
			ID:         user.ID,
		}

		if !matchesFilter(entry, filter) {
			continue
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Name != entries[j].Name {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].MAC < entries[j].MAC
	})

	return entries
}

// matchesFilter reports whether the entry contains the filter substring in any
// human-meaningful field. An empty filter matches everything.
func matchesFilter(entry UserEntry, filter string) bool {
	if filter == "" {
		return true
	}

	needle := strings.ToLower(filter)
	haystacks := []string{entry.Name, entry.Hostname, entry.MAC, entry.FixedIP}

	for _, haystack := range haystacks {
		if strings.Contains(strings.ToLower(haystack), needle) {
			return true
		}
	}

	return false
}

// findUser resolves the identifier to exactly one user record.
func findUser(users []types.User, identifier DeleteIdentifier) (*types.User, error) {
	var matches []*types.User

	for index := range users {
		user := &users[index]
		switch {
		case identifier.MAC != "":
			if strings.EqualFold(user.MAC, identifier.MAC) {
				matches = append(matches, user)
			}
		case identifier.Name != "":
			if user.Name == identifier.Name || user.Hostname == identifier.Name {
				matches = append(matches, user)
			}
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no client found for %s", describeIdentifier(identifier))
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("%s matches %d clients; use --mac to disambiguate", describeIdentifier(identifier), len(matches))
	}

	return matches[0], nil
}

// describeIdentifier renders the identifier for error messages.
func describeIdentifier(identifier DeleteIdentifier) string {
	switch {
	case identifier.MAC != "":
		return "mac " + identifier.MAC
	case identifier.Name != "":
		return "name " + identifier.Name
	default:
		return "no identifier"
	}
}

// DoList writes known-client records to the writer in the requested format.
func DoList(ctx context.Context, client gofi.Client, site string, filter string, options FormatOptions) error {
	entries, err := ListUsers(ctx, client, site, filter)
	if err != nil {
		return err
	}

	if options.JSON {
		return FormatJSON(options.Writer, entries)
	}
	return FormatText(options.Writer, entries)
}

// DoDel removes a known client.
//
// The fixed IP is cleared before the client is forgotten so the address is
// released even when the forget is rejected, rather than leaving a reservation
// stranded on a record the caller believes is gone.
func DoDel(ctx context.Context, client gofi.Client, site string, identifier DeleteIdentifier, dryRun bool, progress io.Writer) (*DeleteResult, error) {
	users, err := client.Users().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	user, err := findUser(users, identifier)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(progress, "  found: name=%q mac=%s fixed=%s\n", user.Name, user.MAC, user.FixedIP)

	result := &DeleteResult{}

	if dryRun {
		if user.UseFixedIP {
			fmt.Fprintf(progress, "  would clear fixed IP %s\n", user.FixedIP)
		}
		fmt.Fprintf(progress, "  would forget client %s\n", user.MAC)
		return result, nil
	}

	if user.UseFixedIP {
		if err := client.Users().ClearFixedIP(ctx, site, user.MAC); err != nil {
			return result, fmt.Errorf("failed to clear fixed IP: %w", err)
		}
		result.ClearedFixedIP = true
		fmt.Fprintf(progress, "  cleared fixed IP %s\n", user.FixedIP)
	}

	if err := client.Clients().Forget(ctx, site, user.MAC); err != nil {
		return result, fmt.Errorf("failed to forget client (fixed IP already released): %w", err)
	}
	result.Forgot = true
	fmt.Fprintf(progress, "  forgot client %s\n", user.MAC)

	return result, nil
}
