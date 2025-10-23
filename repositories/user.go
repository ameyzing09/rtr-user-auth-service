package repositories

import (
	"context"
	"rtr-user-auth-service/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	EmailExists(ctx context.Context, tenantID, email string) (bool, error)
	Create(ctx context.Context, u *models.User) error
	FindByEmail(ctx context.Context, tenantID, email string) (*models.User, error)
	FindByID(ctx context.Context, tenantID, userID string) (*models.User, error)
	ListByTenant(ctx context.Context, tenantID string) ([]models.User, error)
	UpdatePassword(ctx context.Context, tenantID, userID, hashedPassword string, forcePasswordReset *bool) error
}

type GormUserRepo struct {
	db *gorm.DB
}

func NewGormUserRepo(db *gorm.DB) *GormUserRepo {
	return &GormUserRepo{db: db}
}

func (r *GormUserRepo) EmailExists(ctx context.Context, tenantID, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("tenant_id = ? AND email = ? AND is_active = ?", tenantID, email, true).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormUserRepo) Create(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *GormUserRepo) FindByEmail(ctx context.Context, tenantID, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND email = ? AND is_active = ?", tenantID, email, true).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) FindByID(ctx context.Context, tenantID, userID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND is_active = ?", tenantID, userID, true).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) ListByTenant(ctx context.Context, tenantID string) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *GormUserRepo) UpdatePassword(ctx context.Context, tenantID, userID, newHashedPassword string, forcePasswordReset *bool) error {
	q := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND tenant_id = ?", userID, tenantID)

	// If forcePasswordReset is nil, only update password
	if forcePasswordReset == nil {
		return q.Update("password", newHashedPassword).Error
	}

	// If forcePasswordReset is provided, update both password and flag
	return q.Updates(map[string]interface{}{
		"password":             newHashedPassword,
		"force_password_reset": *forcePasswordReset,
	}).Error
}
