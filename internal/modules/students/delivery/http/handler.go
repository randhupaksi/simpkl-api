package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	studentservice "simpkl-api/internal/modules/students/service"
	"simpkl-api/internal/shared/response"
)

type Handler struct{ importer *studentservice.ImportService }

func NewHandler(importer *studentservice.ImportService) *Handler { return &Handler{importer} }

func (h *Handler) Import(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "File Excel wajib dipilih", "FILE_REQUIRED", nil)
		return
	}
	source, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c)
		return
	}
	defer source.Close()
	commit, _ := strconv.ParseBool(c.DefaultPostForm("commit", "false"))
	result, err := h.importer.Import(c.Request.Context(), source, commit)
	if err != nil {
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), "IMPORT_FAILED", nil)
		return
	}
	message := "Validasi impor selesai"
	if commit {
		message = "Impor siswa selesai"
	}
	response.Success(c, http.StatusOK, message, result, nil)
}
