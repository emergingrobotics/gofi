package types

import (
	"encoding/json"
	"testing"
)

func TestDevice_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		check   func(*testing.T, *Device)
	}{
		{
			name: "basic AP device",
			json: `{
				"_id": "device123",
				"mac": "00:11:22:33:44:55",
				"model": "U6-Lite",
				"type": "uap",
				"name": "Office AP",
				"serial": "ABC123DEF456",
				"version": "6.5.54",
				"adopted": true,
				"site_id": "default",
				"state": 1,
				"last_seen": 1642567890,
				"uptime": 86400,
				"upgradable": false,
				"num_sta": 15,
				"radio_table": [
					{
						"radio": "ng",
						"name": "wifi0",
						"builtin_antenna": true
					}
				]
			}`,
			wantErr: false,
			check: func(t *testing.T, d *Device) {
				if d.ID != "device123" {
					t.Errorf("ID = %v, want device123", d.ID)
				}
				if d.MAC != "00:11:22:33:44:55" {
					t.Errorf("MAC = %v, want 00:11:22:33:44:55", d.MAC)
				}
				if d.Type != "uap" {
					t.Errorf("Type = %v, want uap", d.Type)
				}
				if d.State != DeviceStateConnected {
					t.Errorf("State = %v, want Connected", d.State)
				}
				if d.Uptime.Int() != 86400 {
					t.Errorf("Uptime = %v, want 86400", d.Uptime.Int())
				}
				if d.NumSTA != 15 {
					t.Errorf("NumSTA = %v, want 15", d.NumSTA)
				}
				if len(d.RadioTable) != 1 {
					t.Errorf("RadioTable length = %v, want 1", len(d.RadioTable))
				}
			},
		},
		{
			name: "switch device with ports",
			json: `{
				"_id": "switch123",
				"mac": "aa:bb:cc:dd:ee:ff",
				"model": "USW-24-POE",
				"type": "usw",
				"name": "Core Switch",
				"adopted": true,
				"site_id": "default",
				"state": 1,
				"last_seen": 1642567890,
				"port_table": [
					{
						"port_idx": 1,
						"enable": true,
						"up": true,
						"speed": 1000,
						"full_duplex": true,
						"port_poe": true,
						"poe_enable": true
					}
				],
				"total_max_power": 400
			}`,
			wantErr: false,
			check: func(t *testing.T, d *Device) {
				if d.Type != "usw" {
					t.Errorf("Type = %v, want usw", d.Type)
				}
				if len(d.PortTable) != 1 {
					t.Errorf("PortTable length = %v, want 1", len(d.PortTable))
				}
				if d.PortTable[0].PortIdx != 1 {
					t.Errorf("PortIdx = %v, want 1", d.PortTable[0].PortIdx)
				}
				if !d.PortTable[0].PortPoe {
					t.Error("PortPoe should be true")
				}
				if d.TotalMaxPower != 400 {
					t.Errorf("TotalMaxPower = %v, want 400", d.TotalMaxPower)
				}
			},
		},
		{
			name: "UDM gateway device",
			json: `{
				"_id": "udm123",
				"mac": "11:22:33:44:55:66",
				"model": "UDMPRO",
				"type": "udm",
				"name": "Dream Machine Pro",
				"adopted": true,
				"site_id": "default",
				"state": 1,
				"wan_type": "dhcp",
				"speedtest_status": "idle",
				"storage": [
					{
						"mount_point": "/data",
						"name": "HDD",
						"size": "1000000000000",
						"type": "hdd",
						"used": "500000000000"
					}
				]
			}`,
			wantErr: false,
			check: func(t *testing.T, d *Device) {
				if d.Type != "udm" {
					t.Errorf("Type = %v, want udm", d.Type)
				}
				if d.WANType != "dhcp" {
					t.Errorf("WANType = %v, want dhcp", d.WANType)
				}
				if len(d.Storage) != 1 {
					t.Errorf("Storage length = %v, want 1", len(d.Storage))
				}
			},
		},
		{
			name: "device with flex types",
			json: `{
				"_id": "dev456",
				"mac": "00:00:00:00:00:00",
				"model": "test",
				"type": "uap",
				"state": 1,
				"uptime": "12345",
				"rx_bytes": "9876543210",
				"tx_bytes": 1234567890,
				"system-stats": {
					"cpu": "45.5",
					"mem": "60",
					"uptime": 86400
				}
			}`,
			wantErr: false,
			check: func(t *testing.T, d *Device) {
				if d.Uptime.Int() != 12345 {
					t.Errorf("Uptime = %v, want 12345", d.Uptime.Int())
				}
				if d.RxBytes.Int64() != 9876543210 {
					t.Errorf("RxBytes = %v, want 9876543210", d.RxBytes.Int64())
				}
				if d.TxBytes.Int64() != 1234567890 {
					t.Errorf("TxBytes = %v, want 1234567890", d.TxBytes.Int64())
				}
				if d.SystemStats != nil {
					if d.SystemStats.CPU.Float64() != 45.5 {
						t.Errorf("CPU = %v, want 45.5", d.SystemStats.CPU.Float64())
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Device
			err := json.Unmarshal([]byte(tt.json), &d)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, &d)
			}
		})
	}
}

