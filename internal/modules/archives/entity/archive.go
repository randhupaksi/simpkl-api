package entity

import (
	"time"

	"simpkl-api/internal/shared/types"
)

type Archive struct {
	types.BaseModel
	PeriodID   string    `gorm:"type:char(36);not null;uniqueIndex" json:"period_id" validate:"required,uuid"`
	ArchivedBy string    `gorm:"type:char(36);not null;index" json:"archived_by" validate:"required,uuid"`
	ArchivedAt time.Time `gorm:"not null;index" json:"archived_at"`
	Reason     string    `gorm:"type:text" json:"reason"`
	Snapshot   string    `gorm:"type:longtext" json:"snapshot"`
}

func (Archive) TableName() string { return "archives" }
