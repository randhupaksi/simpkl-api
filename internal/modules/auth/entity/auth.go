package entity

import (
	"time"

	"simpkl-api/internal/shared/types"
)

type RefreshSession struct {
	types.BaseModel
	UserID    string     `gorm:"type:char(36);not null;index" json:"user_id"`
	TokenHash string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt *time.Time `gorm:"index" json:"revoked_at"`
	IPAddress string     `gorm:"size:60" json:"ip_address"`
	UserAgent string     `gorm:"size:500" json:"user_agent"`
}

func (RefreshSession) TableName() string { return "refresh_sessions" }

type Profile struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Status      string   `json:"status"`
	MajorID     string   `json:"major_id,omitempty"`
	ClassID     string   `json:"class_id,omitempty"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AuthResult struct {
	User   Profile   `json:"user"`
	Tokens TokenPair `json:"tokens"`
}
