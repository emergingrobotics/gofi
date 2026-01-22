package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/types"
)

const (
	envUsername = "UNIFI_USERNAME"
	envPassword = "UNIFI_PASSWORD"
)

var macRegex = regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)

func main() {
	var (
		host      = flag.String("host", "", "UDM Pro host address (required)")
		port      = flag.Int("port", 443, "UDM Pro port")
		site      = flag.String("site", "default", "Site name")
		insecure  = flag.Bool("insecure", false, "Skip TLS certificate verification")
		mac       = flag.String("mac", "", "MAC address of device (required)")
		ip        = flag.String("ip", "", "Fixed IP address to assign (required)")
		name      = flag.String("name", "", "Hostname/friendly name (required)")
		networkID = flag.String("network", "", "Network ID or name (auto-detect if not specified)")
		force     = flag.Bool("force", false, "Skip conflict checks")
	)

	flag.StringVar(host, "H", "", "UDM Pro host address (shorthand)")
	flag.IntVar(port, "p", 443, "UDM Pro port (shorthand)")
	flag.StringVar(site, "s", "default", "Site name (shorthand)")
	flag.BoolVar(insecure, "k", false, "Skip TLS certificate verification (shorthand)")
	flag.StringVar(mac, "m", "", "MAC address of device (shorthand)")
	flag.StringVar(ip, "i", "", "Fixed IP address to assign (shorthand)")
	flag.StringVar(name, "n", "", "Hostname/friendly name (shorthand)")
	flag.StringVar(networkID, "N", "", "Network ID or name (shorthand)")
	flag.BoolVar(force, "f", false, "Skip conflict checks (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Add a fixed IP assignment for a device.\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tUsername for UDM authentication (required)\n", envUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword for UDM authentication (required)\n\n", envPassword)
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\tUDM Pro host address (required)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port int\tUDM Pro port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -s, --site string\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure\tSkip TLS certificate verification\n")
		fmt.Fprintf(os.Stderr, "  -m, --mac string\tMAC address of device (required)\n")
		fmt.Fprintf(os.Stderr, "  -i, --ip string\tFixed IP address to assign (required)\n")
		fmt.Fprintf(os.Stderr, "  -n, --name string\tHostname/friendly name (required)\n")
		fmt.Fprintf(os.Stderr, "  -N, --network string\tNetwork ID or name (auto-detect if not specified)\n")
		fmt.Fprintf(os.Stderr, "  -f, --force\t\tSkip conflict checks\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\t\tShow this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff -i 192.168.1.100 -n \"My Device\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff -i 192.168.1.100 -n \"My Device\" -N \"LAN\"\n", os.Args[0])
	}

	flag.Parse()

	// Validate required parameters
	if *host == "" {
		exitError("--host is required")
	}
	if *mac == "" {
		exitError("--mac is required")
	}
	if *ip == "" {
		exitError("--ip is required")
	}
	if *name == "" {
		exitError("--name is required")
	}

	// Normalize and validate MAC
	*mac = strings.ToLower(*mac)
	if !macRegex.MatchString(*mac) {
		exitError("invalid MAC address format (expected aa:bb:cc:dd:ee:ff)")
	}

	// Validate IP
	parsedIP := net.ParseIP(*ip)
	if parsedIP == nil {
		exitError("invalid IP address format")
	}
	if parsedIP.To4() == nil {
		exitError("only IPv4 addresses are supported")
	}

	// Get credentials
	username := os.Getenv(envUsername)
	password := os.Getenv(envPassword)
	if username == "" {
		exitError(envUsername + " environment variable is required")
	}
	if password == "" {
		exitError(envPassword + " environment variable is required")
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
		exitError("failed to create client: " + err.Error())
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		exitError("failed to connect: " + err.Error())
	}
	defer client.Disconnect(ctx)

	// Run the assignment
	if err := assignFixedIP(ctx, client, *site, *mac, *ip, *name, *networkID, *force); err != nil {
		exitError(err.Error())
	}
}

