// Package profile captures a site's networks, WLANs, and fixed-IP
// reservations as JSON, and applies one back. Deliberately narrow
// (C-PROFILE-001): devices, firewall, routing, and port profiles are never
// captured, permanently, so "profile" can never silently grow to mean
// "everything on the site."
package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// ErrNoSite reports that Apply was given no target site. It is a sentinel so
// the CLI layer can classify it as a usage error (exit 2) without this
// package importing the CLI's error vocabulary.
var ErrNoSite = errors.New("no target site resolved: pass -S/--site or configure a target")

// Apply applies a profile in the fixed order networks, fixed IPs, WLANs
// (C-PROFILE-003): fixed IPs must land inside an existing subnet, so
// networks go first. A network or WLAN the target site lacks is reported
// and skipped rather than failing the whole run (C-PROFILE-005).
//
// site is the site to write to, resolved by the caller from -S/--site, the
// target's configured site, then "default" -- deliberately not p.Site. A
// profile is portable: it can be applied to a site other than the one it was
// captured from, and a hand-trimmed profile with no "site" field must not
// silently address site "". The mismatch is reported, not refused.
//
// Devices, firewall, and routing are never written, even if such a field
// somehow appears in a hand-edited profile file (I-PROFILE-001) -- Profile's
// struct has no field for them, so there is nothing for Apply to read.
func Apply(ctx context.Context, client gofi.Client, site string, p *Profile, dryRun bool, dnsDomainOverride string, progress io.Writer) error {
	if site == "" {
		return ErrNoSite
	}
	if p.Site != "" && p.Site != site {
		fmt.Fprintf(progress, "note: profile was captured from site %q, applying to site %q\n", p.Site, site)
	}

	targetNetworks, err := network.ListNetworks(ctx, client, site)
	if err != nil {
		return err
	}
	targetNetworksByName := make(map[string]network.NetworkEntry, len(targetNetworks))
	for _, n := range targetNetworks {
		targetNetworksByName[n.Name] = n
	}

	for _, n := range p.Networks {
		if _, ok := targetNetworksByName[n.Name]; !ok {
			fmt.Fprintf(progress, "skipping network %q: not present on the target site\n", n.Name)
			continue
		}
		fmt.Fprintf(progress, "network %q matches target; no write endpoint yet (C-NETWORK-004)\n", n.Name)
	}

	if len(p.FixedIPs) > 0 {
		result, err := ips.DoSet(ctx, client, site, p.FixedIPs, dnsDomainOverride, dryRun, false)
		if err != nil {
			return err
		}
		verb := "fixed IPs"
		if dryRun {
			verb = "fixed IPs (dry run)"
		}
		fmt.Fprintf(progress, "%s: %d created, %d updated, %d skipped, %d errors\n",
			verb, result.Created, result.Updated, result.Skipped, result.Errors)
	}

	targetWLANs, err := client.WLANs().List(ctx, site)
	if err != nil {
		return err
	}
	targetWLANsByName := make(map[string]struct{}, len(targetWLANs))
	for _, w := range targetWLANs {
		targetWLANsByName[w.Name] = struct{}{}
	}

	for _, w := range p.WLANs {
		if _, ok := targetWLANsByName[w.Name]; !ok {
			fmt.Fprintf(progress, "skipping WLAN %q: not present on the target site\n", w.Name)
			continue
		}
		if dryRun {
			fmt.Fprintf(progress, "would apply WLAN %q\n", w.Name)
			continue
		}
		fmt.Fprintf(progress, "WLAN %q: no write endpoint yet\n", w.Name)
	}

	return nil
}
