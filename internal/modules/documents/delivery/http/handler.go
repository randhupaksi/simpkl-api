package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"simpkl-api/internal/modules/documents/entity"
	documentservice "simpkl-api/internal/modules/documents/service"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/pagination"
	"simpkl-api/internal/shared/response"
)

type Handler struct{ service *documentservice.Service }

func NewHandler(service *documentservice.Service) *Handler { return &Handler{service} }

func (h *Handler) List(c *gin.Context) {
	var query pagination.Query
	_ = c.ShouldBindQuery(&query)
	filters := map[string]string{
		"document_type_id": c.Query("document_type_id"), "owner_type": c.Query("owner_type"),
		"owner_id": c.Query("owner_id"), "period_id": c.Query("period_id"),
		"placement_id": c.Query("placement_id"), "status": c.Query("status"),
	}
	items, meta, err := h.service.List(c.Request.Context(), query, filters)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, "Dokumen berhasil diambil", items, meta)
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Dokumen tidak ditemukan", "RESOURCE_NOT_FOUND", nil)
		return
	}
	response.Success(c, http.StatusOK, "Dokumen berhasil diambil", item, nil)
}

func (h *Handler) Upload(c *gin.Context) {
	var request UploadRequest
	if err := c.ShouldBind(&request); err != nil {
		writeDocumentError(c, err)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		writeDocumentError(c, err)
		return
	}
	input := &entity.Document{
		DocumentTypeID: request.DocumentTypeID, OwnerType: request.OwnerType,
		OwnerID: request.OwnerID, PeriodID: request.PeriodID, PlacementID: request.PlacementID,
		Number: request.Number, IssuedAt: request.IssuedAt, ValidFrom: request.ValidFrom,
		ValidUntil: request.ValidUntil, Notes: request.Notes, Status: "pending",
	}
	document, err := h.service.Upload(c.Request.Context(), input, file)
	if err != nil {
		writeDocumentError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Dokumen berhasil diunggah", document, nil)
}

func (h *Handler) Verify(c *gin.Context) {
	var request VerifyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeDocumentError(c, err)
		return
	}
	document, err := h.service.Verify(c.Request.Context(), c.Param("id"), request.Status, c.GetString("user_id"), request.Notes)
	if err != nil {
		writeDocumentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Status dokumen berhasil diperbarui", document, nil)
}

func (h *Handler) Download(c *gin.Context) {
	document, file, err := h.service.Open(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "File dokumen tidak ditemukan", "FILE_NOT_FOUND", nil)
		return
	}
	defer file.Close()
	c.Header("Content-Disposition", `attachment; filename="`+document.OriginalName+`"`)
	c.Header("Content-Type", document.MimeType)
	c.DataFromReader(http.StatusOK, document.Size, document.MimeType, file, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		writeDocumentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Dokumen berhasil dihapus", nil, nil)
}

func writeDocumentError(c *gin.Context, err error) {
	var appError *apperrors.AppError
	if errors.As(err, &appError) {
		response.Error(c, appError.Status, appError.Message, appError.Code, nil)
		return
	}
	response.Error(c, http.StatusUnprocessableEntity, "Dokumen tidak valid", "VALIDATION_ERROR", map[string][]string{"document": {err.Error()}})
}
