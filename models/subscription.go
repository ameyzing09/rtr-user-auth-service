package models

import "time"

type SubscriptionStatus string

const (
	SubTrial     SubscriptionStatus = "TRIAL"
	SubActive    SubscriptionStatus = "ACTIVE"
	SubGrace     SubscriptionStatus = "GRACE"
	SubSuspended SubscriptionStatus = "SUSPENDED"
	SubCanceled  SubscriptionStatus = "CANCELED"
)

type BillingCycle string

const (
	CycleMonthly BillingCycle = "MONTHLY"
	CycleAnnual  BillingCycle = "ANNUAL"
)

type Subscription struct {
	ID            uint64             `gorm:"primaryKey;autoIncrement"`
	TenantID      string             `gorm:"type:char(36);uniqueIndex;not null"`
	Plan          Plan               `gorm:"type:ENUM('BASIC','STARTER','GROWTH','ENTERPRISE','ON_PREM');not null"`
	BillingCycle  BillingCycle       `gorm:"type:ENUM('MONTHLY','ANNUAL');not null;default:'MONTHLY'"`
	Status        SubscriptionStatus `gorm:"type:ENUM('TRIAL','ACTIVE','GRACE','SUSPENDED','CANCELED');not null;default:'TRIAL'"`
	Currency      string             `gorm:"type:char(3);not null;default:'USD'"`
	AmountCents   uint32             `gorm:"not null;default:0"`
	PeriodStart   *time.Time
	PeriodEnd     *time.Time
	TrialEndsAt   *time.Time
	NextRenewalAt *time.Time
	CanceledAt    *time.Time
	UpdatedBy     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Subscription) TableName() string {
	return "subscriptions"
}
