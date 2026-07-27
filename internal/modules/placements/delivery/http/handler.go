package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	placementservice "simpkl-api/internal/modules/placements/service"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/response"
)

type Handler struct {
	transfer *placementservice.TransferService
}

func NewHandler(transfer *placementservice.TransferService) *Handler {
	return &Handler{transfer}
}

func (h *Handler) Transfer(c *gin.Context) {
	var request TransferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Data perpindahan tidak valid", "VALIDATION_ERROR", nil)
		return
	}
	next, err := h.transfer.Transfer(c.Request.Context(), c.Param("id"), request.EndDate, request.Reason, &request.NewPlacement)
	if err != nil {
		var appError *apperrors.AppError
		if errors.As(err, &appError) {
			response.Error(c, appError.Status, appError.Message, appError.Code, nil)
			return
		}
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusCreated, "Siswa berhasil dipindahkan", next, nil)
}
