package mock

import "net/http"

// Scenario defines a test scenario that modifies server behavior.
type Scenario interface {
	// Apply modifies the response based on the scenario.
	Apply(w http.ResponseWriter, r *http.Request) bool
}

// ErrorScenario simulates an error condition.
type ErrorScenario struct {
	Path       string // Path to match (empty = all paths)
	StatusCode int
	RC         string
	Message    string
}

// Apply implements Scenario.
func (e *ErrorScenario) Apply(w http.ResponseWriter, r *http.Request) bool {
	// If path is specified, only apply to matching paths
	if e.Path != "" && r.URL.Path != e.Path {
		return false
	}

	writeAPIError(w, e.StatusCode, e.RC, e.Message)
	return true
}

// Predefined scenarios
var (
	// ScenarioSessionExpired simulates a session expiration.
	ScenarioSessionExpired = &ErrorScenario{
		StatusCode: http.StatusUnauthorized,
		RC:         "error",
		Message:    "Session expired",
	}

	// ScenarioCSRFFailure simulates a CSRF token failure.
	ScenarioCSRFFailure = &ErrorScenario{
		StatusCode: http.StatusForbidden,
		RC:         "error_invalid_csrf_token",
		Message:    "Invalid CSRF token",
	}

	// ScenarioRateLimit simulates rate limiting.
	ScenarioRateLimit = &ErrorScenario{
		StatusCode: http.StatusTooManyRequests,
		RC:         "error",
		Message:    "Too many requests",
	}

	// ScenarioServerError simulates a server error.
	ScenarioServerError = &ErrorScenario{
		StatusCode: http.StatusInternalServerError,
		RC:         "error",
		Message:    "Internal server error",
	}
)
