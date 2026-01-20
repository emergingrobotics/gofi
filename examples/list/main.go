package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/types"
)

const (
	envUsername = "UNIFI_USERNAME"
	envPassword = "UNIFI_PASSWORD"
)

// debugLogger implements gofi.Logger for debug output.
type debugLogger struct{}

func (l *debugLogger) Debug(msg string, keysAndValues ...interface{}) {
	log.Printf("[DEBUG] %s %v", msg, keysAndValues)
}

func (l *debugLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Printf("[INFO] %s %v", msg, keysAndValues)
}

func (l *debugLogger) Warn(msg string, keysAndValues ...interface{}) {
	log.Printf("[WARN] %s %v", msg, keysAndValues)
}

func (l *debugLogger) Error(msg string, keysAndValues ...interface{}) {
	log.Printf("[ERROR] %s %v", msg, keysAndValues)
}

// NetworkInfo holds network information for output.
type NetworkInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Purpose     string `json:"purpose"`
	VLAN        int    `json:"vlan,omitempty"`
	VLANEnabled bool   `json:"vlan_enabled"`
	Subnet      string `json:"subnet"`
	DHCPEnabled bool   `json:"dhcp_enabled"`
	Enabled     bool   `json:"enabled"`
}

func main() {
	// Define command line flags
	var (
		host     = flag.String("host", "", "UDM Pro host address (required)")
		port     = flag.Int("port", 443, "UDM Pro port")
		site     = flag.String("site", "default", "Site name")
		insecure = flag.Bool("insecure", false, "Skip TLS certificate verification")
		jsonOut  = flag.Bool("json", false, "Output in JSON format")
		debug    = flag.Bool("debug", false, "Enable debug output")
	)

	// Add short flag aliases
	flag.StringVar(host, "H", "", "UDM Pro host address (shorthand)")
	flag.IntVar(port, "p", 443, "UDM Pro port (shorthand)")
	flag.StringVar(site, "s", "default", "Site name (shorthand)")
	flag.BoolVar(insecure, "k", false, "Skip TLS certificate verification (shorthand)")
	flag.BoolVar(jsonOut, "j", false, "Output in JSON format (shorthand)")
	flag.BoolVar(debug, "d", false, "Enable debug output (shorthand)")

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "List all networks from a UniFi UDM Pro controller.\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tUsername for UDM authentication (required)\n", envUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword for UDM authentication (required)\n\n", envPassword)
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\tUDM Pro host address (required)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port int\tUDM Pro port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -s, --site string\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure\tSkip TLS certificate verification\n")
		fmt.Fprintf(os.Stderr, "  -j, --json\t\tOutput in JSON format\n")
		fmt.Fprintf(os.Stderr, "  -d, --debug\t\tEnable debug output\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\t\tShow this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s --host 192.168.1.1\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --host 192.168.1.1 --site production --json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -j\n", os.Args[0])
	}

	flag.Parse()

	// Validate required parameters
	if *host == "" {
		fmt.Fprintf(os.Stderr, "Error: --host is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Get credentials from environment
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

	// Create client configuration
	config := &gofi.Config{
		Host:          *host,
		Port:          *port,
		Username:      username,
		Password:      password,
		Site:          *site,
		SkipTLSVerify: *insecure,
	}

	if *debug {
		config.Logger = &debugLogger{}
	}

	// Create client
	client, err := gofi.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Connect to controller
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect(ctx)

	if *debug {
		log.Printf("Connected to UniFi controller at %s:%d", *host, *port)
	}

	// Fetch networks
	networks, err := client.Networks().List(ctx, *site)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list networks: %v\n", err)
		os.Exit(1)
	}

	if *debug {
		log.Printf("Retrieved %d networks", len(networks))
	}

	// Convert to output format
	infos := make([]NetworkInfo, len(networks))
	for i, n := range networks {
		infos[i] = toNetworkInfo(n)
	}

	// Output results
	if *jsonOut {
		outputJSON(infos)
	} else {
		outputText(infos)
	}
}

func toNetworkInfo(n types.Network) NetworkInfo {
	return NetworkInfo{
		ID:          n.ID,
		Name:        n.Name,
		Purpose:     n.Purpose,
		VLAN:        n.VLAN,
		VLANEnabled: n.VLANEnabled,
		Subnet:      n.IPSubnet,
		DHCPEnabled: n.DHCPDEnabled,
		Enabled:     n.Enabled,
	}
}

func outputJSON(infos []NetworkInfo) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(infos); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func outputText(infos []NetworkInfo) {
	if len(infos) == 0 {
		fmt.Println("No networks found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tVLAN\tSUBNET\tDHCP\tENABLED")
	fmt.Fprintln(w, "----\t----\t----\t------\t----\t-------")

	for _, info := range infos {
		vlanStr := "-"
		if info.VLANEnabled {
			vlanStr = fmt.Sprintf("%d", info.VLAN)
		}

		dhcpStr := "No"
		if info.DHCPEnabled {
			dhcpStr = "Yes"
		}

		enabledStr := "No"
		if info.Enabled {
			enabledStr = "Yes"
		}

		subnet := info.Subnet
		if subnet == "" {
			subnet = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			info.Name,
			info.Purpose,
			vlanStr,
			subnet,
			dhcpStr,
			enabledStr,
		)
	}

	w.Flush()
}
