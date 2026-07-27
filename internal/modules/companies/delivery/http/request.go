package http

import "simpkl-api/internal/modules/companies/entity"

type SetMajorCapacitiesRequest struct {
	Items []entity.MajorCapacity `json:"items" binding:"required"`
}