func TestDevice_MarshalJSON(t *testing.T) {
	d := Device{
		ID:       "test123",
		MAC:      "aa:bb:cc:dd:ee:ff",
		Model:    "U6-Lite",
		Type:     "uap",
		Name:     "Test AP",
		State:    DeviceStateConnected,
		Adopted:  true,
		SiteID:   "default",
		LastSeen: 1642567890,
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var d2 Device
	if err := json.Unmarshal(data, &d2); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if d2.ID != d.ID {
		t.Errorf("ID = %v, want %v", d2.ID, d.ID)
	}
	if d2.State != d.State {
		t.Errorf("State = %v, want %v", d2.State, d.State)
	}
}

func TestDeviceBasic_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"mac": "00:11:22:33:44:55",
		"type": "uap",
		"model": "U6-Lite",
		"name": "Office AP",
		"state": 1
	}`

	var db DeviceBasic
	if err := json.Unmarshal([]byte(jsonData), &db); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if db.MAC != "00:11:22:33:44:55" {
		t.Errorf("MAC = %v, want 00:11:22:33:44:55", db.MAC)
	}
	if db.Type != "uap" {
		t.Errorf("Type = %v, want uap", db.Type)
	}
	if db.State != DeviceStateConnected {
		t.Errorf("State = %v, want Connected", db.State)
	}
}

func TestDeviceUplink_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"full_duplex": true,
		"ip": "192.168.1.10",
		"mac": "aa:bb:cc:dd:ee:ff",
		"name": "eth0",
		"speed": 1000,
		"up": true,
		"rx_bytes": "123456789",
		"tx_bytes": 987654321
	}`

	var uplink DeviceUplink
	if err := json.Unmarshal([]byte(jsonData), &uplink); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !uplink.FullDuplex {
		t.Error("FullDuplex should be true")
	}
	if uplink.Speed != 1000 {
		t.Errorf("Speed = %v, want 1000", uplink.Speed)
	}
	if uplink.RxBytes.Int64() != 123456789 {
		t.Errorf("RxBytes = %v, want 123456789", uplink.RxBytes.Int64())
	}
}

func TestRadioTable_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"radio": "ng",
		"name": "wifi0",
		"builtin_antenna": true,
		"max_txpower": 23,
		"nss": 2
	}`

	var radio RadioTable
	if err := json.Unmarshal([]byte(jsonData), &radio); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if radio.Radio != "ng" {
		t.Errorf("Radio = %v, want ng", radio.Radio)
	}
	if !radio.BuiltInAntenna {
		t.Error("BuiltInAntenna should be true")
	}
	if radio.MaxTXPower != 23 {
		t.Errorf("MaxTXPower = %v, want 23", radio.MaxTXPower)
	}
}

func TestRadioTableStats_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"radio": "ng",
		"name": "wifi0",
		"channel": 6,
		"tx_power": 20,
		"num_sta": 10,
		"satisfaction": 95,
		"state": "RUN",
		"tx_packets": "1000000",
		"rx_packets": 500000
	}`

	var stats RadioTableStats
	if err := json.Unmarshal([]byte(jsonData), &stats); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if stats.Channel != 6 {
		t.Errorf("Channel = %v, want 6", stats.Channel)
	}
	if stats.NumSTA != 10 {
		t.Errorf("NumSTA = %v, want 10", stats.NumSTA)
	}
	if stats.TXPackets.Int() != 1000000 {
		t.Errorf("TXPackets = %v, want 1000000", stats.TXPackets.Int())
	}
}

