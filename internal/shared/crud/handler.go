package crud

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/pagination"
	"simpkl-api/internal/shared/response"
	"simpkl-api/internal/shared/types"
	"simpkl-api/internal/shared/validation"
)

type Handler[T any] struct {
	service    *Service[T]
	filterKeys []string
}

func NewHandler[T any](service *Service[T], filterKeys ...string) *Handler[T] {
	return &Handler[T]{service, filterKeys}
}

func (h *Handler[T]) List(c *gin.Context) {
	var query pagination.Query
	if err := c.ShouldBindQuery(&query); err != nil {
		writeError(c, err)
		return
	}
	filters := make(map[string]string, len(h.filterKeys))
	for _, key := range h.filterKeys {
		filters[key] = c.Query(key)
	}
	if contains(h.filterKeys, "major_id") && c.GetString("scope_major_id") != "" {
		filters["major_id"] = c.GetString("scope_major_id")
	}
	if contains(h.filterKeys, "class_id") && c.GetString("scope_class_id") != "" {
		filters["class_id"] = c.GetString("scope_class_id")
	}
	items, meta, err := h.service.List(c.Request.Context(), query, filters)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Data berhasil diambil", items, meta)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (h *Handler[T]) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Data berhasil diambil", item, nil)
}

func (h *Handler[T]) Create(c *gin.Context) {
	var input T
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, err)
		return
	}
	if err := validation.Validator.Struct(input); err != nil {
		writeError(c, err)
		return
	}
	item, err := h.service.Create(c.Request.Context(), &input, auditEvent(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Data berhasil dibuat", item, nil)
}

func (h *Handler[T]) Update(c *gin.Context) {
	var input T
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, err)
		return
	}
	if err := validation.Validator.Struct(input); err != nil {
		writeError(c, err)
		return
	}
	item, err := h.service.Update(
		c.Request.Context(),
		c.Param("id"),
		&input,
		auditEvent(c),
	)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Data berhasil diperbarui", item, nil)
}

func (h *Handler[T]) Delete(c *gin.Context) {
	if err := h.service.Delete(
		c.Request.Context(),
		c.Param("id"),
		auditEvent(c),
	); err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Data berhasil dinonaktifkan", nil, nil)
}

func RegisterRoutes[T any](
	group *gin.RouterGroup,
	handler *Handler[T],
	require func(string) gin.HandlerFunc,
	permissionPrefix string,
) {
	group.GET("", require(permissionPrefix+".view"), handler.List)
	group.GET("/:id", require(permissionPrefix+".view"), handler.Get)
	group.POST("", require(permissionPrefix+".create"), handler.Create)
	group.PUT("/:id", require(permissionPrefix+".update"), handler.Update)
	group.DELETE("/:id", require(permissionPrefix+".delete"), handler.Delete)
}

func auditEvent(c *gin.Context) types.AuditEvent {
	return types.AuditEvent{
		ActorID:   c.GetString("user_id"),
		RequestID: c.GetString("request_id"),
		Reason:    c.GetHeader("X-Change-Reason"),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

func writeError(c *gin.Context, err error) {
	var appError *apperrors.AppError
	if errors.As(err, &appError) {
		response.Error(c, appError.Status, appError.Message, appError.Code, nil)
		return
	}
	response.Error(
		c,
		http.StatusUnprocessableEntity,
		"Data tidak valid",
		"VALIDATION_ERROR",
		map[string][]string{"request": {err.Error()}},
	)
}
