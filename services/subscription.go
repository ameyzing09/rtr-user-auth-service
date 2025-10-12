package services

import (
	"context"
	"errors"
	"time"

	"rtr-user-auth-service/domain"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"
	"rtr-user-auth-service/utils"

	"gorm.io/gorm"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, tenantID string, plan models.Plan, isTrial bool, updatedBy string) (*models.Subscription, error)
	GetSubscription(ctx context.Context, tenantID string) (*models.Subscription, error)
	UpdateSubscription(ctx context.Context, sub *models.Subscription) error
	ActivateSubscription(ctx context.Context, tenantID string, billingCycle models.BillingCycle, amountCents uint32, updatedBy string) error
	SuspendSubscription(ctx context.Context, tenantID string, updatedBy string) error
	ResumeSubscription(ctx context.Context, tenantID string, updatedBy string) error
	CancelSubscription(ctx context.Context, tenantID string, updatedBy string) error
	DeleteSubscription(ctx context.Context, tenantID string) error
}

type subscriptionService struct {
	repo repositories.SubscriptionRepository
}

func NewSubscriptionService(repo repositories.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{repo: repo}
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, tenantID string, plan models.Plan, isTrial bool, updatedBy string) (*models.Subscription, error) {
	now := time.Now().UTC()

	sub := &models.Subscription{
		TenantID:     tenantID,
		Plan:         plan,
		BillingCycle: models.CycleMonthly, // Default to monthly
		Currency:     "USD",
		AmountCents:  0,
		UpdatedBy:    &updatedBy,
	}

	if isTrial {
		sub.Status = models.SubTrial
		trialEndsAt := now.Add(utils.TrialDuration())
		sub.TrialEndsAt = &trialEndsAt
	} else {
		sub.Status = models.SubActive
		sub.PeriodStart = &now
		periodEnd := utils.AddCycle(now, models.CycleMonthly)
		sub.PeriodEnd = &periodEnd
		sub.NextRenewalAt = &periodEnd
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) GetSubscription(ctx context.Context, tenantID string) (*models.Subscription, error) {
	sub, err := s.repo.FindByTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, err
	}
	return sub, nil
}

func (s *subscriptionService) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	return s.repo.Update(ctx, sub)
}

func (s *subscriptionService) ActivateSubscription(ctx context.Context, tenantID string, billingCycle models.BillingCycle, amountCents uint32, updatedBy string) error {
	sub, err := s.repo.FindByTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	sub.Status = models.SubActive
	sub.BillingCycle = billingCycle
	sub.AmountCents = amountCents
	sub.PeriodStart = &now
	periodEnd := utils.AddCycle(now, billingCycle)
	sub.PeriodEnd = &periodEnd
	sub.NextRenewalAt = &periodEnd
	sub.TrialEndsAt = nil // Clear trial end date
	sub.UpdatedBy = &updatedBy

	return s.repo.Update(ctx, sub)
}

func (s *subscriptionService) SuspendSubscription(ctx context.Context, tenantID string, updatedBy string) error {
	sub, err := s.repo.FindByTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	sub.Status = models.SubSuspended
	sub.UpdatedBy = &updatedBy

	return s.repo.Update(ctx, sub)
}

func (s *subscriptionService) ResumeSubscription(ctx context.Context, tenantID string, updatedBy string) error {
	sub, err := s.repo.FindByTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	sub.Status = models.SubActive
	sub.UpdatedBy = &updatedBy

	return s.repo.Update(ctx, sub)
}

func (s *subscriptionService) CancelSubscription(ctx context.Context, tenantID string, updatedBy string) error {
	sub, err := s.repo.FindByTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	sub.Status = models.SubCanceled
	sub.CanceledAt = &now
	sub.UpdatedBy = &updatedBy

	return s.repo.Update(ctx, sub)
}

func (s *subscriptionService) DeleteSubscription(ctx context.Context, tenantID string) error {
	return s.repo.Delete(ctx, tenantID)
}
