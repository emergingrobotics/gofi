package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/unifi-go/gofi"
)

const (
	envUsername     = "UNIFI_USERNAME"
	envPassword     = "UNIFI_PASSWORD"
	envControllerIP = "UNIFI_CONTROLLER_IP"
)

func main() {
	var (
		host     = flag.String("host", "", "UDM Pro host address")
		port     = flag.Int("port", 443, "UDM Pro port")
		site     = flag.String("site", "default", "Site name")
		insecure = flag.Bool("insecure", false, "Skip TLS certificate verification")
		get      = flag.Bool("get", false, "Export fixed IP assignments to stdout in ISC DHCP format")
		set      = flag.Bool("set", false, "Import host declarations from file or stdin")
		add      = flag.Bool("add", false, "Add a single host from ISC DHCP declaration")
		del      = flag.Bool("del", false, "Delete a host by name, MAC, or IP")
		name     = flag.String("name", "", "Hostname for --del")
		mac      = flag.String("mac", "", "MAC address for --del")
		ip       = flag.String("ip", "", "IP address for --del")
		force    = flag.Bool("force", false, "Skip conflict checks; force delete")
		keepDNS  = flag.Bool("keep-dns", false, "Do not delete associated DNS records")
		dryRun   = flag.Bool("dry-run", false, "Show what would be done without making changes")
	)

	flag.StringVar(host, "H", "", "UDM Pro host address (shorthand)")
	flag.IntVar(port, "p", 443, "UDM Pro port (shorthand)")
	flag.StringVar(site, "S", "default", "Site name (shorthand)")
	flag.BoolVar(insecure, "k", false, "Skip TLS certificate verification (shorthand)")
	flag.BoolVar(get, "g", false, "Export (shorthand)")
	flag.BoolVar(set, "s", false, "Import (shorthand)")
	flag.BoolVar(add, "a", false, "Add (shorthand)")
	flag.BoolVar(del, "d", false, "Delete (shorthand)")
	flag.StringVar(name, "n", "", "Hostname for --del (shorthand)")
	flag.StringVar(mac, "m", "", "MAC address for --del (shorthand)")
	flag.StringVar(ip, "i", "", "IP address for --del (shorthand)")
	flag.BoolVar(force, "f", false, "Force (shorthand)")
	flag.BoolVar(keepDNS, "K", false, "Keep DNS (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <mode>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Manage fixed IP + DNS assignments on a UniFi UDM Pro using ISC DHCP format.\n\n")
		fmt.Fprintf(os.Stderr, "Modes:\n")
		fmt.Fprintf(os.Stderr, "  -g, --get\t\tExport assignments to stdout in ISC DHCP format\n")
		fmt.Fprintf(os.Stderr, "  -s, --set\t\tImport assignments from file or stdin\n")
		fmt.Fprintf(os.Stderr, "  -a, --add\t\tAdd a single host from ISC DHCP declaration\n")
		fmt.Fprintf(os.Stderr, "  -d, --del\t\tDelete a host by identifier\n\n")
		fmt.Fprintf(os.Stderr, "Delete identifiers (use exactly one with --del):\n")
		fmt.Fprintf(os.Stderr, "  -n, --name string\tHostname to delete\n")
		fmt.Fprintf(os.Stderr, "  -m, --mac string\tMAC address to delete\n")
		fmt.Fprintf(os.Stderr, "  -i, --ip string\tIP address to delete\n\n")
		fmt.Fprintf(os.Stderr, "Connection:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\tUDM Pro host address (or set %s)\n", envControllerIP)
		fmt.Fprintf(os.Stderr, "  -p, --port int\tUDM Pro port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -S, --site string\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure\tSkip TLS certificate verification\n\n")
		fmt.Fprintf(os.Stderr, "Other:\n")
		fmt.Fprintf(os.Stderr, "  -f, --force\t\tSkip conflict checks; with --set, re-process unchanged entries\n")
		fmt.Fprintf(os.Stderr, "  -K, --keep-dns\tDo not delete DNS records on delete\n")
		fmt.Fprintf(os.Stderr, "      --dry-run\t\tShow what would be done without making changes\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tUsername (required)\n", envUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword (required)\n", envPassword)
		fmt.Fprintf(os.Stderr, "  %s\tUDM host (fallback for -H)\n\n", envControllerIP)
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -g > hosts.conf\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -s hosts.conf\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -a 'host mydev { hardware ethernet aa:bb:cc:dd:ee:ff; fixed-address 192.168.1.50; }'\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -d -n mydev\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -d -m aa:bb:cc:dd:ee:ff\n", os.Args[0])
	}

	flag.Parse()

	// Validate mode exclusivity
	modeCount := boolCount(*get, *set, *add, *del)
	if modeCount == 0 {
		fmt.Fprintf(os.Stderr, "Error: specify a mode (--get, --set, --add, or --del)\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if modeCount > 1 {
		exitError("specify only one mode (--get, --set, --add, or --del)")
	}

	// Go's flag package stops parsing at the first positional argument, so a
	// flag placed after it (e.g. `--set hosts.conf --dry-run`) is silently
	// ignored. Reject that outright so a mistyped --dry-run cannot cause an
	// unintended write.
	if err := checkStrayFlags(flag.Args()); err != nil {
		exitError(err.Error())
	}

	// Resolve host
	if *host == "" {
		*host = os.Getenv(envControllerIP)
	}
	if *host == "" {
		exitError("--host is required (or set " + envControllerIP + ")")
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

	// For --set and --add, parse input before connecting
	var parsedEntries []HostEntry
	var singleEntry *HostEntry

	if *set {
		result, err := parseSetInput(flag.Args())
		if err != nil {
			exitError(err.Error())
		}
		parsedEntries = result.Entries
		if len(parsedEntries) == 0 {
			fmt.Fprintf(os.Stderr, "No entries to process.\n")
			return
		}
		fmt.Fprintf(os.Stderr, "Parsed %d entry/entries from input.\n", len(parsedEntries))
	}

	if *add {
		entry, err := parseAddInput(flag.Args())
		if err != nil {
			exitError(err.Error())
		}
		singleEntry = entry
	}

	if *del {
		idCount := boolCount(*name != "", *mac != "", *ip != "")
		if idCount == 0 {
			exitError("--del requires exactly one of --name, --mac, or --ip")
		}
		if idCount > 1 {
			exitError("--del requires exactly one of --name, --mac, or --ip")
		}
	}

	// Install the live resolver pointed at the UDM itself, so DNS existence
	// checks see device-local (DHCP) DNS as well as static records.
	resolveFQDN = func(ctx context.Context, fqdn string) ([]string, error) {
		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
				dialer := net.Dialer{Timeout: 3 * time.Second}
				return dialer.DialContext(ctx, "udp", net.JoinHostPort(*host, "53"))
			},
		}
		return resolver.LookupHost(ctx, fqdn)
	}

	// Connect
	config := &gofi.Config{
		Host:          *host,
		Port:          *port,
		Username:      username,
		Password:      password,
		Site:          *site,
		SkipTLSVerify: *insecure,
	}

	client, err := gofi.New(config)
	if err != nil {
		exitError("failed to create client: " + err.Error())
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		exitError("failed to connect: " + err.Error())
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: disconnect failed: %v\n", err)
		}
	}()

	// Dispatch
	switch {
	case *get:
		opts := FormatOptions{
			Host: *host,
			Date: time.Now().Format("2006-01-02"),
		}
		if err := DoGet(ctx, client, *site, os.Stdout, opts); err != nil {
			exitError(err.Error())
		}

	case *set:
		result, err := DoSet(ctx, client, *site, parsedEntries, *dryRun, *force)
		if err != nil {
			exitError(err.Error())
		}
		fmt.Fprintf(os.Stderr, "\nSummary: %d processed, %d skipped, %d created, %d updated, %d errors\n",
			len(parsedEntries), result.Skipped, result.Created, result.Updated, result.Errors)
		if result.Errors > 0 {
			os.Exit(1)
		}

	case *add:
		if err := DoAdd(ctx, client, *site, singleEntry, *force); err != nil {
			exitError(err.Error())
		}

	case *del:
		identifier := DeleteIdentifier{Name: *name, MAC: *mac, IP: *ip}
		if err := DoDel(ctx, client, *site, identifier, *force, *keepDNS); err != nil {
			exitError(err.Error())
		}
	}
}

