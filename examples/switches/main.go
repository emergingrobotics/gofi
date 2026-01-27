package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

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

// SwitchListOutput is the top-level JSON output for the list command.
type SwitchListOutput struct {
	Meta     OutputMeta   `json:"meta"`
	Summary  SwitchSummary `json:"summary"`
	Switches []SwitchInfo  `json:"switches"`
}

// OutputMeta contains metadata about the query.
type OutputMeta struct {
	Host      string `json:"host"`
	Site      string `json:"site"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// SwitchSummary contains aggregate statistics.
type SwitchSummary struct {
	TotalSwitches   int `json:"total_switches"`
	TotalPorts      int `json:"total_ports"`
	TotalPoEPorts   int `json:"total_poe_ports"`
	TotalMaxPowerW  int `json:"total_max_power_watts"`
	ConnectedCount  int `json:"connected_count"`
	DisconnectedCount int `json:"disconnected_count"`
}

// SwitchInfo holds switch information for output.
type SwitchInfo struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Model         string     `json:"model"`
	MAC           string     `json:"mac"`
	IP            string     `json:"ip,omitempty"`
	Version       string     `json:"version"`
	State         string     `json:"state"`
	StateCode     int        `json:"state_code"`
	Uptime        int64      `json:"uptime_seconds"`
	UptimeHuman   string     `json:"uptime_human"`
	NumPorts      int        `json:"num_ports"`
	PoEPorts      int        `json:"poe_ports"`
	TotalMaxPower int        `json:"total_max_power_watts,omitempty"`
	Adopted       bool       `json:"adopted"`
	Ports         []PortInfo `json:"ports,omitempty"`
}

// PortInfo holds per-port information.
type PortInfo struct {
	Index       int     `json:"index"`
	Name        string  `json:"name,omitempty"`
	Up          bool    `json:"up"`
	Speed       int     `json:"speed_mbps"`
	PoECapable  bool    `json:"poe_capable"`
	PoEEnabled  bool    `json:"poe_enabled,omitempty"`
	PoEGood     bool    `json:"poe_good,omitempty"`
	PoEMode     string  `json:"poe_mode,omitempty"`
	PoEPowerW   float64 `json:"poe_power_watts,omitempty"`
	PoEVoltageV float64 `json:"poe_voltage_v,omitempty"`
	PoECurrentMA float64 `json:"poe_current_ma,omitempty"`
	PoEClass    string  `json:"poe_class,omitempty"`
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
		timeout  = flag.Duration("timeout", 30*time.Second, "Connection timeout")
		list     = flag.Bool("list", false, "List all switches")
	)

	// Add short flag aliases
	flag.StringVar(host, "H", "", "UDM Pro host address (shorthand)")
	flag.IntVar(port, "p", 443, "UDM Pro port (shorthand)")
	flag.StringVar(site, "s", "default", "Site name (shorthand)")
	flag.BoolVar(insecure, "k", false, "Skip TLS certificate verification (shorthand)")
	flag.BoolVar(jsonOut, "j", false, "Output in JSON format (shorthand)")
	flag.BoolVar(debug, "d", false, "Enable debug output (shorthand)")
	flag.DurationVar(timeout, "t", 30*time.Second, "Connection timeout (shorthand)")
	flag.BoolVar(list, "l", false, "List all switches (shorthand)")

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Manage UniFi switches via UDM Pro controller.\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  -l, --list\t\tList all switches\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tUsername for UDM authentication (required)\n", envUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword for UDM authentication (required)\n\n", envPassword)
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\t\tUDM Pro host address (required)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port int\t\tUDM Pro port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -s, --site string\t\tSite name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure\t\tSkip TLS certificate verification\n")
		fmt.Fprintf(os.Stderr, "  -j, --json\t\t\tOutput in JSON format\n")
		fmt.Fprintf(os.Stderr, "  -d, --debug\t\t\tEnable debug output\n")
		fmt.Fprintf(os.Stderr, "  -t, --timeout duration\tConnection timeout (default 30s)\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\t\t\tShow this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s --host 192.168.1.1 --list\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -l -j\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -l -k --site production\n", os.Args[0])
	}

	flag.Parse()

	// Validate required parameters
	if *host == "" {
		fmt.Fprintf(os.Stderr, "Error: --host is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Require at least one command
	if !*list {
		fmt.Fprintf(os.Stderr, "Error: no command specified (use --list)\n\n")
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
		Timeout:       *timeout,
	}

	if *debug {
		config.Logger = &debugLogger{}
		log.Printf("Connecting to %s:%d as user %s", *host, *port, username)
	}

	// Create client
	client, err := gofi.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Connect to controller
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	if *debug {
		log.Printf("Connected to UniFi controller at %s:%d", *host, *port)
	}

	// Execute command
	if *list {
		if err := listSwitches(ctx, client, *host, *site, *jsonOut, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func listSwitches(ctx context.Context, client gofi.Client, host, site string, jsonOut, debug bool) error {
	// Fetch all devices
	devices, err := client.Devices().List(ctx, site)
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	if debug {
		log.Printf("Retrieved %d devices", len(devices))
	}

	// Filter to switches only (type == "usw")
	var switches []types.Device
	for _, d := range devices {
		if d.Type == "usw" {
			switches = append(switches, d)
		}
	}

	if debug {
		log.Printf("Found %d switches", len(switches))
	}

	// Convert to output format
	infos := make([]SwitchInfo, len(switches))
	for i, sw := range switches {
		infos[i] = toSwitchInfo(sw, jsonOut)
	}

	// Output results
	if jsonOut {
		output := buildJSONOutput(host, site, infos)
		return outputJSON(output)
	}
	outputText(infos)
	return nil
}

func toSwitchInfo(d types.Device, includePorts bool) SwitchInfo {
	// Count total ports and PoE ports
	numPorts := len(d.PortTable)
	poePorts := 0
	for _, p := range d.PortTable {
		if p.PortPoe {
			poePorts++
		}
	}

	info := SwitchInfo{
		ID:            d.ID,
		Name:          d.Name,
		Model:         d.Model,
		MAC:           d.MAC,
		IP:            d.IP,
		Version:       d.Version,
		State:         d.State.String(),
		StateCode:     int(d.State),
		Uptime:        d.Uptime.Int64(),
		UptimeHuman:   formatUptime(d.Uptime.Int64()),
		NumPorts:      numPorts,
		PoEPorts:      poePorts,
		TotalMaxPower: d.TotalMaxPower,
		Adopted:       d.Adopted,
	}

	// Include detailed port info for JSON output
	if includePorts && len(d.PortTable) > 0 {
		info.Ports = make([]PortInfo, len(d.PortTable))
		for i, p := range d.PortTable {
			info.Ports[i] = PortInfo{
				Index:        p.PortIdx,
				Name:         p.Name,
				Up:           p.Up,
				Speed:        p.Speed,
				PoECapable:   p.PortPoe,
				PoEEnabled:   p.PoeEnable,
				PoEGood:      p.PoeGood,
				PoEMode:      p.PoeMode,
				PoEPowerW:    p.PoePower.Float64() / 1000.0,
				PoEVoltageV:  p.PoeVoltage.Float64() / 1000.0,
				PoECurrentMA: p.PoeCurrent.Float64(),
				PoEClass:     p.PoeClass,
			}
		}
	}

	return info
}

func buildJSONOutput(host, site string, switches []SwitchInfo) SwitchListOutput {
	// Calculate summary statistics
	summary := SwitchSummary{
		TotalSwitches: len(switches),
	}

	for _, sw := range switches {
		summary.TotalPorts += sw.NumPorts
		summary.TotalPoEPorts += sw.PoEPorts
		summary.TotalMaxPowerW += sw.TotalMaxPower

		if sw.StateCode == 1 { // Connected
			summary.ConnectedCount++
		} else {
			summary.DisconnectedCount++
		}
	}

	return SwitchListOutput{
		Meta: OutputMeta{
			Host:      host,
			Site:      site,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   "1.0",
		},
		Summary:  summary,
		Switches: switches,
	}
}

func formatUptime(seconds int64) string {
	if seconds <= 0 {
		return "-"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func outputJSON(output SwitchListOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

func outputText(infos []SwitchInfo) {
	if len(infos) == 0 {
		fmt.Println("No switches found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tMODEL\tMAC\tIP\tSTATE\tPORTS\tPOE\tMAX POWER")
	fmt.Fprintln(w, "----\t-----\t---\t--\t-----\t-----\t---\t---------")

	for _, info := range infos {
		ip := info.IP
		if ip == "" {
			ip = "-"
		}

		maxPower := "-"
		if info.TotalMaxPower > 0 {
			maxPower = fmt.Sprintf("%dW", info.TotalMaxPower)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%d\t%s\n",
			info.Name,
			info.Model,
			info.MAC,
			ip,
			info.State,
			info.NumPorts,
			info.PoEPorts,
			maxPower,
		)
	}

	w.Flush()

	fmt.Printf("\nTotal: %d switch(es)\n", len(infos))
}
