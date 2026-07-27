package documents

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	documenthttp "simpkl-api/internal/modules/documents/delivery/http"
	"simpkl-api/internal/modules/documents/entity"
	documentservice "simpkl-api/internal/modules/documents/service"
	"simpkl-api/internal/platform/storage"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(
	api *gin.RouterGroup,
	db *gorm.DB,
	fileStorage storage.Storage,
	auditor types.Auditor,
	require func(string) gin.HandlerFunc,
) {
	service := documentservice.New(db, fileStorage)
	documenthttp.RegisterRoutes(api, documenthttp.NewHandler(service), require)

	typeRepo := crud.NewGormRepository[entity.DocumentType](
		db,
		[]string{"code", "name", "category"},
		map[string]string{"category": "category", "required": "required", "status": "status"},
	)
	typeService := crud.NewService("document_type", typeRepo, auditor, nil, nil)
	typeHandler := crud.NewHandler(typeService, "category", "required", "status")
	crud.RegisterRoutes(api.Group("/document-types"), typeHandler, require, "document_type")
}
