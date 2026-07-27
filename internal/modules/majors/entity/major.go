package entity

import "simpkl-api/internal/shared/types"

type Major struct {
	types.BaseModel
	Code         string `gorm:"size:20;uniqueIndex;not null" json:"code" validate:"required,max=20"`
	Name         string `gorm:"size:150;not null" json:"name" validate:"required,max=150"`
	Abbreviation string `gorm:"size:30;not null" json:"abbreviation" validate:"required,max=30"`
	HeadName     string `gorm:"size:150" json:"head_name"`
	Status       string `gorm:"size:30;not null;default:active;index" json:"status" validate:"required,oneof=active inactive"`
	Description  string `gorm:"type:text" json:"description"`
}

func (Major) TableName() string { return "majors" }