func TestPortTable_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"port_idx": 1,
		"enable": true,
		"up": true,
		"speed": 1000,
		"full_duplex": true,
		"port_poe": true,
		"poe_enable": true,
		"poe_power": "15.5",
		"poe_voltage": "53.2",
		"rx_bytes": 1234567890,
		"tx_bytes": "9876543210"
	}`

	var port PortTable
	if err := json.Unmarshal([]byte(jsonData), &port); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if port.PortIdx != 1 {
		t.Errorf("PortIdx = %v, want 1", port.PortIdx)
	}
	if !port.Enable {
		t.Error("Enable should be true")
	}
	if !port.PortPoe {
		t.Error("PortPoe should be true")
	}
	if port.PoePower.Float64() != 15.5 {
		t.Errorf("PoePower = %v, want 15.5", port.PoePower.Float64())
	}
	if port.RXBytes.Int64() != 1234567890 {
		t.Errorf("RXBytes = %v, want 1234567890", port.RXBytes.Int64())
	}
}

func TestVAPTable_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"bssid": "00:11:22:33:44:55",
		"essid": "MyNetwork",
		"name": "wlan0",
		"radio": "ng",
		"num_sta": 5,
		"up": true,
		"channel": 6,
		"tx_power": 20,
		"rx_bytes": "123456",
		"tx_bytes": 654321
	}`

	var vap VAPTable
	if err := json.Unmarshal([]byte(jsonData), &vap); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if vap.Essid != "MyNetwork" {
		t.Errorf("Essid = %v, want MyNetwork", vap.Essid)
	}
	if vap.NumSTA != 5 {
		t.Errorf("NumSTA = %v, want 5", vap.NumSTA)
	}
	if !vap.Up {
		t.Error("Up should be true")
	}
}

func TestTemperature_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"name": "CPU",
		"type": "cpu",
		"value": "65.5"
	}`

	var temp Temperature
	if err := json.Unmarshal([]byte(jsonData), &temp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if temp.Name != "CPU" {
		t.Errorf("Name = %v, want CPU", temp.Name)
	}
	if temp.Value.Float64() != 65.5 {
		t.Errorf("Value = %v, want 65.5", temp.Value.Float64())
	}
}

func TestStorage_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"mount_point": "/data",
		"name": "HDD",
		"size": "1000000000000",
		"type": "hdd",
		"used": 500000000000
	}`

	var storage Storage
	if err := json.Unmarshal([]byte(jsonData), &storage); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if storage.MountPoint != "/data" {
		t.Errorf("MountPoint = %v, want /data", storage.MountPoint)
	}
	if storage.Size.Int64() != 1000000000000 {
		t.Errorf("Size = %v, want 1000000000000", storage.Size.Int64())
	}
	if storage.Used.Int64() != 500000000000 {
		t.Errorf("Used = %v, want 500000000000", storage.Used.Int64())
	}
}

func TestWAN_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"enable": true,
		"full_duplex": true,
		"gateway": "192.168.1.1",
		"ip": "192.168.1.100",
		"mac": "aa:bb:cc:dd:ee:ff",
		"speed": 1000,
		"up": true,
		"rx_bytes": "9876543210",
		"tx_bytes": 1234567890
	}`

	var wan WAN
	if err := json.Unmarshal([]byte(jsonData), &wan); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !wan.Enable {
		t.Error("Enable should be true")
	}
	if wan.Speed != 1000 {
		t.Errorf("Speed = %v, want 1000", wan.Speed)
	}
	if wan.RXBytes.Int64() != 9876543210 {
		t.Errorf("RXBytes = %v, want 9876543210", wan.RXBytes.Int64())
	}
}
