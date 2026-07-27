package readiness

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	readinesshttp "simpkl-api/internal/modules/readiness/delivery/http"
	"simpkl-api/internal/modules/readiness/entity"
	readinessservice "simpkl-api/internal/modules/readiness/service"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Readiness](
		db,
		[]string{"status", "override_reason"},
		map[string]string{"student_id": "student_id", "period_id": "period_id", "status": "status"},
	)
	baseService := crud.NewService("readiness", repo, auditor, nil, nil)
	baseHandler := crud.NewHandler(baseService, "student_id", "period_id", "status")
	group := api.Group("/readiness")
	group.GET("", require("readiness.view"), baseHandler.List)
	group.GET("/:id", require("readiness.view"), baseHandler.Get)

	readinesshttp.RegisterRoutes(api, readinesshttp.NewHandler(readinessservice.New(db)), require)
}
