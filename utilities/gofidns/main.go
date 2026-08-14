package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	gofi "github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/utilities/internal/conn"
	"github.com/unifi-go/gofi/utilities/internal/dns"
)

func main() {
	var (
		get     = flag.Bool("get", false, "List local DNS records")
		del     = flag.Bool("del", false, "Delete a local DNS record")
		id      = flag.String("id", "", "Record ID to delete")
		name    = flag.String("name", "", "Record name (key) to delete")
		ip      = flag.String("ip", "", "Record value (IP) to delete")
		host    = flag.String("host", "", "UniFi controller address")
		port    = flag.Int("port", 443, "UniFi controller port")
		site    = flag.String("site", "default", "Site name")
		secure  = flag.Bool("secure", false, "Enforce TLS certificate verification")
		jsonOut = flag.Bool("json", false, "Output in JSON format")
		force   = flag.Bool("force", false, "Allow an identifier that matches several records to delete them all")
		dryRun  = flag.Bool("dry-run", false, "Show what would be done without making changes")
	)
	flag.BoolVar(get, "g", false, "List local DNS records (shorthand)")
	flag.BoolVar(del, "d", false, "Delete a local DNS record (shorthand)")
	flag.StringVar(id, "i", "", "Record ID to delete (shorthand)")
	flag.StringVar(name, "n", "", "Record name to delete (shorthand)")
	flag.StringVar(host, "H", "", "UniFi controller address (shorthand)")
	flag.IntVar(port, "p", 443, "UniFi controller port (shorthand)")
	flag.StringVar(site, "S", "default", "Site name (shorthand)")
	flag.BoolVar(secure, "k", false, "Enforce TLS certificate verification (shorthand)")
	flag.BoolVar(jsonOut, "j", false, "JSON output (shorthand)")
	flag.BoolVar(force, "f", false, "Allow a multi-match delete (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <mode>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Manage local (static) DNS records on a UniFi UDM Pro.\n\n")
		fmt.Fprintf(os.Stderr, "Modes:\n")
		fmt.Fprintf(os.Stderr, "  -g, --get\t\tList local DNS records\n")
		fmt.Fprintf(os.Stderr, "  -d, --del\t\tDelete a local DNS record\n\n")
		fmt.Fprintf(os.Stderr, "Delete identifiers (use exactly one with --del):\n")
		fmt.Fprintf(os.Stderr, "  -i, --id string\tRecord ID to delete\n")
		fmt.Fprintf(os.Stderr, "  -n, --name string\tRecord name (key) to delete\n")
		fmt.Fprintf(os.Stderr, "      --ip string\tRecord value (IP) to delete\n\n")
		fmt.Fprintf(os.Stderr, "Output:\n")
		fmt.Fprintf(os.Stderr, "  -j, --json\t\tOutput in JSON format\n\n")
		fmt.Fprintf(os.Stderr, "Connection:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\tUniFi controller address (or set %s)\n", conn.EnvControllerIP)
		fmt.Fprintf(os.Stderr, "  -p, --port int\tUniFi controller port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -S, --site string\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --secure\tEnforce TLS certificate verification (default: accept self-signed)\n\n")
		fmt.Fprintf(os.Stderr, "Other:\n")
		fmt.Fprintf(os.Stderr, "  -f, --force\t\tAllow an identifier matching several records to delete them all\n")
		fmt.Fprintf(os.Stderr, "      --dry-run\t\tShow what would be done without making changes\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tAPI key (preferred; requires %s)\n", conn.EnvAPIKey, conn.EnvConsoleID)
		fmt.Fprintf(os.Stderr, "  %s\tSite Manager console ID (connector mode)\n", conn.EnvConsoleID)
		fmt.Fprintf(os.Stderr, "  %s\tUsername (required if no API key)\n", conn.EnvUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword (required if no API key)\n", conn.EnvPassword)
		fmt.Fprintf(os.Stderr, "  %s\tController host (fallback for -H)\n\n", conn.EnvControllerIP)
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -g\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -g -j\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -d -n stale.example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -d --ip 192.168.1.50 --dry-run\n", os.Args[0])
	}

	flag.Parse()

	if err := validateModes(*get, *del); err != nil {
		exitError(err.Error())
	}

	identifier := dns.DeleteIdentifier{ID: *id, Name: *name, IP: *ip}
	if *del {
		if err := validateIdentifier(identifier); err != nil {
			exitError(err.Error())
		}
	}

	config, err := conn.ResolveConfig(os.Stderr, *host, *port, *site, *secure)
	if err != nil {
		exitError(err.Error())
	}

	if config.ConsoleID != "" {
		fmt.Fprintf(os.Stderr, "Connecting via connector to console %s...\n", config.ConsoleID)
	} else {
		fmt.Fprintf(os.Stderr, "Connecting to %s...\n", config.Host)
	}

	apiClient, err := gofi.New(config)
	if err != nil {
		exitError("failed to create client: " + err.Error())
	}

	ctx := context.Background()
	if err := apiClient.Connect(ctx); err != nil {
		exitError("failed to connect: " + err.Error())
	}
	defer func() {
		if err := apiClient.Disconnect(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: disconnect failed: %v\n", err)
		}
	}()

	switch {
	case *get:
		options := dns.FormatOptions{Writer: os.Stdout, JSON: *jsonOut}
		if err := dns.DoGet(ctx, apiClient, *site, options); err != nil {
			exitError(err.Error())
		}

	case *del:
		result, err := dns.DoDel(ctx, apiClient, *site, identifier, *dryRun, *force, os.Stderr)
		if err != nil {
			exitError(err.Error())
		}
		fmt.Fprintf(os.Stderr, "\nSummary: %d deleted, %d errors\n", result.Deleted, result.Errors)
		if result.Errors > 0 {
			os.Exit(1)
		}
	}
}

// validateModes enforces that exactly one mode flag is given.
func validateModes(get, del bool) error {
	count := 0
	if get {
		count++
	}
	if del {
		count++
	}

	if count == 0 {
		return fmt.Errorf("a mode is required: --get or --del")
	}
	if count > 1 {
		return fmt.Errorf("only one mode may be given: --get or --del")
	}
	return nil
}

// validateIdentifier enforces that exactly one delete identifier is given.
func validateIdentifier(identifier dns.DeleteIdentifier) error {
	count := 0
	if identifier.ID != "" {
		count++
	}
	if identifier.Name != "" {
		count++
	}
	if identifier.IP != "" {
		count++
	}

	if count != 1 {
		return fmt.Errorf("--del requires exactly one of --id, --name, or --ip")
	}
	return nil
}

func exitError(message string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	os.Exit(1)
}
