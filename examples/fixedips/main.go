package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/unifi-go/gofi"
)

const (
	envUsername = "UNIFI_USERNAME"
	envPassword = "UNIFI_PASSWORD"
	envUDMIP    = "UNIFI_UDM_IP"
)

// FixedIPEntry holds information about a fixed IP assignment.
type FixedIPEntry struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname,omitempty"`
	MAC      string `json:"mac"`
	FixedIP  string `json:"fixed_ip"`
}

func main() {
	var (
		host     = flag.String("host", "", "UDM Pro host address (required)")
		port     = flag.Int("port", 443, "UDM Pro port")
		site     = flag.String("site", "default", "Site name")
		insecure = flag.Bool("insecure", false, "Skip TLS certificate verification")
		jsonOut  = flag.Bool("json", false, "Output in JSON format")
	)

	flag.StringVar(host, "H", "", "UDM Pro host address (shorthand)")
	flag.IntVar(port, "p", 443, "UDM Pro port (shorthand)")
	flag.StringVar(site, "s", "default", "Site name (shorthand)")
	flag.BoolVar(insecure, "k", false, "Skip TLS certificate verification (shorthand)")
	flag.BoolVar(jsonOut, "j", false, "Output in JSON format (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "List all clients with fixed IP addresses assigned.\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tUsername for UDM authentication (required)\n", envUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword for UDM authentication (required)\n", envPassword)
		fmt.Fprintf(os.Stderr, "  %s\tUDM Pro host address (optional, can use -H instead)\n\n", envUDMIP)
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\tUDM Pro host address (required unless %s is set)\n", envUDMIP)
		fmt.Fprintf(os.Stderr, "  -p, --port int\tUDM Pro port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -s, --site string\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure\tSkip TLS certificate verification\n")
		fmt.Fprintf(os.Stderr, "  -j, --json\t\tOutput in JSON format\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\t\tShow this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -k  # Uses %s\n", os.Args[0], envUDMIP)
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -j | jq\n", os.Args[0])
	}

	flag.Parse()

	// Check for host from environment variable if not provided
	if *host == "" {
		*host = os.Getenv(envUDMIP)
	}

	if *host == "" {
		fmt.Fprintf(os.Stderr, "Error: --host is required (or set %s environment variable)\n\n", envUDMIP)
		flag.Usage()
		os.Exit(1)
	}

	username := os.Getenv(envUsername)
	password := os.Getenv(envPassword)

	if username == "" {
		fmt.Fprintf(os.Stderr, "Error: %s environment variable is required\n", envUsername)
		os.Exit(1)
	}

	if password == "" {
		fmt.Fprintf(os.Stderr, "Error: %s environment variable is required\n", envPassword)
		os.Exit(1)
	}

	// Create client
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
		fmt.Fprintf(os.Stderr, "Error: failed to create client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect(ctx)

	// List all users (known clients)
	users, err := client.Users().List(ctx, *site)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list users: %v\n", err)
		os.Exit(1)
	}

	// Filter for users with fixed IPs
	var entries []FixedIPEntry
	for _, user := range users {
		if user.UseFixedIP && user.FixedIP != "" {
			name := user.Name
			if name == "" {
				name = user.Hostname
			}
			if name == "" {
				name = user.MAC // fallback to MAC if no name
			}

			entries = append(entries, FixedIPEntry{
				Name:     name,
				Hostname: user.Hostname,
				MAC:      user.MAC,
				FixedIP:  user.FixedIP,
			})
		}
	}

	// Sort by IP address
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FixedIP < entries[j].FixedIP
	})

	// Output
	if *jsonOut {
		outputJSON(entries)
	} else {
		outputTable(entries)
	}
}

func outputJSON(entries []FixedIPEntry) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(entries); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func outputTable(entries []FixedIPEntry) {
	if len(entries) == 0 {
		fmt.Println("No fixed IP assignments found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tHOSTNAME\tMAC\tFIXED IP")
	fmt.Fprintln(w, "----\t--------\t---\t--------")

	for _, e := range entries {
		hostname := e.Hostname
		if hostname == "" {
			hostname = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Name, hostname, e.MAC, e.FixedIP)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d fixed IP assignments\n", len(entries))
}
