package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	api *gin.RouterGroup,
	handler *Handler,
	require func(string) gin.HandlerFunc,
) {
	group := api.Group("/audit-logs")
	group.GET("", require("audit.view"), handler.List)
	group.GET("/:id", require("audit.view"), handler.Get)
}
