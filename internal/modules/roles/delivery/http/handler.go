package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	roleservice "simpkl-api/internal/modules/roles/service"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/response"
)

type Handler struct {
	assignments *roleservice.PermissionAssignmentService
}

func NewHandler(assignments *roleservice.PermissionAssignmentService) *Handler {
	return &Handler{assignments}
}

func (h *Handler) SetPermissions(c *gin.Context) {
	var request SetPermissionsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Daftar permission tidak valid", "VALIDATION_ERROR", nil)
		return
	}
	if err := h.assignments.SetPermissions(c.Request.Context(), c.Param("id"), request.PermissionIDs); err != nil {
		var appError *apperrors.AppError
		if errors.As(err, &appError) {
			response.Error(c, appError.Status, appError.Message, appError.Code, nil)
			return
		}
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, "Permission role berhasil diperbarui", nil, nil)
}
