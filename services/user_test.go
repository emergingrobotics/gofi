package services

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/transport"
	"github.com/unifi-go/gofi/types"
)

func newTestUserTransport(url string) (transport.Transport, error) {
	config := transport.DefaultConfig(url)
	config.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return transport.New(config)
}

func TestUserService_List(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test users
	server.State().AddKnownClient(&types.User{
		ID:   "user1",
		MAC:  "aa:bb:cc:dd:ee:f1",
		Name: "Test User 1",
	})
	server.State().AddKnownClient(&types.User{
		ID:   "user2",
		MAC:  "aa:bb:cc:dd:ee:f2",
		Name: "Test User 2",
	})

	// Create service
	trans, _ := newTestUserTransport(server.URL())
	svc := NewUserService(trans)

	// Test List
	users, err := svc.List(context.Background(), "default")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestUserService_GetByMAC(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test user
	server.State().AddKnownClient(&types.User{
		ID:   "user1",
		MAC:  "aa:bb:cc:dd:ee:ff",
		Name: "Test User",
	})

	// Create service
	trans, _ := newTestUserTransport(server.URL())
	svc := NewUserService(trans)

	// Test GetByMAC
	user, err := svc.GetByMAC(context.Background(), "default", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("GetByMAC failed: %v", err)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got %s", user.Name)
	}
}

func TestUserService_Create(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestUserTransport(server.URL())
	svc := NewUserService(trans)

	// Test Create
	newUser := &types.User{
		MAC:  "aa:bb:cc:dd:ee:f1",
		Name: "New User",
	}

	created, err := svc.Create(context.Background(), "default", newUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.Name != "New User" {
		t.Errorf("Expected name 'New User', got %s", created.Name)
	}

	if created.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestUserService_SetFixedIP(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test user
	server.State().AddKnownClient(&types.User{
		ID:   "user1",
		MAC:  "aa:bb:cc:dd:ee:ff",
		Name: "Test User",
	})

	// Create service
	trans, _ := newTestUserTransport(server.URL())
	svc := NewUserService(trans)

	// Test SetFixedIP
	err := svc.SetFixedIP(context.Background(), "default", "aa:bb:cc:dd:ee:ff", "192.168.1.100", "network1")
	if err != nil {
		t.Fatalf("SetFixedIP failed: %v", err)
	}

	// Verify
	user := server.State().GetKnownClientByMAC("aa:bb:cc:dd:ee:ff")
	if user == nil {
		t.Fatal("User not found")
	}

	if !user.UseFixedIP {
		t.Error("Expected UseFixedIP to be true")
	}

	if user.FixedIP != "192.168.1.100" {
		t.Errorf("Expected FixedIP 192.168.1.100, got %s", user.FixedIP)
	}
}

func TestUserService_ListGroups(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test groups
	server.State().AddUserGroup(&types.UserGroup{
		ID:   "group1",
		Name: "Test Group 1",
	})
	server.State().AddUserGroup(&types.UserGroup{
		ID:   "group2",
		Name: "Test Group 2",
	})

	// Create service
	trans, _ := newTestUserTransport(server.URL())
	svc := NewUserService(trans)

	// Test ListGroups
	groups, err := svc.ListGroups(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}
}

func TestUserService_CreateGroup(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestUserTransport(server.URL())
	svc := NewUserService(trans)

	// Test CreateGroup
	newGroup := &types.UserGroup{
		Name:            "New Group",
		QOSRateMaxDown:  1000,
		QOSRateMaxUp:    500,
	}

	created, err := svc.CreateGroup(context.Background(), "default", newGroup)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if created.Name != "New Group" {
		t.Errorf("Expected name 'New Group', got %s", created.Name)
	}

	if created.ID == "" {
		t.Error("Expected ID to be generated")
	}
}
