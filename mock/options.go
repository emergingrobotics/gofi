package mock

// Option configures a mock server.
type Option func(*Server)

// WithoutAuth disables authentication checks.
func WithoutAuth() Option {
	return func(s *Server) {
		s.requireAuth = false
	}
}

// WithoutCSRF disables CSRF token checks.
func WithoutCSRF() Option {
	return func(s *Server) {
		s.requireCSRF = false
	}
}

// WithFixtures loads fixtures into the server state.
func WithFixtures(fixtures *Fixtures) Option {
	return func(s *Server) {
		if fixtures != nil {
			s.state.LoadFixtures(fixtures)
		}
	}
}

// WithScenario applies a test scenario to the server.
func WithScenario(scenario Scenario) Option {
	return func(s *Server) {
		s.scenario = scenario
	}
}
