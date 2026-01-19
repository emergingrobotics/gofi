package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// networkService implements NetworkService.
type networkService struct {
	transport transport.Transport
}

// NewNetworkService creates a new network service.
func NewNetworkService(transport transport.Transport) NetworkService {
	return &networkService{
		transport: transport,
	}
}

// List returns all networks for a site.
func (s *networkService) List(ctx context.Context, site string) ([]types.Network, error) {
	path := internal.BuildRESTPath(site, "networkconf", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list networks failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Network](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// Get returns a specific network by ID.
func (s *networkService) Get(ctx context.Context, site, id string) (*types.Network, error) {
	path := internal.BuildRESTPath(site, "networkconf", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("get network failed with status %d", resp.StatusCode)
	}

	network, err := internal.ParseSingleResult[types.Network](resp.Body)
	if err != nil {
		return nil, err
	}

	return network, nil
}

// Create creates a new network.
func (s *networkService) Create(ctx context.Context, site string, network *types.Network) (*types.Network, error) {
	path := internal.BuildRESTPath(site, "networkconf", "")
	req := transport.NewRequest("POST", path).WithBody(network)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create network failed with status %d", resp.StatusCode)
	}

	created, err := internal.ParseSingleResult[types.Network](resp.Body)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update updates a network.
func (s *networkService) Update(ctx context.Context, site string, network *types.Network) (*types.Network, error) {
	path := internal.BuildRESTPath(site, "networkconf", network.ID)
	req := transport.NewRequest("PUT", path).WithBody(network)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update network: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update network failed with status %d", resp.StatusCode)
	}

	updated, err := internal.ParseSingleResult[types.Network](resp.Body)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete deletes a network.
func (s *networkService) Delete(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "networkconf", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete network: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete network failed with status %d", resp.StatusCode)
	}

	return nil
}
