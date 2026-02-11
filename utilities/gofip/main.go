package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/types"
)

const (
	envUsername = "UNIFI_USERNAME"
	envPassword = "UNIFI_PASSWORD"
	envUDMIP    = "UNIFI_UDM_IP"
)

var macRegex = regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)

type entry struct {
	IP  string
	MAC string
}

func main() {
	var (
		host     = flag.String("host", "", "UDM Pro host address")
		port     = flag.Int("port", 443, "UDM Pro port")
		site     = flag.String("site", "default", "Site name")
		insecure = flag.Bool("insecure", false, "Skip TLS certificate verification")
		get      = flag.Bool("get", false, "Export fixed IP assignments to stdout")
		set      = flag.Bool("set", false, "Import fixed IP assignments from file or stdin")
	)

	flag.StringVar(host, "H", "", "UDM Pro host address (shorthand)")
	flag.IntVar(port, "p", 443, "UDM Pro port (shorthand)")
	flag.StringVar(site, "S", "default", "Site name (shorthand)")
	flag.BoolVar(insecure, "k", false, "Skip TLS certificate verification (shorthand)")
	flag.BoolVar(get, "g", false, "Export fixed IP assignments to stdout (shorthand)")
	flag.BoolVar(set, "s", false, "Import fixed IP assignments from file or stdin (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] --get\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s [options] --set [filename]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Manage fixed IP (DHCP reservation) assignments on a UniFi UDM Pro.\n\n")
		fmt.Fprintf(os.Stderr, "Modes:\n")
		fmt.Fprintf(os.Stderr, "  -g, --get\t\tExport current assignments to stdout\n")
		fmt.Fprintf(os.Stderr, "  -s, --set\t\tImport assignments from file or stdin\n\n")
		fmt.Fprintf(os.Stderr, "Connection:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\tUDM Pro host address (or set %s)\n", envUDMIP)
		fmt.Fprintf(os.Stderr, "  -p, --port int\tUDM Pro port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -S, --site string\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure\tSkip TLS certificate verification\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tUsername (required)\n", envUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword (required)\n", envPassword)
		fmt.Fprintf(os.Stderr, "  %s\tUDM host (fallback for -H)\n\n", envUDMIP)
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -g > hosts.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -s hosts.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  cat hosts.txt | %s -H 192.168.1.1 -k -s\n", os.Args[0])
	}

	flag.Parse()

	// Validate mode
	if *get == *set {
		if *get {
			exitError("specify only one of --get or --set, not both")
		}
		fmt.Fprintf(os.Stderr, "Error: specify --get or --set\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Resolve host
	if *host == "" {
		*host = os.Getenv(envUDMIP)
	}
	if *host == "" {
		exitError("--host is required (or set " + envUDMIP + ")")
	}

	// Credentials
	username := os.Getenv(envUsername)
	password := os.Getenv(envPassword)
	if username == "" {
		exitError(envUsername + " environment variable is required")
	}
	if password == "" {
		exitError(envPassword + " environment variable is required")
	}

	config := &gofi.Config{
		Host:          *host,
		Port:          *port,
		Username:      username,
		Password:      password,
		Site:          *site,
		SkipTLSVerify: *insecure,
	}

	if *get {
		doGet(config, *site)
	} else {
		doSet(config, *site, flag.Args())
	}
}

// doGet exports current fixed IP assignments to stdout.
func doGet(config *gofi.Config, site string) {
	client, err := gofi.New(config)
	if err != nil {
		exitError("failed to create client: " + err.Error())
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		exitError("failed to connect: " + err.Error())
	}
	defer client.Disconnect(ctx)

	users, err := client.Users().List(ctx, site)
	if err != nil {
		exitError("failed to list users: " + err.Error())
	}

	var entries []entry
	for _, u := range users {
		if u.UseFixedIP && u.FixedIP != "" {
			entries = append(entries, entry{IP: u.FixedIP, MAC: strings.ToLower(u.MAC)})
		}
	}

	sortEntries(entries)

	fmt.Println("# gofip fixed IP assignments")
	fmt.Println("# format: IP MAC")

	if len(entries) == 0 {
		fmt.Println("# No fixed IP assignments found on UDM.")
		fmt.Println("# Example:")
		fmt.Println("# 192.168.1.10 aa:bb:cc:dd:ee:ff")
		fmt.Println("# 192.168.1.11 11:22:33:44:55:66")
		fmt.Fprintf(os.Stderr, "No fixed IP assignments found.\n")
	} else {
		for _, e := range entries {
			fmt.Printf("%s %s\n", e.IP, e.MAC)
		}
		fmt.Fprintf(os.Stderr, "Exported %d fixed IP assignment(s).\n", len(entries))
	}
}

// doSet imports fixed IP assignments from a file or stdin.
func doSet(config *gofi.Config, site string, args []string) {
	// Determine input source
	var scanner *bufio.Scanner
	if len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			exitError("failed to open file: " + err.Error())
		}
		defer f.Close()
		scanner = bufio.NewScanner(f)
		fmt.Fprintf(os.Stderr, "Reading from %s\n", args[0])
	} else {
		scanner = bufio.NewScanner(os.Stdin)
		fmt.Fprintf(os.Stderr, "Reading from stdin\n")
	}

	// Parse and validate all input before connecting
	entries, err := parseInput(scanner)
	if err != nil {
		exitError(err.Error())
	}

	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "No entries to process.\n")
		return
	}

	fmt.Fprintf(os.Stderr, "Parsed %d entry/entries from input.\n", len(entries))

	// Connect
	client, err := gofi.New(config)
	if err != nil {
		exitError("failed to create client: " + err.Error())
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		exitError("failed to connect: " + err.Error())
	}
	defer client.Disconnect(ctx)

	// Fetch existing assignments
	users, err := client.Users().List(ctx, site)
	if err != nil {
		exitError("failed to list users: " + err.Error())
	}

	existing := make(map[string]*types.User) // keyed by lowercase MAC
	for i := range users {
		u := &users[i]
		if u.UseFixedIP && u.FixedIP != "" {
			existing[strings.ToLower(u.MAC)] = u
		}
	}

	// Fetch networks for subnet detection
	networks, err := client.Networks().List(ctx, site)
	if err != nil {
		exitError("failed to list networks: " + err.Error())
	}

	// Process entries
	var created, updated, skipped, errored int

	for _, e := range entries {
		mac := strings.ToLower(e.MAC)

		// Check if already exists with same IP
		if eu, ok := existing[mac]; ok {
			if eu.FixedIP == e.IP {
				fmt.Fprintf(os.Stderr, "  skip: %s %s (unchanged)\n", e.IP, mac)
				skipped++
				continue
			}
			// Different IP — update
			networkID, err := detectNetwork(networks, e.IP)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s %s: %v\n", e.IP, mac, err)
				errored++
				continue
			}
			eu.FixedIP = e.IP
			eu.NetworkID = networkID
			if _, err := client.Users().Update(ctx, site, eu); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s %s: failed to update: %v\n", e.IP, mac, err)
				errored++
				continue
			}
			fmt.Fprintf(os.Stderr, "  updated: %s %s\n", e.IP, mac)
			updated++
			continue
		}

		// New entry — detect network and create
		networkID, err := detectNetwork(networks, e.IP)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error: %s %s: %v\n", e.IP, mac, err)
			errored++
			continue
		}

		// Check if user exists without fixed IP
		existingUser, _ := client.Users().GetByMAC(ctx, site, mac)
		if existingUser != nil {
			existingUser.UseFixedIP = true
			existingUser.FixedIP = e.IP
			existingUser.NetworkID = networkID
			if _, err := client.Users().Update(ctx, site, existingUser); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s %s: failed to update: %v\n", e.IP, mac, err)
				errored++
				continue
			}
			fmt.Fprintf(os.Stderr, "  updated: %s %s (added fixed IP to existing user)\n", e.IP, mac)
			updated++
		} else {
			newUser := &types.User{
				MAC:        mac,
				UseFixedIP: true,
				FixedIP:    e.IP,
				NetworkID:  networkID,
			}
			if _, err := client.Users().Create(ctx, site, newUser); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s %s: failed to create: %v\n", e.IP, mac, err)
				errored++
				continue
			}
			fmt.Fprintf(os.Stderr, "  created: %s %s\n", e.IP, mac)
			created++
		}
	}

	fmt.Fprintf(os.Stderr, "\nSummary: %d processed, %d skipped (unchanged), %d created, %d updated, %d errors\n",
		len(entries), skipped, created, updated, errored)

	if errored > 0 {
		os.Exit(1)
	}
}

