package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// refreshTokenRepository implements RefreshTokenRepository interface
type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, tenantID uuid.UUID, token string) (*entities.RefreshToken, error) {
	var refreshToken entities.RefreshToken
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("token = ? AND tenant_id = ?", token, tenantID).
		First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	var tokens []*entities.RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND is_revoked = ?", userID, tenantID, false).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

func (r *refreshTokenRepository) Update(ctx context.Context, token *entities.RefreshToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.RefreshToken{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Delete(&entities.RefreshToken{}, "expires_at < ?", time.Now()).Error
}