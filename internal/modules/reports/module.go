package reports

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	reporthttp "simpkl-api/internal/modules/reports/delivery/http"
	reportservice "simpkl-api/internal/modules/reports/service"
)

func Register(api *gin.RouterGroup, db *gorm.DB, require func(string) gin.HandlerFunc) {
	reporthttp.RegisterRoutes(api, reporthttp.NewHandler(reportservice.New(db)), require)
}
