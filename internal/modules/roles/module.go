package roles

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	rolehttp "simpkl-api/internal/modules/roles/delivery/http"
	"simpkl-api/internal/modules/roles/entity"
	roleservice "simpkl-api/internal/modules/roles/service"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Role](
		db,
		[]string{"code", "name", "description"},
		map[string]string{"status": "status", "is_system": "is_system"},
	)
	service := crud.NewService("role", repo, auditor, nil, nil)
	handler := crud.NewHandler(service, "status", "is_system")
	group := api.Group("/roles")
	crud.RegisterRoutes(group, handler, require, "role")
	assignments := rolehttp.NewHandler(roleservice.NewPermissionAssignmentService(db))
	group.PUT("/:id/permissions", require("role.manage"), assignments.SetPermissions)
}
