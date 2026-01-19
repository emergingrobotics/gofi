package types

import (
	"encoding/json"
	"testing"
)

func TestClient_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "client123",
		"site_id": "default",
		"mac": "aa:bb:cc:dd:ee:ff",
		"hostname": "laptop",
		"name": "My Laptop",
		"ip": "192.168.1.100",
		"is_guest": false,
		"is_wired": false,
		"first_seen": 1642567890,
		"last_seen": 1642654290,
		"uptime": "86400",
		"ap_mac": "00:11:22:33:44:55",
		"essid": "MyWiFi",
		"channel": 6,
		"radio": "ng",
		"rx_bytes": "1234567890",
		"tx_bytes": 9876543210,
		"signal": "-45",
		"satisfaction": 95,
		"authorized": true,
		"blocked": false
	}`

	var client Client
	if err := json.Unmarshal([]byte(jsonData), &client); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if client.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("MAC = %v, want aa:bb:cc:dd:ee:ff", client.MAC)
	}
	if client.Hostname != "laptop" {
		t.Errorf("Hostname = %v, want laptop", client.Hostname)
	}
	if client.RXBytes.Int64() != 1234567890 {
		t.Errorf("RXBytes = %v, want 1234567890", client.RXBytes.Int64())
	}
	if client.Signal.Int() != -45 {
		t.Errorf("Signal = %v, want -45", client.Signal.Int())
	}
}
