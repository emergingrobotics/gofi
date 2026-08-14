package clients

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Duration units beyond what time.ParseDuration supports. UniFi history windows
// are naturally expressed in days, weeks, and months, so gofimac accepts them.
const (
	durationDay   = 24 * time.Hour
	durationWeek  = 7 * durationDay
	durationMonth = 30 * durationDay
)

// ParseDuration parses a compound duration string supporting s, m, h, d, w, and
// mo units, because the standard time.ParseDuration lacks day/week/month units.
// Examples: "30m", "24h", "7d", "2w", "3mo", "1w2d".
func ParseDuration(input string) (time.Duration, error) {
	if input == "" {
		return 0, fmt.Errorf("empty duration")
	}

	var total time.Duration
	position := 0
	for position < len(input) {
		numberStart := position
		for position < len(input) && input[position] >= '0' && input[position] <= '9' {
			position++
		}
		if position == numberStart {
			return 0, fmt.Errorf("invalid duration %q: expected a number", input)
		}
		value, err := strconv.Atoi(input[numberStart:position])
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", input, err)
		}

		if position >= len(input) {
			return 0, fmt.Errorf("invalid duration %q: number %d has no unit", input, value)
		}

		// "mo" must be checked before the single-character "m" (minutes).
		var unit time.Duration
		switch {
		case strings.HasPrefix(input[position:], "mo"):
			unit = durationMonth
			position += 2
		case input[position] == 's':
			unit = time.Second
			position++
		case input[position] == 'm':
			unit = time.Minute
			position++
		case input[position] == 'h':
			unit = time.Hour
			position++
		case input[position] == 'd':
			unit = durationDay
			position++
		case input[position] == 'w':
			unit = durationWeek
			position++
		default:
			return 0, fmt.Errorf("invalid duration %q: unknown unit %q", input, string(input[position]))
		}

		total += time.Duration(value) * unit
	}

	return total, nil
}

// DurationToHours converts a duration to whole hours, rounding up, for the UDM
// ListAll "within" parameter which is expressed in hours.
func DurationToHours(duration time.Duration) int {
	hours := int(duration / time.Hour)
	if duration%time.Hour != 0 {
		hours++
	}
	return hours
}
