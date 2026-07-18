package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/types"
)

type SetResult struct {
	Created int
	Updated int
	Skipped int
	Errors  int
}

type DeleteIdentifier struct {
	Name string
	MAC  string
	IP   string
}

func DoGet(ctx context.Context, client gofi.Client, site string, writer io.Writer, options FormatOptions) error {
	users, err := client.Users().List(ctx, site)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	// Cross-reference DNS records to warn about hostname mismatches
	dnsRecords, _ := client.DNS().List(ctx, site)
	ipToDNSHostname := make(map[string]string)
	for _, record := range dnsRecords {
		if record.RecordType == types.DNSRecordTypeA && record.Enabled {
			ipToDNSHostname[record.Value] = record.Key
		}
	}

	// Networks give the per-network domain used to build each FQDN for the
	// live drift check below. Non-fatal: without them, drift is simply not audited.
	networks, _ := client.Networks().List(ctx, site)

	var entries []HostEntry
	for _, user := range users {
		if !user.UseFixedIP || user.FixedIP == "" {
			continue
		}

		hostname := resolveHostname(user)

		// Warn if DNS record hostname differs from user hostname. DNS keys are
		// FQDNs (host.domain), so compare on the first label, not the full key.
		if dnsHostname, ok := ipToDNSHostname[user.FixedIP]; ok && !dnsHostnameMatches(dnsHostname, hostname) {
			fmt.Fprintf(os.Stderr, "  warning: %s has DNS hostname %q but user hostname %q\n", user.FixedIP, dnsHostname, hostname)
		}

		// Audit for value drift: the name resolving to an IP other than the
		// fixed IP means DNS and the reservation disagree.
		fqdn := hostname
		if domain := domainForIP(networks, user.FixedIP); domain != "" {
			fqdn = hostname + "." + domain
		}
		if addrs, _ := resolveFQDN(ctx, fqdn); len(addrs) > 0 && !containsString(addrs, user.FixedIP) {
			fmt.Fprintf(os.Stderr, "  warning: %s resolves to %s (drift from fixed IP %s)\n", fqdn, strings.Join(addrs, ","), user.FixedIP)
		}

		entries = append(entries, HostEntry{
			Hostname: hostname,
			MAC:      strings.ToLower(user.MAC),
			IP:       user.FixedIP,
		})
	}

	return Format(writer, entries, options)
}

func resolveHostname(user types.User) string {
	if user.Name != "" && isDNSSafe(user.Name) {
		return user.Name
	}
	if user.Hostname != "" && isDNSSafe(user.Hostname) {
		return user.Hostname
	}
	return strings.ReplaceAll(user.MAC, ":", "-")
}

func displayName(user *types.User) string {
	if user.Name != "" {
		return user.Name
	}
	if user.Hostname != "" {
		return user.Hostname
	}
	return user.MAC
}

