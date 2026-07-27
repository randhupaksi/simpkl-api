package archives

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	archivehttp "simpkl-api/internal/modules/archives/delivery/http"
	"simpkl-api/internal/modules/archives/entity"
	archiveservice "simpkl-api/internal/modules/archives/service"
	"simpkl-api/internal/shared/crud"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	repo := crud.NewGormRepository[entity.Archive](
		db,
		[]string{"period_id", "reason"},
		map[string]string{"period_id": "period_id", "archived_by": "archived_by"},
	)
	baseService := crud.NewService("archive", repo, auditor, nil, nil)
	baseHandler := crud.NewHandler(baseService, "period_id", "archived_by")
	group := api.Group("/archives")
	group.GET("", require("archive.view"), baseHandler.List)
	group.GET("/:id", require("archive.view"), baseHandler.Get)
	archivehttp.RegisterRoutes(api, archivehttp.NewHandler(archiveservice.New(db)), require)
}
