package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	api *gin.RouterGroup,
	handler *Handler,
	authenticate gin.HandlerFunc,
) {
	group := api.Group("/auth")
	group.POST("/login", handler.Login)
	group.POST("/refresh", handler.Refresh)
	group.POST("/logout", authenticate, handler.Logout)
	group.GET("/me", authenticate, handler.Me)
}
