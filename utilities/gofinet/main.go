package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/utilities/internal/conn"
)

func main() {
	var (
		host    = flag.String("host", "", "UniFi controller address")
		port    = flag.Int("port", 443, "UniFi controller port")
		site    = flag.String("site", "default", "Site name")
		secure  = flag.Bool("secure", false, "Enforce TLS certificate verification")
		jsonOut = flag.Bool("json", false, "Output in JSON format")
	)
	flag.StringVar(host, "H", "", "UniFi controller address (shorthand)")
	flag.IntVar(port, "p", 443, "UniFi controller port (shorthand)")
	flag.StringVar(site, "S", "default", "Site name (shorthand)")
	flag.BoolVar(secure, "k", false, "Enforce TLS certificate verification (shorthand)")
	flag.BoolVar(jsonOut, "j", false, "JSON output (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "List UniFi networks with their subnet and DHCP dynamic address pool.\n\n")
		fmt.Fprintf(os.Stderr, "Output:\n")
		fmt.Fprintf(os.Stderr, "  -j, --json\t\tOutput in JSON format\n\n")
		fmt.Fprintf(os.Stderr, "Connection:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\tUniFi controller address (or set %s)\n", conn.EnvControllerIP)
		fmt.Fprintf(os.Stderr, "  -p, --port int\tUniFi controller port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -S, --site string\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --secure\tEnforce TLS certificate verification (default: accept self-signed)\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tAPI key (preferred; requires %s)\n", conn.EnvAPIKey, conn.EnvConsoleID)
		fmt.Fprintf(os.Stderr, "  %s\tSite Manager console ID (connector mode)\n", conn.EnvConsoleID)
		fmt.Fprintf(os.Stderr, "  %s\tUsername (required if no API key)\n", conn.EnvUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword (required if no API key)\n", conn.EnvPassword)
		fmt.Fprintf(os.Stderr, "  %s\tController host (fallback for -H)\n\n", conn.EnvControllerIP)
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -j\n", os.Args[0])
	}

	flag.Parse()

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

	fmt.Fprintf(os.Stderr, "Fetching networks...\n")
	entries, err := ListNetworks(ctx, apiClient, *site)
	if err != nil {
		exitError("failed to list networks: " + err.Error())
	}

	if *jsonOut {
		if err := FormatJSON(os.Stdout, entries); err != nil {
			exitError("failed to write JSON output: " + err.Error())
		}
	} else {
		if err := FormatText(os.Stdout, entries); err != nil {
			exitError("failed to write output: " + err.Error())
		}
	}
}

func exitError(message string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	os.Exit(1)
}
