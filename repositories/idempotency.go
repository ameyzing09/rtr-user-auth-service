package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"rtr-user-auth-service/models"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GormIdempotencyRepo struct {
	db *gorm.DB
}

func NewGormIdempotencyRepo(db *gorm.DB) *GormIdempotencyRepo {
	return &GormIdempotencyRepo{db: db}
}

func (r *GormIdempotencyRepo) UpsertAndGet(ctx context.Context, keyHash, requestHash string) (*models.IdempotencyKey, error) {
	record := &models.IdempotencyKey{
		KeyHash:     keyHash,
		RequestHash: requestHash,
		Status:      models.IdempotencyStatusError,
	}

	err := r.db.WithContext(ctx).Create(record).Error
	if err == nil {
		return record, nil
	}

	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		var existing models.IdempotencyKey
		if err := r.db.WithContext(ctx).
			Where("key_hash = ?", keyHash).
			First(&existing).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}

	return nil, err
}

func (r *GormIdempotencyRepo) SaveResult(ctx context.Context, keyHash string, status models.IdempotencyStatus, response map[string]interface{}) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if response != nil {
		encoded, err := json.Marshal(response)
		if err != nil {
			return err
		}
		updates["response"] = datatypes.JSON(encoded)
	} else {
		updates["response"] = gorm.Expr("NULL")
	}

	return r.db.WithContext(ctx).
		Model(&models.IdempotencyKey{}).
		Where("key_hash = ?", keyHash).
		Updates(updates).Error
}
