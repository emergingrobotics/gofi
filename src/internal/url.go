package internal

import (
	"fmt"
	"path"
)

const (
	// BasePath is the UniFi Network Application base path on UDM Pro
	BasePath = "/proxy/network"

	// APIv1Base is the base path for v1 API endpoints
	APIv1Base = "/api"

	// APIv2Base is the base path for v2 API endpoints
	APIv2Base = "/v2/api"
)

// BuildAPIPath builds a v1 API path for a given site and endpoint.
// Example: BuildAPIPath("default", "stat/device") -> "/proxy/network/api/s/default/stat/device"
func BuildAPIPath(site, endpoint string) string {
	if site == "" {
		site = "default"
	}
	return path.Join(BasePath, APIv1Base, "s", site, endpoint)
}

// BuildV2APIPath builds a v2 API path for a given site and endpoint.
// Example: BuildV2APIPath("default", "site/default/trafficrules") -> "/proxy/network/v2/api/site/default/trafficrules"
func BuildV2APIPath(_, endpoint string) string {
	// v2 API uses full path including site in endpoint
	return path.Join(BasePath, APIv2Base, endpoint)
}

// BuildRESTPath builds a REST API path for a given site, resource, and optional ID.
// Example: BuildRESTPath("default", "networkconf", "abc123") -> "/proxy/network/api/s/default/rest/networkconf/abc123"
func BuildRESTPath(site, resource, id string) string {
	if site == "" {
		site = "default"
	}

	basePath := path.Join(BasePath, APIv1Base, "s", site, "rest", resource)

	if id != "" {
		return path.Join(basePath, id)
	}

	return basePath
}

// BuildCmdPath builds a command API path for a given site and manager.
// Example: BuildCmdPath("default", "device") -> "/proxy/network/api/s/default/cmd/devmgr"
func BuildCmdPath(site, manager string) string {
	if site == "" {
		site = "default"
	}

	// Manager names follow pattern: devmgr, stamgr, sitemgr
	if manager != "" && manager[len(manager)-3:] != "mgr" {
		manager = manager + "mgr"
	}

	return path.Join(BasePath, APIv1Base, "s", site, "cmd", manager)
}

// BuildAuthPath builds an authentication API path.
// Example: BuildAuthPath("login") -> "/api/auth/login"
func BuildAuthPath(endpoint string) string {
	return fmt.Sprintf("/api/auth/%s", endpoint)
}

// BuildSystemPath builds a system API path (no site required).
// Example: BuildSystemPath("status") -> "/api/status"
func BuildSystemPath(endpoint string) string {
	return fmt.Sprintf("/api/%s", endpoint)
}

// BuildWebSocketPath builds a WebSocket path for event streaming.
// Example: BuildWebSocketPath("default") -> "/proxy/network/wss/s/default/events"
func BuildWebSocketPath(site string) string {
	if site == "" {
		site = "default"
	}
	return path.Join(BasePath, "wss", "s", site, "events")
}
