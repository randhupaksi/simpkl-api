package entity

import "simpkl-api/internal/shared/types"

type AuditLog struct {
	types.BaseModel
	ActorID    string `gorm:"type:char(36);index" json:"actor_id"`
	Action     string `gorm:"size:80;not null;index" json:"action"`
	Resource   string `gorm:"size:100;not null;index" json:"resource"`
	ResourceID string `gorm:"type:char(36);index" json:"resource_id"`
	RequestID  string `gorm:"size:100;index" json:"request_id"`
	BeforeJSON string `gorm:"type:longtext" json:"before_json"`
	AfterJSON  string `gorm:"type:longtext" json:"after_json"`
	Reason     string `gorm:"type:text" json:"reason"`
	IPAddress  string `gorm:"size:60" json:"ip_address"`
	UserAgent  string `gorm:"size:500" json:"user_agent"`
}

func (AuditLog) TableName() string { return "audit_logs" }
