package classes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"simpkl-api/internal/modules/classes/entity"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Class](
		db,
		[]string{"name", "homeroom_teacher", "academic_year"},
		map[string]string{"major_id": "major_id", "level": "level", "academic_year": "academic_year", "status": "status"},
	)
	service := crud.NewService("class", repo, auditor, nil, nil)
	handler := crud.NewHandler(service, "major_id", "level", "academic_year", "status")
	crud.RegisterRoutes(api.Group("/classes"), handler, require, "class")
}
