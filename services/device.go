package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// deviceService implements DeviceService.
type deviceService struct {
	transport transport.Transport
}

// NewDeviceService creates a new device service.
func NewDeviceService(transport transport.Transport) DeviceService {
	return &deviceService{
		transport: transport,
	}
}

// List returns all devices for a site.
func (s *deviceService) List(ctx context.Context, site string) ([]types.Device, error) {
	path := internal.BuildAPIPath(site, "stat/device")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list devices failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Device](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// ListBasic returns basic device information for faster queries.
func (s *deviceService) ListBasic(ctx context.Context, site string) ([]types.DeviceBasic, error) {
	path := internal.BuildAPIPath(site, "basicstat/device")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list basic devices: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list basic devices failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.DeviceBasic](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// Get returns a specific device by ID.
func (s *deviceService) Get(ctx context.Context, site, id string) (*types.Device, error) {
	devices, err := s.List(ctx, site)
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		if device.ID == id {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("device not found: %s", id)
}

// GetByMAC returns a specific device by MAC address.
func (s *deviceService) GetByMAC(ctx context.Context, site, mac string) (*types.Device, error) {
	devices, err := s.List(ctx, site)
	if err != nil {
		return nil, err
	}

	// Normalize MAC for comparison
	normalizedMAC := strings.ToLower(strings.ReplaceAll(mac, ":", ""))

	for _, device := range devices {
		deviceMAC := strings.ToLower(strings.ReplaceAll(device.MAC, ":", ""))
		if deviceMAC == normalizedMAC {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("device not found with MAC: %s", mac)
}

// Update updates a device's configuration.
func (s *deviceService) Update(ctx context.Context, site string, device *types.Device) (*types.Device, error) {
	path := internal.BuildRESTPath(site, "device", device.ID)
	req := transport.NewRequest("PUT", path).WithBody(device)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update device failed with status %d", resp.StatusCode)
	}

	updated, err := internal.ParseSingleResult[types.Device](resp.Body)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Adopt adopts a device into the controller.
func (s *deviceService) Adopt(ctx context.Context, site, mac string) error {
	return s.sendCommand(ctx, site, "adopt", mac, nil)
}

// Forget removes a device from the controller.
func (s *deviceService) Forget(ctx context.Context, site, mac string) error {
	return s.sendCommand(ctx, site, "forget", mac, nil)
}

// Restart restarts a device.
func (s *deviceService) Restart(ctx context.Context, site, mac string) error {
	return s.sendCommand(ctx, site, "restart", mac, nil)
}

// ForceProvision forces provisioning of a device.
func (s *deviceService) ForceProvision(ctx context.Context, site, mac string) error {
	return s.sendCommand(ctx, site, "force-provision", mac, nil)
}

// Upgrade upgrades a device to the latest firmware.
func (s *deviceService) Upgrade(ctx context.Context, site, mac string) error {
	return s.sendCommand(ctx, site, "upgrade", mac, nil)
}

// UpgradeExternal upgrades a device using an external firmware URL.
func (s *deviceService) UpgradeExternal(ctx context.Context, site, mac, url string) error {
	return s.sendCommand(ctx, site, "upgrade-external", mac, map[string]interface{}{
		"url": url,
	})
}

// Locate enables the locate LED on a device.
func (s *deviceService) Locate(ctx context.Context, site, mac string) error {
	return s.sendCommand(ctx, site, "set-locate", mac, nil)
}

// Unlocate disables the locate LED on a device.
func (s *deviceService) Unlocate(ctx context.Context, site, mac string) error {
	return s.sendCommand(ctx, site, "unset-locate", mac, nil)
}

// PowerCyclePort power cycles a specific port on a switch.
func (s *deviceService) PowerCyclePort(ctx context.Context, site, switchMAC string, portIdx int) error {
	return s.sendCommand(ctx, site, "power-cycle", switchMAC, map[string]interface{}{
		"port_idx": portIdx,
	})
}

// SetLEDOverride sets the LED override mode for a device.
func (s *deviceService) SetLEDOverride(ctx context.Context, site, mac, mode string) error {
	return s.sendCommand(ctx, site, "set-led-override", mac, map[string]interface{}{
		"mode": mode,
	})
}

// SpectrumScan initiates a spectrum scan on an AP.
func (s *deviceService) SpectrumScan(ctx context.Context, site, mac string) error {
	return s.sendCommand(ctx, site, "spectrum-scan", mac, nil)
}

// sendCommand sends a device command.
func (s *deviceService) sendCommand(ctx context.Context, site, cmd, mac string, params map[string]interface{}) error {
	path := internal.BuildCmdPath(site, "devmgr")

	// Build command request
	cmdReq := map[string]interface{}{
		"cmd": cmd,
		"mac": mac,
	}

	// Add additional parameters
	for k, v := range params {
		cmdReq[k] = v
	}

	req := transport.NewRequest("POST", path).WithBody(cmdReq)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send command %s: %w", cmd, err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("command %s failed with status %d", cmd, resp.StatusCode)
	}

	return nil
}
