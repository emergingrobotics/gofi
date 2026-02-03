package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/types"
)

const (
	envUsername = "UNIFI_USERNAME"
	envPassword = "UNIFI_PASSWORD"
)

// PoE actions
const (
	ActionEnable  = "enable"
	ActionDisable = "disable"
	ActionCycle   = "cycle"
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
	Meta     OutputMeta    `json:"meta"`
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
	TotalSwitches     int `json:"total_switches"`
	TotalPorts        int `json:"total_ports"`
	TotalPoEPorts     int `json:"total_poe_ports"`
	TotalMaxPowerW    int `json:"total_max_power_watts"`
	ConnectedCount    int `json:"connected_count"`
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
	Index        int     `json:"index"`
	Name         string  `json:"name"`
	Up           bool    `json:"up"`
	Speed        int     `json:"speed_mbps"`
	PoECapable   bool    `json:"poe_capable"`
	PoEEnabled   bool    `json:"poe_enabled"`
	PoEGood      bool    `json:"poe_good"`
	PoEMode      string  `json:"poe_mode"`
	PoEPowerW    float64 `json:"poe_power_watts"`
	PoEVoltageV  float64 `json:"poe_voltage_v"`
	PoECurrentMA float64 `json:"poe_current_ma"`
	PoEClass     string  `json:"poe_class"`
}

// PoEControlOutput is the JSON output for PoE control actions.
type PoEControlOutput struct {
	Meta          OutputMeta       `json:"meta"`
	Action        string           `json:"action"`
	Switch        SwitchIdentifier `json:"switch"`
	Port          int              `json:"port"`
	PreviousState string           `json:"previous_state"`
	NewState      string           `json:"new_state"`
	ActualState   string           `json:"actual_state,omitempty"`
	Success       bool             `json:"success"`
	Message       string           `json:"message,omitempty"`
	Duration      string           `json:"duration,omitempty"`
	WaitTime      string           `json:"wait_time,omitempty"`
}

