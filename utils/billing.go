package utils

import (
	"time"

	"rtr-user-auth-service/models"
)

// AddCycle adds a billing cycle to the given start time
func AddCycle(start time.Time, c models.BillingCycle) time.Time {
	if c == models.CycleAnnual {
		return start.AddDate(1, 0, 0)
	}
	return start.AddDate(0, 1, 0)
}
