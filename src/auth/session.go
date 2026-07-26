package auth

import "time"

// Session represents an authenticated session.
type Session struct {
	// Token is the session token/cookie value.
	Token string

	// CSRFToken is the CSRF token for the session.
	CSRFToken string

	// ExpiresAt is the time when the session expires.
	ExpiresAt time.Time

	// Username is the authenticated username.
	Username string

	// CreatedAt is when the session was created.
	CreatedAt time.Time
}

// IsValid returns true if the session is still valid.
func (s *Session) IsValid() bool {
	if s == nil {
		return false
	}

	// Check if session has expired
	if !s.ExpiresAt.IsZero() && time.Now().After(s.ExpiresAt) {
		return false
	}

	// Session must have a token
	return s.Token != ""
}

// NeedsRefresh returns true if the session is getting close to expiration.
// Returns true if less than 10 minutes remaining.
func (s *Session) NeedsRefresh() bool {
	if s == nil || s.ExpiresAt.IsZero() {
		return false
	}

	// Refresh if less than 10 minutes remaining
	refreshThreshold := time.Now().Add(10 * time.Minute)
	return s.ExpiresAt.Before(refreshThreshold)
}

// Age returns how long the session has been active.
func (s *Session) Age() time.Duration {
	if s == nil || s.CreatedAt.IsZero() {
		return 0
	}
	return time.Since(s.CreatedAt)
}

// TimeUntilExpiry returns how much time is left before the session expires.
func (s *Session) TimeUntilExpiry() time.Duration {
	if s == nil || s.ExpiresAt.IsZero() {
		return 0
	}
	return time.Until(s.ExpiresAt)
}
