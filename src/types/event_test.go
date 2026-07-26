package types

import (
	"encoding/json"
	"testing"
)

func TestEvent_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "event123",
		"time": 1642567890000,
		"datetime": "2025-01-15T10:00:00Z",
		"key": "EVT_AP_Connected",
		"msg": "AP connected",
		"site_id": "default",
		"subsystem": "wlan",
		"ap": "00:11:22:33:44:55",
		"ap_name": "Office AP",
		"is_admin": false
	}`

	var event Event
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if event.Key != EventAPConnected {
		t.Errorf("Key = %v, want EVT_AP_Connected", event.Key)
	}
	if event.Subsystem != "wlan" {
		t.Errorf("Subsystem = %v, want wlan", event.Subsystem)
	}
}

func TestAlarm_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "alarm123",
		"time": 1642567890000,
		"datetime": "2025-01-15T10:00:00Z",
		"key": "EVT_IPS_Alert",
		"msg": "IPS alert triggered",
		"site_id": "default",
		"subsystem": "ids",
		"archived": false,
		"handled": false
	}`

	var alarm Alarm
	if err := json.Unmarshal([]byte(jsonData), &alarm); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if alarm.Key != EventIPSAlert {
		t.Errorf("Key = %v, want EVT_IPS_Alert", alarm.Key)
	}
	if alarm.Archived {
		t.Error("Archived should be false")
	}
}
