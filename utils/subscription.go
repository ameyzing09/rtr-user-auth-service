package utils

import (
	"os"
	"strconv"
	"time"
)

// TrialDuration returns the trial duration from environment variable TRIAL_DAYS
// Defaults to 14 days, clamped between 1-60 days
func TrialDuration() time.Duration {
	v := os.Getenv("TRIAL_DAYS")
	if v == "" {
		return 14 * 24 * time.Hour
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 || n > 60 {
		return 14 * 24 * time.Hour
	}
	return time.Duration(n) * 24 * time.Hour
}
