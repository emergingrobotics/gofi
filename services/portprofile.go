package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// portProfileService implements PortProfileService.
type portProfileService struct {
	transport transport.Transport
}

// NewPortProfileService creates a new port profile service.
func NewPortProfileService(transport transport.Transport) PortProfileService {
	return &portProfileService{
		transport: transport,
	}
}

// List returns all port profiles.
func (s *portProfileService) List(ctx context.Context, site string) ([]types.PortProfile, error) {
	path := internal.BuildRESTPath(site, "portconf", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list port profiles: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list port profiles failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.PortProfile](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// Get returns a port profile by ID.
func (s *portProfileService) Get(ctx context.Context, site, id string) (*types.PortProfile, error) {
	path := internal.BuildRESTPath(site, "portconf", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get port profile: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("port profile not found: %s", id)
		}
		return nil, fmt.Errorf("get port profile failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.PortProfile](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("port profile not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// Create creates a new port profile.
func (s *portProfileService) Create(ctx context.Context, site string, profile *types.PortProfile) (*types.PortProfile, error) {
	path := internal.BuildRESTPath(site, "portconf", "")
	req := transport.NewRequest("POST", path).WithBody(profile)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create port profile: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create port profile failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.PortProfile](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create port profile returned no data")
	}

	return &apiResp.Data[0], nil
}

// Update updates an existing port profile.
func (s *portProfileService) Update(ctx context.Context, site string, profile *types.PortProfile) (*types.PortProfile, error) {
	if profile.ID == "" {
		return nil, fmt.Errorf("port profile ID is required for update")
	}

	path := internal.BuildRESTPath(site, "portconf", profile.ID)
	req := transport.NewRequest("PUT", path).WithBody(profile)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update port profile: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("port profile not found: %s", profile.ID)
		}
		return nil, fmt.Errorf("update port profile failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.PortProfile](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update port profile returned no data")
	}

	return &apiResp.Data[0], nil
}

// Delete deletes a port profile.
func (s *portProfileService) Delete(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "portconf", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete port profile: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return fmt.Errorf("port profile not found: %s", id)
		}
		return fmt.Errorf("delete port profile failed with status %d", resp.StatusCode)
	}

	return nil
}
