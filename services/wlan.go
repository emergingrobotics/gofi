package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// wlanService implements WLANService.
type wlanService struct {
	transport transport.Transport
}

// NewWLANService creates a new WLAN service.
func NewWLANService(transport transport.Transport) WLANService {
	return &wlanService{
		transport: transport,
	}
}

// List returns all WLANs for a site.
func (s *wlanService) List(ctx context.Context, site string) ([]types.WLAN, error) {
	path := internal.BuildRESTPath(site, "wlanconf", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list WLANs: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list WLANs failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.WLAN](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// Get returns a specific WLAN by ID.
func (s *wlanService) Get(ctx context.Context, site, id string) (*types.WLAN, error) {
	path := internal.BuildRESTPath(site, "wlanconf", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get WLAN: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("WLAN not found: %s", id)
		}
		return nil, fmt.Errorf("get WLAN failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.WLAN](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("WLAN not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// Create creates a new WLAN.
func (s *wlanService) Create(ctx context.Context, site string, wlan *types.WLAN) (*types.WLAN, error) {
	path := internal.BuildRESTPath(site, "wlanconf", "")
	req := transport.NewRequest("POST", path).WithBody(wlan)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create WLAN: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create WLAN failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.WLAN](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create WLAN returned empty response")
	}

	return &apiResp.Data[0], nil
}

// Update updates an existing WLAN.
func (s *wlanService) Update(ctx context.Context, site string, wlan *types.WLAN) (*types.WLAN, error) {
	if wlan.ID == "" {
		return nil, fmt.Errorf("WLAN ID is required for update")
	}

	path := internal.BuildRESTPath(site, "wlanconf", wlan.ID)
	req := transport.NewRequest("PUT", path).WithBody(wlan)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update WLAN: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update WLAN failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.WLAN](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update WLAN returned empty response")
	}

	return &apiResp.Data[0], nil
}

// Delete deletes a WLAN.
func (s *wlanService) Delete(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "wlanconf", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete WLAN: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete WLAN failed with status %d", resp.StatusCode)
	}

	return nil
}

// Enable enables a WLAN.
func (s *wlanService) Enable(ctx context.Context, site, id string) error {
	wlan, err := s.Get(ctx, site, id)
	if err != nil {
		return err
	}

	wlan.Enabled = true
	_, err = s.Update(ctx, site, wlan)
	return err
}

// Disable disables a WLAN.
func (s *wlanService) Disable(ctx context.Context, site, id string) error {
	wlan, err := s.Get(ctx, site, id)
	if err != nil {
		return err
	}

	wlan.Enabled = false
	_, err = s.Update(ctx, site, wlan)
	return err
}

// SetMACFilter sets MAC filtering on a WLAN.
func (s *wlanService) SetMACFilter(ctx context.Context, site, id, policy string, macs []string) error {
	wlan, err := s.Get(ctx, site, id)
	if err != nil {
		return err
	}

	wlan.MACFilterEnabled = true
	wlan.MACFilterPolicy = policy
	wlan.MACFilterList = macs

	_, err = s.Update(ctx, site, wlan)
	return err
}

// ListGroups returns all WLAN groups for a site.
func (s *wlanService) ListGroups(ctx context.Context, site string) ([]types.WLANGroup, error) {
	path := internal.BuildRESTPath(site, "wlangroup", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list WLAN groups: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list WLAN groups failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.WLANGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// GetGroup returns a specific WLAN group by ID.
func (s *wlanService) GetGroup(ctx context.Context, site, id string) (*types.WLANGroup, error) {
	path := internal.BuildRESTPath(site, "wlangroup", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get WLAN group: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("WLAN group not found: %s", id)
		}
		return nil, fmt.Errorf("get WLAN group failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.WLANGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("WLAN group not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// CreateGroup creates a new WLAN group.
func (s *wlanService) CreateGroup(ctx context.Context, site string, group *types.WLANGroup) (*types.WLANGroup, error) {
	path := internal.BuildRESTPath(site, "wlangroup", "")
	req := transport.NewRequest("POST", path).WithBody(group)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create WLAN group: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create WLAN group failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.WLANGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create WLAN group returned empty response")
	}

	return &apiResp.Data[0], nil
}

// UpdateGroup updates an existing WLAN group.
func (s *wlanService) UpdateGroup(ctx context.Context, site string, group *types.WLANGroup) (*types.WLANGroup, error) {
	if group.ID == "" {
		return nil, fmt.Errorf("WLAN group ID is required for update")
	}

	path := internal.BuildRESTPath(site, "wlangroup", group.ID)
	req := transport.NewRequest("PUT", path).WithBody(group)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update WLAN group: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update WLAN group failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.WLANGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update WLAN group returned empty response")
	}

	return &apiResp.Data[0], nil
}

// DeleteGroup deletes a WLAN group.
func (s *wlanService) DeleteGroup(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "wlangroup", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete WLAN group: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete WLAN group failed with status %d", resp.StatusCode)
	}

	return nil
}
