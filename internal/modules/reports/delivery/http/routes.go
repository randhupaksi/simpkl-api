package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, require func(string) gin.HandlerFunc) {
	group := api.Group("/reports")
	group.GET("/dashboard", require("report.view"), handler.Dashboard)
	group.GET("/placements", require("report.view"), handler.Placements)
	group.GET("/expirations", require("report.view"), handler.Expirations)
}
