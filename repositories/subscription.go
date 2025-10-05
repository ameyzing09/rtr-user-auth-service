package repositories

import (
	"context"
	"rtr-user-auth-service/models"

	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *models.Subscription) error
	FindByTenant(ctx context.Context, tenantID string) (*models.Subscription, error)
	Update(ctx context.Context, s *models.Subscription) error
	Delete(ctx context.Context, tenantID string) error
}

type subscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, s *models.Subscription) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *subscriptionRepo) FindByTenant(ctx context.Context, tenantID string) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepo) Update(ctx context.Context, s *models.Subscription) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *subscriptionRepo) Delete(ctx context.Context, tenantID string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&models.Subscription{}).Error
}
