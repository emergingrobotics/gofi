package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

// dnsService implements DNSService.
type dnsService struct {
	transport transport.Transport
}

// NewDNSService creates a new DNS service.
func NewDNSService(transport transport.Transport) DNSService {
	return &dnsService{
		transport: transport,
	}
}

// buildDNSPath builds the v2 API path for DNS records.
func buildDNSPath(site, id string) string {
	if id != "" {
		return fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns/%s", site, id)
	}
	return fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns", site)
}

// List returns all local DNS records.
func (s *dnsService) List(ctx context.Context, site string) ([]types.DNSRecord, error) {
	path := buildDNSPath(site, "")
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list DNS records failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// v2 API returns array directly, not wrapped in data field
	var records []types.DNSRecord
	if err := json.Unmarshal(resp.Body, &records); err != nil {
		// Try parsing as wrapped response
		var wrapped struct {
			Data []types.DNSRecord `json:"data"`
		}
		if err2 := json.Unmarshal(resp.Body, &wrapped); err2 != nil {
			return nil, fmt.Errorf("failed to parse DNS records response: %w", err)
		}
		records = wrapped.Data
	}

	return records, nil
}

// Get returns a DNS record by ID.
func (s *dnsService) Get(ctx context.Context, site, id string) (*types.DNSRecord, error) {
	path := buildDNSPath(site, id)
	req := transport.NewRequest("GET", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record: %w", err)
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("DNS record not found: %s", id)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("get DNS record failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var record types.DNSRecord
	if err := json.Unmarshal(resp.Body, &record); err != nil {
		return nil, fmt.Errorf("failed to parse DNS record: %w", err)
	}

	return &record, nil
}

// GetByName returns a DNS record by hostname/key.
func (s *dnsService) GetByName(ctx context.Context, site, name string) (*types.DNSRecord, error) {
	records, err := s.List(ctx, site)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.Key == name {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("DNS record not found for name: %s", name)
}

// GetByIP returns DNS records pointing to a specific IP.
func (s *dnsService) GetByIP(ctx context.Context, site, ip string) ([]types.DNSRecord, error) {
	records, err := s.List(ctx, site)
	if err != nil {
		return nil, err
	}

	var matches []types.DNSRecord
	for _, r := range records {
		if r.Value == ip {
			matches = append(matches, r)
		}
	}

	return matches, nil
}

// Create creates a new DNS record.
func (s *dnsService) Create(ctx context.Context, site string, record *types.DNSRecord) (*types.DNSRecord, error) {
	path := buildDNSPath(site, "")
	req := transport.NewRequest("POST", path).WithBody(record)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("create DNS record failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var created types.DNSRecord
	if err := json.Unmarshal(resp.Body, &created); err != nil {
		// Return the input record with success indication
		return record, nil
	}

	return &created, nil
}

// Update updates an existing DNS record.
func (s *dnsService) Update(ctx context.Context, site string, record *types.DNSRecord) (*types.DNSRecord, error) {
	if record.ID == "" {
		return nil, fmt.Errorf("DNS record ID is required for update")
	}

	path := buildDNSPath(site, record.ID)
	req := transport.NewRequest("PUT", path).WithBody(record)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update DNS record: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("update DNS record failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var updated types.DNSRecord
	if err := json.Unmarshal(resp.Body, &updated); err != nil {
		return record, nil
	}

	return &updated, nil
}

// Delete deletes a DNS record by ID.
func (s *dnsService) Delete(ctx context.Context, site, id string) error {
	path := buildDNSPath(site, id)
	req := transport.NewRequest("DELETE", path)

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("delete DNS record failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	return nil
}

// DeleteByName deletes a DNS record by hostname/key.
func (s *dnsService) DeleteByName(ctx context.Context, site, name string) error {
	record, err := s.GetByName(ctx, site, name)
	if err != nil {
		return err
	}

	return s.Delete(ctx, site, record.ID)
}
