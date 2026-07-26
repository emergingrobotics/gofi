package types

import (
	"encoding/json"
	"testing"
)

func TestSite_JSON(t *testing.T) {
	input := `{
		"_id": "507f1f77bcf86cd799439011",
		"name": "default",
		"desc": "Default Site",
		"role": "admin",
		"health": [
			{
				"subsystem": "www",
				"status": "ok",
				"num_user": 5
			}
		]
	}`

	var site Site
	if err := json.Unmarshal([]byte(input), &site); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if site.ID != "507f1f77bcf86cd799439011" {
		t.Errorf("ID = %s, want 507f1f77bcf86cd799439011", site.ID)
	}

	if site.Name != "default" {
		t.Errorf("Name = %s, want default", site.Name)
	}

	if site.Desc != "Default Site" {
		t.Errorf("Desc = %s, want Default Site", site.Desc)
	}

	if len(site.Health) != 1 {
		t.Fatalf("len(Health) = %d, want 1", len(site.Health))
	}

	if site.Health[0].Subsystem != "www" {
		t.Errorf("Health[0].Subsystem = %s, want www", site.Health[0].Subsystem)
	}
}

func TestHealthData_JSON(t *testing.T) {
	input := `{
		"subsystem": "wan",
		"status": "ok",
		"num_gw": 1,
		"num_sta": 10,
		"wan_ip": "192.168.1.1",
		"tx_bytes-r": "1234567",
		"rx_bytes-r": 7654321
	}`

	var health HealthData
	if err := json.Unmarshal([]byte(input), &health); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if health.Subsystem != "wan" {
		t.Errorf("Subsystem = %s, want wan", health.Subsystem)
	}

	if health.TxBytesR.Int() != 1234567 {
		t.Errorf("TxBytesR = %d, want 1234567", health.TxBytesR.Int())
	}

	if health.RxBytesR.Int() != 7654321 {
		t.Errorf("RxBytesR = %d, want 7654321", health.RxBytesR.Int())
	}
}

func TestSysInfo_JSON(t *testing.T) {
	input := `{
		"hostname": "UDM-Pro",
		"version": "7.5.174",
		"uptime": 86400,
		"https_port": 443,
		"cloudkey": false,
		"console": true
	}`

	var sysInfo SysInfo
	if err := json.Unmarshal([]byte(input), &sysInfo); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if sysInfo.Hostname != "UDM-Pro" {
		t.Errorf("Hostname = %s, want UDM-Pro", sysInfo.Hostname)
	}

	if sysInfo.Version != "7.5.174" {
		t.Errorf("Version = %s, want 7.5.174", sysInfo.Version)
	}

	if sysInfo.HTTPSPort != 443 {
		t.Errorf("HTTPSPort = %d, want 443", sysInfo.HTTPSPort)
	}
}

func TestCreateSiteRequest_JSON(t *testing.T) {
	req := CreateSiteRequest{
		Desc: "Test Site",
		Name: "test",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var result CreateSiteRequest
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if result.Desc != "Test Site" {
		t.Errorf("Desc = %s, want Test Site", result.Desc)
	}
}
