package services

import (
	"context"
	"testing"

	"github.com/unifi-go/gofi/mock"
	"github.com/unifi-go/gofi/types"
)

func TestWLANService_List(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLANs
	server.State().AddWLAN(&types.WLAN{
		ID:       "wlan1",
		Name:     "Guest Network",
		Enabled:  true,
		Security: types.SecurityTypeOpen,
		IsGuest:  true,
	})
	server.State().AddWLAN(&types.WLAN{
		ID:       "wlan2",
		Name:     "Corporate Network",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
		IsGuest:  false,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test List
	wlans, err := svc.List(context.Background(), "default")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(wlans) != 2 {
		t.Errorf("Expected 2 WLANs, got %d", len(wlans))
	}
}

func TestWLANService_Get(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN
	server.State().AddWLAN(&types.WLAN{
		ID:       "wlan1",
		Name:     "Test Network",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test Get
	wlan, err := svc.Get(context.Background(), "default", "wlan1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if wlan.Name != "Test Network" {
		t.Errorf("Expected name 'Test Network', got %s", wlan.Name)
	}

	if wlan.Security != types.SecurityTypeWPAPSK {
		t.Errorf("Expected security %s, got %s", types.SecurityTypeWPAPSK, wlan.Security)
	}
}

func TestWLANService_GetNotFound(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test Get non-existent WLAN
	_, err := svc.Get(context.Background(), "default", "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent WLAN")
	}
}

func TestWLANService_Create(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test Create
	newWLAN := &types.WLAN{
		Name:       "New Network",
		Enabled:    true,
		Security:   types.SecurityTypeWPAPSK,
		Passphrase: "testpassword123",
	}

	created, err := svc.Create(context.Background(), "default", newWLAN)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.Name != "New Network" {
		t.Errorf("Expected name 'New Network', got %s", created.Name)
	}

	if created.ID == "" {
		t.Error("Expected ID to be set")
	}
}

func TestWLANService_Update(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN
	server.State().AddWLAN(&types.WLAN{
		ID:       "wlan1",
		Name:     "Old Name",
		Enabled:  false,
		Security: types.SecurityTypeWPAPSK,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test Update
	updatedWLAN := &types.WLAN{
		ID:       "wlan1",
		Name:     "New Name",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
	}

	updated, err := svc.Update(context.Background(), "default", updatedWLAN)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got %s", updated.Name)
	}

	if !updated.Enabled {
		t.Error("Expected WLAN to be enabled")
	}
}

func TestWLANService_Delete(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN
	server.State().AddWLAN(&types.WLAN{
		ID:       "wlan1",
		Name:     "Test Network",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test Delete
	err := svc.Delete(context.Background(), "default", "wlan1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	if wlan := server.State().GetWLAN("wlan1"); wlan != nil {
		t.Error("Expected WLAN to be deleted")
	}
}

func TestWLANService_Enable(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN
	server.State().AddWLAN(&types.WLAN{
		ID:       "wlan1",
		Name:     "Test Network",
		Enabled:  false,
		Security: types.SecurityTypeWPAPSK,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test Enable
	err := svc.Enable(context.Background(), "default", "wlan1")
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}

	// Verify enabled
	wlan := server.State().GetWLAN("wlan1")
	if wlan == nil {
		t.Fatal("WLAN not found")
	}
	if !wlan.Enabled {
		t.Error("Expected WLAN to be enabled")
	}
}

func TestWLANService_Disable(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN
	server.State().AddWLAN(&types.WLAN{
		ID:       "wlan1",
		Name:     "Test Network",
		Enabled:  true,
		Security: types.SecurityTypeWPAPSK,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test Disable
	err := svc.Disable(context.Background(), "default", "wlan1")
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}

	// Verify disabled
	wlan := server.State().GetWLAN("wlan1")
	if wlan == nil {
		t.Fatal("WLAN not found")
	}
	if wlan.Enabled {
		t.Error("Expected WLAN to be disabled")
	}
}

func TestWLANService_SetMACFilter(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN
	server.State().AddWLAN(&types.WLAN{
		ID:                "wlan1",
		Name:              "Test Network",
		Enabled:           true,
		Security:          types.SecurityTypeWPAPSK,
		MACFilterEnabled:  false,
		MACFilterPolicy:   "",
		MACFilterList:     nil,
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test SetMACFilter
	macs := []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"}
	err := svc.SetMACFilter(context.Background(), "default", "wlan1", types.MACFilterPolicyAllow, macs)
	if err != nil {
		t.Fatalf("SetMACFilter failed: %v", err)
	}

	// Verify MAC filter settings
	wlan := server.State().GetWLAN("wlan1")
	if wlan == nil {
		t.Fatal("WLAN not found")
	}
	if !wlan.MACFilterEnabled {
		t.Error("Expected MAC filter to be enabled")
	}
	if wlan.MACFilterPolicy != types.MACFilterPolicyAllow {
		t.Errorf("Expected policy %s, got %s", types.MACFilterPolicyAllow, wlan.MACFilterPolicy)
	}
	if len(wlan.MACFilterList) != 2 {
		t.Errorf("Expected 2 MACs in filter list, got %d", len(wlan.MACFilterList))
	}
}

func TestWLANService_ListGroups(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN groups
	server.State().AddWLANGroup(&types.WLANGroup{
		ID:      "group1",
		Name:    "Group 1",
		Members: []string{"aa:bb:cc:dd:ee:ff"},
	})
	server.State().AddWLANGroup(&types.WLANGroup{
		ID:      "group2",
		Name:    "Group 2",
		Members: []string{},
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test ListGroups
	groups, err := svc.ListGroups(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 WLAN groups, got %d", len(groups))
	}
}

func TestWLANService_GetGroup(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN group
	server.State().AddWLANGroup(&types.WLANGroup{
		ID:      "group1",
		Name:    "Test Group",
		Members: []string{"aa:bb:cc:dd:ee:ff"},
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test GetGroup
	group, err := svc.GetGroup(context.Background(), "default", "group1")
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if group.Name != "Test Group" {
		t.Errorf("Expected name 'Test Group', got %s", group.Name)
	}

	if len(group.Members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(group.Members))
	}
}

func TestWLANService_CreateGroup(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test CreateGroup
	newGroup := &types.WLANGroup{
		Name:    "New Group",
		Members: []string{"aa:bb:cc:dd:ee:ff"},
	}

	created, err := svc.CreateGroup(context.Background(), "default", newGroup)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if created.Name != "New Group" {
		t.Errorf("Expected name 'New Group', got %s", created.Name)
	}

	if created.ID == "" {
		t.Error("Expected ID to be set")
	}
}

func TestWLANService_UpdateGroup(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN group
	server.State().AddWLANGroup(&types.WLANGroup{
		ID:      "group1",
		Name:    "Old Name",
		Members: []string{},
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test UpdateGroup
	updatedGroup := &types.WLANGroup{
		ID:      "group1",
		Name:    "New Name",
		Members: []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"},
	}

	updated, err := svc.UpdateGroup(context.Background(), "default", updatedGroup)
	if err != nil {
		t.Fatalf("UpdateGroup failed: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got %s", updated.Name)
	}

	if len(updated.Members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(updated.Members))
	}
}

func TestWLANService_DeleteGroup(t *testing.T) {
	server := mock.NewServer(mock.WithoutAuth(), mock.WithoutCSRF())
	defer server.Close()

	// Add test WLAN group
	server.State().AddWLANGroup(&types.WLANGroup{
		ID:      "group1",
		Name:    "Test Group",
		Members: []string{},
	})

	// Create service
	trans, _ := newTestTransport(server.URL())
	svc := NewWLANService(trans)

	// Test DeleteGroup
	err := svc.DeleteGroup(context.Background(), "default", "group1")
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	// Verify deletion
	if group := server.State().GetWLANGroup("group1"); group != nil {
		t.Error("Expected WLAN group to be deleted")
	}
}
