package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// clientStationService implements ClientService.
type clientStationService struct {
	transport transport.Transport
}

// NewClientService creates a new client service.
func NewClientService(transport transport.Transport) ClientService {
	return &clientStationService{
		transport: transport,
	}
}

// ListActive returns all currently connected clients.
func (s *clientStationService) ListActive(ctx context.Context, site string) ([]types.Client, error) {
	path := fmt.Sprintf("/api/s/%s/stat/sta", site)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list active clients: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list active clients failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Client](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// ListAll returns all known clients (including historical).
func (s *clientStationService) ListAll(ctx context.Context, site string, opts ...ClientListOption) ([]types.Client, error) {
	options := &clientListOptions{
		withinHours: 8760, // Default: 1 year
	}
	for _, opt := range opts {
		opt(options)
	}

	path := fmt.Sprintf("/api/s/%s/stat/alluser", site)
	if options.withinHours > 0 {
		path += "?within=" + strconv.Itoa(options.withinHours)
	}

	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list all clients: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list all clients failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Client](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// Get returns a client by MAC address.
func (s *clientStationService) Get(ctx context.Context, site, mac string) (*types.Client, error) {
	// Get all active clients and find the one with matching MAC
	clients, err := s.ListActive(ctx, site)
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		if client.MAC == mac {
			return &client, nil
		}
	}

	return nil, fmt.Errorf("client not found: %s", mac)
}

// Block blocks a client from the network.
func (s *clientStationService) Block(ctx context.Context, site, mac string) error {
	return s.executeCommand(ctx, site, "block-sta", mac, nil)
}

// Unblock unblocks a previously blocked client.
func (s *clientStationService) Unblock(ctx context.Context, site, mac string) error {
	return s.executeCommand(ctx, site, "unblock-sta", mac, nil)
}

// Kick disconnects a client from the network.
func (s *clientStationService) Kick(ctx context.Context, site, mac string) error {
	return s.executeCommand(ctx, site, "kick-sta", mac, nil)
}

// AuthorizeGuest authorizes a guest client.
func (s *clientStationService) AuthorizeGuest(ctx context.Context, site, mac string, opts ...GuestAuthOption) error {
	options := &guestAuthOptions{}
	for _, opt := range opts {
		opt(options)
	}

	payload := map[string]interface{}{
		"cmd": "authorize-guest",
		"mac": mac,
	}

	if options.minutes > 0 {
		payload["minutes"] = options.minutes
	}
	if options.up > 0 {
		payload["up"] = options.up
	}
	if options.down > 0 {
		payload["down"] = options.down
	}
	if options.bytes > 0 {
		payload["bytes"] = options.bytes
	}
	if options.apMAC != "" {
		payload["ap_mac"] = options.apMAC
	}

	path := fmt.Sprintf("/api/s/%s/cmd/stamgr", site)
	req := transport.NewRequest("POST", path).WithBody(payload)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to authorize guest: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("authorize guest failed with status %d", resp.StatusCode)
	}

	return nil
}

// UnauthorizeGuest revokes guest authorization.
func (s *clientStationService) UnauthorizeGuest(ctx context.Context, site, mac string) error {
	return s.executeCommand(ctx, site, "unauthorize-guest", mac, nil)
}

// Forget removes a client from the known clients list.
func (s *clientStationService) Forget(ctx context.Context, site, mac string) error {
	return s.executeCommand(ctx, site, "forget-sta", mac, nil)
}

// SetFingerprint overrides the device fingerprint.
func (s *clientStationService) SetFingerprint(ctx context.Context, site, mac string, devID int) error {
	payload := map[string]interface{}{
		"dev_id": devID,
	}
	return s.executeCommand(ctx, site, "set-sta-dev-id", mac, payload)
}

// executeCommand executes a client management command.
func (s *clientStationService) executeCommand(ctx context.Context, site, cmd, mac string, extra map[string]interface{}) error {
	payload := map[string]interface{}{
		"cmd": cmd,
		"mac": mac,
	}

	// Merge extra fields
	for k, v := range extra {
		payload[k] = v
	}

	path := fmt.Sprintf("/api/s/%s/cmd/stamgr", site)
	req := transport.NewRequest("POST", path).WithBody(payload)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute client command %s: %w", cmd, err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("client command %s failed with status %d", cmd, resp.StatusCode)
	}

	return nil
}