func parseSetInput(args []string) (*ParseResult, error) {
	var reader io.Reader
	if len(args) > 0 {
		file, err := os.Open(args[0])
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		fmt.Fprintf(os.Stderr, "Reading from %s\n", args[0])
		reader = file
	} else {
		fmt.Fprintf(os.Stderr, "Reading from stdin\n")
		reader = os.Stdin
	}
	return Parse(reader)
}

func parseAddInput(args []string) (*HostEntry, error) {
	if len(args) > 0 {
		return ParseSingle(strings.Join(args, " "))
	}
	// Read from stdin
	result, err := Parse(os.Stdin)
	if err != nil {
		return nil, err
	}
	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("no host declaration found in input")
	}
	if len(result.Entries) > 1 {
		return nil, fmt.Errorf("expected exactly one host declaration, found %d", len(result.Entries))
	}
	return &result.Entries[0], nil
}

func boolCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

// checkStrayFlags rejects positional arguments that look like flags. Go's flag
// package treats everything after the first positional as further positionals,
// so `--set hosts.conf --dry-run` would silently drop --dry-run; erroring here
// turns that footgun into a clear failure. A bare "-" is allowed.
func checkStrayFlags(args []string) error {
	for _, arg := range args {
		if len(arg) > 1 && strings.HasPrefix(arg, "-") {
			return fmt.Errorf("flag %q must appear before positional arguments (place flags like --dry-run before the filename)", arg)
		}
	}
	return nil
}

func exitError(message string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	os.Exit(1)
}
