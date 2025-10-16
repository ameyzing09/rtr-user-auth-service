package services

import (
	"context"
	"time"

	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"

	"github.com/google/uuid"
)

type AuditLogService interface {
	// Log creates an audit log entry (async-safe)
	Log(ctx context.Context, entry AuditLogEntry) error

	// Query retrieves audit logs with filters
	Query(ctx context.Context, filters repositories.AuditLogFilters, page, pageSize int) ([]models.AuditLog, int64, error)
}

type AuditLogEntry struct {
	Action             string
	ActorID            *string
	ActorTenantID      *string
	ActorRole          *string
	TargetResourceID   *string
	TargetResourceType *string
	TargetTenantID     *string
	Status             models.AuditLogStatus
	Reason             *string
	IPAddress          *string
	UserAgent          *string
	Metadata           map[string]interface{}
}

type auditLogService struct {
	repo repositories.AuditLogRepository
}

func NewAuditLogService(repo repositories.AuditLogRepository) AuditLogService {
	return &auditLogService{repo: repo}
}

func (s *auditLogService) Log(ctx context.Context, entry AuditLogEntry) error {
	log := &models.AuditLog{
		EventID:            uuid.NewString(),
		Timestamp:          time.Now().UTC(),
		Action:             entry.Action,
		ActorID:            entry.ActorID,
		ActorTenantID:      entry.ActorTenantID,
		ActorRole:          entry.ActorRole,
		TargetResourceID:   entry.TargetResourceID,
		TargetResourceType: entry.TargetResourceType,
		TargetTenantID:     entry.TargetTenantID,
		Status:             entry.Status,
		Reason:             entry.Reason,
		IPAddress:          entry.IPAddress,
		UserAgent:          entry.UserAgent,
		Metadata:           entry.Metadata,
	}

	// Use background context to ensure log persists even if request is cancelled
	bgCtx := context.Background()
	return s.repo.Create(bgCtx, log)
}

func (s *auditLogService) Query(ctx context.Context, filters repositories.AuditLogFilters, page, pageSize int) ([]models.AuditLog, int64, error) {
	return s.repo.List(ctx, filters, page, pageSize)
}
