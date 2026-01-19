package types

import (
	"encoding/json"
	"testing"
)

func TestAPIResponse_Unmarshal(t *testing.T) {
	input := `{
		"meta": {
			"rc": "ok",
			"count": 2
		},
		"data": [
			{"id": "1", "name": "test1"},
			{"id": "2", "name": "test2"}
		]
	}`

	type TestData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	var resp APIResponse[TestData]
	if err := json.Unmarshal([]byte(input), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.Meta.RC != "ok" {
		t.Errorf("Meta.RC = %s, want ok", resp.Meta.RC)
	}

	if resp.Meta.Count != 2 {
		t.Errorf("Meta.Count = %d, want 2", resp.Meta.Count)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(resp.Data))
	}

	if resp.Data[0].Name != "test1" {
		t.Errorf("Data[0].Name = %s, want test1", resp.Data[0].Name)
	}
}

func TestMAC_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mac     MAC
		wantErr bool
	}{
		{"valid colon", "aa:bb:cc:dd:ee:ff", false},
		{"valid dash", "aa-bb-cc-dd-ee-ff", false},
		{"valid no separator", "aabbccddeeff", false},
		{"valid uppercase", "AA:BB:CC:DD:EE:FF", false},
		{"empty", "", true},
		{"too short", "aa:bb:cc", true},
		{"invalid chars", "zz:bb:cc:dd:ee:ff", true},
		{"too long", "aa:bb:cc:dd:ee:ff:11", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mac.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeviceState_String(t *testing.T) {
	tests := []struct {
		state DeviceState
		want  string
	}{
		{DeviceStateOffline, "offline"},
		{DeviceStateConnected, "connected"},
		{DeviceStatePending, "pending"},
		{DeviceStateDisconnected, "disconnected"},
		{DeviceStateFirmware, "firmware"},
		{DeviceStateProvisioning, "provisioning"},
		{DeviceStateHeartbeat, "heartbeat"},
		{DeviceStateAdopting, "adopting"},
		{DeviceStateDeleting, "deleting"},
		{DeviceStateInformed, "informed"},
		{DeviceStateUpgrading, "upgrading"},
		{DeviceState(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDeviceState_JSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  DeviceState
	}{
		{"offline", `0`, DeviceStateOffline},
		{"connected", `1`, DeviceStateConnected},
		{"upgrading", `10`, DeviceStateUpgrading},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var state DeviceState
			if err := json.Unmarshal([]byte(tt.input), &state); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if state != tt.want {
				t.Errorf("Unmarshal = %v, want %v", state, tt.want)
			}

			// Test marshaling
			data, err := json.Marshal(state)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}
			if string(data) != tt.input {
				t.Errorf("Marshal = %s, want %s", string(data), tt.input)
			}
		})
	}
}
