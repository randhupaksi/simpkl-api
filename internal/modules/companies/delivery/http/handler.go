package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	companyservice "simpkl-api/internal/modules/companies/service"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/pagination"
	"simpkl-api/internal/shared/response"
)

type Handler struct {
	partnerships *companyservice.PartnershipService
}

func NewHandler(partnerships *companyservice.PartnershipService) *Handler {
	return &Handler{partnerships}
}

func (h *Handler) SetMajorCapacities(c *gin.Context) {
	var request SetMajorCapacitiesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Daftar jurusan dan kuota tidak valid", "VALIDATION_ERROR", nil)
		return
	}
	items, err := h.partnerships.SetMajorCapacities(c.Request.Context(), c.Param("id"), request.Items)
	if err != nil {
		writeCompanyError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Jurusan dan kuota perusahaan berhasil diperbarui", items, nil)
}

func (h *Handler) MajorCapacities(c *gin.Context) {
	items, err := h.partnerships.MajorCapacities(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, "Jurusan dan kuota perusahaan berhasil diambil", items, nil)
}

func (h *Handler) EligibleCompanies(c *gin.Context) {
	var query pagination.Query
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Parameter daftar perusahaan tidak valid", "VALIDATION_ERROR", nil)
		return
	}

	items, meta, err := h.partnerships.EligibleCompanies(
		c.Request.Context(),
		query,
		c.Query("student_id"),
		c.Query("company_id"),
	)
	if err != nil {
		writeCompanyError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Perusahaan yang sesuai berhasil diambil", items, meta)
}

func writeCompanyError(c *gin.Context, err error) {
	var appError *apperrors.AppError
	if errors.As(err, &appError) {
		response.Error(c, appError.Status, appError.Message, appError.Code, appError.Errors)
		return
	}
	response.InternalError(c)
}
