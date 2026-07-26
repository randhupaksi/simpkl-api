package cors

import (
	"time"

	gincors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"simpkl-api/internal/config"
)

func New(cfg config.CORSConfig) gin.HandlerFunc {
	return gincors.New(gincors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
