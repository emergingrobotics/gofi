package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handleDevices routes device-related requests.
func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// stat/device - list all devices
	if strings.Contains(path, "/stat/device") && r.Method == "GET" {
		s.handleDeviceStat(w, r, site)
		return
	}

	// basicstat/device - list device basics
	if strings.Contains(path, "/basicstat/device") && r.Method == "GET" {
		s.handleDeviceBasicStat(w, r, site)
		return
	}

	// rest/device/{id} - update device
	if strings.Contains(path, "/rest/device/") && r.Method == "PUT" {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "device" && i+1 < len(parts) {
				id := parts[i+1]
				s.handleDeviceUpdate(w, r, site, id)
				return
			}
		}
	}

	// cmd/devmgr - device commands
	if strings.Contains(path, "/cmd/devmgr") && r.Method == "POST" {
		s.handleDeviceCommand(w, r, site)
		return
	}

	writeNotFound(w)
}

// handleDeviceStat returns all devices for a site.
func (s *Server) handleDeviceStat(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	devices := s.state.ListDevices()

	// Convert to interface slice
	data := make([]interface{}, len(devices))
	for i, device := range devices {
		data[i] = *device
	}

	writeAPIResponse(w, data)
}

// handleDeviceBasicStat returns basic device info.
func (s *Server) handleDeviceBasicStat(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	devices := s.state.ListDevices()

	// Convert to basic format
	basics := make([]types.DeviceBasic, len(devices))
	for i, device := range devices {
		basics[i] = types.DeviceBasic{
			MAC:   device.MAC,
			Type:  device.Type,
			Model: device.Model,
			Name:  device.Name,
			State: device.State,
		}
	}

	// Convert to interface slice
	data := make([]interface{}, len(basics))
	for i, basic := range basics {
		data[i] = basic
	}

	writeAPIResponse(w, data)
}

// handleDeviceUpdate updates a device.
func (s *Server) handleDeviceUpdate(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "PUT" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	// Get existing device
	device, exists := s.state.GetDevice(id)
	if !exists {
		writeNotFound(w)
		return
	}

	// Parse update request
	var updateReq types.Device
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Update fields (only certain fields can be updated)
	if updateReq.Name != "" {
		device.Name = updateReq.Name
	}
	if updateReq.LEDOverride != "" {
		device.LEDOverride = updateReq.LEDOverride
	}
	if updateReq.LEDOverrideColor != "" {
		device.LEDOverrideColor = updateReq.LEDOverrideColor
	}

	// Save updated device
	s.state.AddDevice(device)

	// Return updated device
	writeAPIResponse(w, []interface{}{*device})
}

// handleDeviceCommand handles device commands.
func (s *Server) handleDeviceCommand(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "POST" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	// Parse command request
	var cmdReq types.CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&cmdReq); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Validate MAC address for most commands
	if cmdReq.MAC == "" && cmdReq.Cmd != "set-default" {
		writeBadRequest(w, "MAC address required")
		return
	}

	// Find device by MAC
	var device *types.Device
	if cmdReq.MAC != "" {
		devices := s.state.ListDevices()
		for i := range devices {
			if strings.EqualFold(devices[i].MAC, cmdReq.MAC) {
				device = devices[i]
				break
			}
		}

		if device == nil {
			writeAPIError(w, http.StatusNotFound, "error", "Device not found")
			return
		}
	}

	// Handle different commands
	switch cmdReq.Cmd {
	case "adopt":
		if device != nil {
			device.Adopted = true
			device.State = types.DeviceStateConnected
			s.state.AddDevice(device)
		}
	case "restart":
		// Simulate restart - no state change needed
	case "force-provision":
		if device != nil {
			device.State = types.DeviceStateProvisioning
			s.state.AddDevice(device)
		}
	case "upgrade":
		if device != nil {
			device.State = types.DeviceStateUpgrading
			s.state.AddDevice(device)
		}
	case "upgrade-external":
		if cmdReq.URL == "" {
			writeBadRequest(w, "URL required for external upgrade")
			return
		}
		if device != nil {
			device.State = types.DeviceStateUpgrading
			s.state.AddDevice(device)
		}
	case "set-locate":
		if device != nil {
			device.LEDOverride = "on"
			s.state.AddDevice(device)
		}
	case "unset-locate":
		if device != nil {
			device.LEDOverride = "default"
			s.state.AddDevice(device)
		}
	case "power-cycle":
		if cmdReq.PortIdx == 0 {
			writeBadRequest(w, "port_idx required")
			return
		}
		// Simulate power cycle - no state change needed
	case "spectrum-scan":
		// Simulate spectrum scan - no state change needed
	default:
		writeBadRequest(w, fmt.Sprintf("Unknown command: %s", cmdReq.Cmd))
		return
	}

	// Return success
	writeAPIResponse(w, []interface{}{})
}
