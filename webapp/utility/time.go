package utility

import (
	"fmt"
	"strings"
	"time"
)

// FormatDurationCoarse returns a pretty printed duration with coarse granularity.
func FormatDurationCoarse(duration time.Duration) string {
	// Negative durations (e.g. future dates) should work too.
	if duration < 0 {
		duration *= -1
	}

	var result = func(text string, v int64) string {
		if v == 1 {
			text = strings.TrimSuffix(text, "s")
		}
		return fmt.Sprintf(text, v)
	}

	if duration.Seconds() < 60.0 {
		return result("%d seconds", int64(duration.Seconds()))
	}

	if duration.Minutes() < 60.0 {
		return result("%d minutes", int64(duration.Minutes()))
	}

	if duration.Hours() < 24.0 {
		return result("%d hours", int64(duration.Hours()))
	}

	days := int64(duration.Hours() / 24)
	if days < 30 {
		return result("%d days", days)
	}

	months := int64(days / 30)
	if months < 12 {
		return result("%d months", months)
	}

	years := int64(days / 365)
	return result("%d years", years)
}
