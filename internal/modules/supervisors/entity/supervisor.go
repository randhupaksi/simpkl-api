package entity

import "simpkl-api/internal/shared/types"

type Supervisor struct {
	types.BaseModel
	EmployeeNumber string `gorm:"size:50;uniqueIndex" json:"employee_number"`
	Name           string `gorm:"size:150;not null;index" json:"name" validate:"required,max=150"`
	Phone          string `gorm:"size:30" json:"phone"`
	Email          string `gorm:"size:150;index" json:"email" validate:"omitempty,email"`
	MajorID        string `gorm:"type:char(36);index" json:"major_id" validate:"omitempty,uuid"`
	Position       string `gorm:"size:100" json:"position"`
	Status         string `gorm:"size:30;not null;default:active;index" json:"status" validate:"required,oneof=active inactive"`
	MaxStudents    int    `gorm:"not null;default:20" json:"max_students" validate:"min=1,max=200"`
}

func (Supervisor) TableName() string { return "supervisors" }
