package types

import (
	"encoding/json"
	"testing"
)

func TestNetwork_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		check   func(*testing.T, *Network)
	}{
		{
			name: "basic corporate network with DHCP",
			json: `{
				"_id": "net123",
				"site_id": "default",
				"name": "LAN",
				"purpose": "corporate",
				"vlan_enabled": false,
				"ip_subnet": "192.168.1.0/24",
				"dhcpd_enabled": true,
				"dhcpd_start": "192.168.1.100",
				"dhcpd_stop": "192.168.1.200",
				"dhcpd_leasetime": 86400,
				"dhcpd_dns_enabled": true,
				"dhcpd_dns_1": "8.8.8.8",
				"dhcpd_gateway_enabled": true,
				"dhcpd_gateway": "192.168.1.1",
				"enabled": true,
				"is_nat": true,
				"networkgroup": "LAN",
				"dhcpguard_enabled": false
			}`,
			wantErr: false,
			check: func(t *testing.T, n *Network) {
				if n.ID != "net123" {
					t.Errorf("ID = %v, want net123", n.ID)
				}
				if n.Name != "LAN" {
					t.Errorf("Name = %v, want LAN", n.Name)
				}
				if n.Purpose != NetworkPurposeCorporate {
					t.Errorf("Purpose = %v, want corporate", n.Purpose)
				}
				if !n.DHCPDEnabled {
					t.Error("DHCPDEnabled should be true")
				}
				if n.DHCPDStart != "192.168.1.100" {
					t.Errorf("DHCPDStart = %v, want 192.168.1.100", n.DHCPDStart)
				}
				if !n.IsNAT {
					t.Error("IsNAT should be true")
				}
			},
		},
		{
			name: "VLAN network",
			json: `{
				"_id": "net456",
				"name": "IoT VLAN",
				"purpose": "corporate",
				"vlan_enabled": true,
				"vlan": 20,
				"ip_subnet": "10.0.20.0/24",
				"dhcpd_enabled": true,
				"dhcpd_start": "10.0.20.10",
				"dhcpd_stop": "10.0.20.250",
				"enabled": true,
				"is_nat": false,
				"networkgroup": "LAN",
				"igmp_snooping": true,
				"dhcpguard_enabled": true
			}`,
			wantErr: false,
			check: func(t *testing.T, n *Network) {
				if !n.VLANEnabled {
					t.Error("VLANEnabled should be true")
				}
				if n.VLAN != 20 {
					t.Errorf("VLAN = %v, want 20", n.VLAN)
				}
				if n.IGMPSnooping != true {
					t.Error("IGMPSnooping should be true")
				}
				if !n.DHCPGuardEnabled {
					t.Error("DHCPGuardEnabled should be true")
				}
			},
		},
		{
			name: "guest network",
			json: `{
				"_id": "net789",
				"name": "Guest WiFi",
				"purpose": "guest",
				"vlan_enabled": true,
				"vlan": 10,
				"ip_subnet": "10.0.10.0/24",
				"dhcpd_enabled": true,
				"enabled": true,
				"is_nat": true,
				"networkgroup": "LAN",
				"dhcpguard_enabled": false
			}`,
			wantErr: false,
			check: func(t *testing.T, n *Network) {
				if n.Purpose != NetworkPurposeGuest {
					t.Errorf("Purpose = %v, want guest", n.Purpose)
				}
				if n.VLAN != 10 {
					t.Errorf("VLAN = %v, want 10", n.VLAN)
				}
			},
		},
		{
			name: "WAN network with DHCP",
			json: `{
				"_id": "wan1",
				"name": "WAN",
				"purpose": "wan",
				"enabled": true,
				"wan_type": "dhcp",
				"networkgroup": "WAN",
				"dhcpguard_enabled": false
			}`,
			wantErr: false,
			check: func(t *testing.T, n *Network) {
				if n.Purpose != NetworkPurposeWAN {
					t.Errorf("Purpose = %v, want wan", n.Purpose)
				}
				if n.WANType != WANTypeDHCP {
					t.Errorf("WANType = %v, want dhcp", n.WANType)
				}
				if n.NetworkGroup != NetworkGroupWAN {
					t.Errorf("NetworkGroup = %v, want WAN", n.NetworkGroup)
				}
			},
		},
		{
			name: "static WAN network",
			json: `{
				"_id": "wan2",
				"name": "Static WAN",
				"purpose": "wan",
				"enabled": true,
				"wan_type": "static",
				"wan_ip": "203.0.113.10",
				"wan_netmask": "255.255.255.0",
				"wan_gateway": "203.0.113.1",
				"wan_dns": ["8.8.8.8", "8.8.4.4"],
				"networkgroup": "WAN",
				"dhcpguard_enabled": false
			}`,
			wantErr: false,
			check: func(t *testing.T, n *Network) {
				if n.WANType != WANTypeStatic {
					t.Errorf("WANType = %v, want static", n.WANType)
				}
				if n.WANIPAddress != "203.0.113.10" {
					t.Errorf("WANIPAddress = %v, want 203.0.113.10", n.WANIPAddress)
				}
				if len(n.WANDNS) != 2 {
					t.Errorf("WANDNS length = %v, want 2", len(n.WANDNS))
				}
			},
		},
		{
			name: "PPPoE WAN network",
			json: `{
				"_id": "wan3",
				"name": "PPPoE WAN",
				"purpose": "wan",
				"enabled": true,
				"wan_type": "pppoe",
				"wan_username": "user@isp.com",
				"wan_password": "secret",
				"networkgroup": "WAN",
				"dhcpguard_enabled": false
			}`,
			wantErr: false,
			check: func(t *testing.T, n *Network) {
				if n.WANType != WANTypePPPoE {
					t.Errorf("WANType = %v, want pppoe", n.WANType)
				}
				if n.WANUsername != "user@isp.com" {
					t.Errorf("WANUsername = %v, want user@isp.com", n.WANUsername)
				}
			},
		},
		{
			name: "network with flex types",
			json: `{
				"_id": "net999",
				"name": "Test Network",
				"purpose": "corporate",
				"ip_subnet": "10.0.0.0/24",
				"enabled": true,
				"networkgroup": "LAN",
				"dhcpguard_enabled": false,
				"rx_bytes": "1234567890",
				"tx_bytes": 9876543210,
				"num_sta": 42
			}`,
			wantErr: false,
			check: func(t *testing.T, n *Network) {
				if n.RXBytes.Int64() != 1234567890 {
					t.Errorf("RXBytes = %v, want 1234567890", n.RXBytes.Int64())
				}
				if n.TXBytes.Int64() != 9876543210 {
					t.Errorf("TXBytes = %v, want 9876543210", n.TXBytes.Int64())
				}
				if n.NumSTA != 42 {
					t.Errorf("NumSTA = %v, want 42", n.NumSTA)
				}
			},
		},
		{
			name: "network with IPv6",
			json: `{
				"_id": "net100",
				"name": "IPv6 Network",
				"purpose": "corporate",
				"ip_subnet": "10.0.0.0/24",
				"enabled": true,
				"networkgroup": "LAN",
				"dhcpguard_enabled": false,
				"ipv6_interface_type": "static",
				"ipv6_ra_enabled": true,
				"ipv6_ra_valid_lifetime": "86400",
				"ipv6_ra_preferred_lifetime": 43200
			}`,
			wantErr: false,
			check: func(t *testing.T, n *Network) {
				if !n.IPv6RAEnabled {
					t.Error("IPv6RAEnabled should be true")
				}
				if n.IPv6RAValidLifetime.Int() != 86400 {
					t.Errorf("IPv6RAValidLifetime = %v, want 86400", n.IPv6RAValidLifetime.Int())
				}
				if n.IPv6RAPreferredLife.Int() != 43200 {
					t.Errorf("IPv6RAPreferredLife = %v, want 43200", n.IPv6RAPreferredLife.Int())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n Network
			err := json.Unmarshal([]byte(tt.json), &n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, &n)
			}
		})
	}
}

