package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	authservice "simpkl-api/internal/modules/auth/service"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/response"
)

type Handler struct{ service *authservice.Service }

func NewHandler(service *authservice.Service) *Handler { return &Handler{service} }

func (h *Handler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Data login tidak valid", "VALIDATION_ERROR", map[string][]string{"request": {err.Error()}})
		return
	}
	result, err := h.service.Login(c.Request.Context(), request.Login, request.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Login berhasil", result, nil)
}

func (h *Handler) Refresh(c *gin.Context) {
	var request RefreshRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Refresh token wajib diisi", "VALIDATION_ERROR", nil)
		return
	}
	result, err := h.service.Refresh(c.Request.Context(), request.RefreshToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Token berhasil diperbarui", result, nil)
}

func (h *Handler) Logout(c *gin.Context) {
	var request LogoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Refresh token wajib diisi", "VALIDATION_ERROR", nil)
		return
	}
	if err := h.service.Logout(c.Request.Context(), request.RefreshToken); err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Logout berhasil", nil, nil)
}

func (h *Handler) Me(c *gin.Context) {
	profile, err := h.service.Me(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Profil berhasil diambil", profile, nil)
}

func writeAuthError(c *gin.Context, err error) {
	var appError *apperrors.AppError
	if errors.As(err, &appError) {
		response.Error(c, appError.Status, appError.Message, appError.Code, nil)
		return
	}
	response.InternalError(c)
}
