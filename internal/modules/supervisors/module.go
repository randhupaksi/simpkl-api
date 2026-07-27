package supervisors

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"simpkl-api/internal/modules/supervisors/entity"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Supervisor](
		db,
		[]string{"employee_number", "name", "email", "phone", "position"},
		map[string]string{"major_id": "major_id", "status": "status"},
	)
	service := crud.NewService("supervisor", repo, auditor, nil, nil)
	handler := crud.NewHandler(service, "major_id", "status")
	crud.RegisterRoutes(api.Group("/supervisors"), handler, require, "supervisor")
}
