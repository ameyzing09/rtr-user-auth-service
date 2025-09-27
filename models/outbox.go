package models

import (
	"time"

	"gorm.io/datatypes"
)

type Outbox struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement"`
	AggregateType string         `gorm:"type:varchar(64);not null"`
	AggregateID   string         `gorm:"type:char(36);not null"`
	Type          string         `gorm:"type:varchar(64);not null"`
	Payload       datatypes.JSON `gorm:"type:json;not null"`
	CreatedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	PublishedAt   *time.Time     `gorm:"type:timestamp"`
}

func (Outbox) TableName() string { return "outbox" }
