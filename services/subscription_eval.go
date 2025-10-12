package services

import (
	"time"

	"rtr-user-auth-service/models"
)

// EffectiveStatus computes the effective subscription status at the given time
// If status is TRIAL and trial_ends_at < now, returns SUSPENDED
func EffectiveStatus(sub *models.Subscription, now time.Time) models.SubscriptionStatus {
	if sub == nil {
		return models.SubSuspended
	}
	if sub.Status == models.SubTrial && sub.TrialEndsAt != nil && now.After(*sub.TrialEndsAt) {
		return models.SubSuspended
	}
	return sub.Status
}