func DoSet(ctx context.Context, client gofi.Client, site string, entries []HostEntry, dryRun, force bool) (*SetResult, error) {
	result := &SetResult{}

	users, err := client.Users().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	existingByMAC := make(map[string]*types.User)
	for index := range users {
		user := &users[index]
		existingByMAC[strings.ToLower(user.MAC)] = user
	}

	networks, err := client.Networks().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	for _, entry := range entries {
		macAddress := strings.ToLower(entry.MAC)

		if existingUser, ok := existingByMAC[macAddress]; ok {
			unchanged := existingUser.UseFixedIP && existingUser.FixedIP == entry.IP &&
				(existingUser.Name == entry.Hostname || existingUser.Hostname == entry.Hostname)

			if unchanged && !force {
				fmt.Fprintf(os.Stderr, "  skip: %s %s %s (user record unchanged)\n", entry.Hostname, entry.IP, macAddress)
				if err := ensureDNSRecord(ctx, client, site, entry.Hostname, domainForIP(networks, entry.IP), entry.IP, dryRun); err != nil {
					fmt.Fprintf(os.Stderr, "  warning: %s: DNS record failed: %v\n", entry.Hostname, err)
				}
				result.Skipped++
				continue
			}

			if dryRun {
				fmt.Fprintf(os.Stderr, "  would update: %s %s %s\n", entry.Hostname, entry.IP, macAddress)
				if err := ensureDNSRecord(ctx, client, site, entry.Hostname, domainForIP(networks, entry.IP), entry.IP, true); err != nil {
					fmt.Fprintf(os.Stderr, "  warning: %s: DNS record failed: %v\n", entry.Hostname, err)
				}
				result.Updated++
				continue
			}

			network, err := detectNetwork(networks, entry.IP)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s %s: %v\n", entry.IP, macAddress, err)
				result.Errors++
				continue
			}

			existingUser.Name = entry.Hostname
			existingUser.UseFixedIP = true
			existingUser.FixedIP = entry.IP
			existingUser.NetworkID = network.ID
			if _, err := client.Users().Update(ctx, site, existingUser); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s %s: failed to update user: %v\n", entry.IP, macAddress, err)
				result.Errors++
				continue
			}

			if err := ensureDNSRecord(ctx, client, site, entry.Hostname, network.DomainName, entry.IP, false); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: %s: DNS record failed: %v\n", entry.Hostname, err)
			}

			fmt.Fprintf(os.Stderr, "  updated: %s %s %s\n", entry.Hostname, entry.IP, macAddress)
			result.Updated++
			continue
		}

		if dryRun {
			fmt.Fprintf(os.Stderr, "  would create: %s %s %s\n", entry.Hostname, entry.IP, macAddress)
			if err := ensureDNSRecord(ctx, client, site, entry.Hostname, domainForIP(networks, entry.IP), entry.IP, true); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: %s: DNS record failed: %v\n", entry.Hostname, err)
			}
			result.Created++
			continue
		}

		network, err := detectNetwork(networks, entry.IP)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error: %s %s: %v\n", entry.IP, macAddress, err)
			result.Errors++
			continue
		}

		existingUser, _ := client.Users().GetByMAC(ctx, site, macAddress)
		if existingUser != nil {
			existingUser.Name = entry.Hostname
			existingUser.UseFixedIP = true
			existingUser.FixedIP = entry.IP
			existingUser.NetworkID = network.ID
			if _, err := client.Users().Update(ctx, site, existingUser); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s %s: failed to update: %v\n", entry.IP, macAddress, err)
				result.Errors++
				continue
			}
			fmt.Fprintf(os.Stderr, "  updated: %s %s %s (added fixed IP to existing user)\n", entry.Hostname, entry.IP, macAddress)
			result.Updated++
		} else {
			newUser := &types.User{
				MAC:        macAddress,
				Name:       entry.Hostname,
				UseFixedIP: true,
				FixedIP:    entry.IP,
				NetworkID:  network.ID,
			}
			if _, err := client.Users().Create(ctx, site, newUser); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s %s: failed to create: %v\n", entry.IP, macAddress, err)
				result.Errors++
				continue
			}
			fmt.Fprintf(os.Stderr, "  created: %s %s %s\n", entry.Hostname, entry.IP, macAddress)
			result.Created++
		}

		if err := ensureDNSRecord(ctx, client, site, entry.Hostname, network.DomainName, entry.IP, false); err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %s: DNS record failed: %v\n", entry.Hostname, err)
		}
	}

	return result, nil
}

func DoAdd(ctx context.Context, client gofi.Client, site string, entry *HostEntry, force bool) error {
	if !force {
		if err := checkAddConflicts(ctx, client, site, entry); err != nil {
			return err
		}
	}

	networks, err := client.Networks().List(ctx, site)
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	network, err := detectNetwork(networks, entry.IP)
	if err != nil {
		return err
	}

	macAddress := strings.ToLower(entry.MAC)
	existingUser, _ := client.Users().GetByMAC(ctx, site, macAddress)

	if existingUser != nil {
		existingUser.Name = entry.Hostname
		existingUser.UseFixedIP = true
		existingUser.FixedIP = entry.IP
		existingUser.NetworkID = network.ID
		if _, err := client.Users().Update(ctx, site, existingUser); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		newUser := &types.User{
			MAC:        macAddress,
			Name:       entry.Hostname,
			UseFixedIP: true,
			FixedIP:    entry.IP,
			NetworkID:  network.ID,
		}
		if _, err := client.Users().Create(ctx, site, newUser); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	if err := ensureDNSRecord(ctx, client, site, entry.Hostname, network.DomainName, entry.IP, false); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: DNS record failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "The fixed IP assignment was created, but DNS must be configured manually.\n")
	}

	fmt.Printf("Created: %s %s %s\n", entry.Hostname, entry.MAC, entry.IP)
	return nil
}

func DoDel(ctx context.Context, client gofi.Client, site string, identifier DeleteIdentifier, force, keepDNS bool) error {
	user, err := findUserByIdentifier(ctx, client, site, identifier)
	if err != nil {
		return err
	}

	userName := displayName(user)

	fmt.Printf("Found: %s %s %s\n", userName, user.MAC, user.FixedIP)

	if !keepDNS && user.UseFixedIP && user.FixedIP != "" {
		dnsRecords, _ := client.DNS().GetByIP(ctx, site, user.FixedIP)
		for _, record := range dnsRecords {
			if err := client.DNS().Delete(ctx, site, record.ID); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete DNS record %s: %v\n", record.Key, err)
			} else {
				fmt.Printf("  Deleted DNS record: %s -> %s\n", record.Key, record.Value)
			}
		}
	}

	if force {
		if err := client.Users().Delete(ctx, site, user.ID); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
		fmt.Printf("  Deleted user entry for %s\n", userName)
	} else {
		if !user.UseFixedIP {
			return fmt.Errorf("user %s has no fixed IP to remove", userName)
		}
		if err := client.Users().ClearFixedIP(ctx, site, user.MAC); err != nil {
			return fmt.Errorf("failed to clear fixed IP: %w", err)
		}
		fmt.Printf("  Removed fixed IP assignment\n")
	}

	return nil
}

