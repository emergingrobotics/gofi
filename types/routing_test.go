package types

import (
	"encoding/json"
	"testing"
)

func TestRoute_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"_id": "route123",
		"site_id": "default",
		"name": "Private Route",
		"enabled": true,
		"type": "nexthop-route",
		"static-route_network": "10.10.0.0/16",
		"static-route_nexthop": "192.168.1.1",
		"static-route_distance": 1
	}`

	var route Route
	if err := json.Unmarshal([]byte(jsonData), &route); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if route.Name != "Private Route" {
		t.Errorf("Name = %v, want Private Route", route.Name)
	}
	if route.Type != RouteTypeNexthop {
		t.Errorf("Type = %v, want nexthop-route", route.Type)
	}
	if !route.Enabled {
		t.Error("Enabled should be true")
	}
}
