package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// userService implements UserService.
type userService struct {
	transport transport.Transport
}

// NewUserService creates a new user service.
func NewUserService(transport transport.Transport) UserService {
	return &userService{
		transport: transport,
	}
}

// List returns all known clients/users.
func (s *userService) List(ctx context.Context, site string) ([]types.User, error) {
	path := internal.BuildRESTPath(site, "user", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list users failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.User](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// Get returns a user by ID.
func (s *userService) Get(ctx context.Context, site, id string) (*types.User, error) {
	path := internal.BuildRESTPath(site, "user", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("get user failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.User](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// GetByMAC returns a user by MAC address.
func (s *userService) GetByMAC(ctx context.Context, site, mac string) (*types.User, error) {
	// List all users and find by MAC
	users, err := s.List(ctx, site)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.MAC == mac {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found with MAC: %s", mac)
}

// Create creates a new user entry.
func (s *userService) Create(ctx context.Context, site string, user *types.User) (*types.User, error) {
	path := internal.BuildRESTPath(site, "user", "")
	req := transport.NewRequest("POST", path).WithBody(user)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create user failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.User](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create user returned empty response")
	}

	return &apiResp.Data[0], nil
}

// Update updates an existing user.
func (s *userService) Update(ctx context.Context, site string, user *types.User) (*types.User, error) {
	if user.ID == "" {
		return nil, fmt.Errorf("user ID is required for update")
	}

	path := internal.BuildRESTPath(site, "user", user.ID)
	req := transport.NewRequest("PUT", path).WithBody(user)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update user failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.User](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update user returned empty response")
	}

	return &apiResp.Data[0], nil
}

// Delete deletes a user by ID.
func (s *userService) Delete(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "user", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete user failed with status %d", resp.StatusCode)
	}

	return nil
}

// DeleteByMAC deletes a user by MAC address.
func (s *userService) DeleteByMAC(ctx context.Context, site, mac string) error {
	// Find user by MAC first
	user, err := s.GetByMAC(ctx, site, mac)
	if err != nil {
		return err
	}

	return s.Delete(ctx, site, user.ID)
}

// SetFixedIP assigns a fixed IP to a user.
func (s *userService) SetFixedIP(ctx context.Context, site, mac, ip, networkID string) error {
	// Get user by MAC
	user, err := s.GetByMAC(ctx, site, mac)
	if err != nil {
		return err
	}

	// Update with fixed IP settings
	user.UseFixedIP = true
	user.FixedIP = ip
	user.NetworkID = networkID

	_, err = s.Update(ctx, site, user)
	return err
}

// ClearFixedIP removes a fixed IP assignment.
func (s *userService) ClearFixedIP(ctx context.Context, site, mac string) error {
	// Get user by MAC
	user, err := s.GetByMAC(ctx, site, mac)
	if err != nil {
		return err
	}

	// Build payload with required fields plus explicit use_fixedip:false
	// The API requires the MAC and typically other fields to be present
	payload := map[string]interface{}{
		"mac":         user.MAC,
		"use_fixedip": false,
	}

	// Include name if present (some API versions require it)
	if user.Name != "" {
		payload["name"] = user.Name
	}

	path := internal.BuildRESTPath(site, "user", user.ID)
	req := transport.NewRequest("PUT", path).WithBody(payload)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to clear fixed IP: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("clear fixed IP failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	return nil
}

// ListGroups returns all user groups.
func (s *userService) ListGroups(ctx context.Context, site string) ([]types.UserGroup, error) {
	path := internal.BuildRESTPath(site, "usergroup", "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list user groups: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list user groups failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.UserGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// GetGroup returns a user group by ID.
func (s *userService) GetGroup(ctx context.Context, site, id string) (*types.UserGroup, error) {
	path := internal.BuildRESTPath(site, "usergroup", id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user group: %w", err)
	}

	if !resp.IsSuccess() {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("user group not found: %s", id)
		}
		return nil, fmt.Errorf("get user group failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.UserGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("user group not found: %s", id)
	}

	return &apiResp.Data[0], nil
}

// CreateGroup creates a new user group.
func (s *userService) CreateGroup(ctx context.Context, site string, group *types.UserGroup) (*types.UserGroup, error) {
	path := internal.BuildRESTPath(site, "usergroup", "")
	req := transport.NewRequest("POST", path).WithBody(group)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user group: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create user group failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.UserGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("create user group returned empty response")
	}

	return &apiResp.Data[0], nil
}

// UpdateGroup updates an existing user group.
func (s *userService) UpdateGroup(ctx context.Context, site string, group *types.UserGroup) (*types.UserGroup, error) {
	if group.ID == "" {
		return nil, fmt.Errorf("user group ID is required for update")
	}

	path := internal.BuildRESTPath(site, "usergroup", group.ID)
	req := transport.NewRequest("PUT", path).WithBody(group)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user group: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update user group failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.UserGroup](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("update user group returned empty response")
	}

	return &apiResp.Data[0], nil
}

// DeleteGroup deletes a user group.
func (s *userService) DeleteGroup(ctx context.Context, site, id string) error {
	path := internal.BuildRESTPath(site, "usergroup", id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete user group: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete user group failed with status %d", resp.StatusCode)
	}

	return nil
}
