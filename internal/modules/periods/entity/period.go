package entity

import (
	"time"

	"simpkl-api/internal/shared/types"
)

type Period struct {
	types.BaseModel
	Name         string    `gorm:"size:180;not null;index" json:"name" validate:"required,max=180"`
	AcademicYear string    `gorm:"size:20;not null;index" json:"academic_year" validate:"required,max=20"`
	Semester     string    `gorm:"size:20;not null" json:"semester" validate:"required,oneof=odd even"`
	StartDate    time.Time `gorm:"type:date;not null;index" json:"start_date" validate:"required"`
	EndDate      time.Time `gorm:"type:date;not null;index" json:"end_date" validate:"required"`
	Cohort       int       `gorm:"not null;index" json:"cohort" validate:"required,min=2000,max=2200"`
	Status       string    `gorm:"size:30;not null;default:draft;index" json:"status" validate:"required,oneof=draft preparation active completed archived"`
	Notes        string    `gorm:"type:text" json:"notes"`
}

func (Period) TableName() string { return "periods" }
