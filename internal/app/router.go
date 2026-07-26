package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"simpkl-api/internal/middleware/auth"
	"simpkl-api/internal/middleware/cors"
	errorhandler "simpkl-api/internal/middleware/error"
	requestlogger "simpkl-api/internal/middleware/logger"
	"simpkl-api/internal/middleware/permission"
	"simpkl-api/internal/middleware/recovery"
	requestid "simpkl-api/internal/middleware/request-id"
	"simpkl-api/internal/modules/archives"
	audithttp "simpkl-api/internal/modules/auditlogs/delivery/http"
	auditrepository "simpkl-api/internal/modules/auditlogs/repository"
	auditservice "simpkl-api/internal/modules/auditlogs/service"
	authhttp "simpkl-api/internal/modules/auth/delivery/http"
	authservice "simpkl-api/internal/modules/auth/service"
	"simpkl-api/internal/modules/classes"
	"simpkl-api/internal/modules/companies"
	"simpkl-api/internal/modules/companycontacts"
	"simpkl-api/internal/modules/documents"
	"simpkl-api/internal/modules/majors"
	"simpkl-api/internal/modules/periods"
	"simpkl-api/internal/modules/permissions"
	"simpkl-api/internal/modules/placements"
	"simpkl-api/internal/modules/readiness"
	"simpkl-api/internal/modules/reports"
	"simpkl-api/internal/modules/roles"
	"simpkl-api/internal/modules/students"
	"simpkl-api/internal/modules/supervisors"
	"simpkl-api/internal/modules/users"
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
	router.StaticFile("/openapi.yaml", "./docs/openapi.yaml")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.URL("/openapi.yaml"),
	))

	api := router.Group("/api/v1")
	authenticate := auth.New(dependencies.Tokens)
	require := permission.New(dependencies.AuthRepository)
	authHandler := authhttp.NewHandler(authservice.New(
		dependencies.AuthRepository,
		dependencies.Tokens,
		dependencies.Config.JWT.AccessTTL,
		dependencies.Config.JWT.RefreshTTL,
	))
	authhttp.RegisterRoutes(api, authHandler, authenticate)

	protected := api.Group("")
	protected.Use(authenticate)
	majors.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	classes.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	students.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	companies.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	companycontacts.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	supervisors.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	periods.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	users.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	roles.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	permissions.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	placements.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	documents.Register(protected, dependencies.Database.GORM, dependencies.Storage, dependencies.Auditor, require)
	readiness.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	archives.Register(protected, dependencies.Database.GORM, dependencies.Auditor, require)
	reports.Register(protected, dependencies.Database.GORM, require)

	auditRepo := auditrepository.NewMySQLRepository(dependencies.Database.GORM)
	audithttp.RegisterRoutes(
		protected,
		audithttp.NewHandler(auditservice.New(auditRepo)),
		require,
	)
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