func TestNetwork_MarshalJSON(t *testing.T) {
	n := Network{
		ID:                  "net123",
		SiteID:              "default",
		Name:                "Test Network",
		Purpose:             NetworkPurposeCorporate,
		VLANEnabled:         true,
		VLAN:                20,
		IPSubnet:            "10.0.20.0/24",
		DHCPDEnabled:        true,
		DHCPDStart:          "10.0.20.10",
		DHCPDStop:           "10.0.20.250",
		DHCPDLeaseTime:      86400,
		DHCPDDNSEnabled:     true,
		DHCPDDNS1:           "8.8.8.8",
		DHCPDGatewayEnabled: true,
		Enabled:             true,
		IsNAT:               true,
		NetworkGroup:        NetworkGroupLAN,
		DHCPGuardEnabled:    false,
	}

	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var n2 Network
	if err := json.Unmarshal(data, &n2); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if n2.ID != n.ID {
		t.Errorf("ID = %v, want %v", n2.ID, n.ID)
	}
	if n2.VLAN != n.VLAN {
		t.Errorf("VLAN = %v, want %v", n2.VLAN, n.VLAN)
	}
	if n2.Purpose != n.Purpose {
		t.Errorf("Purpose = %v, want %v", n2.Purpose, n.Purpose)
	}
}

func TestNetworkPurposeConstants(t *testing.T) {
	purposes := []string{
		NetworkPurposeCorporate,
		NetworkPurposeGuest,
		NetworkPurposeWAN,
		NetworkPurposeVPN,
		NetworkPurposeVLANOnly,
		NetworkPurposeRemoteUser,
	}

	for _, p := range purposes {
		if p == "" {
			t.Errorf("Purpose constant should not be empty")
		}
	}
}

func TestWANTypeConstants(t *testing.T) {
	types := []string{
		WANTypeDHCP,
		WANTypeStatic,
		WANTypePPPoE,
		WANTypeDisabled,
	}

	for _, wt := range types {
		if wt == "" {
			t.Errorf("WAN type constant should not be empty")
		}
	}
}

func TestWANProviderCaps_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"download_kilobits_per_second": "100000",
		"upload_kilobits_per_second": 50000
	}`

	var caps WANProviderCaps
	if err := json.Unmarshal([]byte(jsonData), &caps); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if caps.DownloadKilobitsPerSecond.Int() != 100000 {
		t.Errorf("DownloadKilobitsPerSecond = %v, want 100000", caps.DownloadKilobitsPerSecond.Int())
	}
	if caps.UploadKilobitsPerSecond.Int() != 50000 {
		t.Errorf("UploadKilobitsPerSecond = %v, want 50000", caps.UploadKilobitsPerSecond.Int())
	}
}
