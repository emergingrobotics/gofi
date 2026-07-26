package types

import (
	"encoding/json"
	"testing"
)

func TestStatus_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"up": true,
		"server_version": "10.0.25",
		"hostname": "udm-pro",
		"version": "10.0.25"
	}`

	var status Status
	if err := json.Unmarshal([]byte(jsonData), &status); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !status.Up {
		t.Error("Up should be true")
	}
	if status.ServerVersion != "10.0.25" {
		t.Errorf("ServerVersion = %v, want 10.0.25", status.ServerVersion)
	}
}

func TestAdminUser_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "admin123",
		"unique_id": "uid123",
		"username": "admin",
		"full_name": "Admin User",
		"email": "admin@example.com",
		"isOwner": true,
		"isSuperAdmin": true,
		"roles": [
			{
				"name": "Super Admin",
				"system_role": true,
				"unique_id": "role123",
				"system_key": "super_admin"
			}
		]
	}`

	var admin AdminUser
	if err := json.Unmarshal([]byte(jsonData), &admin); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if admin.Username != "admin" {
		t.Errorf("Username = %v, want admin", admin.Username)
	}
	if !admin.IsSuperAdmin {
		t.Error("IsSuperAdmin should be true")
	}
	if len(admin.Roles) != 1 {
		t.Errorf("Roles length = %v, want 1", len(admin.Roles))
	}
}

func TestSpeedTestStatus_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"status_download": 100,
		"status_upload": 100,
		"latency": 15,
		"xput_download": "500000",
		"xput_upload": 100000,
		"running": false,
		"server_name": "Test Server"
	}`

	var status SpeedTestStatus
	if err := json.Unmarshal([]byte(jsonData), &status); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if status.Latency != 15 {
		t.Errorf("Latency = %v, want 15", status.Latency)
	}
	if status.XputDownload.Int() != 500000 {
		t.Errorf("XputDownload = %v, want 500000", status.XputDownload.Int())
	}
	if status.Running {
		t.Error("Running should be false")
	}
}
