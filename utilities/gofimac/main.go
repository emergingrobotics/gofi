package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/utilities/internal/conn"
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
		host     = flag.String("host", "", "UniFi controller address")
		port     = flag.Int("port", 443, "UniFi controller port")
		site     = flag.String("site", "default", "Site name")
		secure   = flag.Bool("secure", false, "Enforce TLS certificate verification")
		wifi     = flag.Bool("wifi", false, "List only WiFi-connected clients")
		wired    = flag.Bool("wired", false, "List only wired (ethernet) clients")
		all      = flag.Bool("all", false, "List all connected clients (default)")
		jsonOut  = flag.Bool("json", false, "Output in JSON format")
		since    = flag.String("since", "", "Show devices seen within this window (present + gone), e.g. 7d, 24h, 3mo")
		sortKey  = flag.String("sort", "first-seen", "Sort order: first-seen, last-seen, or ip")
		macProbe = flag.String("mac", "", "Probe a single MAC: report present/gone; exit 0 if present, 1 otherwise")
	)
	flag.StringVar(macProbe, "m", "", "Probe a single MAC (shorthand)")

	var gone optionalDuration
	flag.Var(&gone, "gone", "Show only departed devices, optionally within a window (e.g. --gone=30d, default 7d)")

	flag.StringVar(host, "H", "", "UniFi controller address (shorthand)")
	flag.IntVar(port, "p", 443, "UniFi controller port (shorthand)")
	flag.StringVar(site, "S", "default", "Site name (shorthand)")
	flag.BoolVar(secure, "k", false, "Enforce TLS certificate verification (shorthand)")
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
		fmt.Fprintf(os.Stderr, "Probe:\n")
		fmt.Fprintf(os.Stderr, "  -m, --mac string\tProbe one MAC; exit 0 if present, 1 if gone/not found\n\n")
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
		fmt.Fprintf(os.Stderr, "  %s\tUniFi controller (fallback for -H)\n\n", conn.EnvControllerIP)
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

	// --mac probes a single device and is incompatible with the listing modes.
	var probeMAC string
	if *macProbe != "" {
		if *since != "" || gone.set {
			exitError("--mac cannot be combined with --since or --gone")
		}
		parsed, parseErr := net.ParseMAC(*macProbe)
		if parseErr != nil {
			exitError(fmt.Sprintf("invalid MAC %q: %v", *macProbe, parseErr))
		}
		probeMAC = parsed.String()
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

	config, err := conn.ResolveConfig(os.Stderr, *host, *port, *site, *secure)
	if err != nil {
		exitError(err.Error())
	}

	// Load OUI database before connecting
	ouiDatabase, err := LoadOUIDatabase()
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

	now := time.Now().Unix()

	if probeMAC != "" {
		fmt.Fprintf(os.Stderr, "Probing %s...\n", probeMAC)
		entry, probeErr := FindClientByMAC(ctx, apiClient, *site, probeMAC, ouiDatabase)
		if probeErr != nil {
			exitError("failed to probe MAC: " + probeErr.Error())
		}
		code, writeErr := reportProbe(os.Stdout, os.Stderr, entry, probeMAC, *jsonOut, now)
		if writeErr != nil {
			exitError("failed to write output: " + writeErr.Error())
		}
		os.Exit(code)
	}

	if historyMode && filter != FilterAll {
		fmt.Fprintf(os.Stderr, "Warning: link-type filtering may be unreliable for departed devices\n")
	}

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

// reportProbe writes the outcome of a single-MAC probe and returns the process
// exit code: 0 when the device is present, 1 when it is gone or was not found.
// Device output goes to out; the not-found status message goes to errOut.
func reportProbe(out, errOut io.Writer, entry *ClientEntry, mac string, jsonOut bool, now int64) (int, error) {
	if entry == nil {
		fmt.Fprintf(errOut, "MAC %s not found\n", mac)
		if jsonOut {
			return 1, FormatJSON(out, []ClientEntry{})
		}
		return 1, nil
	}

	entries := []ClientEntry{*entry}
	var writeErr error
	if jsonOut {
		writeErr = FormatJSON(out, entries)
	} else {
		writeErr = FormatText(out, entries, true, now)
	}
	if writeErr != nil {
		return 1, writeErr
	}

	if entry.Status == statusPresent {
		return 0, nil
	}
	return 1, nil
}

func exitError(message string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	os.Exit(1)
}
