package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	api *gin.RouterGroup,
	handler *Handler,
	require func(string) gin.HandlerFunc,
) {
	group := api.Group("/documents")
	group.GET("", require("document.view"), handler.List)
	group.GET("/:id", require("document.view"), handler.Get)
	group.POST("", require("document.upload"), handler.Upload)
	group.PUT("/:id/verify", require("document.verify"), handler.Verify)
	group.GET("/:id/download", require("document.download"), handler.Download)
	group.DELETE("/:id", require("document.delete"), handler.Delete)
}
