package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	auditservice "simpkl-api/internal/modules/auditlogs/service"
	"simpkl-api/internal/shared/pagination"
	"simpkl-api/internal/shared/response"
)

type Handler struct{ service *auditservice.Service }

func NewHandler(service *auditservice.Service) *Handler { return &Handler{service} }

func (h *Handler) List(c *gin.Context) {
	var query pagination.Query
	_ = c.ShouldBindQuery(&query)
	filters := map[string]string{
		"actor_id": c.Query("actor_id"), "action": c.Query("action"),
		"resource": c.Query("resource"), "request_id": c.Query("request_id"),
	}
	items, meta, err := h.service.List(c.Request.Context(), query, filters)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, "Audit log berhasil diambil", items, meta)
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Audit log tidak ditemukan", "RESOURCE_NOT_FOUND", nil)
		return
	}
	response.Success(c, http.StatusOK, "Audit log berhasil diambil", item, nil)
}