func checkAddConflicts(ctx context.Context, client gofi.Client, site string, entry *HostEntry) error {
	users, err := client.Users().List(ctx, site)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	macAddress := strings.ToLower(entry.MAC)
	for _, user := range users {
		if user.UseFixedIP && user.FixedIP == entry.IP && strings.ToLower(user.MAC) != macAddress {
			return fmt.Errorf("IP %s is already assigned to %s (%s)\nUse --force to override", entry.IP, displayName(&user), user.MAC)
		}
		if strings.ToLower(user.MAC) == macAddress && user.UseFixedIP && user.FixedIP != entry.IP {
			return fmt.Errorf("MAC %s already has fixed IP %s (assigned to %s)\nUse --force to override", macAddress, user.FixedIP, displayName(&user))
		}
	}

	networks, err := client.Networks().List(ctx, site)
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}
	dnsKey := entry.Hostname
	if domain := domainForIP(networks, entry.IP); domain != "" {
		dnsKey = entry.Hostname + "." + domain
	}
	existingDNS, _ := client.DNS().GetByName(ctx, site, dnsKey)
	if existingDNS != nil && existingDNS.Value != entry.IP {
		return fmt.Errorf("DNS record %s already points to %s (not %s)\nUse --force to override", dnsKey, existingDNS.Value, entry.IP)
	}

	return nil
}

func findUserByIdentifier(ctx context.Context, client gofi.Client, site string, identifier DeleteIdentifier) (*types.User, error) {
	if identifier.MAC != "" {
		user, err := client.Users().GetByMAC(ctx, site, strings.ToLower(identifier.MAC))
		if err != nil {
			return nil, fmt.Errorf("no user found with MAC %s", identifier.MAC)
		}
		return user, nil
	}

	users, err := client.Users().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	if identifier.IP != "" {
		var matches []*types.User
		for index := range users {
			if users[index].UseFixedIP && users[index].FixedIP == identifier.IP {
				matches = append(matches, &users[index])
			}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("no user found with fixed IP %s", identifier.IP)
		}
		if len(matches) > 1 {
			fmt.Fprintf(os.Stderr, "Multiple users found with IP %s:\n", identifier.IP)
			for _, match := range matches {
				fmt.Fprintf(os.Stderr, "  %s (%s)\n", displayName(match), match.MAC)
			}
			return nil, fmt.Errorf("multiple matches for IP %s, use --mac to be specific", identifier.IP)
		}
		return matches[0], nil
	}

	if identifier.Name != "" {
		for index := range users {
			if strings.EqualFold(users[index].Name, identifier.Name) || strings.EqualFold(users[index].Hostname, identifier.Name) {
				return &users[index], nil
			}
		}
		dnsRecord, _ := client.DNS().GetByName(ctx, site, identifier.Name)
		if dnsRecord != nil && dnsRecord.Value != "" {
			for index := range users {
				if users[index].UseFixedIP && users[index].FixedIP == dnsRecord.Value {
					return &users[index], nil
				}
			}
		}
		return nil, fmt.Errorf("no user found with name %q", identifier.Name)
	}

	return nil, fmt.Errorf("no identifier specified (use --name, --mac, or --ip)")
}

// dnsResolver queries the UDM's own DNS, returning the A records it serves for
// a name from BOTH device-local (DHCP) DNS and static records. This is the only
// way to see device-local entries, which are invisible to the static-record API.
type dnsResolver func(ctx context.Context, fqdn string) ([]string, error)

// resolveFQDN is the live UDM resolver. main() installs one pointed at the UDM;
// the default treats every name as unresolved so tests and misconfiguration fall
// back to attempting a static record rather than silently skipping.
var resolveFQDN dnsResolver = func(ctx context.Context, fqdn string) ([]string, error) {
	return nil, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// isOverlapError reports whether a DNS create failed because the name is already
// served by the UDM's device-local DNS. The UDM rejects such static records with
// api.err.StaticDnsOverlapsWithDeviceLocalDns.
func isOverlapError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "StaticDnsOverlaps")
}

