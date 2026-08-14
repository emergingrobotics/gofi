// Package profile captures a site's networks, WLANs, and fixed-IP
// reservations as JSON, and applies one back. Deliberately narrow
// (C-PROFILE-001): devices, firewall, routing, and port profiles are never
// captured, permanently, so "profile" can never silently grow to mean
// "everything on the site."
package profile

import (
	"context"
	"encoding/json"
	"io"

	gofi "github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/utilities/internal/ips"
	"github.com/unifi-go/gofi/utilities/internal/network"
)

// WLANEntry is the portable subset of a WLAN's configuration.
type WLANEntry struct {
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	Passphrase string `json:"passphrase,omitempty"`
}

// Profile is a captured site: networks, WLANs, fixed IPs. Nothing else,
// permanently (C-PROFILE-001).
type Profile struct {
	Site     string                 `json:"site"`
	Captured string                 `json:"captured"`
	Networks []network.NetworkEntry `json:"networks"`
	WLANs    []WLANEntry            `json:"wlans"`
	FixedIPs []ips.HostEntry        `json:"fixed_ips"`
}

// Capture reads the three portable sections from the site. withKeys
// controls whether WLAN passphrases are included in cleartext
// (C-PROFILE-002); when false, Passphrase is left empty on every entry.
func Capture(ctx context.Context, client gofi.Client, site string, withKeys bool) (*Profile, error) {
	networks, err := network.ListNetworks(ctx, client, site)
	if err != nil {
		return nil, err
	}

	wlans, err := client.WLANs().List(ctx, site)
	if err != nil {
		return nil, err
	}
	wlanEntries := make([]WLANEntry, 0, len(wlans))
	for _, w := range wlans {
		entry := WLANEntry{Name: w.Name, Enabled: w.Enabled}
		if withKeys {
			entry.Passphrase = w.Passphrase
		}
		wlanEntries = append(wlanEntries, entry)
	}

	fixedIPs, err := ips.DoGetEntries(ctx, client, site, "")
	if err != nil {
		return nil, err
	}

	return &Profile{
		Site:     site,
		Networks: networks,
		WLANs:    wlanEntries,
		FixedIPs: fixedIPs,
	}, nil
}

// Write serializes the profile as indented JSON.
func (p *Profile) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

// ReadProfile parses a profile written by Write.
func ReadProfile(r io.Reader) (*Profile, error) {
	var p Profile
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}
