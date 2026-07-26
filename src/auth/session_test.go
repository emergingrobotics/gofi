package auth

import (
	"testing"
	"time"
)

func TestSession_IsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		session *Session
		want    bool
	}{
		{
			"valid session",
			&Session{
				Token:     "test-token",
				ExpiresAt: now.Add(1 * time.Hour),
			},
			true,
		},
		{
			"expired session",
			&Session{
				Token:     "test-token",
				ExpiresAt: now.Add(-1 * time.Hour),
			},
			false,
		},
		{
			"no expiry time",
			&Session{
				Token: "test-token",
			},
			true,
		},
		{
			"no token",
			&Session{
				ExpiresAt: now.Add(1 * time.Hour),
			},
			false,
		},
		{
			"nil session",
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.session.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_NeedsRefresh(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		session *Session
		want    bool
	}{
		{
			"needs refresh (5 minutes left)",
			&Session{
				Token:     "test-token",
				ExpiresAt: now.Add(5 * time.Minute),
			},
			true,
		},
		{
			"doesn't need refresh (1 hour left)",
			&Session{
				Token:     "test-token",
				ExpiresAt: now.Add(1 * time.Hour),
			},
			false,
		},
		{
			"no expiry time",
			&Session{
				Token: "test-token",
			},
			false,
		},
		{
			"nil session",
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.session.NeedsRefresh()
			if got != tt.want {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_Age(t *testing.T) {
	now := time.Now()
	createdAt := now.Add(-30 * time.Minute)

	tests := []struct {
		name    string
		session *Session
		wantMin time.Duration
		wantMax time.Duration
	}{
		{
			"30 minutes old",
			&Session{
				Token:     "test-token",
				CreatedAt: createdAt,
			},
			29 * time.Minute,
			31 * time.Minute,
		},
		{
			"no created time",
			&Session{
				Token: "test-token",
			},
			0,
			0,
		},
		{
			"nil session",
			nil,
			0,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.session.Age()
			if tt.wantMax == 0 {
				if got != 0 {
					t.Errorf("Age() = %v, want 0", got)
				}
			} else {
				if got < tt.wantMin || got > tt.wantMax {
					t.Errorf("Age() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
				}
			}
		})
	}
}

func TestSession_TimeUntilExpiry(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		session *Session
		wantMin time.Duration
		wantMax time.Duration
	}{
		{
			"1 hour until expiry",
			&Session{
				Token:     "test-token",
				ExpiresAt: now.Add(1 * time.Hour),
			},
			59 * time.Minute,
			61 * time.Minute,
		},
		{
			"expired",
			&Session{
				Token:     "test-token",
				ExpiresAt: now.Add(-1 * time.Hour),
			},
			-61 * time.Minute,
			-59 * time.Minute,
		},
		{
			"no expiry time",
			&Session{
				Token: "test-token",
			},
			0,
			0,
		},
		{
			"nil session",
			nil,
			0,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.session.TimeUntilExpiry()
			if tt.wantMax == 0 && tt.wantMin == 0 {
				if got != 0 {
					t.Errorf("TimeUntilExpiry() = %v, want 0", got)
				}
			} else {
				if got < tt.wantMin || got > tt.wantMax {
					t.Errorf("TimeUntilExpiry() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
				}
			}
		})
	}
}
