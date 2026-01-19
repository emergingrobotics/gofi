package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// portForwardService implements PortForwardService.
type portForwardService struct {
	transport transport.Transport
}

// NewPortForwardService creates a new port forward service.
func NewPortForwardService(transport transport.Transport) PortForwardService {
	return &portForwardService{
		transport: transport,
	}
}

// List returns all port forwards.
func (s *portForwardService) List(ctx context.Context, site string) ([]types.PortForward, error) {
	path := internal.BuildRESTPath(site, "portforward", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list port forwards: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list port forwards failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.PortForward](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// Get returns a port forward by ID.
func (s *portForwardService) Get(ctx context.Context, site, id string) (*types.PortForward, error) {
	path := internal.BuildRESTPath(site, "portforward", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get port forward: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("port forward not found: %s", id)
		}
		return nil, fmt.Errorf("get port forward failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.PortForward](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("port forward not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// Create creates a new port forward.
func (s *portForwardService) Create(ctx context.Context, site string, forward *types.PortForward) (*types.PortForward, error) {
	path := internal.BuildRESTPath(site, "portforward", "")
	req := transport.NewRequest("POST", path).WithBody(forward)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create port forward: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create port forward failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.PortForward](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create port forward returned no data")
	}

	return &apiResp.Data[0], nil
}

// Update updates an existing port forward.
func (s *portForwardService) Update(ctx context.Context, site string, forward *types.PortForward) (*types.PortForward, error) {
	if forward.ID == "" {
		return nil, fmt.Errorf("port forward ID is required for update")
	}

	path := internal.BuildRESTPath(site, "portforward", forward.ID)
	req := transport.NewRequest("PUT", path).WithBody(forward)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update port forward: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("port forward not found: %s", forward.ID)
		}
		return nil, fmt.Errorf("update port forward failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.PortForward](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update port forward returned no data")
	}

	return &apiResp.Data[0], nil
}

// Delete deletes a port forward.
func (s *portForwardService) Delete(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "portforward", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete port forward: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return fmt.Errorf("port forward not found: %s", id)
		}
		return fmt.Errorf("delete port forward failed with status %d", resp.StatusCode)
	}

	return nil
}

// Enable enables a port forward.
func (s *portForwardService) Enable(ctx context.Context, site, id string) error {
	forward, err := s.Get(ctx, site, id)
	if err != nil {
		return err
	}

	forward.Enabled = true
	_, err = s.Update(ctx, site, forward)
	return err
}

// Disable disables a port forward.
func (s *portForwardService) Disable(ctx context.Context, site, id string) error {
	forward, err := s.Get(ctx, site, id)
	if err != nil {
		return err
	}

	forward.Enabled = false
	_, err = s.Update(ctx, site, forward)
	return err
}
