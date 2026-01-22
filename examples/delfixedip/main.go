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
		host     = flag.String("host", "", "UDM Pro host address (required)")
		port     = flag.Int("port", 443, "UDM Pro port")
		site     = flag.String("site", "default", "Site name")
		insecure = flag.Bool("insecure", false, "Skip TLS certificate verification")
		mac      = flag.String("mac", "", "MAC address of device")
		ip       = flag.String("ip", "", "Fixed IP address to look up")
		delUser  = flag.Bool("delete", false, "Delete the user entry entirely (not just the fixed IP)")
	)

	flag.StringVar(host, "H", "", "UDM Pro host address (shorthand)")
	flag.IntVar(port, "p", 443, "UDM Pro port (shorthand)")
	flag.StringVar(site, "s", "default", "Site name (shorthand)")
	flag.BoolVar(insecure, "k", false, "Skip TLS certificate verification (shorthand)")
	flag.StringVar(mac, "m", "", "MAC address of device (shorthand)")
	flag.StringVar(ip, "i", "", "Fixed IP address to look up (shorthand)")
	flag.BoolVar(delUser, "D", false, "Delete the user entry entirely (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Remove a fixed IP assignment, allowing the device to get a dynamic address.\n")
		fmt.Fprintf(os.Stderr, "Specify the device by MAC address (-m) or by its current fixed IP (-i).\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tUsername for UDM authentication (required)\n", envUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword for UDM authentication (required)\n\n", envPassword)
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\tUDM Pro host address (required)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port int\tUDM Pro port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -s, --site string\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure\tSkip TLS certificate verification\n")
		fmt.Fprintf(os.Stderr, "  -m, --mac string\tMAC address of device\n")
		fmt.Fprintf(os.Stderr, "  -i, --ip string\tFixed IP address to look up\n")
		fmt.Fprintf(os.Stderr, "  -D, --delete\t\tDelete the user entry entirely (not just the fixed IP)\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\t\tShow this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -i 192.168.1.100\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -m aa:bb:cc:dd:ee:ff -D  # Delete user entirely\n", os.Args[0])
	}

	flag.Parse()

	// Validate required parameters
	if *host == "" {
		exitError("--host is required")
	}
	if *mac == "" && *ip == "" {
		exitError("either --mac or --ip is required")
	}

	// Validate MAC if provided
	if *mac != "" {
		*mac = strings.ToLower(*mac)
		if !macRegex.MatchString(*mac) {
			exitError("invalid MAC address format (expected aa:bb:cc:dd:ee:ff)")
		}
	}

	// Validate IP if provided
	if *ip != "" {
		parsedIP := net.ParseIP(*ip)
		if parsedIP == nil {
			exitError("invalid IP address format")
		}
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

	// Find the user
	user, err := findUser(ctx, client, *site, *mac, *ip)
	if err != nil {
		exitError(err.Error())
	}

	// Display what we found
	userName := user.Name
	if userName == "" {
		userName = user.Hostname
	}
	if userName == "" {
		userName = user.MAC
	}

	fmt.Printf("Found user:\n")
	fmt.Printf("  Name:     %s\n", userName)
	fmt.Printf("  MAC:      %s\n", user.MAC)
	if user.UseFixedIP {
		fmt.Printf("  Fixed IP: %s\n", user.FixedIP)
	} else {
		fmt.Printf("  Fixed IP: (none)\n")
	}

	// Delete or clear fixed IP
	if *delUser {
		// Delete the user entry entirely
		if err := client.Users().Delete(ctx, *site, user.ID); err != nil {
			exitError("failed to delete user: " + err.Error())
		}
		fmt.Printf("\nDeleted user entry for %s\n", userName)
	} else {
		// Just clear the fixed IP
		if !user.UseFixedIP {
			fmt.Printf("\nNo fixed IP assignment to remove.\n")
			return
		}

		err := client.Users().ClearFixedIP(ctx, *site, user.MAC)
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "LocalDnsRecordRequiresFixedIp") {
				fmt.Fprintf(os.Stderr, "\nError: This device has a local DNS record that depends on the fixed IP.\n")
				fmt.Fprintf(os.Stderr, "You must delete the DNS record first:\n")
				fmt.Fprintf(os.Stderr, "  1. Go to UniFi Network UI -> Settings -> DNS\n")
				fmt.Fprintf(os.Stderr, "  2. Find and delete the DNS record for '%s'\n", userName)
				fmt.Fprintf(os.Stderr, "  3. Then run this command again\n")
				os.Exit(1)
			}
			exitError("failed to clear fixed IP: " + err.Error())
		}
		fmt.Printf("\nRemoved fixed IP assignment. Device will now use DHCP.\n")
	}
}

func findUser(ctx context.Context, client gofi.Client, site, mac, ip string) (*types.User, error) {
	// If MAC provided, look up directly
	if mac != "" {
		user, err := client.Users().GetByMAC(ctx, site, mac)
		if err != nil {
			return nil, fmt.Errorf("no user found with MAC %s", mac)
		}
		return user, nil
	}

	// If IP provided, search through users
	users, err := client.Users().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	for _, u := range users {
		if u.UseFixedIP && u.FixedIP == ip {
			return &u, nil
		}
	}

	return nil, fmt.Errorf("no user found with fixed IP %s", ip)
}

func exitError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}
