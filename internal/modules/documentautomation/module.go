package documentautomation

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	automationhttp "simpkl-api/internal/modules/documentautomation/delivery/http"
	automationservice "simpkl-api/internal/modules/documentautomation/service"
	"simpkl-api/internal/platform/storage"
	"simpkl-api/internal/shared/types"
)

func Register(api *gin.RouterGroup, db *gorm.DB, fileStorage storage.Storage, auditor types.Auditor, require func(string) gin.HandlerFunc) {
	service := automationservice.New(db, fileStorage, auditor)
	automationhttp.RegisterRoutes(api, automationhttp.NewHandler(service), require)
}
