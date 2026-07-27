package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	archiveservice "simpkl-api/internal/modules/archives/service"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/response"
)

type Handler struct{ service *archiveservice.Service }

func NewHandler(service *archiveservice.Service) *Handler { return &Handler{service} }

func (h *Handler) Archive(c *gin.Context) {
	var request ArchiveRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Periode wajib dipilih", "VALIDATION_ERROR", nil)
		return
	}
	archive, err := h.service.Archive(c.Request.Context(), request.PeriodID, c.GetString("user_id"), request.Reason)
	if err != nil {
		var appError *apperrors.AppError
		if errors.As(err, &appError) {
			response.Error(c, appError.Status, appError.Message, appError.Code, nil)
			return
		}
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusCreated, "Periode berhasil diarsipkan", archive, nil)
}
