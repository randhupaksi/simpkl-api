package entity

import "simpkl-api/internal/shared/types"

type Class struct {
	types.BaseModel
	Name            string `gorm:"size:100;not null;index" json:"name" validate:"required,max=100"`
	Level           int    `gorm:"not null;index" json:"level" validate:"required,min=10,max=13"`
	MajorID         string `gorm:"type:char(36);not null;index" json:"major_id" validate:"required,uuid"`
	HomeroomTeacher string `gorm:"size:150" json:"homeroom_teacher"`
	AcademicYear    string `gorm:"size:20;not null;index" json:"academic_year" validate:"required,max=20"`
	Status          string `gorm:"size:30;not null;default:active;index" json:"status" validate:"required,oneof=active inactive"`
}

func (Class) TableName() string { return "classes" }
