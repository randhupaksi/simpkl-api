package entity

import "simpkl-api/internal/shared/types"

type Permission struct {
	types.BaseModel
	Code        string `gorm:"size:120;uniqueIndex;not null" json:"code" validate:"required,max=120"`
	Name        string `gorm:"size:150;not null" json:"name" validate:"required,max=150"`
	Module      string `gorm:"size:80;not null;index" json:"module" validate:"required,max=80"`
	Description string `gorm:"type:text" json:"description"`
}

func (Permission) TableName() string { return "permissions" }

type RolePermission struct {
	RoleID       string `gorm:"type:char(36);primaryKey"`
	PermissionID string `gorm:"type:char(36);primaryKey"`
}

func (RolePermission) TableName() string { return "role_permissions" }
