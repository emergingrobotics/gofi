package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// settingService implements SettingService.
type settingService struct {
	transport transport.Transport
}

// NewSettingService creates a new setting service.
func NewSettingService(transport transport.Transport) SettingService {
	return &settingService{
		transport: transport,
	}
}

// Get returns a setting by key.
func (s *settingService) Get(ctx context.Context, site, key string) (interface{}, error) {
	path := fmt.Sprintf("/proxy/network/api/s/%s/rest/setting/%s", site, key)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("setting not found: %s", key)
		}
		return nil, fmt.Errorf("get setting failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Setting](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("setting not found: %s", key)
	}

	return &apiResp.Data[0], nil
}

// Update updates a setting.
func (s *settingService) Update(ctx context.Context, site string, setting interface{}) error {
	// Extract key from setting (must be a types.Setting or compatible struct)
	var key string
	if set, ok := setting.(*types.Setting); ok {
		key = set.Key
	} else {
		// Try to get key via reflection or type assertion for typed settings
		return fmt.Errorf("invalid setting type")
	}

	path := fmt.Sprintf("/proxy/network/api/s/%s/rest/setting/%s", site, key)
	req := transport.NewRequest("PUT", path).WithBody(setting)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("update setting failed with status %d", resp.StatusCode)
	}

	return nil
}

// ListRadiusProfiles returns all RADIUS profiles.
func (s *settingService) ListRadiusProfiles(ctx context.Context, site string) ([]types.RADIUSProfile, error) {
	path := internal.BuildRESTPath(site, "radiusprofile", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list RADIUS profiles: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list RADIUS profiles failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.RADIUSProfile](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// GetRadiusProfile returns a RADIUS profile by ID.
func (s *settingService) GetRadiusProfile(ctx context.Context, site, id string) (*types.RADIUSProfile, error) {
	path := internal.BuildRESTPath(site, "radiusprofile", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get RADIUS profile: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("RADIUS profile not found: %s", id)
		}
		return nil, fmt.Errorf("get RADIUS profile failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.RADIUSProfile](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("RADIUS profile not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// CreateRadiusProfile creates a new RADIUS profile.
func (s *settingService) CreateRadiusProfile(ctx context.Context, site string, profile *types.RADIUSProfile) (*types.RADIUSProfile, error) {
	path := internal.BuildRESTPath(site, "radiusprofile", "")
	req := transport.NewRequest("POST", path).WithBody(profile)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create RADIUS profile: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create RADIUS profile failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.RADIUSProfile](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create RADIUS profile returned no data")
	}

	return &apiResp.Data[0], nil
}

// UpdateRadiusProfile updates an existing RADIUS profile.
func (s *settingService) UpdateRadiusProfile(ctx context.Context, site string, profile *types.RADIUSProfile) (*types.RADIUSProfile, error) {
	if profile.ID == "" {
		return nil, fmt.Errorf("RADIUS profile ID is required for update")
	}

	path := internal.BuildRESTPath(site, "radiusprofile", profile.ID)
	req := transport.NewRequest("PUT", path).WithBody(profile)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update RADIUS profile: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("RADIUS profile not found: %s", profile.ID)
		}
		return nil, fmt.Errorf("update RADIUS profile failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.RADIUSProfile](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update RADIUS profile returned no data")
	}

	return &apiResp.Data[0], nil
}

// DeleteRadiusProfile deletes a RADIUS profile.
func (s *settingService) DeleteRadiusProfile(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "radiusprofile", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete RADIUS profile: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return fmt.Errorf("RADIUS profile not found: %s", id)
		}
		return fmt.Errorf("delete RADIUS profile failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetDynamicDNS returns the Dynamic DNS configuration.
func (s *settingService) GetDynamicDNS(ctx context.Context, site string) (*types.DynamicDNS, error) {
	path := internal.BuildRESTPath(site, "dynamicdns", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get Dynamic DNS: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("get Dynamic DNS failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.DynamicDNS](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		// Not configured
		return nil, nil
	}

	return &apiResp.Data[0], nil
}

// UpdateDynamicDNS updates the Dynamic DNS configuration.
func (s *settingService) UpdateDynamicDNS(ctx context.Context, site string, ddns *types.DynamicDNS) error {
	path := internal.BuildRESTPath(site, "dynamicdns", "")
	req := transport.NewRequest("PUT", path).WithBody(ddns)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update Dynamic DNS: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("update Dynamic DNS failed with status %d", resp.StatusCode)
	}

	return nil
}
