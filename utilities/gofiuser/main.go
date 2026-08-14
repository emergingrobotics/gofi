package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	gofi "github.com/unifi-go/gofi/src"
	"github.com/unifi-go/gofi/utilities/internal/conn"
	"github.com/unifi-go/gofi/utilities/internal/users"
)

func main() {
	var (
		list    = flag.Bool("list", false, "List known-client records")
		del     = flag.Bool("del", false, "Remove a known client")
		mac     = flag.String("mac", "", "Client MAC to remove")
		name    = flag.String("name", "", "Client name or hostname to remove")
		filter  = flag.String("filter", "", "Only list entries containing this substring")
		host    = flag.String("host", "", "UniFi controller address")
		port    = flag.Int("port", 443, "UniFi controller port")
		site    = flag.String("site", "default", "Site name")
		secure  = flag.Bool("secure", false, "Enforce TLS certificate verification")
		jsonOut = flag.Bool("json", false, "Output in JSON format")
		dryRun  = flag.Bool("dry-run", false, "Show what would be done without making changes")
	)
	flag.BoolVar(list, "l", false, "List known-client records (shorthand)")
	flag.BoolVar(del, "d", false, "Remove a known client (shorthand)")
	flag.StringVar(mac, "m", "", "Client MAC to remove (shorthand)")
	flag.StringVar(name, "n", "", "Client name to remove (shorthand)")
	flag.StringVar(host, "H", "", "UniFi controller address (shorthand)")
	flag.IntVar(port, "p", 443, "UniFi controller port (shorthand)")
	flag.StringVar(site, "S", "default", "Site name (shorthand)")
	flag.BoolVar(secure, "k", false, "Enforce TLS certificate verification (shorthand)")
	flag.BoolVar(jsonOut, "j", false, "JSON output (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <mode>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Inspect and remove known-client (user) records on a UniFi UDM Pro.\n\n")
		fmt.Fprintf(os.Stderr, "Modes:\n")
		fmt.Fprintf(os.Stderr, "  -l, --list\t\tList known-client records\n")
		fmt.Fprintf(os.Stderr, "  -d, --del\t\tRemove a known client (clears fixed IP, then forgets it)\n\n")
		fmt.Fprintf(os.Stderr, "Delete identifiers (use exactly one with --del):\n")
		fmt.Fprintf(os.Stderr, "  -m, --mac string\tClient MAC to remove\n")
		fmt.Fprintf(os.Stderr, "  -n, --name string\tClient name or hostname to remove\n\n")
		fmt.Fprintf(os.Stderr, "Output:\n")
		fmt.Fprintf(os.Stderr, "      --filter string\tOnly list entries containing this substring\n")
		fmt.Fprintf(os.Stderr, "  -j, --json\t\tOutput in JSON format\n\n")
		fmt.Fprintf(os.Stderr, "Connection:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\tUniFi controller address (or set %s)\n", conn.EnvControllerIP)
		fmt.Fprintf(os.Stderr, "  -p, --port int\tUniFi controller port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -S, --site string\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --secure\tEnforce TLS certificate verification (default: accept self-signed)\n\n")
		fmt.Fprintf(os.Stderr, "Other:\n")
		fmt.Fprintf(os.Stderr, "      --dry-run\t\tShow what would be done without making changes\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tAPI key (preferred; requires %s)\n", conn.EnvAPIKey, conn.EnvConsoleID)
		fmt.Fprintf(os.Stderr, "  %s\tSite Manager console ID (connector mode)\n", conn.EnvConsoleID)
		fmt.Fprintf(os.Stderr, "  %s\tUsername (required if no API key)\n", conn.EnvUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword (required if no API key)\n", conn.EnvPassword)
		fmt.Fprintf(os.Stderr, "  %s\tController host (fallback for -H)\n\n", conn.EnvControllerIP)
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -l\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -l --filter tapo\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -d -m aa:bb:cc:dd:ee:ff --dry-run\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -d -n oldphone\n", os.Args[0])
	}

	flag.Parse()

	if err := validateModes(*list, *del); err != nil {
		exitError(err.Error())
	}

	identifier := users.DeleteIdentifier{MAC: *mac, Name: *name}
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
	case *list:
		options := users.FormatOptions{Writer: os.Stdout, JSON: *jsonOut}
		if err := users.DoList(ctx, apiClient, *site, *filter, options); err != nil {
			exitError(err.Error())
		}

	case *del:
		result, err := users.DoDel(ctx, apiClient, *site, identifier, *dryRun, os.Stderr)
		if err != nil {
			exitError(err.Error())
		}
		fmt.Fprintf(os.Stderr, "\nSummary: cleared_fixed_ip=%t forgot=%t\n", result.ClearedFixedIP, result.Forgot)
	}
}

// validateModes enforces that exactly one mode flag is given.
func validateModes(list, del bool) error {
	count := 0
	if list {
		count++
	}
	if del {
		count++
	}

	if count == 0 {
		return fmt.Errorf("a mode is required: --list or --del")
	}
	if count > 1 {
		return fmt.Errorf("only one mode may be given: --list or --del")
	}
	return nil
}

// validateIdentifier enforces that exactly one delete identifier is given.
func validateIdentifier(identifier users.DeleteIdentifier) error {
	count := 0
	if identifier.MAC != "" {
		count++
	}
	if identifier.Name != "" {
		count++
	}

	if count != 1 {
		return fmt.Errorf("--del requires exactly one of --mac or --name")
	}
	return nil
}

func exitError(message string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	os.Exit(1)
}
