package mock

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/unifi-go/gofi/types"
)

var testRoutingHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func TestHandleListRoutes(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test routes
	server.state.AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route 1",
		Enabled:             true,
		StaticRouteNetwork:  "10.0.0.0/24",
		StaticRouteNexthop:  "192.168.1.1",
		Type:                types.RouteTypeNexthop,
	})
	server.state.AddRoute(&types.Route{
		ID:                  "route2",
		Name:                "Test Route 2",
		Enabled:             false,
		StaticRouteNetwork:  "10.1.0.0/24",
		Type:                types.RouteTypeBlackhole,
	})

	// Test list routes
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/rest/routing", nil)
	resp, err := testRoutingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to list routes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.Route `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 2 {
		t.Fatalf("Expected 2 routes, got %d", len(apiResp.Data))
	}
}

func TestHandleGetRoute(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test route
	server.state.AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route",
		Enabled:             true,
		StaticRouteNetwork:  "10.0.0.0/24",
		StaticRouteNexthop:  "192.168.1.1",
	})

	// Test get route
	req, _ := http.NewRequest("GET", server.URL()+"/api/s/default/rest/routing/route1", nil)
	resp, err := testRoutingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get route: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.Route `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].Name != "Test Route" {
		t.Errorf("Expected name 'Test Route', got %s", apiResp.Data[0].Name)
	}
}

func TestHandleCreateRoute(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Create route
	newRoute := types.Route{
		Name:                "New Route",
		Enabled:             true,
		StaticRouteNetwork:  "10.2.0.0/24",
		StaticRouteNexthop:  "192.168.1.1",
		Type:                types.RouteTypeNexthop,
	}

	body, _ := json.Marshal(newRoute)
	req, _ := http.NewRequest("POST", server.URL()+"/api/s/default/rest/routing", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testRoutingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp struct {
		Data []types.Route `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apiResp.Data) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(apiResp.Data))
	}

	if apiResp.Data[0].ID == "" {
		t.Error("Expected ID to be generated")
	}

	if apiResp.Data[0].Name != "New Route" {
		t.Errorf("Expected name 'New Route', got %s", apiResp.Data[0].Name)
	}
}

func TestHandleUpdateRoute(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test route
	server.state.AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route",
		Enabled:             true,
		StaticRouteNetwork:  "10.0.0.0/24",
		StaticRouteNexthop:  "192.168.1.1",
	})

	// Update route
	update := types.Route{
		Name:                "Updated Route",
		Enabled:             false,
		StaticRouteNetwork:  "10.0.0.0/24",
		StaticRouteNexthop:  "192.168.1.2",
	}

	body, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", server.URL()+"/api/s/default/rest/routing/route1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testRoutingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to update route: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify
	route := server.state.GetRoute("route1")
	if route == nil {
		t.Fatal("Route not found")
	}

	if route.Name != "Updated Route" {
		t.Errorf("Expected name 'Updated Route', got %s", route.Name)
	}
}

func TestHandleDeleteRoute(t *testing.T) {
	server := NewServer(WithoutAuth(), WithoutCSRF())
	defer server.Close()

	// Add test route
	server.state.AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route",
		Enabled:             true,
		StaticRouteNetwork:  "10.0.0.0/24",
	})

	// Delete route
	req, _ := http.NewRequest("DELETE", server.URL()+"/api/s/default/rest/routing/route1", nil)
	resp, err := testRoutingHTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to delete route: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify
	route := server.state.GetRoute("route1")
	if route != nil {
		t.Error("Expected route to be deleted")
	}
}
