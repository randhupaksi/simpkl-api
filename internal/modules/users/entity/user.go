package entity

import (
	"time"

	"simpkl-api/internal/shared/types"
)

type User struct {
	types.BaseModel
	Name         string     `gorm:"size:150;not null;index" json:"name" validate:"required,max=150"`
	Email        string     `gorm:"size:180;uniqueIndex;not null" json:"email" validate:"required,email,max=180"`
	Username     string     `gorm:"size:80;uniqueIndex;not null" json:"username" validate:"required,max=80"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Password     string     `gorm:"-" json:"password,omitempty" validate:"omitempty,min=8,max=72"`
	MajorID      string     `gorm:"type:char(36);index" json:"major_id" validate:"omitempty,uuid"`
	ClassID      string     `gorm:"type:char(36);index" json:"class_id" validate:"omitempty,uuid"`
	Status       string     `gorm:"size:30;not null;default:active;index" json:"status" validate:"required,oneof=active inactive locked"`
	LastLoginAt  *time.Time `gorm:"type:datetime" json:"last_login_at"`
	RoleIDs      []string   `gorm:"-" json:"role_ids"`
	Roles        []string   `gorm:"-" json:"roles,omitempty"`
	Permissions  []string   `gorm:"-" json:"permissions,omitempty"`
}

func (User) TableName() string { return "users" }
