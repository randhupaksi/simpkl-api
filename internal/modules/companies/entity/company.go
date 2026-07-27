package entity

import (
	"time"

	"simpkl-api/internal/shared/types"
)

type Company struct {
	types.BaseModel
	Name             string     `gorm:"size:180;not null;index" json:"name" validate:"required,max=180"`
	BusinessType     string     `gorm:"size:80" json:"business_type"`
	Industry         string     `gorm:"size:150;index" json:"industry" validate:"required,max=150"`
	Description      string     `gorm:"type:text" json:"description"`
	Address          string     `gorm:"type:text;not null" json:"address" validate:"required"`
	District         string     `gorm:"size:100" json:"district"`
	City             string     `gorm:"size:100;index" json:"city" validate:"required,max=100"`
	Province         string     `gorm:"size:100" json:"province"`
	PostalCode       string     `gorm:"size:15" json:"postal_code"`
	Phone            string     `gorm:"size:30;index" json:"phone"`
	Email            string     `gorm:"size:150;index" json:"email" validate:"omitempty,email"`
	Website          string     `gorm:"size:255" json:"website" validate:"omitempty,url"`
	MapsURL          string     `gorm:"size:500" json:"maps_url" validate:"omitempty,url"`
	Status           string     `gorm:"size:40;not null;default:candidate;index" json:"status" validate:"required,oneof=candidate verifying active inactive expired not_recommended blocked"`
	Capacity         int        `gorm:"not null;default:0" json:"capacity" validate:"min=0"`
	CooperationStart *time.Time `json:"cooperation_start"`
	CooperationEnd   *time.Time `gorm:"index" json:"cooperation_end"`
	Notes            string     `gorm:"type:text" json:"notes"`
}

func (Company) TableName() string { return "companies" }

type MajorCapacity struct {
	CompanyID string `gorm:"type:char(36);primaryKey" json:"company_id"`
	MajorID   string `gorm:"type:char(36);primaryKey" json:"major_id" validate:"required,uuid"`
	Capacity  int    `gorm:"not null;default:0" json:"capacity" validate:"min=0"`
}

func (MajorCapacity) TableName() string { return "company_major_capacities" }
