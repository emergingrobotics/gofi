package services

import (
	"context"
	"fmt"

	"github.com/unifi-go/gofi/internal"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// systemService implements SystemService.
type systemService struct {
	transport transport.Transport
}

// NewSystemService creates a new system service.
func NewSystemService(transport transport.Transport) SystemService {
	return &systemService{
		transport: transport,
	}
}

// Status returns the system status (non-authenticated endpoint).
func (s *systemService) Status(ctx context.Context) (*types.Status, error) {
	path := "/api/status"
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("get status failed with status %d", resp.StatusCode)
	}

	var status types.Status
	if err := resp.Parse(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

// Self returns the current user information.
func (s *systemService) Self(ctx context.Context) (*types.AdminUser, error) {
	path := "/api/self"
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get self: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("get self failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.AdminUser](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("no user data returned")
	}

	return &apiResp.Data[0], nil
}

// Reboot reboots the controller.
func (s *systemService) Reboot(ctx context.Context) error {
	path := "/proxy/network/api/cmd/system"
	req := transport.NewRequest("POST", path).WithBody(map[string]string{
		"cmd": "reboot",
	})

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to reboot: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("reboot failed with status %d", resp.StatusCode)
	}

	return nil
}

// SpeedTest initiates a speed test.
func (s *systemService) SpeedTest(ctx context.Context, site string) error {
	path := fmt.Sprintf("/proxy/network/api/s/%s/cmd/speedtest", site)
	req := transport.NewRequest("POST", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start speed test: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("speed test failed with status %d", resp.StatusCode)
	}

	return nil
}

// SpeedTestStatus returns the speed test status.
func (s *systemService) SpeedTestStatus(ctx context.Context, site string) (*types.SpeedTestStatus, error) {
	path := fmt.Sprintf("/proxy/network/api/s/%s/stat/speedtest", site)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get speed test status: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("get speed test status failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.SpeedTestStatus](resp.Body)
	if err != nil {
		return nil, err
	}

	if len(apiResp.Data) == 0 {
		// No speed test run yet
		return nil, nil
	}

	return &apiResp.Data[0], nil
}

// ListBackups returns all backup files.
func (s *systemService) ListBackups(ctx context.Context) ([]types.Backup, error) {
	path := "/proxy/network/api/cmd/backup"
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list backups failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.Backup](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// CreateBackup creates a new backup.
func (s *systemService) CreateBackup(ctx context.Context) error {
	path := "/proxy/network/api/cmd/backup"
	req := transport.NewRequest("POST", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("create backup failed with status %d", resp.StatusCode)
	}

	return nil
}

// DeleteBackup deletes a backup file.
func (s *systemService) DeleteBackup(ctx context.Context, filename string) error {
	path := fmt.Sprintf("/proxy/network/api/cmd/backup/%s", filename)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete backup failed with status %d", resp.StatusCode)
	}

	return nil
}

// ListAdmins returns all admin users.
func (s *systemService) ListAdmins(ctx context.Context) ([]types.AdminUser, error) {
	path := "/proxy/network/api/stat/admin"
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list admins: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list admins failed with status %d", resp.StatusCode)
	}

	apiResp, err := internal.ParseAPIResponse[types.AdminUser](resp.Body)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}
