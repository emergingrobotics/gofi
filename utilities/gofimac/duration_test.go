package main

import (
	"testing"
	"time"
)

func TestParseDuration_SingleUnits(t *testing.T) {
	cases := []struct {
		input    string
		expected time.Duration
	}{
		{"45s", 45 * time.Second},
		{"30m", 30 * time.Minute},
		{"24h", 24 * time.Hour},
		{"7d", 7 * durationDay},
		{"2w", 2 * durationWeek},
		{"3mo", 3 * durationMonth},
	}
	for _, testCase := range cases {
		result, err := parseDuration(testCase.input)
		if err != nil {
			t.Errorf("parseDuration(%q) unexpected error: %v", testCase.input, err)
			continue
		}
		if result != testCase.expected {
			t.Errorf("parseDuration(%q) = %v, want %v", testCase.input, result, testCase.expected)
		}
	}
}

func TestParseDuration_Compound(t *testing.T) {
	result, err := parseDuration("1w2d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := durationWeek + 2*durationDay
	if result != expected {
		t.Errorf("parseDuration(\"1w2d\") = %v, want %v", result, expected)
	}
}

func TestParseDuration_MonthNotMinute(t *testing.T) {
	month, err := parseDuration("1mo")
	if err != nil {
		t.Fatalf("unexpected error for 1mo: %v", err)
	}
	if month != durationMonth {
		t.Errorf("parseDuration(\"1mo\") = %v, want %v", month, durationMonth)
	}

	minute, err := parseDuration("1m")
	if err != nil {
		t.Fatalf("unexpected error for 1m: %v", err)
	}
	if minute != time.Minute {
		t.Errorf("parseDuration(\"1m\") = %v, want %v", minute, time.Minute)
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	invalid := []string{"", "7", "d", "7x", "abc", "7dd7"}
	for _, input := range invalid {
		if _, err := parseDuration(input); err == nil {
			t.Errorf("parseDuration(%q) expected error, got nil", input)
		}
	}
}

func TestDurationToHours_RoundsUp(t *testing.T) {
	cases := []struct {
		duration time.Duration
		expected int
	}{
		{2 * time.Hour, 2},
		{90 * time.Minute, 2},
		{30 * time.Minute, 1},
		{durationDay, 24},
		{7 * durationDay, 168},
		{0, 0},
	}
	for _, testCase := range cases {
		result := durationToHours(testCase.duration)
		if result != testCase.expected {
			t.Errorf("durationToHours(%v) = %d, want %d", testCase.duration, result, testCase.expected)
		}
	}
}
