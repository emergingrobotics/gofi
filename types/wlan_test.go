package types

import (
	"encoding/json"
	"testing"
)

func TestWLAN_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		check   func(*testing.T, *WLAN)
	}{
		{
			name: "basic WPA2 network",
			json: `{
				"_id": "wlan123",
				"site_id": "default",
				"name": "MyWiFi",
				"enabled": true,
				"security": "wpapsk",
				"wpa_mode": "wpa2",
				"wpa_enc": "ccmp",
				"x_passphrase": "SecurePassword123!",
				"hide_ssid": false,
				"is_guest": false,
				"networkconf_id": "net123",
				"usergroup_id": "user123",
				"ap_group_ids": ["apgroup1"],
				"wlan_bands": ["2g", "5g"],
				"fast_roaming_enabled": true,
				"uapsd_enabled": true,
				"mac_filter_enabled": false,
				"l2_isolation": false,
				"iapp_enabled": true
			}`,
			wantErr: false,
			check: func(t *testing.T, w *WLAN) {
				if w.ID != "wlan123" {
					t.Errorf("ID = %v, want wlan123", w.ID)
				}
				if w.Name != "MyWiFi" {
					t.Errorf("Name = %v, want MyWiFi", w.Name)
				}
				if w.Security != SecurityTypeWPAPSK {
					t.Errorf("Security = %v, want wpapsk", w.Security)
				}
				if w.WPAMode != WPAModeWPA2 {
					t.Errorf("WPAMode = %v, want wpa2", w.WPAMode)
				}
				if !w.FastRoamingEnabled {
					t.Error("FastRoamingEnabled should be true")
				}
				if len(w.WLANBands) != 2 {
					t.Errorf("WLANBands length = %v, want 2", len(w.WLANBands))
				}
			},
		},
		{
			name: "open guest network",
			json: `{
				"_id": "wlan456",
				"name": "Guest WiFi",
				"enabled": true,
				"security": "open",
				"hide_ssid": false,
				"is_guest": true,
				"networkconf_id": "guest_net",
				"l2_isolation": true,
				"wlan_bands": ["2g"],
				"fast_roaming_enabled": false,
				"uapsd_enabled": false,
				"mac_filter_enabled": false,
				"iapp_enabled": true
			}`,
			wantErr: false,
			check: func(t *testing.T, w *WLAN) {
				if w.Security != SecurityTypeOpen {
					t.Errorf("Security = %v, want open", w.Security)
				}
				if !w.IsGuest {
					t.Error("IsGuest should be true")
				}
				if !w.L2Isolation {
					t.Error("L2Isolation should be true")
				}
			},
		},
		{
			name: "WPA3 network",
			json: `{
				"_id": "wlan789",
				"name": "Secure WPA3",
				"enabled": true,
				"security": "wpapsk",
				"wpa_mode": "wpa3",
				"wpa_enc": "ccmp",
				"x_passphrase": "VerySecure!@#",
				"wpa3_support": true,
				"wpa3_transition": true,
				"pmf_mode": "required",
				"hide_ssid": false,
				"is_guest": false,
				"wlan_bands": ["5g", "6g"],
				"fast_roaming_enabled": true,
				"uapsd_enabled": true,
				"mac_filter_enabled": false,
				"iapp_enabled": true,
				"l2_isolation": false
			}`,
			wantErr: false,
			check: func(t *testing.T, w *WLAN) {
				if !w.WPA3Support {
					t.Error("WPA3Support should be true")
				}
				if !w.WPA3Transition {
					t.Error("WPA3Transition should be true")
				}
				if w.PMFMode != PMFModeRequired {
					t.Errorf("PMFMode = %v, want required", w.PMFMode)
				}
			},
		},
		{
			name: "enterprise network with RADIUS",
			json: `{
				"_id": "wlan101",
				"name": "Enterprise WiFi",
				"enabled": true,
				"security": "wpaeap",
				"wpa_mode": "wpa2",
				"wpa_enc": "ccmp",
				"hide_ssid": false,
				"is_guest": false,
				"radius_mac_auth_enabled": true,
				"radius_profile_id": "radius123",
				"wlan_bands": ["2g", "5g"],
				"fast_roaming_enabled": true,
				"uapsd_enabled": true,
				"mac_filter_enabled": false,
				"iapp_enabled": true,
				"l2_isolation": false
			}`,
			wantErr: false,
			check: func(t *testing.T, w *WLAN) {
				if w.Security != SecurityTypeWPAEAP {
					t.Errorf("Security = %v, want wpaeap", w.Security)
				}
				if !w.RADIUSMACAuthEnabled {
					t.Error("RADIUSMACAuthEnabled should be true")
				}
				if w.RADIUSProfileID != "radius123" {
					t.Errorf("RADIUSProfileID = %v, want radius123", w.RADIUSProfileID)
				}
			},
		},
		{
			name: "network with MAC filtering",
			json: `{
				"_id": "wlan202",
				"name": "MAC Filtered",
				"enabled": true,
				"security": "wpapsk",
				"wpa_mode": "wpa2",
				"x_passphrase": "password",
				"hide_ssid": false,
				"is_guest": false,
				"mac_filter_enabled": true,
				"mac_filter_policy": "allow",
				"mac_filter_list": ["aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"],
				"wlan_bands": ["2g"],
				"fast_roaming_enabled": false,
				"uapsd_enabled": false,
				"iapp_enabled": true,
				"l2_isolation": false
			}`,
			wantErr: false,
			check: func(t *testing.T, w *WLAN) {
				if !w.MACFilterEnabled {
					t.Error("MACFilterEnabled should be true")
				}
				if w.MACFilterPolicy != MACFilterPolicyAllow {
					t.Errorf("MACFilterPolicy = %v, want allow", w.MACFilterPolicy)
				}
				if len(w.MACFilterList) != 2 {
					t.Errorf("MACFilterList length = %v, want 2", len(w.MACFilterList))
				}
			},
		},
		{
			name: "network with schedule",
			json: `{
				"_id": "wlan303",
				"name": "Scheduled WiFi",
				"enabled": true,
				"security": "wpapsk",
				"wpa_mode": "wpa2",
				"x_passphrase": "password",
				"hide_ssid": false,
				"is_guest": false,
				"schedule_enabled": true,
				"schedule": ["mon|08:00-17:00", "tue|08:00-17:00"],
				"wlan_bands": ["2g"],
				"fast_roaming_enabled": false,
				"uapsd_enabled": false,
				"mac_filter_enabled": false,
				"iapp_enabled": true,
				"l2_isolation": false
			}`,
			wantErr: false,
			check: func(t *testing.T, w *WLAN) {
				if !w.ScheduleEnabled {
					t.Error("ScheduleEnabled should be true")
				}
				if len(w.Schedule) != 2 {
					t.Errorf("Schedule length = %v, want 2", len(w.Schedule))
				}
			},
		},
		{
			name: "network with data rates",
			json: `{
				"_id": "wlan404",
				"name": "Rate Limited WiFi",
				"enabled": true,
				"security": "wpapsk",
				"wpa_mode": "wpa2",
				"x_passphrase": "password",
				"hide_ssid": false,
				"is_guest": false,
				"minrate_ng_enabled": true,
				"minrate_ng_data_rate_kbps": 12000,
				"minrate_na_enabled": true,
				"minrate_na_data_rate_kbps": 24000,
				"wlan_bands": ["2g", "5g"],
				"fast_roaming_enabled": false,
				"uapsd_enabled": false,
				"mac_filter_enabled": false,
				"iapp_enabled": true,
				"l2_isolation": false
			}`,
			wantErr: false,
			check: func(t *testing.T, w *WLAN) {
				if !w.MinrateNGEnabled {
					t.Error("MinrateNGEnabled should be true")
				}
				if w.MinrateNGDataRateKbps != 12000 {
					t.Errorf("MinrateNGDataRateKbps = %v, want 12000", w.MinrateNGDataRateKbps)
				}
			},
		},
		{
			name: "network with statistics",
			json: `{
				"_id": "wlan505",
				"name": "Stats WiFi",
				"enabled": true,
				"security": "open",
				"hide_ssid": false,
				"is_guest": false,
				"num_sta": 25,
				"rx_bytes": "9876543210",
				"tx_bytes": 1234567890,
				"wlan_bands": ["2g"],
				"fast_roaming_enabled": false,
				"uapsd_enabled": false,
				"mac_filter_enabled": false,
				"iapp_enabled": true,
				"l2_isolation": false
			}`,
			wantErr: false,
			check: func(t *testing.T, w *WLAN) {
				if w.NumSTA != 25 {
					t.Errorf("NumSTA = %v, want 25", w.NumSTA)
				}
				if w.RXBytes.Int64() != 9876543210 {
					t.Errorf("RXBytes = %v, want 9876543210", w.RXBytes.Int64())
				}
				if w.TXBytes.Int64() != 1234567890 {
					t.Errorf("TXBytes = %v, want 1234567890", w.TXBytes.Int64())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w WLAN
			err := json.Unmarshal([]byte(tt.json), &w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, &w)
			}
		})
	}
}

