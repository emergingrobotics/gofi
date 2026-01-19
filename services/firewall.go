package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// firewallService implements FirewallService.
type firewallService struct {
	transport transport.Transport
}

// NewFirewallService creates a new firewall service.
func NewFirewallService(transport transport.Transport) FirewallService {
	return &firewallService{
		transport: transport,
	}
}

// ListRules returns all firewall rules for a site.
func (s *firewallService) ListRules(ctx context.Context, site string) ([]types.FirewallRule, error) {
	path := internal.BuildRESTPath(site, "firewallrule", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list firewall rules: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list firewall rules failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.FirewallRule](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// GetRule returns a specific firewall rule by ID.
func (s *firewallService) GetRule(ctx context.Context, site, id string) (*types.FirewallRule, error) {
	path := internal.BuildRESTPath(site, "firewallrule", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get firewall rule: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("firewall rule not found: %s", id)
		}
		return nil, fmt.Errorf("get firewall rule failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.FirewallRule](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("firewall rule not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// CreateRule creates a new firewall rule.
func (s *firewallService) CreateRule(ctx context.Context, site string, rule *types.FirewallRule) (*types.FirewallRule, error) {
	path := internal.BuildRESTPath(site, "firewallrule", "")
	req := transport.NewRequest("POST", path).WithBody(rule)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall rule: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create firewall rule failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.FirewallRule](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create firewall rule returned empty response")
	}

	return &apiResp.Data[0], nil
}

// UpdateRule updates an existing firewall rule.
func (s *firewallService) UpdateRule(ctx context.Context, site string, rule *types.FirewallRule) (*types.FirewallRule, error) {
	if rule.ID == "" {
		return nil, fmt.Errorf("firewall rule ID is required for update")
	}

	path := internal.BuildRESTPath(site, "firewallrule", rule.ID)
	req := transport.NewRequest("PUT", path).WithBody(rule)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update firewall rule: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update firewall rule failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.FirewallRule](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update firewall rule returned empty response")
	}

	return &apiResp.Data[0], nil
}

// DeleteRule deletes a firewall rule.
func (s *firewallService) DeleteRule(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "firewallrule", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete firewall rule: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete firewall rule failed with status %d", resp.StatusCode)
	}

	return nil
}

// EnableRule enables a firewall rule.
func (s *firewallService) EnableRule(ctx context.Context, site, id string) error {
	rule, err := s.GetRule(ctx, site, id)
	if err != nil {
		return err
	}

	rule.Enabled = true
	_, err = s.UpdateRule(ctx, site, rule)
	return err
}

// DisableRule disables a firewall rule.
func (s *firewallService) DisableRule(ctx context.Context, site, id string) error {
	rule, err := s.GetRule(ctx, site, id)
	if err != nil {
		return err
	}

	rule.Enabled = false
	_, err = s.UpdateRule(ctx, site, rule)
	return err
}

// ReorderRules reorders firewall rules within a ruleset.
func (s *firewallService) ReorderRules(ctx context.Context, site, ruleset string, updates []types.FirewallRuleIndexUpdate) error {
	path := internal.BuildRESTPath(site, "firewallrule", "reorder")
	req := transport.NewRequest("POST", path).WithBody(updates)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to reorder firewall rules: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("reorder firewall rules failed with status %d", resp.StatusCode)
	}

	return nil
}

// ListGroups returns all firewall groups for a site.
func (s *firewallService) ListGroups(ctx context.Context, site string) ([]types.FirewallGroup, error) {
	path := internal.BuildRESTPath(site, "firewallgroup", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list firewall groups: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list firewall groups failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.FirewallGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// GetGroup returns a specific firewall group by ID.
func (s *firewallService) GetGroup(ctx context.Context, site, id string) (*types.FirewallGroup, error) {
	path := internal.BuildRESTPath(site, "firewallgroup", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get firewall group: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("firewall group not found: %s", id)
		}
		return nil, fmt.Errorf("get firewall group failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.FirewallGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("firewall group not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// CreateGroup creates a new firewall group.
func (s *firewallService) CreateGroup(ctx context.Context, site string, group *types.FirewallGroup) (*types.FirewallGroup, error) {
	path := internal.BuildRESTPath(site, "firewallgroup", "")
	req := transport.NewRequest("POST", path).WithBody(group)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall group: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create firewall group failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.FirewallGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create firewall group returned empty response")
	}

	return &apiResp.Data[0], nil
}

// UpdateGroup updates an existing firewall group.
func (s *firewallService) UpdateGroup(ctx context.Context, site string, group *types.FirewallGroup) (*types.FirewallGroup, error) {
	if group.ID == "" {
		return nil, fmt.Errorf("firewall group ID is required for update")
	}

	path := internal.BuildRESTPath(site, "firewallgroup", group.ID)
	req := transport.NewRequest("PUT", path).WithBody(group)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update firewall group: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update firewall group failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.FirewallGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update firewall group returned empty response")
	}

	return &apiResp.Data[0], nil
}

// DeleteGroup deletes a firewall group.
func (s *firewallService) DeleteGroup(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "firewallgroup", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete firewall group: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete firewall group failed with status %d", resp.StatusCode)
	}

	return nil
}

// ListTrafficRules returns all traffic rules for a site.
func (s *firewallService) ListTrafficRules(ctx context.Context, site string) ([]types.TrafficRule, error) {
	path := internal.BuildV2APIPath(site, fmt.Sprintf("site/%s/trafficrule", site))
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list traffic rules: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list traffic rules failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.TrafficRule](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// GetTrafficRule returns a specific traffic rule by ID.
func (s *firewallService) GetTrafficRule(ctx context.Context, site, id string) (*types.TrafficRule, error) {
	path := internal.BuildV2APIPath(site, fmt.Sprintf("site/%s/trafficrule/%s", site, id))
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic rule: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("traffic rule not found: %s", id)
		}
		return nil, fmt.Errorf("get traffic rule failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.TrafficRule](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("traffic rule not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// CreateTrafficRule creates a new traffic rule.
func (s *firewallService) CreateTrafficRule(ctx context.Context, site string, rule *types.TrafficRule) (*types.TrafficRule, error) {
	path := internal.BuildV2APIPath(site, fmt.Sprintf("site/%s/trafficrule", site))
	req := transport.NewRequest("POST", path).WithBody(rule)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create traffic rule: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create traffic rule failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.TrafficRule](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create traffic rule returned empty response")
	}

	return &apiResp.Data[0], nil
}

// UpdateTrafficRule updates an existing traffic rule.
func (s *firewallService) UpdateTrafficRule(ctx context.Context, site string, rule *types.TrafficRule) (*types.TrafficRule, error) {
	if rule.ID == "" {
		return nil, fmt.Errorf("traffic rule ID is required for update")
	}

	path := internal.BuildV2APIPath(site, fmt.Sprintf("site/%s/trafficrule/%s", site, rule.ID))
	req := transport.NewRequest("PUT", path).WithBody(rule)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update traffic rule: %w", err)
	}

	// Note: v2 API returns 201 for PUT operations
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("update traffic rule failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.TrafficRule](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update traffic rule returned empty response")
	}

	return &apiResp.Data[0], nil
}

// DeleteTrafficRule deletes a traffic rule.
func (s *firewallService) DeleteTrafficRule(ctx context.Context, site, id string) error {
	path := internal.BuildV2APIPath(site, fmt.Sprintf("site/%s/trafficrule/%s", site, id))
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete traffic rule: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete traffic rule failed with status %d", resp.StatusCode)
	}

	return nil
}
