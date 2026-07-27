package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	readinessservice "simpkl-api/internal/modules/readiness/service"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/response"
)

type Handler struct{ service *readinessservice.Service }

func NewHandler(service *readinessservice.Service) *Handler { return &Handler{service} }

func (h *Handler) Recalculate(c *gin.Context) {
	var request RecalculateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Data kesiapan tidak valid", "VALIDATION_ERROR", nil)
		return
	}
	result, err := h.service.Recalculate(c.Request.Context(), request.StudentID, request.PeriodID)
	if err != nil {
		writeReadinessError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Kesiapan administrasi berhasil dihitung", result, nil)
}

func (h *Handler) Override(c *gin.Context) {
	var request OverrideRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Alasan pengecualian wajib diisi", "VALIDATION_ERROR", nil)
		return
	}
	result, err := h.service.Override(c.Request.Context(), request.StudentID, request.PeriodID, request.Reason)
	if err != nil {
		writeReadinessError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Pengecualian kesiapan berhasil diberikan", result, nil)
}

func writeReadinessError(c *gin.Context, err error) {
	var appError *apperrors.AppError
	if errors.As(err, &appError) {
		response.Error(c, appError.Status, appError.Message, appError.Code, nil)
		return
	}
	response.InternalError(c)
}