// ensureDNSRecord makes the FQDN (hostname + "." + domain) resolve to the given
// IP. It prefers what already works: if the UDM's device-local DNS already
// answers the name correctly, no static record is created (the UDM would reject
// it as overlapping anyway). A static record is created only when the name does
// not resolve at all. If the name resolves to a different IP via device-local
// DNS, that is reported as drift because a static record cannot override it.
// Any legacy bare-hostname static record we previously wrote is removed first.
// When domain is empty the FQDN degrades to the bare hostname.
func ensureDNSRecord(ctx context.Context, client gofi.Client, site, hostname, domain, ipAddress string, dryRun bool) error {
	fqdn := hostname
	if domain != "" {
		fqdn = hostname + "." + domain
	}

	if fqdn != hostname {
		if stale, _ := client.DNS().GetByName(ctx, site, hostname); stale != nil && stale.Key == hostname {
			if dryRun {
				fmt.Fprintf(os.Stderr, "    would replace bare DNS record: %s -> %s\n", stale.Key, stale.Value)
			} else {
				if err := client.DNS().Delete(ctx, site, stale.ID); err != nil {
					return fmt.Errorf("failed to delete stale DNS record %s: %w", stale.Key, err)
				}
			}
		}
	}

	// A static record we own: reconcile its value.
	existing, _ := client.DNS().GetByName(ctx, site, fqdn)
	if existing != nil {
		if existing.Value == ipAddress {
			return nil
		}
		if dryRun {
			fmt.Fprintf(os.Stderr, "    would update DNS record: %s -> %s\n", fqdn, ipAddress)
			return nil
		}
		existing.Value = ipAddress
		if _, err := client.DNS().Update(ctx, site, existing); err != nil {
			return fmt.Errorf("failed to update DNS record: %w", err)
		}
		fmt.Fprintf(os.Stderr, "    updated DNS record: %s -> %s\n", fqdn, ipAddress)
		return nil
	}

	// No static record. Consult the UDM's live view (device-local DNS + static).
	addrs, _ := resolveFQDN(ctx, fqdn)
	if containsString(addrs, ipAddress) {
		fmt.Fprintf(os.Stderr, "    ok: %s already resolves to %s (device local DNS)\n", fqdn, ipAddress)
		return nil
	}
	if len(addrs) > 0 {
		return fmt.Errorf("%s resolves to %s via device local DNS, not %s; correct the DHCP reservation", fqdn, strings.Join(addrs, ","), ipAddress)
	}

	if dryRun {
		fmt.Fprintf(os.Stderr, "    would create DNS record: %s -> %s\n", fqdn, ipAddress)
		return nil
	}

	record := &types.DNSRecord{
		Key:        fqdn,
		Value:      ipAddress,
		RecordType: types.DNSRecordTypeA,
		Enabled:    true,
	}
	if _, err := client.DNS().Create(ctx, site, record); err != nil {
		if isOverlapError(err) {
			// Device-local DNS registered the name between our lookup and create.
			fmt.Fprintf(os.Stderr, "    ok: %s served by device local DNS (static overlaps)\n", fqdn)
			return nil
		}
		return fmt.Errorf("failed to create DNS record: %w", err)
	}
	fmt.Fprintf(os.Stderr, "    created DNS record: %s -> %s\n", fqdn, ipAddress)
	return nil
}

func detectNetwork(networks []types.Network, ipAddress string) (*types.Network, error) {
	parsedIP := net.ParseIP(ipAddress)
	for index := range networks {
		network := &networks[index]
		if network.IPSubnet == "" {
			continue
		}
		_, subnet, err := net.ParseCIDR(network.IPSubnet)
		if err != nil {
			continue
		}
		if subnet.Contains(parsedIP) {
			return network, nil
		}
	}
	return nil, fmt.Errorf("no network found containing IP %s", ipAddress)
}

// domainForIP returns the local domain of the network that owns the IP, or an
// empty string if the IP is not in any network or that network has no domain.
// Used on the DNS-repair path where a missing network is non-fatal.
func domainForIP(networks []types.Network, ipAddress string) string {
	network, err := detectNetwork(networks, ipAddress)
	if err != nil {
		return ""
	}
	return network.DomainName
}

// dnsHostnameMatches reports whether a DNS record key belongs to the given
// hostname, accepting either the bare hostname or its FQDN form
// (hostname + "." + domain). The first label of the key must equal the hostname.
func dnsHostnameMatches(dnsKey, hostname string) bool {
	if dnsKey == hostname {
		return true
	}
	firstLabel := dnsKey
	if index := strings.Index(dnsKey, "."); index >= 0 {
		firstLabel = dnsKey[:index]
	}
	return firstLabel == hostname
}
