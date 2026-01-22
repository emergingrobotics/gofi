package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/unifi-go/gofi"
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

func main() {
	// Define command line flags
	var (
		host     = flag.String("host", "", "UDM Pro host address (required)")
		port     = flag.Int("port", 443, "UDM Pro port")
		site     = flag.String("site", "default", "Site name")
		insecure = flag.Bool("insecure", false, "Skip TLS certificate verification")
		debug    = flag.Bool("debug", false, "Enable debug output")
		timeout  = flag.Duration("timeout", 30*time.Second, "Connection timeout")
	)

	// Add short flag aliases
	flag.StringVar(host, "H", "", "UDM Pro host address (shorthand)")
	flag.IntVar(port, "p", 443, "UDM Pro port (shorthand)")
	flag.StringVar(site, "s", "default", "Site name (shorthand)")
	flag.BoolVar(insecure, "k", false, "Skip TLS certificate verification (shorthand)")
	flag.BoolVar(debug, "d", false, "Enable debug output (shorthand)")
	flag.DurationVar(timeout, "t", 30*time.Second, "Connection timeout (shorthand)")

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Basic example demonstrating gofi library usage.\n")
		fmt.Fprintf(os.Stderr, "Lists sites, devices, networks, and health status.\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tUsername for UDM authentication (required)\n", envUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword for UDM authentication (required)\n\n", envPassword)
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\t\tUDM Pro host address (required)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port int\t\tUDM Pro port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -s, --site string\t\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure\t\tSkip TLS certificate verification\n")
		fmt.Fprintf(os.Stderr, "  -d, --debug\t\t\tEnable debug output\n")
		fmt.Fprintf(os.Stderr, "  -t, --timeout duration\tConnection timeout (default 30s)\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\t\t\tShow this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s --host 192.168.1.1 --debug\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -d\n", os.Args[0])
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

	if *debug {
		log.Printf("Connecting to %s:%d as user %s", *host, *port, username)
	}

	// Create client configuration
	config := &gofi.Config{
		Host:          *host,
		Port:          *port,
		Username:      username,
		Password:      password,
		Site:          *site,
		SkipTLSVerify: *insecure,
		Timeout:       *timeout,
	}

	if *debug {
		config.Logger = &debugLogger{}
		log.Println("Creating client...")
	}

	// Create client
	client, err := gofi.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create client: %v\n", err)
		os.Exit(1)
	}

	if *debug {
		log.Println("Client created, attempting to connect...")
	}

	// Connect to controller with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	if *debug {
		log.Println("Connected successfully!")
	}

	fmt.Println("Connected to UniFi controller!")

	// List all sites
	if *debug {
		log.Println("Fetching sites...")
	}

	sites, err := client.Sites().List(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list sites: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFound %d site(s):\n", len(sites))
	for _, s := range sites {
		fmt.Printf("  - %s (%s)\n", s.Desc, s.Name)
	}

	// List devices on specified site
	if *debug {
		log.Printf("Fetching devices for site %s...", *site)
	}

	devices, err := client.Devices().List(ctx, *site)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list devices: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFound %d device(s):\n", len(devices))
	for _, device := range devices {
		fmt.Printf("  - %s (%s) - %s - State: %s\n",
			device.Name,
			device.Model,
			device.MAC,
			device.State.String(),
		)
	}

	// List networks
	if *debug {
		log.Printf("Fetching networks for site %s...", *site)
	}

	networks, err := client.Networks().List(ctx, *site)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list networks: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFound %d network(s):\n", len(networks))
	for _, network := range networks {
		fmt.Printf("  - %s (VLAN Enabled: %t)\n", network.Name, network.VLANEnabled)
	}

	// Get health information
	if *debug {
		log.Printf("Fetching health for site %s...", *site)
	}

	health, err := client.Sites().Health(ctx, *site)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get health: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nHealth Status:\n")
	for _, h := range health {
		fmt.Printf("  - %s: %s\n", h.Subsystem, h.Status)
	}

	if *debug {
		log.Println("Done!")
	}
}