// parseInput reads and validates all entries from the scanner.
// Returns an error if any line is malformed or there are duplicates.
func parseInput(scanner *bufio.Scanner) ([]entry, error) {
	var entries []entry
	seenIPs := make(map[string]int)   // IP -> line number
	seenMACs := make(map[string]int)  // MAC -> line number
	var dupIPs, dupMACs []string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and blank lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("line %d: expected 'IP MAC', got %d field(s): %s", lineNum, len(fields), line)
		}

		ip := fields[0]
		mac := strings.ToLower(fields[1])

		// Validate IP
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			return nil, fmt.Errorf("line %d: invalid IP address: %s", lineNum, ip)
		}
		if parsedIP.To4() == nil {
			return nil, fmt.Errorf("line %d: only IPv4 is supported: %s", lineNum, ip)
		}

		// Validate MAC
		if !macRegex.MatchString(mac) {
			return nil, fmt.Errorf("line %d: invalid MAC address: %s (expected aa:bb:cc:dd:ee:ff)", lineNum, fields[1])
		}

		// Track duplicates
		if prev, ok := seenIPs[ip]; ok {
			dupIPs = append(dupIPs, fmt.Sprintf("  IP %s on lines %d and %d", ip, prev, lineNum))
		}
		seenIPs[ip] = lineNum

		if prev, ok := seenMACs[mac]; ok {
			dupMACs = append(dupMACs, fmt.Sprintf("  MAC %s on lines %d and %d", mac, prev, lineNum))
		}
		seenMACs[mac] = lineNum

		entries = append(entries, entry{IP: ip, MAC: mac})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	// Report all duplicates at once
	if len(dupIPs) > 0 || len(dupMACs) > 0 {
		var msgs []string
		if len(dupIPs) > 0 {
			msgs = append(msgs, "duplicate IP addresses:\n"+strings.Join(dupIPs, "\n"))
		}
		if len(dupMACs) > 0 {
			msgs = append(msgs, "duplicate MAC addresses:\n"+strings.Join(dupMACs, "\n"))
		}
		return nil, fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	return entries, nil
}

// detectNetwork finds which network contains the given IP.
func detectNetwork(networks []types.Network, ip string) (string, error) {
	parsedIP := net.ParseIP(ip)

	for _, n := range networks {
		if n.IPSubnet == "" {
			continue
		}

		_, subnet, err := net.ParseCIDR(n.IPSubnet)
		if err != nil {
			continue
		}

		if subnet.Contains(parsedIP) {
			return n.ID, nil
		}
	}

	return "", fmt.Errorf("no network found containing IP %s", ip)
}

// sortEntries sorts entries by IP address numerically.
func sortEntries(entries []entry) {
	sort.Slice(entries, func(i, j int) bool {
		return ipToUint32(entries[i].IP) < ipToUint32(entries[j].IP)
	})
}

// ipToUint32 converts an IPv4 address string to a uint32 for numeric sorting.
func ipToUint32(ipStr string) uint32 {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ip4)
}

func exitError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}