// SwitchIdentifier holds switch identification info.
type SwitchIdentifier struct {
	Name string `json:"name"`
	MAC  string `json:"mac"`
	ID   string `json:"id"`
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

		// Commands
		list = flag.Bool("list", false, "List all switches")

		// PoE control
		poeAction     = flag.String("poe", "", "PoE action: enable, disable, cycle")
		switchName    = flag.String("switch", "", "Switch name or MAC address")
		portNum       = flag.Int("port-num", 0, "Port number for PoE control")
		cycleDuration = flag.Duration("duration", 500*time.Millisecond, "Power cycle duration (for cycle action)")
		wait          = flag.Bool("wait", false, "Wait for state change to complete")
		waitTimeout   = flag.Duration("wait-timeout", 30*time.Second, "Timeout when waiting for state change")
		settleTime    = flag.Duration("settle", 2*time.Second, "Hardware settle time after config applies")
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
	flag.StringVar(poeAction, "P", "", "PoE action (shorthand)")
	flag.StringVar(switchName, "S", "", "Switch name or MAC (shorthand)")
	flag.IntVar(portNum, "n", 0, "Port number (shorthand)")
	flag.DurationVar(cycleDuration, "D", 500*time.Millisecond, "Cycle duration (shorthand)")
	flag.BoolVar(wait, "w", false, "Wait for state change (shorthand)")
	flag.DurationVar(waitTimeout, "W", 30*time.Second, "Wait timeout (shorthand)")
	flag.DurationVar(settleTime, "T", 2*time.Second, "Settle time (shorthand)")

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Manage UniFi switches via UDM Pro controller.\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  -l, --list                    List all switches\n")
		fmt.Fprintf(os.Stderr, "  -P, --poe <action>            PoE control: enable, disable, cycle\n\n")
		fmt.Fprintf(os.Stderr, "PoE Control Options (required with --poe):\n")
		fmt.Fprintf(os.Stderr, "  -S, --switch <name|mac>       Switch name or MAC address\n")
		fmt.Fprintf(os.Stderr, "  -n, --port-num <num>          Port number\n")
		fmt.Fprintf(os.Stderr, "  -D, --duration <duration>     Cycle duration (default 500ms)\n")
		fmt.Fprintf(os.Stderr, "  -w, --wait                    Wait for state change to complete\n")
		fmt.Fprintf(os.Stderr, "  -W, --wait-timeout <duration> Timeout for wait (default 30s)\n")
		fmt.Fprintf(os.Stderr, "  -T, --settle <duration>       Hardware settle time (default 2s)\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s               Username for UDM authentication (required)\n", envUsername)
		fmt.Fprintf(os.Stderr, "  %s               Password for UDM authentication (required)\n\n", envPassword)
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -H, --host <host>             UDM Pro host address (required)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port <port>             UDM Pro port (default 443)\n")
		fmt.Fprintf(os.Stderr, "  -s, --site <site>             Site name (default \"default\")\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure                Skip TLS certificate verification\n")
		fmt.Fprintf(os.Stderr, "  -j, --json                    Output in JSON format\n")
		fmt.Fprintf(os.Stderr, "  -d, --debug                   Enable debug output\n")
		fmt.Fprintf(os.Stderr, "  -t, --timeout <duration>      Connection timeout (default 30s)\n")
		fmt.Fprintf(os.Stderr, "  -h, --help                    Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -l\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -l -j\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -P enable -S OfficeSW -n 5\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -P disable -S OfficeSW -n 5 -w\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -P cycle -S OfficeSW -n 5 -D 2s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -H 192.168.1.1 -P enable -S fc:ec:da:40:22:f9 -n 6 -w -W 60s\n", os.Args[0])
	}

	flag.Parse()

	// Validate required parameters
	if *host == "" {
		fmt.Fprintf(os.Stderr, "Error: --host is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate commands
	hasListCmd := *list
	hasPoeCmd := *poeAction != ""

	if !hasListCmd && !hasPoeCmd {
		fmt.Fprintf(os.Stderr, "Error: no command specified (use --list or --poe)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate PoE command options
	if hasPoeCmd {
		action := strings.ToLower(*poeAction)
		if action != ActionEnable && action != ActionDisable && action != ActionCycle {
			fmt.Fprintf(os.Stderr, "Error: --poe must be one of: enable, disable, cycle\n\n")
			flag.Usage()
			os.Exit(1)
		}
		if *switchName == "" {
			fmt.Fprintf(os.Stderr, "Error: --switch is required with --poe\n\n")
			flag.Usage()
			os.Exit(1)
		}
		if *portNum <= 0 {
			fmt.Fprintf(os.Stderr, "Error: --port-num is required with --poe (must be > 0)\n\n")
			flag.Usage()
			os.Exit(1)
		}
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

	// Execute commands
	if hasListCmd {
		if err := listSwitches(ctx, client, *host, *site, *jsonOut, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	if hasPoeCmd {
		action := strings.ToLower(*poeAction)
		if err := controlPoE(ctx, client, *host, *site, action, *switchName, *portNum, *cycleDuration, *wait, *waitTimeout, *settleTime, *jsonOut, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func controlPoE(ctx context.Context, client gofi.Client, host, site, action, switchNameOrMAC string, portNum int, cycleDuration time.Duration, wait bool, waitTimeout, settleTime time.Duration, jsonOut, debug bool) error {
	// Find the switch
	sw, err := findSwitch(ctx, client, site, switchNameOrMAC)
	if err != nil {
		return err
	}

	if debug {
		log.Printf("Found switch: %s (%s)", sw.Name, sw.MAC)
	}

	// Validate port exists and is PoE capable
	var targetPort *types.PortTable
	for i := range sw.PortTable {
		if sw.PortTable[i].PortIdx == portNum {
			targetPort = &sw.PortTable[i]
			break
		}
	}

	if targetPort == nil {
		return fmt.Errorf("port %d not found on switch %s", portNum, sw.Name)
	}

	if !targetPort.PortPoe {
		return fmt.Errorf("port %d on switch %s is not PoE capable", portNum, sw.Name)
	}

	// Get previous state
	previousState := "disabled"
	if targetPort.PoeEnable {
		previousState = "enabled"
	}

	// Execute action
	var newState string
	var expectedEnabled bool
	var message string

	switch action {
	case ActionEnable:
		if err := setPoEMode(ctx, client, site, sw, portNum, "auto"); err != nil {
			return err
		}
		newState = "enabled"
		expectedEnabled = true
		message = fmt.Sprintf("PoE enabled on port %d", portNum)

	case ActionDisable:
		if err := setPoEMode(ctx, client, site, sw, portNum, "off"); err != nil {
			return err
		}
		newState = "disabled"
		expectedEnabled = false
		message = fmt.Sprintf("PoE disabled on port %d", portNum)

	case ActionCycle:
		if err := cyclePoE(ctx, client, site, sw, portNum, cycleDuration, debug); err != nil {
			return err
		}
		newState = previousState // Returns to previous state after cycle
		expectedEnabled = targetPort.PoeEnable
		message = fmt.Sprintf("PoE power cycled on port %d (duration: %v)", portNum, cycleDuration)
	}

	// Wait for state change if requested (not applicable to cycle)
	var actualState string
	var waitTime time.Duration
	var waitErr error
	if wait && action != ActionCycle {
		waitStart := time.Now()
		actualState, waitErr = waitForState(ctx, client, site, sw.MAC, portNum, expectedEnabled, waitTimeout, settleTime, debug)
		waitTime = time.Since(waitStart)
		if waitErr != nil {
			actualState = "unknown"
			message = fmt.Sprintf("%s (wait failed: %v)", message, waitErr)
		} else {
			message = fmt.Sprintf("%s (confirmed after %v)", message, waitTime.Round(time.Millisecond))
		}
	}

	// Output result
	if jsonOut {
		output := PoEControlOutput{
			Meta: OutputMeta{
				Host:      host,
				Site:      site,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Version:   "1.0",
			},
			Action: action,
			Switch: SwitchIdentifier{
				Name: sw.Name,
				MAC:  sw.MAC,
				ID:   sw.ID,
			},
			Port:          portNum,
			PreviousState: previousState,
			NewState:      newState,
			Success:       true,
			Message:       message,
		}
		if action == ActionCycle {
			output.Duration = cycleDuration.String()
		}
		if wait && action != ActionCycle {
			output.ActualState = actualState
			output.WaitTime = waitTime.Round(time.Millisecond).String()
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	// Text output
	fmt.Printf("Switch:    %s (%s)\n", sw.Name, sw.MAC)
	fmt.Printf("Port:      %d\n", portNum)
	fmt.Printf("Action:    %s\n", action)
	fmt.Printf("Previous:  %s\n", previousState)
	fmt.Printf("New State: %s\n", newState)
	if action == ActionCycle {
		fmt.Printf("Duration:  %v\n", cycleDuration)
	}
	if wait && action != ActionCycle {
		fmt.Printf("Actual:    %s\n", actualState)
		fmt.Printf("Wait Time: %v\n", waitTime.Round(time.Millisecond))
		if waitErr != nil {
			fmt.Printf("Wait Err:  %v\n", waitErr)
		}
	}
	fmt.Printf("Status:    Success\n")

	return nil
}

// waitForState polls the switch until the port config is applied, then waits for hardware settle time.
func waitForState(ctx context.Context, client gofi.Client, site, mac string, portNum int, expectedEnabled bool, timeout, settleTime time.Duration, debug bool) (string, error) {
	const pollInterval = 100 * time.Millisecond
	deadline := time.Now().Add(timeout)

	expectedMode := "off"
	if expectedEnabled {
		expectedMode = "auto"
	}

	// Phase 1: Wait for config to be applied (poe_mode changes)
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timeout waiting for config to apply")
		}

		// Fetch fresh device data
		sw, err := client.Devices().GetByMAC(ctx, site, mac)
		if err != nil {
			if debug {
				log.Printf("Error fetching device: %v", err)
			}
			time.Sleep(pollInterval)
			continue
		}

		// Find the port
		for _, p := range sw.PortTable {
			if p.PortIdx == portNum {
				if debug {
					log.Printf("Port %d: poe_mode=%s (want %s)", portNum, p.PoeMode, expectedMode)
				}

				// Check if poe_mode matches expected
				modeMatches := (expectedEnabled && p.PoeMode != "off" && p.PoeMode != "") ||
					(!expectedEnabled && p.PoeMode == "off")

				if modeMatches {
					if debug {
						log.Printf("Config applied, waiting %v for hardware settle", settleTime)
					}
					// Phase 2: Wait for hardware to settle
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case <-time.After(settleTime):
						// Done - config applied and hardware had time to settle
						if expectedEnabled {
							return "enabled", nil
						}
						return "disabled", nil
					}
				}
				break
			}
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(pollInterval):
			// Continue polling
		}
	}
}

func findSwitch(ctx context.Context, client gofi.Client, site, nameOrMAC string) (*types.Device, error) {
	devices, err := client.Devices().List(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Normalize search term
	searchLower := strings.ToLower(nameOrMAC)
	searchMAC := strings.ToLower(strings.ReplaceAll(nameOrMAC, ":", ""))

	for _, d := range devices {
		if d.Type != "usw" {
			continue
		}

		// Match by name (case-insensitive)
		if strings.ToLower(d.Name) == searchLower {
			return &d, nil
		}

		// Match by MAC (normalized)
		deviceMAC := strings.ToLower(strings.ReplaceAll(d.MAC, ":", ""))
		if deviceMAC == searchMAC {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("switch not found: %s", nameOrMAC)
}

func setPoEMode(ctx context.Context, client gofi.Client, site string, sw *types.Device, portIdx int, mode string) error {
	// Build port overrides - we need to preserve existing overrides and update/add ours
	overrides := make([]types.PortOverride, 0, len(sw.PortOverrides)+1)

	found := false
	for _, po := range sw.PortOverrides {
		if po.PortIdx == portIdx {
			// Update existing override
			po.PoeMode = mode
			overrides = append(overrides, po)
			found = true
		} else {
			overrides = append(overrides, po)
		}
	}

	if !found {
		// Add new override
		overrides = append(overrides, types.PortOverride{
			PortIdx: portIdx,
			PoeMode: mode,
		})
	}

	// Create update request - must include name to preserve it
	updateReq := &types.Device{
		ID:            sw.ID,
		Name:          sw.Name, // Preserve existing name
		PortOverrides: overrides,
	}

	_, err := client.Devices().Update(ctx, site, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	// Force provision to apply the change immediately
	if err := client.Devices().ForceProvision(ctx, site, sw.MAC); err != nil {
		return fmt.Errorf("failed to provision device: %w", err)
	}

	return nil
}

func cyclePoE(ctx context.Context, client gofi.Client, site string, sw *types.Device, portIdx int, duration time.Duration, debug bool) error {
	// Use the built-in power-cycle command
	if err := client.Devices().PowerCyclePort(ctx, site, sw.MAC, portIdx); err != nil {
		return fmt.Errorf("failed to power cycle port: %w", err)
	}

	if debug {
		log.Printf("Power cycle initiated")
	}

	// Force provision to ensure the command is applied
	if err := client.Devices().ForceProvision(ctx, site, sw.MAC); err != nil {
		return fmt.Errorf("failed to provision device: %w", err)
	}

	// Wait for the specified duration if longer than default
	if duration > time.Second {
		if debug {
			log.Printf("Waiting %v for extended cycle duration", duration)
		}
		time.Sleep(duration)
	}

	return nil
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
			portInfo := PortInfo{
				Index:      p.PortIdx,
				Name:       p.Name,
				Up:         p.Up,
				Speed:      p.Speed,
				PoECapable: p.PortPoe,
			}

			if p.PortPoe {
				// PoE-capable port: populate all PoE fields
				portInfo.PoEEnabled = p.PoeEnable
				portInfo.PoEGood = p.PoeGood
				portInfo.PoEMode = p.PoeMode
				if portInfo.PoEMode == "" {
					portInfo.PoEMode = "off"
				}
				portInfo.PoEPowerW = p.PoePower.Float64()
				portInfo.PoEVoltageV = p.PoeVoltage.Float64()
				portInfo.PoECurrentMA = p.PoeCurrent.Float64()
				portInfo.PoEClass = p.PoeClass
				if portInfo.PoEClass == "" {
					portInfo.PoEClass = "Unknown"
				}
			} else {
				// Non-PoE port: use N/A indicators
				portInfo.PoEMode = "N/A"
				portInfo.PoEClass = "N/A"
			}

			info.Ports[i] = portInfo
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
