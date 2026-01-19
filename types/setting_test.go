package types

import (
	"encoding/json"
	"testing"
)

func TestSettingMgmt_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "set123",
		"site_id": "default",
		"key": "mgmt",
		"led_enabled": true,
		"alert_enabled": true,
		"auto_upgrade": false
	}`

	var setting SettingMgmt
	if err := json.Unmarshal([]byte(jsonData), &setting); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if setting.Key != SettingKeyMgmt {
		t.Errorf("Key = %v, want mgmt", setting.Key)
	}
	if !setting.LEDEnabled {
		t.Error("LEDEnabled should be true")
	}
}

func TestRADIUSProfile_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "radius123",
		"site_id": "default",
		"name": "Main RADIUS",
		"auth_servers": [
			{
				"ip": "192.168.1.10",
				"port": 1812,
				"x_secret": "secret123"
			}
		],
		"vlan_enabled": true
	}`

	var profile RADIUSProfile
	if err := json.Unmarshal([]byte(jsonData), &profile); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if profile.Name != "Main RADIUS" {
		t.Errorf("Name = %v, want Main RADIUS", profile.Name)
	}
	if len(profile.AuthServers) != 1 {
		t.Errorf("AuthServers length = %v, want 1", len(profile.AuthServers))
	}
	if !profile.VLANEnabled {
		t.Error("VLANEnabled should be true")
	}
}

func TestDynamicDNS_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "ddns123",
		"site_id": "default",
		"service": "dyndns",
		"enabled": true,
		"host": "myhost.dyndns.org",
		"login": "username"
	}`

	var ddns DynamicDNS
	if err := json.Unmarshal([]byte(jsonData), &ddns); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if ddns.Service != "dyndns" {
		t.Errorf("Service = %v, want dyndns", ddns.Service)
	}
	if !ddns.Enabled {
		t.Error("Enabled should be true")
	}
}
