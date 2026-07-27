package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, require func(string) gin.HandlerFunc) {
	group := api.Group("/readiness")
	group.POST("/recalculate", require("readiness.update"), handler.Recalculate)
	group.POST("/override", require("readiness.override"), handler.Override)
}
