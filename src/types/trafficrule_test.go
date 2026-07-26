package types

import (
	"encoding/json"
	"testing"
)

func TestTrafficRule_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "traffic123",
		"site_id": "default",
		"name": "Limit Guest Bandwidth",
		"enabled": true,
		"action": "LIMIT",
		"matching_target": "NETWORK",
		"network_ids": ["net123"],
		"bandwidth": {
			"download_enabled": true,
			"download_limit_kbps": "10000",
			"upload_enabled": true,
			"upload_limit_kbps": 5000
		}
	}`

	var rule TrafficRule
	if err := json.Unmarshal([]byte(jsonData), &rule); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if rule.Name != "Limit Guest Bandwidth" {
		t.Errorf("Name = %v, want Limit Guest Bandwidth", rule.Name)
	}
	if rule.Action != TrafficActionLimit {
		t.Errorf("Action = %v, want LIMIT", rule.Action)
	}
	if rule.Bandwidth == nil {
		t.Fatal("Bandwidth should not be nil")
	}
	if rule.Bandwidth.DownloadLimit.Int() != 10000 {
		t.Errorf("DownloadLimit = %v, want 10000", rule.Bandwidth.DownloadLimit.Int())
	}
}

func TestSchedule_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"mode": "TIME_RANGE",
		"date_start": "2025-01-01",
		"date_end": "2025-12-31",
		"days_of_week": ["MON", "TUE", "WED"],
		"repeat_on_days": true
	}`

	var schedule Schedule
	if err := json.Unmarshal([]byte(jsonData), &schedule); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if schedule.Mode != ScheduleModeTimeRange {
		t.Errorf("Mode = %v, want TIME_RANGE", schedule.Mode)
	}
	if len(schedule.DaysOfWeek) != 3 {
		t.Errorf("DaysOfWeek length = %v, want 3", len(schedule.DaysOfWeek))
	}
}

func TestTargetDevice_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"client_mac": "aa:bb:cc:dd:ee:ff",
		"type": "CLIENT"
	}`

	var target TargetDevice
	if err := json.Unmarshal([]byte(jsonData), &target); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if target.Type != MatchingTargetClient {
		t.Errorf("Type = %v, want CLIENT", target.Type)
	}
	if target.ClientMAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("ClientMAC = %v, want aa:bb:cc:dd:ee:ff", target.ClientMAC)
	}
}
