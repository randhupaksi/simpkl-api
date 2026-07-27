package permissions

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"simpkl-api/internal/modules/permissions/entity"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Permission](
		db,
		[]string{"code", "name", "module", "description"},
		map[string]string{"module": "module"},
	)
	service := crud.NewService("permission", repo, auditor, nil, nil)
	handler := crud.NewHandler(service, "module")
	crud.RegisterRoutes(api.Group("/permissions"), handler, require, "permission")
}
