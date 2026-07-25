// Package conn resolves UniFi connection configuration from flags and the
// environment, shared by the gofips/gofimac/gofinet CLIs.
package conn

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/unifi-go/gofi"
)

// Environment variable names used by all three CLIs.
const (
	EnvUsername     = "UNIFI_USERNAME"
	EnvPassword     = "UNIFI_PASSWORD"
	EnvControllerIP = "UNIFI_CONTROLLER_IP"
	EnvAPIKey       = "UNIFI_API_KEY"
	EnvConsoleID    = "UNIFI_CONSOLE_ID"
)

// ResolveConfig builds a *gofi.Config from the given flag values and the
// environment. When UNIFI_API_KEY is set it uses connector mode (requiring
// UNIFI_CONSOLE_ID) and ignores username/password, writing a note to w if
// those were also set. Otherwise it uses local username/password auth,
// falling back to UNIFI_CONTROLLER_IP for the host.
func ResolveConfig(w io.Writer, host string, port int, site string, secure bool) (*gofi.Config, error) {
	apiKey := os.Getenv(EnvAPIKey)
	consoleID := os.Getenv(EnvConsoleID)

	if apiKey != "" {
		if os.Getenv(EnvUsername) != "" || os.Getenv(EnvPassword) != "" {
			fmt.Fprintf(w, "Note: %s is set; ignoring %s/%s.\n", EnvAPIKey, EnvUsername, EnvPassword)
		}
		if consoleID == "" {
			return nil, fmt.Errorf("%s is required when %s is set (connector mode)", EnvConsoleID, EnvAPIKey)
		}
		return &gofi.Config{
			APIKey:    apiKey,
			ConsoleID: consoleID,
			Site:      site,
			// api.ui.com has a valid CA-signed certificate; always verify TLS
			// here regardless of -k/--secure, which only accommodates
			// self-signed certs on LAN UDMs in the local-auth branch below.
			SkipTLSVerify: false,
			// The cloud connector throttles bulk writes with HTTP 429; retry
			// with generous backoff (honoring Retry-After) so a large --set
			// converges instead of erroring out.
			RetryConfig: &gofi.RetryConfig{
				MaxRetries:     8,
				InitialBackoff: 1 * time.Second,
				MaxBackoff:     30 * time.Second,
			},
		}, nil
	}

	username := os.Getenv(EnvUsername)
	password := os.Getenv(EnvPassword)
	if username == "" && password == "" {
		return nil, fmt.Errorf("no credentials: set %s (+ %s), or %s/%s", EnvAPIKey, EnvConsoleID, EnvUsername, EnvPassword)
	}
	if username == "" {
		return nil, fmt.Errorf("%s environment variable is required", EnvUsername)
	}
	if password == "" {
		return nil, fmt.Errorf("%s environment variable is required", EnvPassword)
	}
	if host == "" {
		host = os.Getenv(EnvControllerIP)
	}
	if host == "" {
		return nil, fmt.Errorf("--host is required (or set %s)", EnvControllerIP)
	}
	return &gofi.Config{
		Host:          host,
		Port:          port,
		Username:      username,
		Password:      password,
		Site:          site,
		SkipTLSVerify: !secure,
		// Modest retry for transient 5xx/429 on the local controller.
		RetryConfig: &gofi.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 200 * time.Millisecond,
			MaxBackoff:     5 * time.Second,
		},
	}, nil
}
