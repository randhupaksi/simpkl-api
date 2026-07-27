package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	userservice "simpkl-api/internal/modules/users/service"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/response"
)

type Handler struct {
	assignments *userservice.RoleAssignmentService
}

func NewHandler(assignments *userservice.RoleAssignmentService) *Handler {
	return &Handler{assignments}
}

func (h *Handler) SetRoles(c *gin.Context) {
	var request SetRolesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Daftar role tidak valid", "VALIDATION_ERROR", nil)
		return
	}
	if err := h.assignments.SetRoles(c.Request.Context(), c.Param("id"), request.RoleIDs); err != nil {
		var appError *apperrors.AppError
		if errors.As(err, &appError) {
			response.Error(c, appError.Status, appError.Message, appError.Code, nil)
			return
		}
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, "Role pengguna berhasil diperbarui", nil, nil)
}
