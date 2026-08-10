package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler, require func(string) gin.HandlerFunc) {
	group := api.Group("/document-automation")
	group.GET("/profile", require("automation.view"), handler.GetProfile)
	group.PUT("/profile", require("automation.manage"), handler.UpdateProfile)
	group.GET("/signatories", require("automation.view"), handler.ListSignatories)
	group.POST("/signatories", require("automation.manage"), handler.CreateSignatory)
	group.PUT("/signatories/:id", require("automation.manage"), handler.UpdateSignatory)
	group.DELETE("/signatories/:id", require("automation.manage"), handler.DeleteSignatory)
	group.GET("/templates", require("automation.view"), handler.ListTemplates)
	group.POST("/templates", require("automation.manage"), handler.CreateTemplate)
	group.PUT("/templates/:id", require("automation.manage"), handler.UpdateTemplate)
	group.PUT("/templates/:id/active", require("automation.manage"), handler.SetTemplateActive)
	group.POST("/preview", require("automation.view"), handler.Preview)
	group.POST("/generate", require("automation.generate"), handler.Generate)
	group.GET("/batches", require("automation.view"), handler.ListBatches)
	group.GET("/batches/:id", require("automation.view"), handler.GetBatch)
	group.GET("/batches/:id/download", require("automation.download"), handler.DownloadBatch)
	group.GET("/documents", require("automation.view"), handler.ListDocuments)
	group.GET("/documents/:id/download", require("automation.download"), handler.DownloadDocument)
}