func TestWLAN_MarshalJSON(t *testing.T) {
	w := WLAN{
		ID:                 "wlan123",
		SiteID:             "default",
		Name:               "Test WiFi",
		Enabled:            true,
		Security:           SecurityTypeWPAPSK,
		WPAMode:            WPAModeWPA2,
		WPAEnc:             WPAEncCCMP,
		Passphrase:         "TestPassword123",
		HideSSID:           false,
		IsGuest:            false,
		WLANBands:          []string{WLANBand2G, WLANBand5G},
		FastRoamingEnabled: true,
		UAPSDEnabled:       true,
		MACFilterEnabled:   false,
		IAPPEnabled:        true,
		L2Isolation:        false,
	}

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var w2 WLAN
	if err := json.Unmarshal(data, &w2); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if w2.ID != w.ID {
		t.Errorf("ID = %v, want %v", w2.ID, w.ID)
	}
	if w2.Security != w.Security {
		t.Errorf("Security = %v, want %v", w2.Security, w.Security)
	}
	if w2.WPAMode != w.WPAMode {
		t.Errorf("WPAMode = %v, want %v", w2.WPAMode, w.WPAMode)
	}
}

func TestWLANGroup_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "group123",
		"site_id": "default",
		"name": "Ground Floor APs",
		"attr_hidden_id": ["aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"],
		"attr_no_delete": false
	}`

	var wg WLANGroup
	if err := json.Unmarshal([]byte(jsonData), &wg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if wg.ID != "group123" {
		t.Errorf("ID = %v, want group123", wg.ID)
	}
	if wg.Name != "Ground Floor APs" {
		t.Errorf("Name = %v, want Ground Floor APs", wg.Name)
	}
	if len(wg.Members) != 2 {
		t.Errorf("Members length = %v, want 2", len(wg.Members))
	}
}

func TestWLANSchedule_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"day": "mon",
		"start_hour": 8,
		"start_min": 0,
		"end_hour": 17,
		"end_min": 30
	}`

	var schedule WLANSchedule
	if err := json.Unmarshal([]byte(jsonData), &schedule); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if schedule.Day != "mon" {
		t.Errorf("Day = %v, want mon", schedule.Day)
	}
	if schedule.StartHour != 8 {
		t.Errorf("StartHour = %v, want 8", schedule.StartHour)
	}
	if schedule.EndHour != 17 {
		t.Errorf("EndHour = %v, want 17", schedule.EndHour)
	}
}

func TestSecurityTypeConstants(t *testing.T) {
	types := []string{
		SecurityTypeOpen,
		SecurityTypeWPAPSK,
		SecurityTypeWPAEAP,
		SecurityTypeWPA3,
	}

	for _, st := range types {
		if st == "" {
			t.Errorf("Security type constant should not be empty")
		}
	}
}

func TestWPAModeConstants(t *testing.T) {
	modes := []string{
		WPAModeWPA,
		WPAModeWPA2,
		WPAModeWPA3,
		WPAModeBoth,
	}

	for _, mode := range modes {
		if mode == "" {
			t.Errorf("WPA mode constant should not be empty")
		}
	}
}

func TestMACFilterPolicyConstants(t *testing.T) {
	policies := []string{
		MACFilterPolicyAllow,
		MACFilterPolicyDeny,
	}

	for _, policy := range policies {
		if policy == "" {
			t.Errorf("MAC filter policy constant should not be empty")
		}
	}
}

func TestWLANBandConstants(t *testing.T) {
	bands := []string{
		WLANBand2G,
		WLANBand5G,
		WLANBand6G,
	}

	for _, band := range bands {
		if band == "" {
			t.Errorf("WLAN band constant should not be empty")
		}
	}
}
