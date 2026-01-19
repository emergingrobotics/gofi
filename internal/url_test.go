package internal

import "testing"

func TestBuildAPIPath(t *testing.T) {
	tests := []struct {
		name     string
		site     string
		endpoint string
		want     string
	}{
		{"default site", "default", "stat/device", "/proxy/network/api/s/default/stat/device"},
		{"custom site", "mysite", "stat/device", "/proxy/network/api/s/mysite/stat/device"},
		{"empty site defaults", "", "stat/device", "/proxy/network/api/s/default/stat/device"},
		{"complex endpoint", "default", "rest/networkconf", "/proxy/network/api/s/default/rest/networkconf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildAPIPath(tt.site, tt.endpoint)
			if got != tt.want {
				t.Errorf("BuildAPIPath(%q, %q) = %q, want %q", tt.site, tt.endpoint, got, tt.want)
			}
		})
	}
}

func TestBuildV2APIPath(t *testing.T) {
	tests := []struct {
		name     string
		site     string
		endpoint string
		want     string
	}{
		{"traffic rules", "default", "site/default/trafficrules", "/proxy/network/v2/api/site/default/trafficrules"},
		{"custom site", "mysite", "site/mysite/trafficrules", "/proxy/network/v2/api/site/mysite/trafficrules"},
		{"empty site", "", "site/default/notifications", "/proxy/network/v2/api/site/default/notifications"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildV2APIPath(tt.site, tt.endpoint)
			if got != tt.want {
				t.Errorf("BuildV2APIPath(%q, %q) = %q, want %q", tt.site, tt.endpoint, got, tt.want)
			}
		})
	}
}

func TestBuildRESTPath(t *testing.T) {
	tests := []struct {
		name     string
		site     string
		resource string
		id       string
		want     string
	}{
		{"list networks", "default", "networkconf", "", "/proxy/network/api/s/default/rest/networkconf"},
		{"get network", "default", "networkconf", "abc123", "/proxy/network/api/s/default/rest/networkconf/abc123"},
		{"custom site", "mysite", "wlanconf", "xyz789", "/proxy/network/api/s/mysite/rest/wlanconf/xyz789"},
		{"empty site", "", "portconf", "port1", "/proxy/network/api/s/default/rest/portconf/port1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildRESTPath(tt.site, tt.resource, tt.id)
			if got != tt.want {
				t.Errorf("BuildRESTPath(%q, %q, %q) = %q, want %q", tt.site, tt.resource, tt.id, got, tt.want)
			}
		})
	}
}

func TestBuildCmdPath(t *testing.T) {
	tests := []struct {
		name    string
		site    string
		manager string
		want    string
	}{
		{"device manager", "default", "devmgr", "/proxy/network/api/s/default/cmd/devmgr"},
		{"station manager", "default", "stamgr", "/proxy/network/api/s/default/cmd/stamgr"},
		{"without mgr suffix", "default", "dev", "/proxy/network/api/s/default/cmd/devmgr"},
		{"custom site", "mysite", "sitemgr", "/proxy/network/api/s/mysite/cmd/sitemgr"},
		{"empty site", "", "devmgr", "/proxy/network/api/s/default/cmd/devmgr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildCmdPath(tt.site, tt.manager)
			if got != tt.want {
				t.Errorf("BuildCmdPath(%q, %q) = %q, want %q", tt.site, tt.manager, got, tt.want)
			}
		})
	}
}

func TestBuildAuthPath(t *testing.T) {
	tests := []struct {
		endpoint string
		want     string
	}{
		{"login", "/api/auth/login"},
		{"logout", "/api/auth/logout"},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			got := BuildAuthPath(tt.endpoint)
			if got != tt.want {
				t.Errorf("BuildAuthPath(%q) = %q, want %q", tt.endpoint, got, tt.want)
			}
		})
	}
}

func TestBuildSystemPath(t *testing.T) {
	tests := []struct {
		endpoint string
		want     string
	}{
		{"status", "/api/status"},
		{"self", "/api/self"},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			got := BuildSystemPath(tt.endpoint)
			if got != tt.want {
				t.Errorf("BuildSystemPath(%q) = %q, want %q", tt.endpoint, got, tt.want)
			}
		})
	}
}

func TestBuildWebSocketPath(t *testing.T) {
	tests := []struct {
		name string
		site string
		want string
	}{
		{"default site", "default", "/proxy/network/wss/s/default/events"},
		{"custom site", "mysite", "/proxy/network/wss/s/mysite/events"},
		{"empty site", "", "/proxy/network/wss/s/default/events"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildWebSocketPath(tt.site)
			if got != tt.want {
				t.Errorf("BuildWebSocketPath(%q) = %q, want %q", tt.site, got, tt.want)
			}
		})
	}
}
