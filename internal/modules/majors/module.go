package majors

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"simpkl-api/internal/modules/majors/entity"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Major](
		db,
		[]string{"code", "name", "abbreviation", "head_name"},
		map[string]string{"status": "status"},
	)
	service := crud.NewService("major", repo, auditor, nil, nil)
	handler := crud.NewHandler(service, "status")
	crud.RegisterRoutes(api.Group("/majors"), handler, require, "major")
}
