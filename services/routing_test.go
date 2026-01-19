package services

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

func newTestRoutingTransport(url string) (transport.Transport, error) {
	config := transport.DefaultConfig(url)
	config.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return transport.New(config)
}

func TestRoutingService_List(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test routes
	server.State().AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route 1",
		Enabled:             true,
		StaticRouteNetwork:  "10.0.0.0/24",
		StaticRouteNexthop:  "192.168.1.1",
		Type:                types.RouteTypeNexthop,
	})
	server.State().AddRoute(&types.Route{
		ID:                  "route2",
		Name:                "Test Route 2",
		Enabled:             false,
		StaticRouteNetwork:  "10.1.0.0/24",
		Type:                types.RouteTypeBlackhole,
	})

	// Create service
	trans, _ := newTestRoutingTransport(server.URL())
	svc := NewRoutingService(trans)

	// Test List
	routes, err := svc.List(context.Background(), "default")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(routes))
	}
}

func TestRoutingService_Get(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test route
	server.State().AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route",
		Enabled:             true,
		StaticRouteNetwork:  "10.0.0.0/24",
		StaticRouteNexthop:  "192.168.1.1",
	})

	// Create service
	trans, _ := newTestRoutingTransport(server.URL())
	svc := NewRoutingService(trans)

	// Test Get
	route, err := svc.Get(context.Background(), "default", "route1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if route.Name != "Test Route" {
		t.Errorf("Expected name 'Test Route', got %s", route.Name)
	}

	if !route.Enabled {
		t.Error("Expected route to be enabled")
	}
}

func TestRoutingService_Create(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestRoutingTransport(server.URL())
	svc := NewRoutingService(trans)

	// Test Create
	newRoute := &types.Route{
		Name:                "New Route",
		Enabled:             true,
		StaticRouteNetwork:  "10.2.0.0/24",
		StaticRouteNexthop:  "192.168.1.1",
		Type:                types.RouteTypeNexthop,
	}

	created, err := svc.Create(context.Background(), "default", newRoute)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.Name != "New Route" {
		t.Errorf("Expected name 'New Route', got %s", created.Name)
	}

	if created.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestRoutingService_Update(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test route
	server.State().AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route",
		Enabled:             true,
		StaticRouteNetwork:  "10.0.0.0/24",
		StaticRouteNexthop:  "192.168.1.1",
	})

	// Create service
	trans, _ := newTestRoutingTransport(server.URL())
	svc := NewRoutingService(trans)

	// Test Update
	route, _ := svc.Get(context.Background(), "default", "route1")
	route.Name = "Updated Route"
	route.StaticRouteNexthop = "192.168.1.2"

	updated, err := svc.Update(context.Background(), "default", route)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != "Updated Route" {
		t.Errorf("Expected name 'Updated Route', got %s", updated.Name)
	}

	if updated.StaticRouteNexthop != "192.168.1.2" {
		t.Errorf("Expected nexthop 192.168.1.2, got %s", updated.StaticRouteNexthop)
	}
}

func TestRoutingService_Delete(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test route
	server.State().AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route",
		Enabled:             true,
		StaticRouteNetwork:  "10.0.0.0/24",
	})

	// Create service
	trans, _ := newTestRoutingTransport(server.URL())
	svc := NewRoutingService(trans)

	// Test Delete
	err := svc.Delete(context.Background(), "default", "route1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify
	_, err = svc.Get(context.Background(), "default", "route1")
	if err == nil {
		t.Error("Expected error when getting deleted route")
	}
}

func TestRoutingService_Enable(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add disabled route
	server.State().AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route",
		Enabled:             false,
		StaticRouteNetwork:  "10.0.0.0/24",
	})

	// Create service
	trans, _ := newTestRoutingTransport(server.URL())
	svc := NewRoutingService(trans)

	// Test Enable
	err := svc.Enable(context.Background(), "default", "route1")
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}

	// Verify
	route := server.State().GetRoute("route1")
	if route == nil {
		t.Fatal("Route not found")
	}

	if !route.Enabled {
		t.Error("Expected route to be enabled")
	}
}

func TestRoutingService_Disable(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add enabled route
	server.State().AddRoute(&types.Route{
		ID:                  "route1",
		Name:                "Test Route",
		Enabled:             true,
		StaticRouteNetwork:  "10.0.0.0/24",
	})

	// Create service
	trans, _ := newTestRoutingTransport(server.URL())
	svc := NewRoutingService(trans)

	// Test Disable
	err := svc.Disable(context.Background(), "default", "route1")
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}

	// Verify
	route := server.State().GetRoute("route1")
	if route == nil {
		t.Fatal("Route not found")
	}

	if route.Enabled {
		t.Error("Expected route to be disabled")
	}
}
