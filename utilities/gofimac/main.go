package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/unifi-go/gofi"
)

const (
	envUsername = "UNIFI_USERNAME"
	envPassword = "UNIFI_PASSWORD"
	envUDMIP    = "UNIFI_UDM_IP"
)

// defaultGoneWindow is the history window used when --gone is given without a duration.
const defaultGoneWindow = "7d"

// optionalDuration is a flag value that may be supplied bare (--gone) or with a
// value (--gone=30d). Implementing IsBoolFlag lets the stdlib flag package accept
// the bare form; a bare use records set=true with an empty value.
type optionalDuration struct {
	set   bool
	value string
}

func (o *optionalDuration) String() string { return o.value }

func (o *optionalDuration) Set(raw string) error {
	o.set = true
	if raw == "true" { // stdlib passes "true" for a bare bool-style flag
		o.value = ""
	} else {
		o.value = raw
	}
	return nil
}

func (o *optionalDuration) IsBoolFlag() bool { return true }

func main() {
	var (
		host     = flag.String("host", "", "UDM Pro host address")
		port     = flag.Int("port", 443, "UDM Pro port")
		site     = flag.String("site", "default", "Site name")
		insecure = flag.Bool("insecure", false, "Skip TLS certificate verification")
		wifi     = flag.Bool("wifi", false, "List only WiFi-connected clients")
		wired    = flag.Bool("wired", false, "List only wired (ethernet) clients")
		all      = flag.Bool("all", false, "List all connected clients (default)")
		jsonOut  = flag.Bool("json", false, "Output in JSON format")
		since    = flag.String("since", "", "Show devices seen within this window (present + gone), e.g. 7d, 24h, 3mo")
		sortKey  = flag.String("sort", "first-seen", "Sort order: first-seen, last-seen, or ip")
	)

	var gone optionalDuration
	flag.Var(&gone, "gone", "Show only departed devices, optionally within a window (e.g. --gone=30d, default 7d)")

	flag.StringVar(host, "H", "", "UDM Pro host address (shorthand)")
	flag.IntVar(port, "p", 443, "UDM Pro port (shorthand)")
	flag.StringVar(site, "S", "default", "Site name (shorthand)")
	flag.BoolVar(insecure, "k", false, "Skip TLS certificate verification (shorthand)")
	flag.BoolVar(wifi, "w", false, "WiFi only (shorthand)")
	flag.BoolVar(wired, "e", false, "Wired only (shorthand)")
	flag.BoolVar(all, "a", false, "All clients (shorthand)")
	flag.BoolVar(jsonOut, "j", false, "JSON output (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "List connected clients on a UniFi UDM Pro with OUI manufacturer lookup.\n\n")
		fmt.Fprintf(os.Stderr, "Filter:\n")
		fmt.Fprintf(os.Stderr, "  -w, --wifi\t\tList only WiFi clients\n")
		fmt.Fprintf(os.Stderr, "  -e, --wired\t\tList only wired clients\n")
		fmt.Fprintf(os.Stderr, "  -a, --all\t\tList all clients (default)\n\n")
		fmt.Fprintf(os.Stderr, "History:\n")
		fmt.Fprintf(os.Stderr, "  --since string\tShow devices seen within window (present + gone), e.g. 7d, 24h, 3mo\n")
		fmt.Fprintf(os.Stderr, "  --gone[=string]\tShow only departed devices, optional window (default %s)\n", defaultGoneWindow)
		fmt.Fprintf(os.Stderr, "  --sort string\t\tSort order: first-seen (default), last-seen, ip\n\n")
		fmt.Fprintf(os.Stderr, "Output:\n")
		fmt.Fprintf(os.Stderr, "  -j, --json\t\tOutput in JSON format\n\n")
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
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -w\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k -e -j\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k --gone=30d\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -k --since 7d\n", os.Args[0])
	}

	flag.Parse()

	// Validate filter exclusivity
	if *wifi && *wired {
		exitError("--wifi and --wired are mutually exclusive")
	}
	if *since != "" && gone.set {
		exitError("--since and --gone are mutually exclusive")
	}

	sortMode, err := parseSortMode(*sortKey)
	if err != nil {
		exitError(err.Error())
	}

	filter := FilterAll
	if *wifi {
		filter = FilterWifi
	} else if *wired {
		filter = FilterWired
	}

	historyMode := *since != "" || gone.set
	var withinHours int
	if historyMode {
		windowString := *since
		if gone.set {
			windowString = gone.value
			if windowString == "" {
				windowString = defaultGoneWindow
			}
		}
		window, parseErr := parseDuration(windowString)
		if parseErr != nil {
			exitError(parseErr.Error())
		}
		withinHours = durationToHours(window)
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

	// Load OUI database before connecting
	ouiDatabase, err := LoadOUIDatabase()
	if err != nil {
		exitError(err.Error())
	}

	fmt.Fprintf(os.Stderr, "Connecting to %s...\n", *host)
	config := &gofi.Config{
		Host:          *host,
		Port:          *port,
		Username:      username,
		Password:      password,
		Site:          *site,
		SkipTLSVerify: *insecure,
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

	if historyMode && filter != FilterAll {
		fmt.Fprintf(os.Stderr, "Warning: link-type filtering may be unreliable for departed devices\n")
	}

	now := time.Now().Unix()
	var entries []ClientEntry
	if historyMode {
		fmt.Fprintf(os.Stderr, "Fetching client history...\n")
		entries, err = ListClientsHistory(ctx, apiClient, *site, filter, sortMode, withinHours, gone.set, ouiDatabase)
	} else {
		fmt.Fprintf(os.Stderr, "Fetching active clients...\n")
		entries, err = ListClients(ctx, apiClient, *site, filter, sortMode, ouiDatabase)
	}
	if err != nil {
		exitError("failed to list clients: " + err.Error())
	}

	if *jsonOut {
		if err := FormatJSON(os.Stdout, entries); err != nil {
			exitError("failed to write JSON output: " + err.Error())
		}
	} else {
		if err := FormatText(os.Stdout, entries, historyMode, now); err != nil {
			exitError("failed to write output: " + err.Error())
		}
	}
}

func exitError(message string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	os.Exit(1)
}