func assignFixedIP(ctx context.Context, client gofi.Client, site, mac, ip, name, networkHint string, force bool) error {
	// Step 1: Check for conflicts (unless --force)
	if !force {
		if err := checkConflicts(ctx, client, site, mac, ip); err != nil {
			return err
		}
	}

	// Step 2: Determine network ID
	networkID, networkName, err := resolveNetwork(ctx, client, site, ip, networkHint)
	if err != nil {
		return err
	}

	// Step 3: Check if MAC already exists as a user
	existingUser, _ := client.Users().GetByMAC(ctx, site, mac)

	if existingUser != nil {
		// Update existing user
		if existingUser.UseFixedIP && existingUser.FixedIP != "" && existingUser.FixedIP != ip {
			fmt.Printf("Note: Updating existing fixed IP assignment (%s -> %s)\n", existingUser.FixedIP, ip)
		}

		existingUser.Name = name
		existingUser.UseFixedIP = true
		existingUser.FixedIP = ip
		existingUser.NetworkID = networkID

		_, err := client.Users().Update(ctx, site, existingUser)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		fmt.Printf("Updated fixed IP assignment:\n")
	} else {
		// Create new user
		newUser := &types.User{
			MAC:        mac,
			Name:       name,
			UseFixedIP: true,
			FixedIP:    ip,
			NetworkID:  networkID,
		}

		_, err := client.Users().Create(ctx, site, newUser)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		fmt.Printf("Created fixed IP assignment:\n")
	}

	fmt.Printf("  Name:    %s\n", name)
	fmt.Printf("  MAC:     %s\n", mac)
	fmt.Printf("  IP:      %s\n", ip)
	fmt.Printf("  Network: %s\n", networkName)

	return nil
}

func checkConflicts(ctx context.Context, client gofi.Client, site, mac, ip string) error {
	// Check active clients for IP in use
	activeClients, err := client.Clients().ListActive(ctx, site)
	if err != nil {
		return fmt.Errorf("failed to list active clients: %w", err)
	}

	for _, c := range activeClients {
		if c.IP == ip && c.MAC != mac {
			clientName := c.Name
			if clientName == "" {
				clientName = c.Hostname
			}
			if clientName == "" {
				clientName = "unknown"
			}
			return fmt.Errorf("IP %s is currently in use by %s (%s)\nUse --force to skip this check", ip, clientName, c.MAC)
		}
	}

	// Check existing fixed IP reservations
	users, err := client.Users().List(ctx, site)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	for _, u := range users {
		if u.UseFixedIP && u.FixedIP == ip && u.MAC != mac {
			userName := u.Name
			if userName == "" {
				userName = u.Hostname
			}
			if userName == "" {
				userName = "unknown"
			}
			return fmt.Errorf("IP %s is already reserved for %s (%s)\nUse --force to skip this check", ip, userName, u.MAC)
		}
	}

	return nil
}

func resolveNetwork(ctx context.Context, client gofi.Client, site, ip, networkHint string) (string, string, error) {
	networks, err := client.Networks().List(ctx, site)
	if err != nil {
		return "", "", fmt.Errorf("failed to list networks: %w", err)
	}

	// If network hint provided, try to find by ID or name
	if networkHint != "" {
		for _, n := range networks {
			if n.ID == networkHint || strings.EqualFold(n.Name, networkHint) {
				return n.ID, n.Name, nil
			}
		}
		return "", "", fmt.Errorf("network not found: %s", networkHint)
	}

	// Auto-detect network from IP subnet
	parsedIP := net.ParseIP(ip)

	for _, n := range networks {
		if n.IPSubnet == "" {
			continue
		}

		// Parse the network subnet (format: "192.168.1.1/24" or "192.168.1.0/24")
		_, subnet, err := net.ParseCIDR(n.IPSubnet)
		if err != nil {
			// Try treating it as gateway/prefix
			parts := strings.Split(n.IPSubnet, "/")
			if len(parts) == 2 {
				gatewayIP := net.ParseIP(parts[0])
				if gatewayIP != nil {
					// Reconstruct CIDR from gateway
					_, subnet, err = net.ParseCIDR(n.IPSubnet)
					if err != nil {
						continue
					}
				}
			}
			if subnet == nil {
				continue
			}
		}

		if subnet.Contains(parsedIP) {
			return n.ID, n.Name, nil
		}
	}

	// List available networks for user
	fmt.Fprintf(os.Stderr, "Could not auto-detect network for IP %s\n", ip)
	fmt.Fprintf(os.Stderr, "Available networks:\n")
	for _, n := range networks {
		if n.IPSubnet != "" {
			fmt.Fprintf(os.Stderr, "  - %s (ID: %s, Subnet: %s)\n", n.Name, n.ID, n.IPSubnet)
		}
	}

	return "", "", fmt.Errorf("please specify network with --network flag")
}

func exitError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}
