package entity

import (
	"time"

	"simpkl-api/internal/shared/types"
)

type Document struct {
	types.BaseModel
	DocumentTypeID string     `gorm:"type:char(36);not null;index" json:"document_type_id" validate:"required,uuid"`
	OwnerType      string     `gorm:"size:30;not null;index" json:"owner_type" validate:"required,oneof=student company placement period"`
	OwnerID        string     `gorm:"type:char(36);not null;index" json:"owner_id" validate:"required,uuid"`
	PeriodID       string     `gorm:"type:char(36);index" json:"period_id" validate:"omitempty,uuid"`
	PlacementID    string     `gorm:"type:char(36);index" json:"placement_id" validate:"omitempty,uuid"`
	Number         string     `gorm:"size:100;index" json:"number"`
	OriginalName   string     `gorm:"size:255;not null" json:"original_name" validate:"required,max=255"`
	StoredName     string     `gorm:"size:255;not null" json:"stored_name" validate:"required,max=255"`
	Path           string     `gorm:"size:500;not null" json:"-"`
	MimeType       string     `gorm:"size:120;not null" json:"mime_type"`
	Size           int64      `gorm:"not null" json:"size" validate:"min=0"`
	IssuedAt       *time.Time `json:"issued_at"`
	ValidFrom      *time.Time `json:"valid_from"`
	ValidUntil     *time.Time `gorm:"index" json:"valid_until"`
	Status         string     `gorm:"size:40;not null;default:uploaded;index" json:"status" validate:"required,oneof=draft uploaded pending valid revision_required rejected expired superseded"`
	Version        int        `gorm:"not null;default:1" json:"version" validate:"min=1"`
	VerifiedBy     string     `gorm:"type:char(36);index" json:"verified_by"`
	VerifiedAt     *time.Time `json:"verified_at"`
	Notes          string     `gorm:"type:text" json:"notes"`
}

func (Document) TableName() string { return "documents" }

type DocumentType struct {
	types.BaseModel
	Code        string `gorm:"size:60;uniqueIndex;not null" json:"code" validate:"required,max=60"`
	Name        string `gorm:"size:150;not null" json:"name" validate:"required,max=150"`
	Category    string `gorm:"size:40;not null;index" json:"category" validate:"required,oneof=student company school placement completion"`
	Required    bool   `gorm:"not null;default:false;index" json:"required"`
	HasExpiry   bool   `gorm:"not null;default:false" json:"has_expiry"`
	MaxSize     int64  `gorm:"not null;default:10485760" json:"max_size"`
	AllowedMIME string `gorm:"size:500" json:"allowed_mime"`
	Status      string `gorm:"size:30;not null;default:active;index" json:"status" validate:"required,oneof=active inactive"`
}

func (DocumentType) TableName() string { return "document_types" }
