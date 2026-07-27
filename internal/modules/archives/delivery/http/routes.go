package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, require func(string) gin.HandlerFunc) {
	api.POST("/archives", require("period.archive"), handler.Archive)
}
