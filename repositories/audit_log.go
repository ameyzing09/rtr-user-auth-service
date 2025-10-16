package repositories

import (
	"context"
	"time"

	"rtr-user-auth-service/models"

	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	List(ctx context.Context, filters AuditLogFilters, page, pageSize int) ([]models.AuditLog, int64, error)
}

type AuditLogFilters struct {
	TenantID      *string
	ActorID       *string
	Action        *string
	Status        *models.AuditLogStatus
	StartTime     *time.Time
	EndTime       *time.Time
	ResourceType  *string
	ResourceID    *string
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditLogRepository) List(ctx context.Context, filters AuditLogFilters, page, pageSize int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AuditLog{})

	// Apply filters
	if filters.TenantID != nil {
		query = query.Where("target_tenant_id = ? OR actor_tenant_id = ?", *filters.TenantID, *filters.TenantID)
	}
	if filters.ActorID != nil {
		query = query.Where("actor_id = ?", *filters.ActorID)
	}
	if filters.Action != nil {
		query = query.Where("action = ?", *filters.Action)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.StartTime != nil {
		query = query.Where("timestamp >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("timestamp <= ?", *filters.EndTime)
	}
	if filters.ResourceType != nil {
		query = query.Where("target_resource_type = ?", *filters.ResourceType)
	}
	if filters.ResourceID != nil {
		query = query.Where("target_resource_id = ?", *filters.ResourceID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginate
	offset := (page - 1) * pageSize
	if err := query.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
