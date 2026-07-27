package http

import "time"

type UploadRequest struct {
	DocumentTypeID string     `form:"document_type_id" binding:"required"`
	OwnerType      string     `form:"owner_type" binding:"required"`
	OwnerID        string     `form:"owner_id" binding:"required"`
	PeriodID       string     `form:"period_id"`
	PlacementID    string     `form:"placement_id"`
	Number         string     `form:"number"`
	IssuedAt       *time.Time `form:"issued_at" time_format:"2006-01-02"`
	ValidFrom      *time.Time `form:"valid_from" time_format:"2006-01-02"`
	ValidUntil     *time.Time `form:"valid_until" time_format:"2006-01-02"`
	Notes          string     `form:"notes"`
}

type VerifyRequest struct {
	Status string `json:"status" binding:"required"`
	Notes  string `json:"notes"`
}
