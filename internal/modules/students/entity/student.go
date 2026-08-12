package entity

import "simpkl-api/internal/shared/types"

type Student struct {
	types.BaseModel
	NIS         string `gorm:"size:30;uniqueIndex;not null" json:"nis" validate:"required,max=30"`
	NISN        string `gorm:"size:30;uniqueIndex" json:"nisn" validate:"omitempty,max=30"`
	Name        string `gorm:"size:150;not null;index" json:"name" validate:"required,max=150"`
	Nickname    string `gorm:"size:80" json:"nickname"`
	Gender      string `gorm:"size:20" json:"gender" validate:"omitempty,oneof=male female"`
	ClassID     string `gorm:"type:char(36);not null;index" json:"class_id" validate:"required,uuid"`
	MajorID     string `gorm:"type:char(36);not null;index" json:"major_id" validate:"required,uuid"`
	Phone       string `gorm:"size:30" json:"phone"`
	Email       string `gorm:"size:150;index" json:"email" validate:"omitempty,email"`
	Address     string `gorm:"type:text" json:"address"`
	ParentName  string `gorm:"size:150" json:"parent_name"`
	ParentPhone string `gorm:"size:30" json:"parent_phone"`
	Status      string `gorm:"size:30;not null;default:active;index" json:"status" validate:"required,oneof=active inactive graduated transferred withdrawn"`
	PKLStatus   string `gorm:"size:50;not null;default:unplaced;index" json:"pkl_status" validate:"required,oneof=unregistered unplaced placement_process awaiting_documents ready active completed cancelled transferred not_participating"`
	Notes       string `gorm:"type:text" json:"notes"`
}

func (Student) TableName() string { return "students" }
