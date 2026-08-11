package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	reportservice "simpkl-api/internal/modules/reports/service"
	"simpkl-api/internal/shared/response"
)

type Handler struct{ service *reportservice.Service }

func NewHandler(service *reportservice.Service) *Handler { return &Handler{service} }

func (h *Handler) Dashboard(c *gin.Context) {
	result, err := h.service.Dashboard(c.Request.Context(), c.Query("period_id"), map[string]string{
		"major_id": c.GetString("scope_major_id"),
		"class_id": c.GetString("scope_class_id"),
	})
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, "Ringkasan dashboard berhasil diambil", result, nil)
}

func (h *Handler) Placements(c *gin.Context) {
	filters := reportFilters(c)
	format := c.DefaultQuery("format", "json")
	if format == "json" {
		rows, err := h.service.Placements(c.Request.Context(), filters)
		if err != nil {
			response.InternalError(c)
			return
		}
		response.Success(c, http.StatusOK, "Laporan penempatan berhasil diambil", rows, nil)
		return
	}
	if format != "xlsx" && format != "pdf" {
		response.Error(c, http.StatusUnprocessableEntity, "Format laporan harus json, xlsx, atau pdf", "INVALID_REPORT_FORMAT", nil)
		return
	}
	data, contentType, err := h.service.ExportPlacements(c.Request.Context(), format, filters)
	if err != nil {
		response.InternalError(c)
		return
	}
	filename := fmt.Sprintf("penempatan-pkl-%s.%s", time.Now().Format("20060102-150405"), format)
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, contentType, data)
}

func (h *Handler) Expirations(c *gin.Context) {
	days := 30
	_, _ = fmt.Sscanf(c.DefaultQuery("days", "30"), "%d", &days)
	items, err := h.service.Expirations(c.Request.Context(), days)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, "Pengingat masa berlaku berhasil diambil", items, nil)
}

func reportFilters(c *gin.Context) map[string]string {
	filters := map[string]string{
		"period_id": c.Query("period_id"), "major_id": c.Query("major_id"),
		"class_id": c.Query("class_id"), "company_id": c.Query("company_id"),
		"supervisor_id": c.Query("supervisor_id"), "status": c.Query("status"),
	}
	if c.GetString("scope_major_id") != "" {
		filters["major_id"] = c.GetString("scope_major_id")
	}
	if c.GetString("scope_class_id") != "" {
		filters["class_id"] = c.GetString("scope_class_id")
	}
	return filters
}
