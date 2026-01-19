package types

// Route represents a static route configuration.
type Route struct {
	ID                      string  `json:"_id,omitempty"`
	SiteID                  string  `json:"site_id,omitempty"`
	Name                    string  `json:"name"`
	Enabled                 bool    `json:"enabled"`
	Type                    string  `json:"type"` // "nexthop-route", "blackhole"
	StaticRouteDistance     int     `json:"static-route_distance,omitempty"`
	StaticRouteInterface    string  `json:"static-route_interface,omitempty"`
	StaticRouteNexthop      string  `json:"static-route_nexthop,omitempty"`
	StaticRouteNetwork      string  `json:"static-route_network"`
	StaticRouteType         string  `json:"static-route_type,omitempty"`
	GatewayType             string  `json:"gateway_type,omitempty"`
	GatewayDevice           string  `json:"gateway_device,omitempty"`
	PfRule                  string  `json:"pfrule,omitempty"`
}

// Route type constants.
const (
	RouteTypeNexthop  = "nexthop-route"
	RouteTypeBlackhole = "blackhole"
)
