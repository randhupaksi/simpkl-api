package http

import (
	"time"

	"simpkl-api/internal/modules/placements/entity"
)

type TransferRequest struct {
	EndDate      time.Time        `json:"end_date" binding:"required"`
	Reason       string           `json:"reason" binding:"required"`
	NewPlacement entity.Placement `json:"new_placement" binding:"required"`
}
