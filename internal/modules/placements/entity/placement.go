package entity

import (
	"time"

	"simpkl-api/internal/shared/types"
)

type Placement struct {
	types.BaseModel
	PeriodID            string    `gorm:"type:char(36);not null;index" json:"period_id" validate:"required,uuid"`
	StudentID           string    `gorm:"type:char(36);not null;index" json:"student_id" validate:"required,uuid"`
	CompanyID           string    `gorm:"type:char(36);not null;index" json:"company_id" validate:"required,uuid"`
	CompanyContactID    string    `gorm:"type:char(36);index;default:null" json:"company_contact_id" validate:"omitempty,uuid"`
	SupervisorID        string    `gorm:"type:char(36);index;default:null" json:"supervisor_id" validate:"omitempty,uuid"`
	PreviousPlacementID string    `gorm:"type:char(36);index;default:null" json:"previous_placement_id" validate:"omitempty,uuid"`
	Division            string    `gorm:"size:120" json:"division"`
	Position            string    `gorm:"size:120" json:"position"`
	WorkSystem          string    `gorm:"size:30;not null" json:"work_system" validate:"required,oneof=wfo wfh hybrid company_policy"`
	Address             string    `gorm:"type:text" json:"address"`
	StartDate           time.Time `gorm:"type:date;not null;index" json:"start_date" validate:"required"`
	EndDate             time.Time `gorm:"type:date;not null;index" json:"end_date" validate:"required"`
	Status              string    `gorm:"size:40;not null;default:draft;index" json:"status" validate:"required,oneof=draft pending_verification approved ready active completed cancelled transferred"`
	Source              string    `gorm:"size:40;not null;default:school" json:"source" validate:"required,oneof=school self_submission teacher_recommendation company_recruitment previous_partnership"`
	OverrideReason      string    `gorm:"type:text" json:"override_reason"`
	TransferReason      string    `gorm:"type:text" json:"transfer_reason"`
	Notes               string    `gorm:"type:text" json:"notes"`
}

func (Placement) TableName() string { return "placements" }
