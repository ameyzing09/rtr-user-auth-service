package models

import "time"

type AuditLogStatus string

const (
	AuditStatusSuccess AuditLogStatus = "success"
	AuditStatusDenied  AuditLogStatus = "denied"
	AuditStatusError   AuditLogStatus = "error"
)

type AuditLog struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement"`
	EventID            string         `gorm:"type:char(36);not null;uniqueIndex"`
	Timestamp          time.Time      `gorm:"not null;index:idx_timestamp"`
	Action             string         `gorm:"type:varchar(100);not null;index:idx_action"`
	ActorID            *string        `gorm:"type:char(36);index:idx_actor_id"`
	ActorTenantID      *string        `gorm:"type:char(36)"`
	ActorRole          *string        `gorm:"type:varchar(50)"`
	TargetResourceID   *string        `gorm:"type:varchar(255)"`
	TargetResourceType *string        `gorm:"type:varchar(50)"`
	TargetTenantID     *string        `gorm:"type:char(36);index:idx_target_tenant"`
	Status             AuditLogStatus `gorm:"type:enum('success','denied','error');not null"`
	Reason             *string        `gorm:"type:varchar(255)"`
	IPAddress          *string        `gorm:"type:varchar(45)"`
	UserAgent          *string        `gorm:"type:text"`
	Metadata           JSONMap        `gorm:"type:json"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
