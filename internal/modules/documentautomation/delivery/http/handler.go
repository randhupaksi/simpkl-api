package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"simpkl-api/internal/modules/documentautomation/entity"
	automationservice "simpkl-api/internal/modules/documentautomation/service"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/response"
)

type Handler struct{ service *automationservice.Service }

func NewHandler(service *automationservice.Service) *Handler { return &Handler{service: service} }

func (h *Handler) GetProfile(c *gin.Context) {
	item, err := h.service.Profile(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Profil institusi berhasil diambil", item, nil)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var input entity.SchoolProfile
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, err)
		return
	}
	item, err := h.service.UpdateProfile(c.Request.Context(), input)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Profil institusi berhasil diperbarui", item, nil)
}

func (h *Handler) ListSignatories(c *gin.Context) {
	items, err := h.service.Signatories(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Penandatangan berhasil diambil", items, nil)
}

func (h *Handler) CreateSignatory(c *gin.Context) { h.saveSignatory(c, "") }
func (h *Handler) UpdateSignatory(c *gin.Context) { h.saveSignatory(c, c.Param("id")) }
func (h *Handler) saveSignatory(c *gin.Context, id string) {
	var input entity.Signatory
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, err)
		return
	}
	item, err := h.service.SaveSignatory(c.Request.Context(), id, input)
	if err != nil {
		writeError(c, err)
		return
	}
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
	}
	response.Success(c, status, "Penandatangan berhasil disimpan", item, nil)
}

func (h *Handler) DeleteSignatory(c *gin.Context) {
	if err := h.service.DeleteSignatory(c.Request.Context(), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Penandatangan berhasil dihapus", nil, nil)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	items, err := h.service.Templates(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Template dokumen berhasil diambil", items, nil)
}

func (h *Handler) CreateTemplate(c *gin.Context) { h.saveTemplate(c, "") }
func (h *Handler) UpdateTemplate(c *gin.Context) { h.saveTemplate(c, c.Param("id")) }
func (h *Handler) saveTemplate(c *gin.Context, id string) {
	var input entity.DocumentTemplate
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, err)
		return
	}
	item, err := h.service.SaveTemplate(c.Request.Context(), id, input)
	if err != nil {
		writeError(c, err)
		return
	}
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
	}
	response.Success(c, status, "Versi template berhasil disimpan", item, nil)
}

func (h *Handler) SetTemplateActive(c *gin.Context) {
	var input struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, err)
		return
	}
	if err := h.service.SetTemplateActive(c.Request.Context(), c.Param("id"), input.Active); err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Status template berhasil diperbarui", nil, nil)
}

func (h *Handler) Preview(c *gin.Context) {
	var input struct {
		Filters       automationservice.Filters `json:"filters"`
		TemplateCodes []string                  `json:"template_codes"`
		Formats       []string                  `json:"formats"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, err)
		return
	}
	preview, err := h.service.Preview(c.Request.Context(), input.Filters, input.TemplateCodes, input.Formats)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Pratinjau dokumen berhasil dihitung", preview, nil)
}

func (h *Handler) Generate(c *gin.Context) {
	var input automationservice.GenerateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, err)
		return
	}
	input.RequestedBy = c.GetString("user_id")
	batch, err := h.service.Generate(c.Request.Context(), input)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Dokumen berhasil dibuat", batch, nil)
}

func (h *Handler) ListBatches(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	items, err := h.service.Batches(c.Request.Context(), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Riwayat batch berhasil diambil", items, nil)
}

func (h *Handler) GetBatch(c *gin.Context) {
	batch, documents, err := h.service.Batch(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Detail batch berhasil diambil", gin.H{"batch": batch, "documents": documents}, nil)
}

func (h *Handler) ListDocuments(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	filters := map[string]string{"batch_id": c.Query("batch_id"), "student_id": c.Query("student_id"), "period_id": c.Query("period_id"), "template_code": c.Query("template_code"), "format": c.Query("format")}
	items, err := h.service.Documents(c.Request.Context(), filters, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Riwayat dokumen berhasil diambil", items, nil)
}

func (h *Handler) DownloadDocument(c *gin.Context) {
	item, file, err := h.service.OpenDocument(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	defer file.Close()
	c.Header("Content-Disposition", `attachment; filename="`+item.OriginalName+`"`)
	c.DataFromReader(http.StatusOK, item.Size, item.MimeType, file, nil)
}

func (h *Handler) DownloadBatch(c *gin.Context) {
	batch, file, err := h.service.OpenBatch(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	defer file.Close()
	c.Header("Content-Disposition", `attachment; filename="`+batch.ArchiveName+`"`)
	c.DataFromReader(http.StatusOK, batch.ArchiveSize, "application/zip", file, nil)
}

func writeError(c *gin.Context, err error) {
	var appError *apperrors.AppError
	if errors.As(err, &appError) {
		response.Error(c, appError.Status, appError.Message, appError.Code, nil)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "Data otomasi dokumen tidak ditemukan", "RESOURCE_NOT_FOUND", nil)
		return
	}
	response.Error(c, http.StatusUnprocessableEntity, "Permintaan otomasi dokumen tidak dapat diproses", "AUTOMATION_ERROR", nil)
}
