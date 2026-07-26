package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"simpkl-api/internal/middleware/cors"
	errorhandler "simpkl-api/internal/middleware/error"
	requestlogger "simpkl-api/internal/middleware/logger"
	"simpkl-api/internal/middleware/recovery"
	requestid "simpkl-api/internal/middleware/request-id"
	"simpkl-api/internal/shared/response"
)

type HealthResponse struct {
	Status   string `json:"status"`
	Service  string `json:"service"`
	Database string `json:"database"`
	Time     string `json:"time"`
}

type databasePinger interface {
	PingContext(context.Context) error
}

func NewRouter(dependencies *Dependencies) *gin.Engine {
	if dependencies.Config.App.Env.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		requestid.New(),
		requestlogger.New(dependencies.Logger),
		recovery.New(dependencies.Logger),
		errorhandler.New(dependencies.Logger),
		cors.New(dependencies.Config.CORS),
	)

	healthHandler := newHealthHandler(
		dependencies.Config.App.Name,
		dependencies.Database.SQL,
	)
	router.GET("/health", healthHandler)
	router.GET("/api/v1/health", healthHandler)

	return router
}

func newHealthHandler(serviceName string, database databasePinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			response.Error(
				c,
				http.StatusServiceUnavailable,
				"Database tidak tersedia",
				"DATABASE_UNAVAILABLE",
				nil,
			)
			return
		}

		response.Success(c, http.StatusOK, "Layanan berjalan normal", HealthResponse{
			Status:   "ok",
			Service:  serviceName,
			Database: "connected",
			Time:     time.Now().UTC().Format(time.RFC3339),
		}, nil)
	}
}
