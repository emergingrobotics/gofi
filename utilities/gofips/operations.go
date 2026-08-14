package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/src/types"
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

func DoGet(ctx context.Context, client gofi.Client, site, dnsDomainOverride string, writer io.Writer, options FormatOptions) error {
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

	// Networks supply the domain used to build each FQDN for the live drift
	// check below. Non-fatal: export is a read, so an unresolvable domain costs
	// the audit, not the export.
	networks, _ := client.Networks().List(ctx, site)
	domain, domainErr := dnsDomain(networks, dnsDomainOverride)
	if domainErr != nil {
		fmt.Fprintf(os.Stderr, "  warning: DNS drift not audited: %v\n", domainErr)
	}

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
		if domainErr == nil {
			fqdn := hostname + "." + domain
			if addrs, _ := resolveFQDN(ctx, fqdn); len(addrs) > 0 && !containsString(addrs, user.FixedIP) {
				fmt.Fprintf(os.Stderr, "  warning: %s resolves to %s (drift from fixed IP %s)\n", fqdn, strings.Join(addrs, ","), user.FixedIP)
			}
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

func DoSet(ctx context.Context, client gofi.Client, site string, entries []HostEntry, dnsDomainOverride string, dryRun, force bool) (*SetResult, error) {
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

	// Resolve the DNS domain before writing anything. Continuing without one
	// would write bare keys that no qualified query ever reaches, so the run
	// would report success while changing nothing a client can see.
	domain, err := dnsDomain(networks, dnsDomainOverride)
	if err != nil {
		return nil, err
	}
	reconciler, err := newDNSReconciler(ctx, client, site, domain, dryRun)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		macAddress := strings.ToLower(entry.MAC)

		if existingUser, ok := existingByMAC[macAddress]; ok {
			unchanged := existingUser.UseFixedIP && existingUser.FixedIP == entry.IP &&
				(existingUser.Name == entry.Hostname || existingUser.Hostname == entry.Hostname)

			if unchanged && !force {
				fmt.Fprintf(os.Stderr, "  skip: %s %s %s (user record unchanged)\n", entry.Hostname, entry.IP, macAddress)
				if err := reconciler.ensure(ctx, entry.Hostname, entry.IP); err != nil {
					fmt.Fprintf(os.Stderr, "  warning: %s: DNS record failed: %v\n", entry.Hostname, err)
				}
				result.Skipped++
				continue
			}

			if dryRun {
				fmt.Fprintf(os.Stderr, "  would update: %s %s %s\n", entry.Hostname, entry.IP, macAddress)
				if err := reconciler.ensure(ctx, entry.Hostname, entry.IP); err != nil {
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

			if err := reconciler.ensure(ctx, entry.Hostname, entry.IP); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: %s: DNS record failed: %v\n", entry.Hostname, err)
			}

			fmt.Fprintf(os.Stderr, "  updated: %s %s %s\n", entry.Hostname, entry.IP, macAddress)
			result.Updated++
			continue
		}

		if dryRun {
			fmt.Fprintf(os.Stderr, "  would create: %s %s %s\n", entry.Hostname, entry.IP, macAddress)
			if err := reconciler.ensure(ctx, entry.Hostname, entry.IP); err != nil {
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

		if err := reconciler.ensure(ctx, entry.Hostname, entry.IP); err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %s: DNS record failed: %v\n", entry.Hostname, err)
		}
	}

	return result, nil
}

func DoAdd(ctx context.Context, client gofi.Client, site string, entry *HostEntry, dnsDomainOverride string, force bool) error {
	networks, err := client.Networks().List(ctx, site)
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	domain, err := dnsDomain(networks, dnsDomainOverride)
	if err != nil {
		return err
	}

	if !force {
		if err := checkAddConflicts(ctx, client, site, entry, domain); err != nil {
			return err
		}
	}

	network, err := detectNetwork(networks, entry.IP)
	if err != nil {
		return err
	}

	reconciler, err := newDNSReconciler(ctx, client, site, domain, false)
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

	if err := reconciler.ensure(ctx, entry.Hostname, entry.IP); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: DNS record failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "The fixed IP assignment was created, but DNS must be configured manually.\n")
	}

	fmt.Printf("Created: %s %s %s\n", entry.Hostname, entry.MAC, entry.IP)
	return nil
}

func DoDel(ctx context.Context, client gofi.Client, site string, identifier DeleteIdentifier, dnsDomainOverride string, force, keepDNS bool) error {
	user, err := findUserByIdentifier(ctx, client, site, identifier)
	if err != nil {
		return err
	}

	userName := displayName(user)

	fmt.Printf("Found: %s %s %s\n", userName, user.MAC, user.FixedIP)

	if !keepDNS {
		// Matching on the hostname as well as the address catches records left
		// under an address the host no longer holds -- the ones a value-only
		// lookup silently walks past.
		networks, _ := client.Networks().List(ctx, site)
		domain, _ := dnsDomain(networks, dnsDomainOverride)
		reconciler, err := newDNSReconciler(ctx, client, site, domain, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		} else if err := reconciler.removeHost(ctx, resolveHostname(*user), user.FixedIP, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
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

func checkAddConflicts(ctx context.Context, client gofi.Client, site string, entry *HostEntry, domain string) error {
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

	dnsKey := entry.Hostname + "." + domain
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
// dnsReconciler owns a site's static DNS records for the life of one command.
//
// Records are listed once and every lookup scans that snapshot by hostname.
// Locating a record by exact key or by the host's current IP -- as earlier
// versions did -- makes precisely the records that need repair invisible, since
// a record needing repair is by definition one that disagrees with current
// state.
type dnsReconciler struct {
	client  gofi.Client
	site    string
	domain  string
	dryRun  bool
	records []types.DNSRecord
}

// newDNSReconciler snapshots the site's records. A domain is required to write
// a key but not to find one by hostname, so an empty domain is accepted here
// and rejected by ensure.
func newDNSReconciler(ctx context.Context, client gofi.Client, site, domain string, dryRun bool) (*dnsReconciler, error) {
	records, err := client.DNS().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	return &dnsReconciler{
		client:  client,
		site:    site,
		domain:  domain,
		dryRun:  dryRun,
		records: records,
	}, nil
}

func (r *dnsReconciler) fqdn(hostname string) string {
	return hostname + "." + r.domain
}

// forHost returns every A record whose first label is hostname, whatever its
// suffix, so a record stranded under an old domain is still reachable.
func (r *dnsReconciler) forHost(hostname string) []types.DNSRecord {
	var matches []types.DNSRecord
	for _, record := range r.records {
		if record.RecordType != types.DNSRecordTypeA {
			continue
		}
		if dnsHostnameMatches(record.Key, hostname) {
			matches = append(matches, record)
		}
	}
	return matches
}

// ensure makes hostname resolve to ipAddress and nothing else.
func (r *dnsReconciler) ensure(ctx context.Context, hostname, ipAddress string) error {
	if r.domain == "" {
		return fmt.Errorf("no DNS domain resolved; a bare DNS key is never served for a qualified query")
	}

	fqdn := r.fqdn(hostname)

	var current *types.DNSRecord
	var strays []types.DNSRecord
	for _, record := range r.forHost(hostname) {
		if record.Key == fqdn && current == nil {
			kept := record
			current = &kept
			continue
		}
		strays = append(strays, record)
	}

	// A stray is a bare key, a key under a suffix we no longer use, or a second
	// record at the same key. None can be reached by a later lookup, and a
	// duplicate makes the name round-robin onto a dead address.
	for _, stray := range strays {
		if err := r.remove(ctx, stray, "stale"); err != nil {
			return err
		}
	}

	if current != nil {
		if current.Value == ipAddress {
			return nil
		}
		if r.dryRun {
			fmt.Fprintf(os.Stderr, "    would update DNS record: %s -> %s\n", fqdn, ipAddress)
			return nil
		}
		updated := *current
		updated.Value = ipAddress
		if _, err := r.client.DNS().Update(ctx, r.site, &updated); err != nil {
			return fmt.Errorf("failed to update DNS record: %w", err)
		}
		r.replace(updated)
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

	if r.dryRun {
		fmt.Fprintf(os.Stderr, "    would create DNS record: %s -> %s\n", fqdn, ipAddress)
		return nil
	}

	created, err := r.client.DNS().Create(ctx, r.site, &types.DNSRecord{
		Key:        fqdn,
		Value:      ipAddress,
		RecordType: types.DNSRecordTypeA,
		Enabled:    true,
	})
	if err != nil {
		if isOverlapError(err) {
			// Device-local DNS registered the name between our lookup and create.
			fmt.Fprintf(os.Stderr, "    ok: %s served by device local DNS (static overlaps)\n", fqdn)
			return nil
		}
		return fmt.Errorf("failed to create DNS record: %w", err)
	}
	r.records = append(r.records, *created)
	fmt.Fprintf(os.Stderr, "    created DNS record: %s -> %s\n", fqdn, ipAddress)
	return nil
}

// removeHost deletes every record belonging to a host: those keyed on its name
// under any suffix, plus any record still pointing at the address it is giving
// up. Progress goes to writer because delete reports to stdout.
func (r *dnsReconciler) removeHost(ctx context.Context, hostname, ipAddress string, writer io.Writer) error {
	doomed := r.forHost(hostname)
	claimed := make(map[string]bool, len(doomed))
	for _, record := range doomed {
		claimed[record.ID] = true
	}
	if ipAddress != "" {
		for _, record := range r.records {
			if record.Value == ipAddress && !claimed[record.ID] {
				claimed[record.ID] = true
				doomed = append(doomed, record)
			}
		}
	}

	for _, record := range doomed {
		if r.dryRun {
			fmt.Fprintf(writer, "  Would delete DNS record: %s -> %s\n", record.Key, record.Value)
			continue
		}
		if err := r.client.DNS().Delete(ctx, r.site, record.ID); err != nil {
			return fmt.Errorf("failed to delete DNS record %s: %w", record.Key, err)
		}
		r.drop(record.ID)
		fmt.Fprintf(writer, "  Deleted DNS record: %s -> %s\n", record.Key, record.Value)
	}
	return nil
}

func (r *dnsReconciler) remove(ctx context.Context, record types.DNSRecord, reason string) error {
	if r.dryRun {
		fmt.Fprintf(os.Stderr, "    would delete %s DNS record: %s -> %s\n", reason, record.Key, record.Value)
		return nil
	}
	if err := r.client.DNS().Delete(ctx, r.site, record.ID); err != nil {
		return fmt.Errorf("failed to delete %s DNS record %s: %w", reason, record.Key, err)
	}
	r.drop(record.ID)
	fmt.Fprintf(os.Stderr, "    deleted %s DNS record: %s -> %s\n", reason, record.Key, record.Value)
	return nil
}

func (r *dnsReconciler) replace(record types.DNSRecord) {
	for index := range r.records {
		if r.records[index].ID == record.ID {
			r.records[index] = record
			return
		}
	}
}

func (r *dnsReconciler) drop(id string) {
	for index := range r.records {
		if r.records[index].ID == id {
			r.records = append(r.records[:index], r.records[index+1:]...)
			return
		}
	}
}

// ensureDNSRecord reconciles a single hostname against a freshly listed record
// set. Batch callers build one reconciler instead, so the record list is
// fetched once per command rather than once per host.
func ensureDNSRecord(ctx context.Context, client gofi.Client, site, hostname, domain, ipAddress string, dryRun bool) error {
	reconciler, err := newDNSReconciler(ctx, client, site, domain, dryRun)
	if err != nil {
		return err
	}
	return reconciler.ensure(ctx, hostname, ipAddress)
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

// dnsDomain resolves the single suffix used for every static DNS key.
//
// Static DNS records carry no network association -- the controller stores only
// a key and a value -- so deriving the suffix from a per-network domain_name
// breaks the moment a host moves between VLANs: the key changes, and the record
// left under the old suffix becomes unreachable to every later lookup. A network
// with no domain_name yields a bare key that the resolver never serves for a
// qualified query, which fails silently. One site-wide domain keeps a host's key
// stable wherever the host lives.
func dnsDomain(networks []types.Network, override string) (string, error) {
	if override != "" {
		return override, nil
	}

	var domains []string
	for _, network := range networks {
		if network.DomainName == "" || containsString(domains, network.DomainName) {
			continue
		}
		domains = append(domains, network.DomainName)
	}

	switch len(domains) {
	case 1:
		return domains[0], nil
	case 0:
		return "", fmt.Errorf("no network defines a domain name; set one on a network or pass --dns-domain")
	default:
		return "", fmt.Errorf("networks define %d different domain names (%s); pass --dns-domain to choose one",
			len(domains), strings.Join(domains, ", "))
	}
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
