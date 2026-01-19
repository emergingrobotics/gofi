package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// siteService implements SiteService.
type siteService struct {
	transport transport.Transport
}

// NewSiteService creates a new site service.
func NewSiteService(transport transport.Transport) SiteService {
	return &siteService{
		transport: transport,
	}
}

// List returns all sites.
func (s *siteService) List(ctx context.Context) ([]types.Site, error) {
	req := transport.NewRequest("GET", "/api/self/sites")

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list sites: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list sites failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Site](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// Get returns a specific site.
func (s *siteService) Get(ctx context.Context, id string) (*types.Site, error) {
	// List all sites and find the one we want
	sites, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, site := range sites {
		if site.ID == id || site.Name == id {
			return &site, nil
		}
	}

	return nil, fmt.Errorf("site not found: %s", id)
}

// Create creates a new site.
func (s *siteService) Create(ctx context.Context, desc, name string) (*types.Site, error) {
	createReq := types.CreateSiteRequest{
		Desc: desc,
		Name: name,
	}

	req := transport.NewRequest("POST", "/api/self/sites").
		WithBody(createReq)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create site: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create site failed with status %d", resp.StatusCode)
	}

	site, err := internal.ParseSingleResult[types.Site](resp.Body)
	if err != nil {
		return nil, err
	}

	return site, nil
}

// Update updates a site.
func (s *siteService) Update(ctx context.Context, site *types.Site) (*types.Site, error) {
	updateReq := types.UpdateSiteRequest{
		Desc: site.Desc,
		Name: site.Name,
	}

	path := fmt.Sprintf("/api/self/sites/%s", site.ID)
	req := transport.NewRequest("PUT", path).
		WithBody(updateReq)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update site: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update site failed with status %d", resp.StatusCode)
	}

	result, err := internal.ParseSingleResult[types.Site](resp.Body)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete deletes a site.
func (s *siteService) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/self/sites/%s", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete site: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete site failed with status %d", resp.StatusCode)
	}

	return nil
}

// Health returns health information for a site.
func (s *siteService) Health(ctx context.Context, site string) ([]types.HealthData, error) {
	path := internal.BuildAPIPath(site, "stat/health")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get health: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("get health failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.HealthData](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// SysInfo returns system information.
func (s *siteService) SysInfo(ctx context.Context, site string) (*types.SysInfo, error) {
	path := internal.BuildAPIPath(site, "stat/sysinfo")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get sysinfo: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("get sysinfo failed with status %d", resp.StatusCode)
	}

	sysInfo, err := internal.ParseSingleResult[types.SysInfo](resp.Body)
	if err != nil {
		return nil, err
	}

	return sysInfo, nil
}
