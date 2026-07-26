package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	return nil
}

type AuditEvent struct {
	ActorID    string
	Action     string
	Resource   string
	ResourceID string
	RequestID  string
	Before     any
	After      any
	Reason     string
	IPAddress  string
	UserAgent  string
}

type Auditor interface {
	Record(event AuditEvent) error
}

type NoopAuditor struct{}

func (NoopAuditor) Record(AuditEvent) error { return nil }
