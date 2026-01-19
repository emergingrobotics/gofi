package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// routingService implements RoutingService.
type routingService struct {
	transport transport.Transport
}

// NewRoutingService creates a new routing service.
func NewRoutingService(transport transport.Transport) RoutingService {
	return &routingService{
		transport: transport,
	}
}

// List returns all routes.
func (s *routingService) List(ctx context.Context, site string) ([]types.Route, error) {
	path := internal.BuildRESTPath(site, "routing", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list routes failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Route](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// Get returns a route by ID.
func (s *routingService) Get(ctx context.Context, site, id string) (*types.Route, error) {
	path := internal.BuildRESTPath(site, "routing", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("route not found: %s", id)
		}
		return nil, fmt.Errorf("get route failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Route](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("route not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// Create creates a new route.
func (s *routingService) Create(ctx context.Context, site string, route *types.Route) (*types.Route, error) {
	path := internal.BuildRESTPath(site, "routing", "")
	req := transport.NewRequest("POST", path).WithBody(route)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create route failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Route](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create route returned no data")
	}

	return &apiResp.Data[0], nil
}

// Update updates an existing route.
func (s *routingService) Update(ctx context.Context, site string, route *types.Route) (*types.Route, error) {
	if route.ID == "" {
		return nil, fmt.Errorf("route ID is required for update")
	}

	path := internal.BuildRESTPath(site, "routing", route.ID)
	req := transport.NewRequest("PUT", path).WithBody(route)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update route: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("route not found: %s", route.ID)
		}
		return nil, fmt.Errorf("update route failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Route](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update route returned no data")
	}

	return &apiResp.Data[0], nil
}

// Delete deletes a route.
func (s *routingService) Delete(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "routing", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return fmt.Errorf("route not found: %s", id)
		}
		return fmt.Errorf("delete route failed with status %d", resp.StatusCode)
	}

	return nil
}

// Enable enables a route.
func (s *routingService) Enable(ctx context.Context, site, id string) error {
	route, err := s.Get(ctx, site, id)
	if err != nil {
		return err
	}

	route.Enabled = true
	_, err = s.Update(ctx, site, route)
	return err
}

// Disable disables a route.
func (s *routingService) Disable(ctx context.Context, site, id string) error {
	route, err := s.Get(ctx, site, id)
	if err != nil {
		return err
	}

	route.Enabled = false
	_, err = s.Update(ctx, site, route)
	return err
}
