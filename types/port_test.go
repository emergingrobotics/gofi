package types

import (
	"encoding/json"
	"testing"
)

func TestPortForward_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "pf123",
		"site_id": "default",
		"name": "SSH Forward",
		"enabled": true,
		"proto": "tcp",
		"src": "wan",
		"dst_port": "22",
		"fwd": "192.168.1.10",
		"fwd_port": "22"
	}`

	var pf PortForward
	if err := json.Unmarshal([]byte(jsonData), &pf); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if pf.Name != "SSH Forward" {
		t.Errorf("Name = %v, want SSH Forward", pf.Name)
	}
	if pf.Protocol != ProtocolTCP {
		t.Errorf("Protocol = %v, want tcp", pf.Protocol)
	}
	if pf.FwdIP != "192.168.1.10" {
		t.Errorf("FwdIP = %v, want 192.168.1.10", pf.FwdIP)
	}
}

func TestPortProfile_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "profile123",
		"site_id": "default",
		"name": "Trunk Port",
		"forward": "customize",
		"native_networkconf_id": "net1",
		"tagged_networkconf_ids": ["net2", "net3"],
		"poe_mode": "auto"
	}`

	var profile PortProfile
	if err := json.Unmarshal([]byte(jsonData), &profile); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if profile.Name != "Trunk Port" {
		t.Errorf("Name = %v, want Trunk Port", profile.Name)
	}
	if len(profile.TaggedNetworkConfIDs) != 2 {
		t.Errorf("TaggedNetworkConfIDs length = %v, want 2", len(profile.TaggedNetworkConfIDs))
	}
}
