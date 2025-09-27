package models

import (
	"time"

	"gorm.io/datatypes"
)

type IdempotencyStatus string

const (
	IdempotencyStatusSuccess IdempotencyStatus = "SUCCESS"
	IdempotencyStatusError   IdempotencyStatus = "ERROR"
)

type IdempotencyKey struct {
	ID          uint64            `gorm:"primaryKey;autoIncrement"`
	KeyHash     string            `gorm:"type:char(64);not null;uniqueIndex:ux_idemp_key"`
	RequestHash string            `gorm:"type:char(64);not null"`
	Response    datatypes.JSON    `gorm:"type:json"`
	Status      IdempotencyStatus `gorm:"type:enum('SUCCESS','ERROR');not null"`
	CreatedAt   time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (IdempotencyKey) TableName() string { return "idempotency_keys" }
