package types

import (
	"encoding/json"
	"testing"
)

func TestUser_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "user123",
		"site_id": "default",
		"mac": "aa:bb:cc:dd:ee:ff",
		"hostname": "server",
		"name": "My Server",
		"use_fixedip": true,
		"network_id": "net123",
		"fixed_ip": "192.168.1.50",
		"blocked": false
	}`

	var user User
	if err := json.Unmarshal([]byte(jsonData), &user); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if user.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("MAC = %v, want aa:bb:cc:dd:ee:ff", user.MAC)
	}
	if !user.UseFixedIP {
		t.Error("UseFixedIP should be true")
	}
	if user.FixedIP != "192.168.1.50" {
		t.Errorf("FixedIP = %v, want 192.168.1.50", user.FixedIP)
	}
}

func TestUserGroup_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "group123",
		"site_id": "default",
		"name": "Limited Users",
		"qos_rate_max_down": 10000,
		"qos_rate_max_up": 5000
	}`

	var group UserGroup
	if err := json.Unmarshal([]byte(jsonData), &group); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if group.Name != "Limited Users" {
		t.Errorf("Name = %v, want Limited Users", group.Name)
	}
	if group.QOSRateMaxDown != 10000 {
		t.Errorf("QOSRateMaxDown = %v, want 10000", group.QOSRateMaxDown)
	}
}
