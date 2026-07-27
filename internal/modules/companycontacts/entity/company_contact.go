package entity

import "simpkl-api/internal/shared/types"

type CompanyContact struct {
	types.BaseModel
	CompanyID string `gorm:"type:char(36);not null;index" json:"company_id" validate:"required,uuid"`
	Name      string `gorm:"size:150;not null;index" json:"name" validate:"required,max=150"`
	Position  string `gorm:"size:100" json:"position"`
	Division  string `gorm:"size:100" json:"division"`
	Phone     string `gorm:"size:30" json:"phone"`
	Email     string `gorm:"size:150" json:"email" validate:"omitempty,email"`
	IsPrimary bool   `gorm:"not null;default:false;index" json:"is_primary"`
	Notes     string `gorm:"type:text" json:"notes"`
}

func (CompanyContact) TableName() string { return "company_contacts" }
