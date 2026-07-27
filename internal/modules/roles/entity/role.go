package entity

import "simpkl-api/internal/shared/types"

type Role struct {
	types.BaseModel
	Code          string   `gorm:"size:80;uniqueIndex;not null" json:"code" validate:"required,max=80"`
	Name          string   `gorm:"size:120;not null" json:"name" validate:"required,max=120"`
	Description   string   `gorm:"type:text" json:"description"`
	IsSystem      bool     `gorm:"not null;default:false" json:"is_system"`
	Status        string   `gorm:"size:30;not null;default:active;index" json:"status" validate:"required,oneof=active inactive"`
	PermissionIDs []string `gorm:"-" json:"permission_ids"`
}

func (Role) TableName() string { return "roles" }

type UserRole struct {
	UserID string `gorm:"type:char(36);primaryKey"`
	RoleID string `gorm:"type:char(36);primaryKey"`
}

func (UserRole) TableName() string { return "user_roles" }
