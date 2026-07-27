package entity

import "simpkl-api/internal/shared/types"

type Readiness struct {
	types.BaseModel
	StudentID               string  `gorm:"type:char(36);not null;uniqueIndex:idx_readiness_student_period" json:"student_id" validate:"required,uuid"`
	PeriodID                string  `gorm:"type:char(36);not null;uniqueIndex:idx_readiness_student_period" json:"period_id" validate:"required,uuid"`
	PlacementID             string  `gorm:"type:char(36);index" json:"placement_id" validate:"omitempty,uuid"`
	DataComplete            bool    `json:"data_complete"`
	CompanyAssigned         bool    `json:"company_assigned"`
	ContactAvailable        bool    `json:"contact_available"`
	SupervisorAssigned      bool    `json:"supervisor_assigned"`
	DatesSet                bool    `json:"dates_set"`
	AcceptanceLetterValid   bool    `json:"acceptance_letter_valid"`
	ParentPermissionValid   bool    `json:"parent_permission_valid"`
	IntroductionLetterValid bool    `json:"introduction_letter_valid"`
	RequiredCount           int     `gorm:"not null;default:8" json:"required_count"`
	CompletedCount          int     `gorm:"not null;default:0" json:"completed_count"`
	Percentage              float64 `gorm:"type:decimal(5,2);not null;default:0" json:"percentage"`
	Status                  string  `gorm:"size:30;not null;default:incomplete;index" json:"status" validate:"required,oneof=incomplete attention ready started completed"`
	OverrideReason          string  `gorm:"type:text" json:"override_reason"`
}

func (Readiness) TableName() string { return "administrative_readiness" }
